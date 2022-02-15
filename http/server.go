package http

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/etl"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/search"
	"github.com/knakk/sirkulator/sql"
	"golang.org/x/text/language"
)

//go:embed assets/*.css
var embeddedFS embed.FS

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

func NewServer(ctx context.Context, assetsDir string, db *sqlitex.Pool, idx *search.Index) *Server {
	s := Server{
		Addr: "localhost:0", // assign random port as default, usefull for testing
		db:   db,
		idx:  idx,
	}
	s.srv = &http.Server{
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		ReadHeaderTimeout: 20 * time.Second,
		IdleTimeout:       120 * time.Second,
		//ErrorLog: TODO
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
	//r.Get("/favicon.ico", ...)

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
			r.Get("/person/{id}", s.pagePerson)
			//r.Get("/publication/{id}", s.pagePublication)
		})
	})

	return r

}

func (s *Server) Open() (err error) {
	if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
		return err
	}

	go s.srv.Serve(s.ln)
	return nil
}

func (s *Server) Close() error {
	// TODO verify that this is run
	ctx, cancel := context.WithTimeout(s.srv.BaseContext(s.ln), 1*time.Second)
	defer cancel()
	if s.idx != nil {
		s.idx.Close() // TODO check err?
	}
	return s.srv.Shutdown(ctx)
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

func (s *Server) pagePerson(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		// TODO which statuscode/response is appropriate?
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypePerson, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	tmpl := html.PersonTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource: res,
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
	numOK := 0
	for _, id := range strings.Split(ids, "\n") {
		// TODO detect type of ID: ISBN/EAN/ISSN
		entry := html.ImportResultEntry{
			IDType: "ISBN",
			ID:     id,
		}
		err := ing.IngestISBN(r.Context(), id)
		if err != nil {
			entry.Err = err.Error()
		} else {
			numOK++
		}
	}
	tmpl := html.ImportResultsTmpl{
		NumOK:   numOK, // TODO remove
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
	ing.ImageAsync = true
	var res []html.ImportPreviewEntry
	for _, id := range strings.Split(ids, "\n") {
		// TODO detect type of ID: ISBN/EAN/ISSN
		p := html.ImportPreviewEntry{
			IDType: "ISBN",
			ID:     id,
		}
		data, err := ing.PreviewISBN(r.Context(), id)
		if err == nil {
			p.Data = data
		} else {
			p.Err = err.Error()
		}
		res = append(res, p)

	}

	tmpl := html.ImportPreviewTmpl{
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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	tmpl := html.SearchResultsTmpl{
		Results: res,
		SortBy:  sortBy,
		SortAsc: sortAsc,
	}
	tmpl.Render(r.Context(), w)

}
