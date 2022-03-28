package http

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/knakk/sirkulator"
	"github.com/knakk/sirkulator/http/html"
	"github.com/knakk/sirkulator/internal/localizer"
)

func (s *Server) pageMaintenance(w http.ResponseWriter, r *http.Request) {
	tmpl := html.MaintenanceTemplate{
		Page: html.Page{
			Lang: s.Lang,
			Path: r.URL.Path,
		},
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewJobRuns(w http.ResponseWriter, r *http.Request) {
	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		limit = 10 // default size
	}

	runs, err := s.runner.JobRuns(r.Context(), limit)
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewJobRuns{
		Runs:      runs,
		Localizer: r.Context().Value("localizer").(localizer.Localizer),
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) viewSchedules(w http.ResponseWriter, r *http.Request) {

	schedules, err := s.runner.Schedules(r.Context())
	if err != nil {
		ServerError(w, err)
		return
	}

	tmpl := html.ViewSchedules{
		JobNames:  s.runner.JobNames(),
		Schedules: schedules,
		Localizer: r.Context().Value("localizer").(localizer.Localizer),
	}
	tmpl.Render(r.Context(), w)
}

func (s *Server) scheduleJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	jobName := r.PostForm.Get("job_name")
	cronExpr := r.PostForm.Get("cron_expr")
	if jobName == "" || cronExpr == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	_, err := s.runner.ParseCron(cronExpr)
	if err != nil {
		log.Println(err) // TODO investigate what kind of errors
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.runner.ScheduleJob(r.Context(), jobName, cronExpr); err != nil {
		ServerError(w, err)
	}
	w.Header().Add("HX-Trigger", "jobScheduled")
}

func (s *Server) deleteSchedule(w http.ResponseWriter, r *http.Request) {
	idparam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := s.runner.DeleteSchedule(r.Context(), int64(id)); err != nil {
		if errors.Is(err, sirkulator.ErrNotFound) {
			http.NotFound(w, r)
		} else {
			ServerError(w, err)
		}
		return
	}
	w.Header().Add("HX-Trigger", "scheduleDeleted")
}

func (s *Server) viewJobRunOutput(w http.ResponseWriter, r *http.Request) {
	idparam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idparam)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	run, err := s.runner.GetJobRun(r.Context(), int64(id))
	if errors.Is(err, sirkulator.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		ServerError(w, err)
		return
	}
	w.Header().Add("Content-Type", "text/plain")
	io.WriteString(w, run.Output)
}

func (s *Server) runJob(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	jobName := r.PostForm.Get("job_name")
	if jobName == "" {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if _, _, err := s.runner.RunJob(context.Background(), jobName); err != nil {
		ServerError(w, err)
		return
	}
	w.Header().Add("HX-Trigger", "runTriggered")
}
