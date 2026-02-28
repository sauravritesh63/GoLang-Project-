// Package logging provides a centralized, structured logger for the distributed
// task scheduler using zerolog. It supports context-enriched log entries with
// workflow and task fields for end-to-end request tracing.
package logging

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type contextKey int

const loggerKey contextKey = 0

// Logger is the package-level default logger. It writes JSON to stdout.
var Logger zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

// New returns a zerolog.Logger that writes to the supplied writer with
// timestamps. Pass os.Stderr for console-style output or a file for
// persistent log storage.
func New(w io.Writer) zerolog.Logger {
	return zerolog.New(w).With().Timestamp().Logger()
}

// WithContext returns a copy of ctx with the logger embedded.
// Retrieve it later with FromContext.
func WithContext(ctx context.Context, l zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}

// FromContext returns the logger stored in ctx, or the package-level default
// Logger if none was set.
func FromContext(ctx context.Context) zerolog.Logger {
	if l, ok := ctx.Value(loggerKey).(zerolog.Logger); ok {
		return l
	}
	return Logger
}

// WithWorkflow returns a logger with "workflow_id" and "workflow_name" fields
// pre-set. Use this when logging events scoped to a single workflow.
func WithWorkflow(l zerolog.Logger, id, name string) zerolog.Logger {
	return l.With().Str("workflow_id", id).Str("workflow_name", name).Logger()
}

// WithTask returns a logger with "task_id" and "task_name" fields pre-set.
// Use this when logging events scoped to a single task or task-run.
func WithTask(l zerolog.Logger, id, name string) zerolog.Logger {
	return l.With().Str("task_id", id).Str("task_name", name).Logger()
}

// WithWorker returns a logger with a "worker_id" field pre-set.
func WithWorker(l zerolog.Logger, id string) zerolog.Logger {
	return l.With().Str("worker_id", id).Logger()
}
