package store

import (
	"database/sql"
	"time"
)

// WatchOptions controls which runs are returned by Watch.
type WatchOptions struct {
	Command string
	Since   time.Time
}

// WatchResult holds a single run returned from Watch.
type WatchResult struct {
	Run
	Tags []string
}

// Watch returns runs that completed after opts.Since, optionally filtered by
// command. Results are ordered oldest-first so callers can tail the log.
func Watch(db *sql.DB, opts WatchOptions) ([]WatchResult, error) {
	query := `
		SELECT r.id, r.command, r.started_at, r.finished_at,
		       r.exit_code, r.stdout, r.stderr
		FROM runs r
		WHERE r.finished_at > ?
	`
	args := []any{opts.Since.UTC().Format(time.RFC3339Nano)}

	if opts.Command != "" {
		query += " AND r.command = ?"
		args = append(args, opts.Command)
	}
	query += " ORDER BY r.finished_at ASC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []WatchResult
	for rows.Next() {
		var wr WatchResult
		if err := rows.Scan(
			&wr.ID, &wr.Command, &wr.StartedAt, &wr.FinishedAt,
			&wr.ExitCode, &wr.Stdout, &wr.Stderr,
		); err != nil {
			return nil, err
		}
		tags, err := ListTags(db, wr.ID)
		if err != nil {
			return nil, err
		}
		wr.Tags = tags
		results = append(results, wr)
	}
	return results, rows.Err()
}
