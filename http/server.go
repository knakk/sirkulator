package http

import (
	"context"
	"embed"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/dewey"
	"github.com/knakk/sirkulator/etl"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/runner"
	"github.com/knakk/sirkulator/search"
	"github.com/knakk/sirkulator/sql"
	"golang.org/x/text/language"
)

//go:embed assets/*.css
var embeddedFS embed.FS

// Server represents the HTTP server responsible for serving Sirkulators admin interface.
type Server struct {
	ln     net.Listener
	srv    *http.Server
	db     *sqlitex.Pool
	idx    *search.Index
	runner *runner.Runner

	// The follwing fields should be set before calls to Open:

	// Addr is the bind address for the tcp listener.
	Addr string
	// Default language
	Lang language.Tag
}

// NewServer returns a new Server with the given database and index and assets settings.
func NewServer(ctx context.Context, assetsDir string, db *sqlitex.Pool, idx *search.Index) *Server {
	s := Server{
		Addr:   "localhost:0", // assign random port as default, useful for testing
		db:     db,
		idx:    idx,
		runner: runner.New(db),
	}

	s.runner.Register(&dewey.ImportAllJob{
		DB:        db,
		Idx:       idx,
		BatchSize: 100,
	})

	s.srv = &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		IdleTimeout:       120 * time.Second,
		// ErrorLog: TODO
	}
	s.srv.Handler = s.router(assetsDir)
	return &s
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem, assetDir string) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		if assetDir == "" {
			http.FileServer(root).ServeHTTP(w, r)
		} else {
			fs := http.StripPrefix(pathPrefix, http.FileServer(root))
			fs.ServeHTTP(w, r)
		}
	})
}

// WithLocalizer is a middleware which stores a Localizer in the request context,
// with language extracted from the Accept-Language HTTP headers if present.
func WithLocalizer() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			l := localizer.GetFromAcceptLang(r.Header.Get("Accept-Language"))
			r = r.WithContext(context.WithValue(r.Context(), "localizer", l))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (s *Server) router(assetsDir string) chi.Router {
	r := chi.NewRouter()

	// Static assets
	fs := http.FS(embeddedFS)
	if assetsDir != "" {
		fs = http.Dir(assetsDir)
	}
	FileServer(r, "/assets", fs, assetsDir)

	r.Get("/image/{id}", s.image)

	// Main UI routes
	r.Route("/", func(r chi.Router) {
		r.Use(WithLocalizer())

		r.Get("/", s.pageHome)
		r.Get("/circulation", s.pageCirculation)

		r.Route("/metadata", func(r chi.Router) {
			r.Get("/", s.pageMetadata)
			r.Get("/reviews", s.viewReviews)
			r.Post("/import", s.importResources) // s.tmplImportResponse ?
			r.Post("/preview", s.importPreview)
			r.Post("/search", s.searchResources)
			r.Get("/corporation/{id}", s.pageCorporation)
			r.Post("/person/{id}", s.savePerson)
			r.Get("/person/{id}", s.pagePerson)
			r.Post("/person/{id}/contributions", s.viewContributions)
			r.Get("/publication/{id}", s.pagePublication)
			r.Get("/dewey/{id}", s.pageDewey)
			r.Get("/dewey/{id}/partsof", s.viewDeweyPartsOf)
		})

		r.Route("/maintenance", func(r chi.Router) {
			r.Get("/", s.pageMaintenance)
			r.Get("/runs", s.viewJobRuns)
			r.Post("/schedule", s.scheduleJob)
			r.Get("/schedules", s.viewSchedules)
			r.Delete("/schedule/{id}", s.deleteSchedule)
			r.Route("/run", func(r chi.Router) {
				r.Post("/", s.runJob)
				r.Get("/{id}/output", s.viewJobRunOutput)
			})
		})
	})

	return r
}

// Open starts to listen at the Server's host:port address, and starts
// serving incomming connections.
func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	go s.srv.Serve(s.ln)
	return nil
}

// Close closes the Server, and peform a graceful shutdown (TODO verify)
func (s *Server) Close() error {
	// TODO verify that this is run
	ctx, cancel := context.WithTimeout(s.srv.BaseContext(s.ln), 1*time.Second)
	defer cancel()
	if s.idx != nil {
		s.idx.Close() // TODO check err?
	}
	return s.srv.Shutdown(ctx)
}

func (s *Server) indexResources(res []sirkulator.Resource) {
	if s.idx == nil {
		return
	}

	var docs []search.Document
	for _, r := range res {
		docs = append(docs, search.Document{
			ID:        r.ID,
			Type:      r.Type.String(),
			Label:     r.Label,
			Gain:      1.0,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	if err := s.idx.Store(docs...); err != nil {
		log.Println(err) // TODO or not
	}
}

func (s *Server) image(w http.ResponseWriter, r *http.Request) {
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	id := chi.URLParam(r, "id")
	var rowID int64
	var imgType string
	fn := func(stmt *sqlite.Stmt) error {
		rowID = stmt.ColumnInt64(0)
		imgType = stmt.ColumnText(1)
		return nil
	}
	const q = "SELECT rowid, type FROM files.image WHERE id=?"
	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		http.NotFound(w, r)
		return
	}
	if rowID == 0 {
		http.NotFound(w, r)
		return
	}
	blob, err := conn.OpenBlob("files", "image", "data", rowID, false)
	if err != nil {
		ServerError(w, err)
		return
	}
	defer blob.Close()

	w.Header().Set("Content-Type", "image/"+imgType)
	io.Copy(w, blob)
}

func (s *Server) pageHome(w http.ResponseWriter, r *http.Request) {
	// 404 not found handler goes here
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := html.HomeTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) pageCirculation(w http.ResponseWriter, r *http.Request) {
	tmpl := html.CircTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) pageMetadata(w http.ResponseWriter, r *http.Request) {
	tmpl := html.MetadataTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewReviews(w http.ResponseWriter, r *http.Request) {
	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10 // default size
	}

	res, err := sql.GetAllReviews(conn, limit)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewReviews{
		Reviews:   res,
		Localizer: r.Context().Value("localizer").(localizer.Localizer),
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) importResources(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ids := r.PostForm.Get("identifiers")
	if ids == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	ing := etl.NewIngestor(s.db, s.idx)
	ing.ImageDownload = true
	ing.ImageAsync = true
	var res []html.ImportResultEntry
	for _, id := range strings.Split(ids, "\n") {
		if len(strings.TrimSpace(id)) < 10 {
			// TODO proper validation and detection of type of ID: ISBN/GTIN/ISSN
			continue
		}

		entry := html.ImportResultEntry{
			IDType: "ISBN",
			ID:     id,
			Data:   ing.IngestISBN(r.Context(), id, true),
		}
		res = append(res, entry)
	}
	tmpl := html.ImportResultsTmpl{
		Entries: res,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) importPreview(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ids := r.PostForm.Get("identifiers")
	if ids == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	ing := etl.NewPreviewIngestor(s.db)
	ing.ImageDownload = false
	var res []html.ImportResultEntry

	for _, id := range strings.Split(ids, "\n") {
		if len(strings.TrimSpace(id)) < 10 {
			// TODO proper validation and detection of type of ID: ISBN/GTIN/ISSN
			continue
		}

		p := html.ImportResultEntry{
			IDType: "ISBN",
			ID:     id,
		}
		entry := ing.IngestISBN(r.Context(), id, false)
		p.Data = entry
		res = append(res, p)

	}

	tmpl := html.ImportResultsTmpl{
		Entries: res,
	}

	tmpl.Render(r.Context(), w)
}

func (s *Server) searchResources(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	q := r.PostForm.Get("q")
	resType := r.PostForm.Get("type")

	// TODO sortby/direction is rather cumbersome without client side state in javascript;
	// consider something like alpine.js (but only if there are many other use cases)
	sortBy := r.PostForm.Get("sort_by")
	sortAsc := false
	sortDir := "-" // descending
	if r.PostForm.Get("sort_asc") == "false" {
		sortAsc = true // toggle
		sortDir = ""   // ascending
	}

	res, err := s.idx.Search(r.Context(), q, resType, sortBy, sortDir, 10)
	if err != nil {
		// TODO do we filter out all user errors above in parseform?
		ServerError(w, err)
		return
	}

	tmpl := html.SearchResultsTmpl{
		Results: res,
		SortBy:  sortBy,
		SortAsc: sortAsc,
	}
	tmpl.Render(r.Context(), w)
}

// ServerError logs the given error before responding to the request
// with an Internal Server error.
// TODO consider taking an error code and (optional) message,
// which can be conveyed to client, ie.:
//   "Error code: xyz. If the problem persist, please inform system administrator
//   with references to the error code."
func ServerError(w http.ResponseWriter, err error) {
	// Internal Server errors are normally something that "should not happen",
	// and therefor interesting to log.
	log.Println(err)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintln(w, http.StatusText(http.StatusInternalServerError))
}
