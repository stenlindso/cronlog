package store

import (
	"database/sql"
	"fmt"
)

// PruneOptions controls which records are deleted.
type PruneOptions struct {
	// KeepLast retains the N most recent runs per command label.
	// Zero means no limit-based pruning.
	KeepLast int
	// OlderThanDays removes runs older than N days.
	// Zero means no age-based pruning.
	OlderThanDays int
}

// Prune deletes old run records from the database according to opts.
// It returns the number of rows deleted.
func Prune(db *sql.DB, opts PruneOptions) (int64, error) {
	var total int64

	if opts.OlderThanDays > 0 {
		res, err := db.Exec(
			`DELETE FROM runs WHERE started_at < datetime('now', ?)`,
			fmt.Sprintf("-%d days", opts.OlderThanDays),
		)
		if err != nil {
			return total, fmt.Errorf("prune by age: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}

	if opts.KeepLast > 0 {
		res, err := db.Exec(
			`DELETE FROM runs
			 WHERE id NOT IN (
			   SELECT id FROM runs
			   ORDER BY started_at DESC
			   LIMIT ?
			 )`,
			opts.KeepLast,
		)
		if err != nil {
			return total, fmt.Errorf("prune by count: %w", err)
		}
		n, _ := res.RowsAffected()
		total += n
	}

	return total, nil
}
