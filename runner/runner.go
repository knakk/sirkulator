package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"crawshaw.io/sqlite"
	"crawshaw.io/sqlite/sqlitex"
	"github.com/knakk/sirkulator"
	"github.com/robfig/cron/v3"
)

// Job represents a runnable job, typically set up to perform some
// side effect or interacting with sirkulator data.
type Job interface {

	// Name of the job.
	Name() string

	// Run runs the job.
	// Any logging or other relevant output produced should be written to the given writer.
	Run(ctx context.Context, w io.Writer) error
}

// JobRun represents a run of a job that has been started. It corresponds
// to a row in the job_run table.
type JobRun struct {
	ID      int64
	Name    string
	StartAt time.Time
	StopAt  time.Time
	Status  string // running, ok, failed, cancelled
	Output  string
}

// Runner executes jobs, either ad-hoc or by cron-schedule.
type Runner struct {
	db   *sqlitex.Pool
	cron *cron.Cron

	mu   sync.RWMutex
	jobs map[string]Job
	//running map[int]Job
}

func New(db *sqlitex.Pool) *Runner {
	r := Runner{
		db:   db,
		cron: cron.New(cron.WithLogger(cron.DefaultLogger)), // cron.WithChain(wrapStoreInDB, wrapNotifyDone etc)
		jobs: make(map[string]Job),
	}
	r.cron.Start()
	return &r
}

func (r *Runner) Stop() context.Context {
	return r.cron.Stop()
}

// Register the given job. If a Job exists with the same name,
// it will be overwritten
func (r *Runner) Register(job Job) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[job.Name()] = job
}

// Start a job immedeatly. Returns the job run ID and a channel which will send
// when the job is completed.
func (r *Runner) Start(ctx context.Context, jobName string) (int64, chan struct{}, error) {
	r.mu.RLock()
	job, ok := r.jobs[jobName]
	defer r.mu.RUnlock()
	if !ok {
		return 0, nil, sirkulator.ErrNotFound // TODO annotate
	}
	// cron does not support to run a job immediately, so we register it and
	// start it manually
	now := time.Now()
	id, err := r.startJob(ctx,
		JobRun{
			Name:    jobName,
			StartAt: now,
			Status:  "running",
		})
	if err != nil {
		return 0, nil, fmt.Errorf("Runner.Start(%q) %w", jobName, err)
	}

	done := make(chan struct{})
	go func() {
		var b bytes.Buffer
		runErr := job.Run(ctx, &b)
		if err := r.doneJob(context.Background(), id, b.Bytes(), runErr); err != nil {
			log.Println(err) // TODO where to propagate this error
		}
		done <- struct{}{}
	}()

	return id, done, nil
}

func (r *Runner) startJob(ctx context.Context, job JobRun) (int64, error) {
	conn := r.db.Get(ctx)
	if conn == nil {
		return 0, context.Canceled
	}
	defer r.db.Put(conn)

	stmt := conn.Prep(`
		INSERT INTO job_run (name, start_at, status)
			VALUES ($name, $start_at, 'running')
		RETURNING id
	`)
	stmt.SetText("$name", job.Name)
	stmt.SetInt64("$start_at", job.StartAt.Unix())
	if ok, err := stmt.Step(); err != nil {
		return 0, err // TODO annotate
	} else if !ok {
		return 0, errors.New("startJob: no id returned")
	}

	id := stmt.GetInt64("id")
	stmt.Reset() // TODO why necessary?

	return id, nil
}

func (r *Runner) doneJob(ctx context.Context, id int64, output []byte, err error) error {
	conn := r.db.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer r.db.Put(conn)

	status := "ok"
	if errors.Is(err, context.Canceled) {
		status = "cancelled"
	} else if err != nil {
		status = "failed"
	}
	stmt := conn.Prep(`
		UPDATE job_run
		SET stop_at=$stop_at,
			status=$status,
			output=$output
		WHERE id=$id
	`)
	stmt.SetInt64("$stop_at", time.Now().Unix())
	stmt.SetText("$status", status)
	stmt.SetBytes("$output", output)
	stmt.SetInt64("$id", id)
	if _, err := stmt.Step(); err != nil {
		return fmt.Errorf("doneJob(%d): %w", id, err)
	}
	return nil
}

func (r *Runner) getJobRun(ctx context.Context, id int64) (JobRun, error) {
	var run JobRun
	conn := r.db.Get(ctx)
	if conn == nil {
		return run, context.Canceled
	}
	defer r.db.Put(conn)

	const q = `
		SELECT
			name,
			start_at,
			stop_at,
			status,
			output
		 FROM job_run
		WHERE id=?`

	fn := func(stmt *sqlite.Stmt) error {
		run.ID = id
		run.Name = stmt.ColumnText(0)
		run.StartAt = time.Unix(stmt.ColumnInt64(1), 0)
		run.StopAt = time.Unix(stmt.ColumnInt64(2), 0)
		run.Status = stmt.ColumnText(3)
		run.Output = stmt.ColumnText(4)
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, id); err != nil {
		return run, fmt.Errorf("getJobRun(%d): %w", id, err)
	}

	return run, nil
}

/*
func (r *Runner) ListJobs() []string
func (r *Runner) ListScheduledJobs() [2]string
func (r *Runner) ScheduleJob(jobName string, cronExpr string) error
func (r *Runner) Cancel(int)
*/
