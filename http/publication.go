package http

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/marc"
	"github.com/knakk/sirkulator/sql"
	"github.com/knakk/sirkulator/vocab"
)

func (s *Server) pagePublication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

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

	img, _ := sql.GetImage(conn, id) // img is nil if err != nil TODO log err if err != ErrNotFound?

	tmpl := html.PublicationTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
		Resource: res,
		Image:    img,
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewPublicationRelations(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	l, _ := r.Context().Value("localizer").(localizer.Localizer)

	conn := s.db.Get(r.Context())
	if conn == nil {
		http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
		return
	}
	defer s.db.Put(conn)

	rel, err := sql.GetPublicationRelations(conn, id)
	if err != nil {
		ServerError(w, err)
		return
	}

	// Localize type label, and group contributor nodes to one relation per agent
	for i := len(rel) - 1; i >= 0; i-- {
		// We loop backwards, which make it easier to remove relation if necessary
		r := rel[i]
		if r.Type == "has_contributor" {
			newAgent := true
			var roleLabel string
			if role, ok := r.Data["role"].(string); ok {
				relator, err := marc.ParseRelator(role)
				if err == nil {
					roleLabel = relator.Label(l.Lang)
				}
			}

			// Check if we allready have a relation to agent
			for j := len(rel) - 1; j > i; j-- {
				if r.ToID == rel[j].ToID && rel[j].Data["role"] != nil { // TODO find a more reliable way of matching "has_contributor" relations
					// Append role to relation type label
					rel[j].Relation.Type = fmt.Sprintf("%s, %s", rel[j].Relation.Type, roleLabel)

					// Remove current relation
					rel = append(rel[:i], rel[i+1:]...)

					newAgent = false
					break
				}
			}

			if newAgent {
				rel[i].Relation.Type = roleLabel
			}
		} else {
			rel[i].Relation.Type = vocab.ParseRelation(r.Type).Label(l.Lang)
		}
	}

	sort.Slice(rel, func(i, j int) bool {
		return rel[i].Type < rel[j].Type
	})

	tmpl := html.ViewPublicationRelations{
		Relations: rel,
	}
	tmpl.Render(r.Context(), w)
}
