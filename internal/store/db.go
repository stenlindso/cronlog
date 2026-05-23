package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps a SQLite database connection for cronlog.
type DB struct {
	conn *sql.DB
}

// JobRun represents a single recorded execution of a cron job.
type JobRun struct {
	ID        int64
	Name      string
	Command   string
	StartedAt time.Time
	Duration  time.Duration
	ExitCode  int
	Output    string
}

// Open opens (or creates) the SQLite database at the given path and
// applies the schema migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("store: open %q: %w", path, err)
	}
	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, err
	}
	return db, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// migrate creates the job_runs table if it does not already exist.
func (db *DB) migrate() error {
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS job_runs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT    NOT NULL,
			command     TEXT    NOT NULL,
			started_at  INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			exit_code   INTEGER NOT NULL,
			output      TEXT    NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("store: migrate: %w", err)
	}
	return nil
}

// Insert persists a JobRun to the database and sets its ID.
func (db *DB) Insert(run *JobRun) error {
	res, err := db.conn.Exec(
		`INSERT INTO job_runs (name, command, started_at, duration_ms, exit_code, output)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		run.Name,
		run.Command,
		run.StartedAt.UnixMilli(),
		run.Duration.Milliseconds(),
		run.ExitCode,
		run.Output,
	)
	if err != nil {
		return fmt.Errorf("store: insert: %w", err)
	}
	run.ID, _ = res.LastInsertId()
	return nil
}

// List returns all job runs ordered by most recent first.
func (db *DB) List() ([]JobRun, error) {
	rows, err := db.conn.Query(
		`SELECT id, name, command, started_at, duration_ms, exit_code, output
		 FROM job_runs ORDER BY started_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("store: list: %w", err)
	}
	defer rows.Close()

	var runs []JobRun
	for rows.Next() {
		var r JobRun
		var startedAtMs, durationMs int64
		if err := rows.Scan(&r.ID, &r.Name, &r.Command, &startedAtMs, &durationMs, &r.ExitCode, &r.Output); err != nil {
			return nil, fmt.Errorf("store: list scan: %w", err)
		}
		r.StartedAt = time.UnixMilli(startedAtMs)
		r.Duration = time.Duration(durationMs) * time.Millisecond
		runs = append(runs, r)
	}
	return runs, rows.Err()
}
