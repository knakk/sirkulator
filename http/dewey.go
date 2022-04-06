package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/sql"
)

func (s *Server) pageDewey(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)
	res, err := sql.GetResource(conn, sirkulator.TypeDewey, id)
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}

	parents, err := sql.GetDeweyParents(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	children, err := sql.GetDeweyChildren(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	parts, err := sql.GetDeweyParts(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	numPartsOf, err := sql.GetDeweyPartsOfCount(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	pubCount, err := sql.GetDeweyPublicationsCount(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	pubSubCount, err := sql.GetDeweySubPublicationsCount(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.DeweyTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource:             res,
		Parents:              parents,
		Children:             children,
		Parts:                parts,
		PartsOfCount:         numPartsOf,
		PublicationsCount:    pubCount,
		PublicationsSubCount: pubSubCount,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewDeweyPartsOf(w http.ResponseWriter, r *http.Request) {
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

	id := chi.URLParam(r, "id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	partsOf, hasMore, err := sql.GetDeweyPartsOf(conn, id, from, to, limit)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewDeweyPartsOf{
		ID:      id,
		From:    from,
		To:      to,
		HasMore: hasMore,
		PartsOf: partsOf,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewDeweyPublications(w http.ResponseWriter, r *http.Request) {
	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	/*limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10 // default size
	}*/

	id := chi.URLParam(r, "id")
	inclSub := r.URL.Query().Get("include_subdewey") != ""

	publications, err := sql.GetDeweyPublications(conn, id, inclSub)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewDeweyPublications{
		Publications: publications,
	}
	tmpl.Render(r.Context(), w)
}
