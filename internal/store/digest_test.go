package store_test

import (
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func seedDigestRuns(t *testing.T, db interface{ Exec(string, ...any) (interface{}, error) }) {
	t.Helper()
}

func TestGetDigestAllCommands(t *testing.T) {
	db := tempDB(t)
	now := time.Now().UTC()

	insertRun := func(cmd string, exitCode int, durMS int64, at time.Time) {
		t.Helper()
		_, err := db.Exec(
			`INSERT INTO runs (command, output, exit_code, duration_ms, started_at) VALUES (?, '', ?, ?, ?)`,
			cmd, exitCode, durMS, at.Format(time.RFC3339),
		)
		if err != nil {
			t.Fatalf("seed insert: %v", err)
		}
	}

	insertRun("backup.sh", 0, 1000, now.Add(-1*time.Hour))
	insertRun("backup.sh", 0, 2000, now.Add(-2*time.Hour))
	insertRun("backup.sh", 1, 500, now.Add(-3*time.Hour))
	insertRun("sync.sh", 0, 300, now.Add(-30*time.Minute))

	opts := store.DigestOptions{
		Since: now.Add(-6 * time.Hour),
		Until: now.Add(time.Minute),
	}
	entries, err := store.GetDigest(db, opts)
	if err != nil {
		t.Fatalf("GetDigest: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// backup.sh has a failure, should be first
	if entries[0].Command != "backup.sh" {
		t.Errorf("expected backup.sh first, got %s", entries[0].Command)
	}
	if entries[0].TotalRuns != 3 {
		t.Errorf("expected 3 total runs for backup.sh, got %d", entries[0].TotalRuns)
	}
	if entries[0].Failures != 1 {
		t.Errorf("expected 1 failure for backup.sh, got %d", entries[0].Failures)
	}
	if entries[0].Successes != 2 {
		t.Errorf("expected 2 successes for backup.sh, got %d", entries[0].Successes)
	}
}

func TestGetDigestFilterByCommand(t *testing.T) {
	db := tempDB(t)
	now := time.Now().UTC()

	for _, cmd := range []string{"job.sh", "other.sh"} {
		_, err := db.Exec(
			`INSERT INTO runs (command, output, exit_code, duration_ms, started_at) VALUES (?, '', 0, 100, ?)`,
			cmd, now.Add(-1*time.Hour).Format(time.RFC3339),
		)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	opts := store.DigestOptions{
		Command: "job.sh",
		Since:   now.Add(-6 * time.Hour),
		Until:   now.Add(time.Minute),
	}
	entries, err := store.GetDigest(db, opts)
	if err != nil {
		t.Fatalf("GetDigest: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Command != "job.sh" {
		t.Errorf("expected job.sh, got %s", entries[0].Command)
	}
}

func TestGetDigestEmptyWindow(t *testing.T) {
	db := tempDB(t)
	now := time.Now().UTC()

	_, err := db.Exec(
		`INSERT INTO runs (command, output, exit_code, duration_ms, started_at) VALUES ('old.sh', '', 0, 50, ?)`,
		now.Add(-48*time.Hour).Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}

	opts := store.DigestOptions{
		Since: now.Add(-1 * time.Hour),
		Until: now,
	}
	entries, err := store.GetDigest(db, opts)
	if err != nil {
		t.Fatalf("GetDigest: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}
