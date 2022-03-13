package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/sql"
	"github.com/knakk/sirkulator/vocab"
)

func (s *Server) pagePublication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	l, _ := r.Context().Value("localizer").(localizer.Localizer)

	conn := s.db.Get(r.Context())
	if conn == nil {
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

	rel, err := sql.GetPublcationRelations(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}
	// Localize type label
	for i, r := range rel {
		rel[i].Relation.Type = vocab.ParseRelation(r.Type).Label(l.Lang)
	}
	// TODO:Group contribution relations by agent

	img, _ := sql.GetImage(conn, id) // img is nil if err != nil TODO log err if err != ErrNotFound?

	tmpl := html.PublicationTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:  res,
		Relations: rel,
		Image:     img,
	}
	tmpl.Render(r.Context(), w)
}
