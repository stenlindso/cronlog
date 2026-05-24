package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Baseline holds the expected duration baseline for a command.
type Baseline struct {
	Command     string
	AvgDuration float64 // seconds
	SampleCount int
	UpdatedAt   time.Time
}

// UpsertBaseline computes and stores the average duration baseline for a command
// using the most recent `sampleSize` successful runs.
func UpsertBaseline(db *sql.DB, command string, sampleSize int) (*Baseline, error) {
	if sampleSize <= 0 {
		return nil, fmt.Errorf("sampleSize must be > 0")
	}

	row := db.QueryRow(`
		SELECT AVG(duration_ms), COUNT(*) FROM (
			SELECT duration_ms FROM runs
			WHERE command = ? AND exit_code = 0
			ORDER BY started_at DESC
			LIMIT ?
		)
	`, command, sampleSize)

	var avg sql.NullFloat64
	var count int
	if err := row.Scan(&avg, &count); err != nil {
		return nil, fmt.Errorf("baseline query: %w", err)
	}
	if count == 0 {
		return nil, fmt.Errorf("no successful runs found for command %q", command)
	}

	now := time.Now().UTC()
	_, err := db.Exec(`
		INSERT INTO baselines (command, avg_duration_ms, sample_count, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(command) DO UPDATE SET
			avg_duration_ms = excluded.avg_duration_ms,
			sample_count    = excluded.sample_count,
			updated_at      = excluded.updated_at
	`, command, avg.Float64, count, now)
	if err != nil {
		return nil, fmt.Errorf("baseline upsert: %w", err)
	}

	return &Baseline{
		Command:     command,
		AvgDuration: avg.Float64 / 1000.0,
		SampleCount: count,
		UpdatedAt:   now,
	}, nil
}

// GetBaseline retrieves the stored baseline for a command.
func GetBaseline(db *sql.DB, command string) (*Baseline, error) {
	row := db.QueryRow(`
		SELECT command, avg_duration_ms, sample_count, updated_at
		FROM baselines WHERE command = ?
	`, command)

	var b Baseline
	var avgMs float64
	var updatedAt string
	if err := row.Scan(&b.Command, &avgMs, &b.SampleCount, &updatedAt); err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("get baseline: %w", err)
	}
	b.AvgDuration = avgMs / 1000.0
	b.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &b, nil
}
