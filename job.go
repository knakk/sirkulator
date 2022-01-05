package sirkulator

import (
	"context"
	"io"
)

// Job represents a runnable job, typically set up to perform some
// side effect or interacting with sirkulator data.
// TODO arguments to job? either in Run or separat init method
type Job interface {

	// Name of the job.
	Name() string

	// Run runs the job.
	// Any logging or other relevant output produced should be written to the given writer.
	Run(ctx context.Context, w io.Writer) error
}
