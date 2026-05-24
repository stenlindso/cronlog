package store

import (
	"database/sql"
	"strings"
)

// SearchOptions controls how runs are searched.
type SearchOptions struct {
	Command    string // substring match on command
	ExitCode   *int   // if set, filter by exit code
	Limit      int    // max results, 0 = default 50
}

// Search returns runs matching the given options, ordered most recent first.
func Search(db *sql.DB, opts SearchOptions) ([]Run, error) {
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	var conditions []string
	var args []interface{}

	if opts.Command != "" {
		conditions = append(conditions, "command LIKE ?")
		args = append(args, "%"+opts.Command+"%")
	}
	if opts.ExitCode != nil {
		conditions = append(conditions, "exit_code = ?")
		args = append(args, *opts.ExitCode)
	}

	query := "SELECT id, command, started_at, duration_ms, exit_code, output FROM runs"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY started_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []Run
	for rows.Next() {
		var r Run
		if err := rows.Scan(&r.ID, &r.Command, &r.StartedAt, &r.DurationMs, &r.ExitCode, &r.Output); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}
