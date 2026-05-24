package store

import (
	"database/sql"
	"fmt"
)

// RetentionPolicy defines rules for automatic data retention.
type RetentionPolicy struct {
	MaxAgeDays int
	MaxCount   int
	Command    string // empty means apply to all commands
}

// UpsertRetentionPolicy inserts or replaces a retention policy for a command (or global).
func UpsertRetentionPolicy(db *sql.DB, p RetentionPolicy) error {
	if p.MaxAgeDays == 0 && p.MaxCount == 0 {
		return fmt.Errorf("retention policy must specify at least one of max_age_days or max_count")
	}
	_, err := db.Exec(`
		INSERT INTO retention_policies (command, max_age_days, max_count)
		VALUES (?, ?, ?)
		ON CONFLICT(command) DO UPDATE SET
			max_age_days = excluded.max_age_days,
			max_count    = excluded.max_count
	`, p.Command, p.MaxAgeDays, p.MaxCount)
	return err
}

// ListRetentionPolicies returns all configured retention policies.
func ListRetentionPolicies(db *sql.DB) ([]RetentionPolicy, error) {
	rows, err := db.Query(`SELECT command, max_age_days, max_count FROM retention_policies ORDER BY command`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []RetentionPolicy
	for rows.Next() {
		var p RetentionPolicy
		if err := rows.Scan(&p.Command, &p.MaxAgeDays, &p.MaxCount); err != nil {
			return nil, err
		}
		policies = append(policies, p)
	}
	return policies, rows.Err()
}

// DeleteRetentionPolicy removes the policy for the given command (or global if empty).
func DeleteRetentionPolicy(db *sql.DB, command string) error {
	res, err := db.Exec(`DELETE FROM retention_policies WHERE command = ?`, command)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("no retention policy found for command %q", command)
	}
	return nil
}

// ApplyRetentionPolicies runs all configured policies using Prune.
func ApplyRetentionPolicies(db *sql.DB) (int64, error) {
	policies, err := ListRetentionPolicies(db)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, p := range policies {
		n, err := Prune(db, PruneOptions{
			OlderThanDays: p.MaxAgeDays,
			KeepLast:      p.MaxCount,
			Command:       p.Command,
		})
		if err != nil {
			return total, fmt.Errorf("applying policy for %q: %w", p.Command, err)
		}
		total += n
	}
	return total, nil
}
