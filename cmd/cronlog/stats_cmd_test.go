package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func TestRunStatsNoArgs(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	now := time.Now()
	for i := 0; i < 3; i++ {
		_ = store.Insert(db, store.Run{
			Command:   "backup.sh",
			StartedAt: now,
			Duration:  2 * time.Second,
			ExitCode:  0,
			Output:    "ok",
		})
	}
	_ = store.Insert(db, store.Run{
		Command:   "backup.sh",
		StartedAt: now,
		Duration:  time.Second,
		ExitCode:  1,
		Output:    "fail",
	})

	var buf bytes.Buffer
	err = runStats(db, []string{}, &buf)
	if err != nil {
		t.Fatalf("runStats: %v", err)
	}

	out := buf.String()
	if out == "" {
		t.Error("expected non-empty output")
	}
	for _, want := range []string{"COMMAND", "RUNS", "backup.sh"} {
		if !bytes.Contains([]byte(out), []byte(want)) {
			t.Errorf("expected output to contain %q, got:\n%s", want, out)
		}
	}
}

func TestRunStatsFilterByCommand(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	now := time.Now()
	_ = store.Insert(db, store.Run{Command: "job-a", StartedAt: now, Duration: time.Second, ExitCode: 0})
	_ = store.Insert(db, store.Run{Command: "job-b", StartedAt: now, Duration: time.Second, ExitCode: 0})

	var buf bytes.Buffer
	err = runStats(db, []string{"--command", "job-a"}, &buf)
	if err != nil {
		t.Fatalf("runStats: %v", err)
	}

	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("job-a")) {
		t.Errorf("expected job-a in output, got:\n%s", out)
	}
	if bytes.Contains([]byte(out), []byte("job-b")) {
		t.Errorf("did not expect job-b in output, got:\n%s", out)
	}
}

func TestRunStatsEmptyDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var buf bytes.Buffer
	err = runStats(db, []string{}, &buf)
	if err != nil {
		t.Fatalf("runStats: %v", err)
	}

	_ = os.Stdout // suppress unused import
	out := buf.String()
	if !bytes.Contains([]byte(out), []byte("no stats")) {
		t.Errorf("expected 'no stats' message, got:\n%s", out)
	}
}
