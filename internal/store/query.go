package store

import (
	"database/sql"
	"time"
)

// Run represents a single recorded cron job execution.
type Run struct {
	ID        int64
	Command   string
	StartedAt time.Time
	Duration  time.Duration
	ExitCode  int
	Stdout    string
	Stderr    string
}

// Insert records a completed run into the database.
func Insert(db *sql.DB, r Run) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO runs (command, started_at, duration_ms, exit_code, stdout, stderr)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		r.Command,
		r.StartedAt.UTC().Format(time.RFC3339Nano),
		r.Duration.Milliseconds(),
		r.ExitCode,
		r.Stdout,
		r.Stderr,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// List returns the most recent `limit` runs, newest first.
func List(db *sql.DB, limit int) ([]Run, error) {
	rows, err := db.Query(
		`SELECT id, command, started_at, duration_ms, exit_code, stdout, stderr
		 FROM runs ORDER BY started_at DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		var startedAtStr string
		var durationMs int64
		if err := rows.Scan(&r.ID, &r.Command, &startedAtStr, &durationMs, &r.ExitCode, &r.Stdout, &r.Stderr); err != nil {
			return nil, err
		}
		r.StartedAt, err = time.Parse(time.RFC3339Nano, startedAtStr)
		if err != nil {
			return nil, err
		}
		r.Duration = time.Duration(durationMs) * time.Millisecond
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

// GetByID retrieves a single run by its primary key.
func GetByID(db *sql.DB, id int64) (*Run, error) {
	row := db.QueryRow(
		`SELECT id, command, started_at, duration_ms, exit_code, stdout, stderr
		 FROM runs WHERE id = ?`, id,
	)
	var r Run
	var startedAtStr string
	var durationMs int64
	err := row.Scan(&r.ID, &r.Command, &startedAtStr, &durationMs, &r.ExitCode, &r.Stdout, &r.Stderr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.StartedAt, err = time.Parse(time.RFC3339Nano, startedAtStr)
	if err != nil {
		return nil, err
	}
	r.Duration = time.Duration(durationMs) * time.Millisecond
	return &r, nil
}
