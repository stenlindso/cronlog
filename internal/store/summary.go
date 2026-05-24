package store

import (
	"database/sql"
	"time"
)

// DailySummary holds aggregated run statistics for a single day.
type DailySummary struct {
	Date       string
	TotalRuns  int
	Failures   int
	Successes  int
	AvgDurMs   float64
	Commands   int
}

// DailySummaryOptions controls the range and granularity of the summary query.
type DailySummaryOptions struct {
	Since   time.Time
	Command string
	Limit   int
}

// DailySummaries returns per-day aggregated stats ordered most-recent first.
func DailySummaries(db *sql.DB, opts DailySummaryOptions) ([]DailySummary, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 30
	}

	args := []any{opts.Since.UTC().Format(time.RFC3339)}
	filter := ""
	if opts.Command != "" {
		filter = " AND command = ?"
		args = append(args, opts.Command)
	}
	args = append(args, limit)

	q := `
		SELECT
			date(started_at) AS day,
			COUNT(*) AS total,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failures,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS successes,
			AVG(duration_ms) AS avg_dur,
			COUNT(DISTINCT command) AS commands
		FROM runs
		WHERE started_at >= ?` + filter + `
		GROUP BY day
		ORDER BY day DESC
		LIMIT ?
	`

	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []DailySummary
	for rows.Next() {
		var s DailySummary
		if err := rows.Scan(&s.Date, &s.TotalRuns, &s.Failures, &s.Successes, &s.AvgDurMs, &s.Commands); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
