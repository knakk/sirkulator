package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/knakk/sirkulator/sql"
)

type TestJob struct {
	dir      string
	filename string
	msg      string
	errMsg   string
	shodFail bool
	slow     bool
}

func (j TestJob) Name() string {
	return "test_job"
}

func (j TestJob) Run(ctx context.Context, w io.Writer) error {
	if j.shodFail {
		err := fmt.Errorf("failed to create file: %s", j.filename)
		fmt.Fprint(w, err.Error())
		return err
	}

	delay := 0 * time.Second
	if j.slow {
		delay = 1 * time.Second
	}
	select {
	case <-ctx.Done():
		fmt.Fprint(w, "\nJob cancelled.")
		return context.Canceled
	case <-time.After(delay):
		break
	}

	f, err := os.Create(filepath.Join(j.dir, j.filename))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(w, "file created: %s", j.filename)

	return nil
}

func TestRunner(t *testing.T) {
	dir, err := ioutil.TempDir("", "sirkulator")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	db, err := sql.OpenMem()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	t.Run("succesfull run", func(t *testing.T) {
		filename := "test.txt"
		job := TestJob{
			dir:      dir,
			filename: filename,
			msg:      fmt.Sprintf("file created: %s", filename),
		}
		r := New(db)
		r.Register(job)
		id, done, err := r.Start(context.Background(), job.Name())
		if err != nil {
			t.Fatal(err)
		}

		<-done

		// Verify that file was created.
		_, err = os.Stat(filepath.Join(dir, filename))
		if errors.Is(err, os.ErrNotExist) {
			t.Errorf("file %q not found", filename)
		}

		// Verify job run was stored in DB.
		run, err := r.getJobRun(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		// Reset timestamps as not to fail comparision below.
		run.StartAt = time.Time{}
		run.StopAt = time.Time{}

		want := JobRun{
			ID:     id,
			Name:   job.Name(),
			Status: "ok",
			Output: fmt.Sprintf("file created: %s", filename),
		}

		if diff := cmp.Diff(want, run); diff != "" {
			t.Errorf("JobRun mismatch (-want +got):\n%s", diff)
		}

	})

	t.Run("failed run", func(t *testing.T) {
		filename := "test2.txt"
		job := TestJob{
			shodFail: true,
			dir:      dir,
			filename: filename,
			errMsg:   fmt.Sprintf("failed to create file: %s", filename),
		}
		r := New(db)
		r.Register(job)
		id, done, err := r.Start(context.Background(), job.Name())
		if err != nil {
			t.Fatal(err)
		}

		<-done

		// Verify that file was not created.
		_, err = os.Stat(filepath.Join(dir, filename))
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("file %q found", filename)
		}

		// Verify job run was stored in DB.
		run, err := r.getJobRun(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		// Reset timestamps as not to fail comparision below.
		run.StartAt = time.Time{}
		run.StopAt = time.Time{}

		want := JobRun{
			ID:     id,
			Name:   job.Name(),
			Status: "failed",
			Output: job.errMsg,
		}

		if diff := cmp.Diff(want, run); diff != "" {
			t.Errorf("JobRun mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("cancel run", func(t *testing.T) {
		filename := "test3.txt"
		job := TestJob{
			slow:     true,
			dir:      dir,
			filename: filename,
			errMsg:   "\nJob cancelled.",
		}
		r := New(db)
		r.Register(job)
		ctx, cancel := context.WithCancel(context.Background())
		id, done, err := r.Start(ctx, job.Name())
		if err != nil {
			t.Fatal(err)
		}
		cancel()

		<-done

		// Verify that file was not created.
		_, err = os.Stat(filepath.Join(dir, filename))
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("file %q found", filename)
		}

		// Verify job run was stored in DB.
		run, err := r.getJobRun(context.Background(), id)
		if err != nil {
			t.Fatal(err)
		}
		// Reset timestamps as not to fail comparision below.
		run.StartAt = time.Time{}
		run.StopAt = time.Time{}

		want := JobRun{
			ID:     id,
			Name:   job.Name(),
			Status: "cancelled",
			Output: job.errMsg,
		}

		if diff := cmp.Diff(want, run); diff != "" {
			t.Errorf("JobRun mismatch (-want +got):\n%s", diff)
		}
	})
}
