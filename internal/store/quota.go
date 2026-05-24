package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Quota defines a run-rate limit for a given command.
type Quota struct {
	ID        int64
	Command   string
	MaxRuns   int
	WindowSec int // window duration in seconds
	CreatedAt time.Time
}

// QuotaViolation is returned when a command exceeds its quota.
type QuotaViolation struct {
	Command string
	MaxRuns int
	Window  time.Duration
	Actual  int
}

func (v *QuotaViolation) Error() string {
	return fmt.Sprintf("quota exceeded for %q: %d runs in %s (limit %d)",
		v.Command, v.Actual, v.Window, v.MaxRuns)
}

// UpsertQuota inserts or replaces a quota rule for a command.
func UpsertQuota(db *sql.DB, command string, maxRuns, windowSec int) error {
	if command == "" {
		return errors.New("command is required")
	}
	if maxRuns <= 0 {
		return errors.New("max_runs must be positive")
	}
	if windowSec <= 0 {
		return errors.New("window_sec must be positive")
	}
	_, err := db.Exec(`
		INSERT INTO quota_rules (command, max_runs, window_sec)
		VALUES (?, ?, ?)
		ON CONFLICT(command) DO UPDATE SET max_runs=excluded.max_runs, window_sec=excluded.window_sec`,
		command, maxRuns, windowSec)
	return err
}

// ListQuotas returns all quota rules.
func ListQuotas(db *sql.DB) ([]Quota, error) {
	rows, err := db.Query(`SELECT id, command, max_runs, window_sec, created_at FROM quota_rules ORDER BY command`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Quota
	for rows.Next() {
		var q Quota
		if err := rows.Scan(&q.ID, &q.Command, &q.MaxRuns, &q.WindowSec, &q.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, q)
	}
	return out, rows.Err()
}

// DeleteQuota removes a quota rule by command.
func DeleteQuota(db *sql.DB, command string) error {
	res, err := db.Exec(`DELETE FROM quota_rules WHERE command = ?`, command)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("no quota rule found for %q", command)
	}
	return nil
}

// CheckQuota returns a QuotaViolation if the command has exceeded its quota
// within the configured window, or nil if within limits.
func CheckQuota(db *sql.DB, command string) (*QuotaViolation, error) {
	var maxRuns, windowSec int
	err := db.QueryRow(`SELECT max_runs, window_sec FROM quota_rules WHERE command = ?`, command).
		Scan(&maxRuns, &windowSec)
	if err == sql.ErrNoRows {
		return nil, nil // no quota configured
	}
	if err != nil {
		return nil, err
	}
	since := time.Now().UTC().Add(-time.Duration(windowSec) * time.Second)
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM runs WHERE command = ? AND started_at >= ?`, command, since).Scan(&count); err != nil {
		return nil, err
	}
	if count >= maxRuns {
		return &QuotaViolation{Command: command, MaxRuns: maxRuns, Window: time.Duration(windowSec) * time.Second, Actual: count}, nil
	}
	return nil, nil
}
