package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/sql"
)

func (s *Server) pageCorporation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
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

	relations, err := sql.GetRelationsAsObject(conn, id)
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
		Relations:     relations,
	}
	tmpl.Render(r.Context(), w)
}
