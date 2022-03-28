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
	Status  string // running|ok|failed|cancelled
	Output  string
	// The following fields are only present in active jobs (scheduled or currently running)
	CronID cron.EntryID       // id needed for removing scheduled job from cron
	Cancel context.CancelFunc // to cancel a job
	Done   chan struct{}      // if you need to wait until job is done
}

// Schedule corresponds to a row in the job_schedule table.
type Schedule struct {
	ID   int64
	Name string
	Cron string
}

// Runner executes jobs, either ad-hoc or by cron-schedule.
type Runner struct {
	db   *sqlitex.Pool
	cron *cron.Cron

	mu         sync.RWMutex           // mu protectes the following 3 maps
	jobs       map[string]Job         // available jobs
	running    map[int64]JobRun       // currently running jobs
	id2entryID map[int64]cron.EntryID // job.shedule.id => EntryID in cron
}

func New(db *sqlitex.Pool) *Runner {
	r := Runner{
		db:         db,
		cron:       cron.New(cron.WithLogger(cron.DefaultLogger), cron.WithSeconds()), // TODO use our own logger
		jobs:       make(map[string]Job),
		running:    make(map[int64]JobRun),
		id2entryID: make(map[int64]cron.EntryID),
	}

	return &r
}

func (r *Runner) Start(ctx context.Context) error {
	r.registerDefaultJobs()
	if err := r.loadSchedules(ctx); err != nil {
		return err
	}
	// In case server stopped/crashed while jobs where running, they will still have status 'running' in db
	// Close those as 'crashed'
	if err := r.closeActiveRuns(context.Background()); err != nil {
		return err
	}

	r.cron.Start()
	return nil
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

func (r *Runner) registerDefaultJobs() *Runner {
	//r.Register(TestJobQuick{})
	//r.Register(TestJobSlow{})
	//r.Register(TestJobSlowest{})
	return r
}

func (r *Runner) ParseCron(s string) (cron.Schedule, error) {
	return cron.NewParser(
		cron.Second |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.Dow |
			cron.Descriptor).Parse(s)
}

func (r *Runner) GetJob(name string) (Job, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	job, ok := r.jobs[name]
	return job, ok
}

func (r *Runner) mapScheduleID(id int64, cronID cron.EntryID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.id2entryID[id] = cronID
}

func (r *Runner) unmapScheduleID(id int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.id2entryID, id)
}

func (r *Runner) getCronID(id int64) (cron.EntryID, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cronID, ok := r.id2entryID[id]
	return cronID, ok
}

func (r *Runner) ScheduleJob(ctx context.Context, jobName string, cronExpr string) error {
	job, ok := r.GetJob(jobName)
	if !ok {
		return sirkulator.ErrNotFound // TODO annotate?
	}

	schedule, err := r.ParseCron(cronExpr)
	if err != nil {
		return err
	}

	id, err := r.insertSchedule(ctx, jobName, cronExpr)
	if err != nil {
		return err // TODO annotate
	}

	var cronID cron.EntryID
	cronID = r.cron.Schedule(schedule, WrapForCron(r, job, &cronID))
	r.mapScheduleID(id, cronID)

	return nil
}

func (r *Runner) isRunning(job Job) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, run := range r.running {
		if run.Name == job.Name() {
			return true
		}
	}
	return false
}

func WrapForCron(r *Runner, job Job, cronID *cron.EntryID) cron.Job {
	return cron.FuncJob(func() {
		if r.isRunning(job) {
			log.Printf("runner: skipping job %q; already running", job.Name())
			return
		}

		ctx := context.Background() // TODO or use Server.srv.BaseContext?

		run, err := r.insertRun(ctx, job)
		if err != nil {
			log.Printf("runner: error inserting job %q in DB: %v", job.Name(), err)
			return
		}

		ctx, cancel := context.WithCancel(ctx)
		run.Cancel = cancel
		done := make(chan struct{})
		run.Done = done
		run.CronID = *cronID
		var b bytes.Buffer // TODO consider streaming to DB using conn.OpenBlob?

		r.setRunning(run) // TODO merge with isRunning at start of fn, to guarantee no race runs?
		runErr := job.Run(ctx, &b)
		if runErr != nil {
			fmt.Fprintf(&b, "\nfailed with: %v", runErr)
		}
		if err := r.doneJob(context.Background(), run.ID, b.Bytes(), runErr); err != nil {
			log.Println(err) // TODO where to propagate this error
		}
		r.unsetRunning(run)

		done <- struct{}{}
	})

}

// RunJob runs a job immedeatly. Returns the job run ID and a channel which will send
// when the job is completed.
// TODO return JobRun instead, align with cron scheduled jobs (running map etc)
func (r *Runner) RunJob(ctx context.Context, jobName string) (int64, chan struct{}, error) {
	r.mu.RLock()
	job, ok := r.jobs[jobName]
	defer r.mu.RUnlock()
	if !ok {
		return 0, nil, sirkulator.ErrNotFound // TODO annotate
	}
	// cron does not support to run a job immediately, so we register it and
	// start it manually
	// TODO schedule using cron, 1 second from time.Now?
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
		if runErr != nil {
			fmt.Fprintf(&b, "\nfailed with: %v", runErr)
		}
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

func (r *Runner) insertRun(ctx context.Context, job Job) (JobRun, error) {
	var run JobRun
	conn := r.db.Get(ctx)
	if conn == nil {
		return run, context.Canceled
	}
	defer r.db.Put(conn)

	startAt := time.Now()
	stmt := conn.Prep(`
		INSERT INTO job_run (name, start_at, status)
			VALUES ($name, $start_at, 'running')
		RETURNING id
	`)
	stmt.SetText("$name", job.Name())
	stmt.SetInt64("$start_at", startAt.Unix())
	if ok, err := stmt.Step(); err != nil {
		return run, err // TODO annotate
	} else if !ok {
		return run, errors.New("insertRun: no id returned")
	}

	run.ID = stmt.GetInt64("id")
	run.Name = job.Name()
	run.StartAt = startAt
	run.Status = "running"
	stmt.Reset()

	return run, nil
}

func (r *Runner) insertSchedule(ctx context.Context, jobName, cronExpr string) (int64, error) {
	conn := r.db.Get(ctx)
	if conn == nil {
		return 0, context.Canceled
	}
	defer r.db.Put(conn)

	stmt := conn.Prep(`
		INSERT INTO job_schedule (name, cron)
			VALUES ($name, $cron)
		RETURNING id
	`)
	stmt.SetText("$name", jobName)
	stmt.SetText("$cron", cronExpr)

	if ok, err := stmt.Step(); err != nil {
		return 0, err // TODO annotate
	} else if !ok {
		return 0, errors.New("insertSchedule: no id returned")
	}

	id := stmt.GetInt64("id")
	stmt.Reset()

	return id, nil
}

func (r *Runner) DeleteSchedule(ctx context.Context, id int64) error {
	conn := r.db.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer r.db.Put(conn)

	stmt := conn.Prep("DELETE FROM job_schedule WHERE id=$id RETURNING id")
	stmt.SetInt64("$id", id)
	if ok, err := stmt.Step(); err != nil {
		return err // TODO annotate
	} else if !ok {
		return sirkulator.ErrNotFound
	}
	stmt.Reset()

	if cronID, ok := r.getCronID(id); ok {
		r.unmapScheduleID(id)
		r.cron.Remove(cronID)
	} else {
		// TODO return error?
		log.Printf("DeleteSchedule: no corresponding cron entry for schedule id=%d", id)
	}

	return nil
}

func (r *Runner) loadSchedules(ctx context.Context) error {
	conn := r.db.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer r.db.Put(conn)

	const q = "SELECT id, name, cron FROM job_schedule"
	fn := func(stmt *sqlite.Stmt) error {
		id := stmt.ColumnInt64(0)
		name := stmt.ColumnText(1)
		cronExpr := stmt.ColumnText(2)

		job, ok := r.GetJob(name)
		if !ok {
			return sirkulator.ErrNotFound
		}

		schedule, err := r.ParseCron(cronExpr)
		if err != nil {
			return err
		}

		var cronID cron.EntryID
		cronID = r.cron.Schedule(schedule, WrapForCron(r, job, &cronID))
		r.mapScheduleID(id, cronID)
		return nil

	}
	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return fmt.Errorf("loadSchedules: %w", err)
	}

	return nil
}

func (r *Runner) closeActiveRuns(ctx context.Context) error {
	conn := r.db.Get(ctx)
	if conn == nil {
		return context.Canceled
	}
	defer r.db.Put(conn)

	return sqlitex.ExecScript(conn, "UPDATE job_run SET status='crashed' WHERE status='running'")
}

func (r *Runner) Schedules(ctx context.Context) ([]Schedule, error) {
	var res []Schedule
	conn := r.db.Get(ctx)
	if conn == nil {
		return res, context.Canceled
	}
	defer r.db.Put(conn)

	const q = "SELECT id, name, cron FROM job_schedule ORDER BY name"
	fn := func(stmt *sqlite.Stmt) error {
		res = append(res, Schedule{
			ID:   stmt.ColumnInt64(0),
			Name: stmt.ColumnText(1),
			Cron: stmt.ColumnText(2),
		})
		return nil
	}

	if err := sqlitex.Exec(conn, q, fn); err != nil {
		return res, fmt.Errorf("Schedules: %w", err)
	}

	return res, nil
}

func (r *Runner) setRunning(run JobRun) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.running[run.ID] = run
}

func (r *Runner) unsetRunning(run JobRun) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.running, run.ID)
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

func (r *Runner) GetJobRun(ctx context.Context, id int64) (JobRun, error) {
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
		return run, fmt.Errorf("GetJobRun(%d): %w", id, err)
	}

	if run.ID == 0 {
		return run, sirkulator.ErrNotFound
	}

	return run, nil
}

func (r *Runner) JobRuns(ctx context.Context, limit int) ([]JobRun, error) {
	var runs []JobRun
	conn := r.db.Get(ctx)
	if conn == nil {
		return runs, context.Canceled
	}
	defer r.db.Put(conn)

	const q = `
		SELECT
			id,
			name,
			start_at,
			stop_at,
			status,
			output
		 FROM job_run
		 ORDER BY start_at DESC
		 LIMIT ?
		`

	fn := func(stmt *sqlite.Stmt) error {
		var run JobRun
		run.ID = stmt.ColumnInt64(0)
		run.Name = stmt.ColumnText(1)
		run.StartAt = time.Unix(stmt.ColumnInt64(2), 0)
		run.StopAt = time.Unix(stmt.ColumnInt64(3), 0)
		run.Status = stmt.ColumnText(4)
		run.Output = stmt.ColumnText(5)
		runs = append(runs, run)
		return nil
	}
	if err := sqlitex.Exec(conn, q, fn, limit); err != nil {
		return runs, fmt.Errorf("JobRuns(%d): %w", limit, err)
	}

	return runs, nil
}

func (r *Runner) JobNames() []string {
	var jobs []string
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, job := range r.jobs {
		jobs = append(jobs, job.Name())
	}
	return jobs
}
