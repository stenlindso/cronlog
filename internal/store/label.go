package store

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Label represents a key-value metadata pair attached to a run.
type Label struct {
	RunID int64  `db:"run_id"`
	Key   string `db:"key"`
	Value string `db:"value"`
}

// SetLabels upserts one or more key=value labels on a run.
// Each entry in pairs must be in "key=value" format.
func SetLabels(db *sqlx.DB, runID int64, pairs []string) error {
	if len(pairs) == 0 {
		return nil
	}

	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	for _, p := range pairs {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return fmt.Errorf("invalid label %q: must be key=value", p)
		}
		_, err := tx.Exec(
			`INSERT INTO labels (run_id, key, value) VALUES (?, ?, ?)
			 ON CONFLICT(run_id, key) DO UPDATE SET value = excluded.value`,
			runID, parts[0], parts[1],
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetLabels returns all labels for a given run.
func GetLabels(db *sqlx.DB, runID int64) ([]Label, error) {
	var labels []Label
	err := db.Select(&labels,
		`SELECT run_id, key, value FROM labels WHERE run_id = ? ORDER BY key`,
		runID,
	)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

// DeleteLabel removes a single label key from a run.
func DeleteLabel(db *sqlx.DB, runID int64, key string) error {
	res, err := db.Exec(`DELETE FROM labels WHERE run_id = ? AND key = ?`, runID, key)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
