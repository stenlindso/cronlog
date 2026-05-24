package store

import (
	"database/sql"
	"fmt"
	"time"
)

// DigestEntry represents a periodic digest of run outcomes for a command.
type DigestEntry struct {
	Command    string
	PeriodStart time.Time
	PeriodEnd   time.Time
	TotalRuns  int
	Failures   int
	Successes  int
	AvgDurMS   int64
	MaxDurMS   int64
}

// DigestOptions controls the range and grouping for a digest query.
type DigestOptions struct {
	Command string
	Since   time.Time
	Until   time.Time
}

// GetDigest returns aggregated run statistics grouped by command within the
// specified time window. If Command is set, results are filtered to that command.
func GetDigest(db *sql.DB, opts DigestOptions) ([]DigestEntry, error) {
	if opts.Until.IsZero() {
		opts.Until = time.Now()
	}
	if opts.Since.IsZero() {
		opts.Since = opts.Until.Add(-24 * time.Hour)
	}

	query := `
		SELECT
			command,
			COUNT(*) AS total_runs,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failures,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS successes,
			AVG(duration_ms) AS avg_dur_ms,
			MAX(duration_ms) AS max_dur_ms
		FROM runs
		WHERE started_at >= ? AND started_at <= ?
`
	args := []any{opts.Since.UTC().Format(time.RFC3339), opts.Until.UTC().Format(time.RFC3339)}

	if opts.Command != "" {
		query += " AND command = ?"
		args = append(args, opts.Command)
	}
	query += " GROUP BY command ORDER BY failures DESC, total_runs DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("digest query: %w", err)
	}
	defer rows.Close()

	var entries []DigestEntry
	for rows.Next() {
		var e DigestEntry
		if err := rows.Scan(&e.Command, &e.TotalRuns, &e.Failures, &e.Successes, &e.AvgDurMS, &e.MaxDurMS); err != nil {
			return nil, fmt.Errorf("digest scan: %w", err)
		}
		e.PeriodStart = opts.Since
		e.PeriodEnd = opts.Until
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
