package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
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

func (s *Server) savePublisher(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	// Load resource
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

	l, _ := r.Context().Value("localizer").(localizer.Localizer)

	// Validate input
	valid := true
	changed := false
	var newP sirkulator.Publisher
	oldP := res.Data.(*sirkulator.Publisher)

	newP.Name = strings.TrimSpace(r.PostFormValue("name"))
	if newP.Name == "" {
		valid = false
	}
	newP.Notes = splitAndClean(r.PostFormValue("notes"))
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

	if diff := cmp.Diff(oldP, &newP); diff != "" {
		changed = true
		fmt.Println(diff) // TODO remove
	}

	if !valid {
		//w.WriteHeader(http.StatusBadRequest)
		tmpl := html.PublisherForm{
			Publisher:   &newP,
			UpdatedAt:   res.UpdatedAt.Unix(),
			Localizer:   l,
			SaveMessage: l.Translate("Validation failed. Check input fields."),
		}
		tmpl.Render(r.Context(), w)
		return
	}

	if !changed {
		// No changes to resource, no point in saving to DB
		tmpl := html.PublisherForm{
			Publisher:   &newP,
			UpdatedAt:   res.UpdatedAt.Unix(),
			Localizer:   l,
			SaveMessage: l.Translate("No changes."),
		}
		tmpl.Render(r.Context(), w)
		return
	}

	// Check that resource hasn't been updated by some other process
	updatedAtStr := r.PostForm.Get("updated_at")
	updatedAt, err := strconv.ParseInt(updatedAtStr, 10, 0)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	if updatedAt != res.UpdatedAt.Unix() {
		var b bytes.Buffer
		io.WriteString(&b, l.Translate("Not saved. Resource has been updated by some else."))
		io.WriteString(&b, `<a href="/metadata/publisher/`+id+`" target="_blank">`)
		io.WriteString(&b, l.Translate("Open this page in a new tab"))
		io.WriteString(&b, "</a> ")
		io.WriteString(&b, l.Translate("to verify and redo your changes."))
		tmpl := html.PublisherForm{
			Publisher:   &newP,
			UpdatedAt:   updatedAt,
			Localizer:   l,
			SaveMessage: b.String(),
		}
		tmpl.Render(r.Context(), w)
		return
	}

	// Validation passed, save resource
	if err := sql.UpdateResource(conn, sirkulator.Resource{
		ID:   id,
		Type: sirkulator.TypePublisher,
		Data: newP,
	}, newP.Label()); err != nil {
		ServerError(w, err)
		return
	}

	// Load resource from DB again
	// TODO or make sql.UpdateResource return updated resource?
	res, err = sql.GetResource(conn, sirkulator.TypePublisher, id)
	if err != nil {
		ServerError(w, err)
		return
	}
	go s.indexResources([]sirkulator.Resource{res})

	tmpl := html.PublisherForm{
		Publisher:   res.Data.(*sirkulator.Publisher),
		UpdatedAt:   res.UpdatedAt.Unix(),
		Localizer:   l,
		SaveMessage: l.Translate("OK, saved."),
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
