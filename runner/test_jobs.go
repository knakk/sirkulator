package runner

import (
	"context"
	"fmt"
	"io"
	"time"
)

type TestJobQuick struct{}
type TestJobSlow struct{}
type TestJobSlowest struct{}

func (j TestJobQuick) Name() string   { return "test_job_quick" }
func (j TestJobSlow) Name() string    { return "test_job_slow" }
func (j TestJobSlowest) Name() string { return "test_job_slowest" }

func (j TestJobQuick) Run(ctx context.Context, w io.Writer) error   { return testJobRun(ctx, w, j, 0) }
func (j TestJobSlow) Run(ctx context.Context, w io.Writer) error    { return testJobRun(ctx, w, j, 10) }
func (j TestJobSlowest) Run(ctx context.Context, w io.Writer) error { return testJobRun(ctx, w, j, 60) }

func testJobRun(ctx context.Context, w io.Writer, job Job, secDelay int) error {
	delay := time.Duration(secDelay) * time.Second
	fmt.Fprintf(w, "starting job %q\n", job.Name())
	select {
	case <-ctx.Done():
		fmt.Fprintf(w, "cancelled job %q\n", job.Name())
		return context.Canceled
	case <-time.After(delay):
		break
	}

	fmt.Fprintf(w, "done job %q\n", job.Name())

	return nil
}
