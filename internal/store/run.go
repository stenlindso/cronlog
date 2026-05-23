package store

import "time"

// Run represents a single recorded cron job execution.
type Run struct {
	ID         int64
	Command    string
	StartedAt  time.Time
	DurationMs int64
	ExitCode   int
	Output     string
}
