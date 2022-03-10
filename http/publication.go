package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/sql"
)

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
