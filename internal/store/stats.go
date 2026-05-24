package store

import (
	"database/sql"
	"fmt"
)

// CommandStats holds aggregate statistics for a specific command.
type CommandStats struct {
	Command      string
	TotalRuns    int
	SuccessCount int
	FailureCount int
	AvgDurationMs float64
	LastRunAt    string
}

// Stats returns aggregate statistics grouped by command.
// If command is non-empty, results are filtered to that command.
func Stats(db *sql.DB, command string) ([]CommandStats, error) {
	query := `
		SELECT
			command,
			COUNT(*) AS total_runs,
			SUM(CASE WHEN exit_code = 0 THEN 1 ELSE 0 END) AS success_count,
			SUM(CASE WHEN exit_code != 0 THEN 1 ELSE 0 END) AS failure_count,
			AVG(duration_ms) AS avg_duration_ms,
			MAX(started_at) AS last_run_at
		FROM runs
	`

	args := []interface{}{}
	if command != "" {
		query += " WHERE command = ?"
		args = append(args, command)
	}
	query += " GROUP BY command ORDER BY last_run_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("stats query: %w", err)
	}
	defer rows.Close()

	var results []CommandStats
	for rows.Next() {
		var s CommandStats
		if err := rows.Scan(
			&s.Command,
			&s.TotalRuns,
			&s.SuccessCount,
			&s.FailureCount,
			&s.AvgDurationMs,
			&s.LastRunAt,
		); err != nil {
			return nil, fmt.Errorf("stats scan: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}
