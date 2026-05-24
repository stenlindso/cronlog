package store

import (
	"database/sql"
	"time"
)

// NotifyRule defines a condition that triggers a notification record.
type NotifyRule struct {
	ID        int64
	Command   string
	OnFailure bool
	OnSuccess bool
	CreatedAt time.Time
}

// AddNotifyRule inserts a notification rule for the given command.
func AddNotifyRule(db *sql.DB, command string, onFailure, onSuccess bool) (int64, error) {
	res, err := db.Exec(
		`INSERT INTO notify_rules (command, on_failure, on_success, created_at)
		 VALUES (?, ?, ?, ?)`,
		command, onFailure, onSuccess, time.Now().UTC(),
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListNotifyRules returns all notification rules, optionally filtered by command.
func ListNotifyRules(db *sql.DB, command string) ([]NotifyRule, error) {
	query := `SELECT id, command, on_failure, on_success, created_at FROM notify_rules`
	args := []interface{}{}
	if command != "" {
		query += ` WHERE command = ?`
		args = append(args, command)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []NotifyRule
	for rows.Next() {
		var r NotifyRule
		if err := rows.Scan(&r.ID, &r.Command, &r.OnFailure, &r.OnSuccess, &r.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

// DeleteNotifyRule removes a notification rule by ID.
func DeleteNotifyRule(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM notify_rules WHERE id = ?`, id)
	return err
}

// MatchingRules returns rules that match the given command and exit code.
func MatchingRules(db *sql.DB, command string, exitCode int) ([]NotifyRule, error) {
	rows, err := db.Query(
		`SELECT id, command, on_failure, on_success, created_at
		 FROM notify_rules
		 WHERE command = ? AND (on_failure = 1 OR on_success = 1)`,
		command,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matched []NotifyRule
	for rows.Next() {
		var r NotifyRule
		if err := rows.Scan(&r.ID, &r.Command, &r.OnFailure, &r.OnSuccess, &r.CreatedAt); err != nil {
			return nil, err
		}
		if (exitCode != 0 && r.OnFailure) || (exitCode == 0 && r.OnSuccess) {
			matched = append(matched, r)
		}
	}
	return matched, rows.Err()
}
