package store

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Open opens (or creates) the SQLite database at the given path and runs migrations.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS runs (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		command    TEXT    NOT NULL,
		started_at DATETIME NOT NULL,
		duration   INTEGER NOT NULL,
		exit_code  INTEGER NOT NULL,
		stdout     TEXT    NOT NULL DEFAULT '',
		stderr     TEXT    NOT NULL DEFAULT ''
	);
	CREATE TABLE IF NOT EXISTS tags (
		run_id INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
		tag    TEXT    NOT NULL,
		PRIMARY KEY (run_id, tag)
	);
	CREATE TABLE IF NOT EXISTS annotations (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id     INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
		note       TEXT    NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS notify_rules (
		id        INTEGER PRIMARY KEY AUTOINCREMENT,
		command   TEXT NOT NULL DEFAULT '',
		condition TEXT NOT NULL,
		target    TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS retention_policies (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		command    TEXT NOT NULL DEFAULT '',
		keep_days  INTEGER,
		keep_count INTEGER,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		UNIQUE(command)
	);
	CREATE TABLE IF NOT EXISTS alerts (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		run_id     INTEGER NOT NULL REFERENCES runs(id) ON DELETE CASCADE,
		command    TEXT NOT NULL,
		reason     TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS baselines (
		command      TEXT PRIMARY KEY,
		avg_duration REAL NOT NULL,
		stddev       REAL NOT NULL,
		sample_size  INTEGER NOT NULL,
		updated_at   DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	CREATE TABLE IF NOT EXISTS quota_rules (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		command    TEXT NOT NULL UNIQUE,
		max_runs   INTEGER NOT NULL,
		window_sec INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now'))
	);
	`)
	if err != nil {
		return fmt.Errorf("create tables: %w", err)
	}
	return nil
}
