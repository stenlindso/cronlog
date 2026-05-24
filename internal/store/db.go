package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Open opens (or creates) the SQLite database at the given path.
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create db dir: %w", err)
	}
	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			command     TEXT    NOT NULL,
			started_at  DATETIME NOT NULL,
			duration_ms INTEGER NOT NULL,
			exit_code   INTEGER NOT NULL,
			stdout      TEXT,
			stderr      TEXT
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
			created_at DATETIME NOT NULL
		);
		CREATE TABLE IF NOT EXISTS notify_rules (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			command    TEXT    NOT NULL,
			on_failure INTEGER NOT NULL DEFAULT 0,
			on_success INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL
		);
	`)
	return err
}
