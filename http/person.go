package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
	"github.com/knakk/sirkulator/sql"
	"github.com/knakk/sirkulator/vocab"
)

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
	}, newP.Label()); err != nil {
		ServerError(w, err)
		return
	}

	// Load resource from DB again
	// TODO or make sql.UpdateResource return updated resource?
	res, err = sql.GetResource(conn, sirkulator.TypePerson, id)
	if err != nil {
		ServerError(w, err)
		return
	}
	go s.indexResources([]sirkulator.Resource{res})

	tmpl := html.PersonForm{
		Person:    res.Data.(*sirkulator.Person),
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

// TODO also used by corporation - move out to agent.go?
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
