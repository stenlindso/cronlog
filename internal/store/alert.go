package store

import (
	"database/sql"
	"errors"
	"time"
)

// Alert represents a triggered alert for a run that matched a notify rule.
type Alert struct {
	ID        int64
	RunID     int64
	Command   string
	Condition string
	SentAt    time.Time
	Channel   string
}

// RecordAlert inserts an alert record into the alerts table.
func RecordAlert(db *sql.DB, runID int64, command, condition, channel string) error {
	_, err := db.Exec(
		`INSERT INTO alerts (run_id, command, condition, channel, sent_at)
		 VALUES (?, ?, ?, ?, ?)`,
		runID, command, condition, channel, time.Now().UTC(),
	)
	return err
}

// ListAlerts returns alerts, optionally filtered by command.
func ListAlerts(db *sql.DB, command string, limit int) ([]Alert, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows *sql.Rows
	var err error
	if command != "" {
		rows, err = db.Query(
			`SELECT id, run_id, command, condition, channel, sent_at
			 FROM alerts WHERE command = ?
			 ORDER BY sent_at DESC LIMIT ?`,
			command, limit,
		)
	} else {
		rows, err = db.Query(
			`SELECT id, run_id, command, condition, channel, sent_at
			 FROM alerts ORDER BY sent_at DESC LIMIT ?`,
			limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		if err := rows.Scan(&a.ID, &a.RunID, &a.Command, &a.Condition, &a.Channel, &a.SentAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// DeleteAlert removes an alert by ID. Returns error if not found.
func DeleteAlert(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM alerts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("alert not found")
	}
	return nil
}
