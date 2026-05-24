package store

import (
	"database/sql"
	"fmt"
)

// ReplayOptions controls which runs are replayed.
type ReplayOptions struct {
	Command string
	Limit   int
	DryRun  bool
}

// ReplayResult holds the command string retrieved for re-execution.
type ReplayResult struct {
	ID      int64
	Command string
	Args    []string
}

// GetLastRun returns the most recent run matching the given command prefix.
// If command is empty, it returns the most recent run overall.
func GetLastRun(db *sql.DB, command string) (*Run, error) {
	query := `
		SELECT id, command, args, exit_code, stdout, stderr, started_at, duration_ms
		FROM runs
	`
	args := []interface{}{}
	if command != "" {
		query += " WHERE command = ?"
		args = append(args, command)
	}
	query += " ORDER BY started_at DESC LIMIT 1"

	row := db.QueryRow(query, args...)
	var r Run
	err := row.Scan(&r.ID, &r.Command, &r.Args, &r.ExitCode, &r.Stdout, &r.Stderr, &r.StartedAt, &r.DurationMs)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no runs found")
	}
	if err != nil {
		return nil, fmt.Errorf("query last run: %w", err)
	}
	return &r, nil
}

// GetReplayCommands returns up to limit recent distinct commands for replay.
func GetReplayCommands(db *sql.DB, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := db.Query(`
		SELECT DISTINCT command
		FROM runs
		ORDER BY started_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query replay commands: %w", err)
	}
	defer rows.Close()

	var commands []string
	for rows.Next() {
		var cmd string
		if err := rows.Scan(&cmd); err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	return commands, rows.Err()
}
