package http

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
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
	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/etl"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/search"
	"github.com/knakk/sirkulator/sql"
	"github.com/knakk/sirkulator/vocab"
	"golang.org/x/text/language"
)

//go:embed assets/*.css
var embeddedFS embed.FS

// Server represents the HTTP server responsible for serving Sirkulators admin interface.
type Server struct {
	ln  net.Listener
	srv *http.Server
	db  *sqlitex.Pool
	idx *search.Index

	// The follwing fields should be set before calls to Open:

	// Addr is the bind address for the tcp listener.
	Addr string
	// Default language
	Lang language.Tag
}

// NewServer returns a new Server with the given database and index and assets settings.
func NewServer(ctx context.Context, assetsDir string, db *sqlitex.Pool, idx *search.Index) *Server {
	s := Server{
		Addr: "localhost:0", // assign random port as default, useful for testing
		db:   db,
		idx:  idx,
	}
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
			r.Post("/import", s.importResources) // s.tmplImportResponse ?
			r.Post("/preview", s.importPreview)
			r.Post("/search", s.searchResources)
			r.Get("/corporation/{id}", s.pageCorporation)
			r.Post("/person/{id}", s.savePerson)
			r.Get("/person/{id}", s.pagePerson)
			r.Post("/person/{id}/contributions", s.viewContributions)
			r.Get("/publication/{id}", s.pagePublication)
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

func (s *Server) image(w http.ResponseWriter, r *http.Request) {
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}

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

func splitAndClean(s string) []string {
	var res []string
	for _, line := range strings.Split(s, "\n") {
		if entry := strings.TrimSpace(line); entry != "" {
			res = append(res, entry)
		}
	}
	return res
}

func joinWith(ss []string, prefix string) []string {
	for i, s := range ss {
		ss[i] = prefix + s
	}
	return ss
}

func (s *Server) savePerson(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Load resource
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypePerson, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	// Check that resource hasn't been updated by some other process
	l, _ := r.Context().Value("localizer").(localizer.Localizer)
	if updatedAt := r.PostForm.Get("updated_at"); updatedAt != strconv.Itoa(int(res.UpdatedAt.Unix())) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(l.Translate("Not saved. Resource has been updated by some else.")))
		w.Write([]byte(`<a href="/metadata/person/` + id + `" target="_blank">`))
		w.Write([]byte(l.Translate("Open this page in a new tab")))
		w.Write([]byte("</a> "))
		w.Write([]byte(l.Translate("to verify and redo your changes.")))
		return
	}

	// Validate input
	valid := true
	changed := false
	var newP sirkulator.Person
	oldP := res.Data.(*sirkulator.Person)

	newP.Name = strings.TrimSpace(r.PostFormValue("name"))
	if newP.Name == "" {
		valid = false
	}
	newP.Description = strings.TrimSpace(r.PostFormValue("description"))
	newP.NameVariations = splitAndClean(r.PostFormValue("name_variations"))
	newP.YearRange.From = json.Number(strings.TrimSpace(r.PostFormValue("year_range.from")))
	newP.YearRange.To = json.Number(strings.TrimSpace(r.PostFormValue("year_range.to")))
	if r.PostFormValue("year_range.approx") == "on" {
		newP.YearRange.Approx = true
	}
	if !newP.YearRange.Valid() {
		valid = false
	}
	newP.Gender = vocab.ParseGender(r.PostFormValue("gender"))
	newP.Countries = joinWith(r.PostForm["countries"], "iso3166/")
	newP.Nationalities = joinWith(r.PostForm["nationalities"], "bs/")

	if diff := cmp.Diff(oldP, &newP); diff != "" {
		changed = true
		fmt.Println(diff) // TODO remove
	}

	if !valid {
		w.WriteHeader(http.StatusBadRequest)
		tmpl := html.PersonForm{
			Person:    &newP,
			UpdatedAt: res.UpdatedAt.Unix(),
			Localizer: l,
		}
		tmpl.Render(r.Context(), w)
		return
	}

	if !changed {
		// No changes to resource, no point in saving to DB
		http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
		return
	}

	// Validation passed, save resource
	if err := sql.UpdateResource(conn, sirkulator.Resource{
		ID:   id,
		Type: sirkulator.TypePerson,
		Data: newP,
	}); err != nil {
		ServerError(w, err)
		return
	}

	// TODO load resource from DB again?
	// or make sql.UpdateResource return updated resource?

	tmpl := html.PersonForm{
		Person:    &newP,
		UpdatedAt: res.UpdatedAt.Unix(),
		Localizer: l,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) pagePerson(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypePerson, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	contrib, err := sql.GetAgentContributions(conn, id, "year", false)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.PersonTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:      res,
		Contributions: contrib,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewContributions(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	sortBy := r.PostForm.Get("sort_by")
	sortAsc := false
	if r.PostForm.Get("sort_asc") == "false" {
		sortAsc = true // toggle
	}

	contrib, err := sql.GetAgentContributions(conn, id, sortBy, sortAsc)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewContributions{
		Contributions: contrib,
		SortBy:        sortBy,
		SortAsc:       sortAsc,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) pagePublication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypePublication, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	contrib, err := sql.GetPublcationContributors(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}
	reviews, err := sql.GetPublcationReviews(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	img, _ := sql.GetImage(conn, id) // img is nil if err != nil TODO log err if err != ErrNotFound?

	tmpl := html.PublicationTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:      res,
		Contributions: contrib,
		Reviews:       reviews,
		Image:         img,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) pageCorporation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypeCorporation, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	contrib, err := sql.GetAgentContributions(conn, id, "year", false)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.CorporationTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:      res,
		Contributions: contrib,
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
