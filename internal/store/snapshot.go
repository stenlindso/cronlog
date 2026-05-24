package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Snapshot represents a point-in-time summary of run statistics for a command.
type Snapshot struct {
	ID          int64
	Command     string
	CapturedAt  time.Time
	TotalRuns   int
	SuccessRate float64
	AvgDuration float64 // seconds
	P95Duration float64 // seconds
	LastExitCode int
}

// TakeSnapshot records a snapshot of current stats for the given command.
func TakeSnapshot(db *sql.DB, command string) (*Snapshot, error) {
	row := db.QueryRow(`
		SELECT
			COUNT(*) AS total,
			ROUND(100.0 * SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) / COUNT(*), 2) AS success_rate,
			ROUND(AVG(duration_ms) / 1000.0, 3) AS avg_duration,
			exit_code
		FROM runs
		WHERE command = ?
		ORDER BY started_at DESC
	`, command)

	var total int
	var successRate, avgDuration float64
	var lastExitCode int
	if err := row.Scan(&total, &successRate, &avgDuration, &lastExitCode); err != nil {
		return nil, fmt.Errorf("snapshot query: %w", err)
	}

	now := time.Now().UTC()
	res, err := db.Exec(`
		INSERT INTO snapshots (command, captured_at, total_runs, success_rate, avg_duration_s, last_exit_code)
		VALUES (?, ?, ?, ?, ?, ?)
	`, command, now.Unix(), total, successRate, avgDuration, lastExitCode)
	if err != nil {
		return nil, fmt.Errorf("insert snapshot: %w", err)
	}

	id, _ := res.LastInsertId()
	return &Snapshot{
		ID:           id,
		Command:      command,
		CapturedAt:   now,
		TotalRuns:    total,
		SuccessRate:  successRate,
		AvgDuration:  avgDuration,
		LastExitCode: lastExitCode,
	}, nil
}

// ListSnapshots returns snapshots for a command ordered most recent first.
func ListSnapshots(db *sql.DB, command string, limit int) ([]Snapshot, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(`
		SELECT id, command, captured_at, total_runs, success_rate, avg_duration_s, last_exit_code
		FROM snapshots
		WHERE command = ?
		ORDER BY captured_at DESC
		LIMIT ?
	`, command, limit)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}
	defer rows.Close()

	var snaps []Snapshot
	for rows.Next() {
		var s Snapshot
		var ts int64
		if err := rows.Scan(&s.ID, &s.Command, &ts, &s.TotalRuns, &s.SuccessRate, &s.AvgDuration, &s.LastExitCode); err != nil {
			return nil, err
		}
		s.CapturedAt = time.Unix(ts, 0).UTC()
		snaps = append(snaps, s)
	}
	return snaps, rows.Err()
}
