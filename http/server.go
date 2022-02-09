package http

import (
	"context"
	"embed"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"crawshaw.io/sqlite/sqlitex"
	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator/etl"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"golang.org/x/text/language"
)

//go:embed assets/*.css
var embeddedFS embed.FS

type Server struct {
	ln  net.Listener
	srv *http.Server
	db  *sqlitex.Pool

	// The follwing fields should be set before calls to Open:

	// Addr is the bind address for the tcp listener.
	Addr string
	// Default language
	Lang language.Tag
}

func NewServer(ctx context.Context, assetsDir string, db *sqlitex.Pool) *Server {
	s := Server{
		Addr: "localhost:0", // assign random port as default, usefull for testing
		db:   db,
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

		r.Get("/", s.tmplHome)
		r.Get("/circulation", s.tmplCirculation)
		r.Route("/metadata", func(r chi.Router) {
			r.Get("/", s.tmplMetadata)           // s.pageMetadata ?
			r.Post("/import", s.importResources) // s.tmplImportResponse ?
			r.Post("/preview", s.importPreview)
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
	ctx, cancel := context.WithTimeout(s.srv.BaseContext(s.ln), 1*time.Second)
	defer cancel()
	return s.srv.Shutdown(ctx)
}

func (s *Server) tmplHome(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) tmplCirculation(w http.ResponseWriter, r *http.Request) {
	tmpl := html.CircTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) tmplMetadata(w http.ResponseWriter, r *http.Request) {
	tmpl := html.MetadataTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
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
	ing := etl.NewIngestor(s.db)
	ing.ImageDownload = true
	ing.ImageAsync = true
	numOK := 0
	for _, id := range strings.Split(ids, "\n") {
		// TODO detect type of ID: ISBN/EAN/ISSN
		if err := ing.IngestISBN(r.Context(), id); err == nil {
			numOK++
		}
	}
	fmt.Fprintf(w, "<div><h3>%d Imported OK!</h3></div>", numOK)
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
