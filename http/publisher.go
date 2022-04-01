package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/sql"
)

func (s *Server) pagePublisher(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypePublisher, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	pubs, err := sql.GetPublisherPublications(conn, id, "year", false)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.PublisherTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:     res,
		Publications: pubs,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewPublisherPublications(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	sortBy := r.PostForm.Get("sort_by")
	sortAsc := false
	if r.PostForm.Get("sort_asc") == "false" {
		sortAsc = true // toggle
	}

	pubs, err := sql.GetPublisherPublications(conn, id, sortBy, sortAsc)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewPublisherPublications{
		Publications: pubs,
		SortBy:       sortBy,
		SortAsc:      sortAsc,
	}
	tmpl.Render(r.Context(), w)
}
