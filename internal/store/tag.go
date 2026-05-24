package store

import (
	"database/sql"
	"fmt"
	"strings"
)

// Tag represents a label attached to a run.
type Tag struct {
	RunID int64
	Name  string
}

// AddTags associates one or more tag names with a run by ID.
func AddTags(db *sql.DB, runID int64, tags []string) error {
	if len(tags) == 0 {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("addtags begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO tags (run_id, name) VALUES (?, ?)`)
	if err != nil {
		return fmt.Errorf("addtags prepare: %w", err)
	}
	defer stmt.Close()

	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t == "" {
			continue
		}
		if _, err := stmt.Exec(runID, t); err != nil {
			return fmt.Errorf("addtags exec tag %q: %w", t, err)
		}
	}
	return tx.Commit()
}

// ListTags returns all tag names for the given run ID.
func ListTags(db *sql.DB, runID int64) ([]string, error) {
	rows, err := db.Query(`SELECT name FROM tags WHERE run_id = ? ORDER BY name`, runID)
	if err != nil {
		return nil, fmt.Errorf("listtags query: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("listtags scan: %w", err)
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

// ListRunsByTag returns runs that have the given tag, most recent first.
func ListRunsByTag(db *sql.DB, tag string, limit int) ([]Run, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := db.Query(`
		SELECT r.id, r.command, r.started_at, r.duration_ms, r.exit_code, r.output
		FROM runs r
		JOIN tags t ON t.run_id = r.id
		WHERE t.name = ?
		ORDER BY r.started_at DESC
		LIMIT ?`, tag, limit)
	if err != nil {
		return nil, fmt.Errorf("listrunsbytag query: %w", err)
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Command, &r.StartedAt, &r.DurationMs, &r.ExitCode, &r.Output); err != nil {
			return nil, fmt.Errorf("listrunsbytag scan: %w", err)
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}
