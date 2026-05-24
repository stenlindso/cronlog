package store_test

import (
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func seedReplayRuns(t *testing.T, db interface{ Helper() }, path string) {
	t.Helper()
	d, err := store.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	base := time.Now().Add(-10 * time.Minute)
	runs := []store.Run{
		{Command: "backup.sh", Args: "", ExitCode: 0, Stdout: "ok", Stderr: "", StartedAt: base, DurationMs: 100},
		{Command: "sync.sh", Args: "", ExitCode: 1, Stdout: "", Stderr: "fail", StartedAt: base.Add(2 * time.Minute), DurationMs: 50},
		{Command: "backup.sh", Args: "--full", ExitCode: 0, Stdout: "done", Stderr: "", StartedAt: base.Add(5 * time.Minute), DurationMs: 200},
	}
	for _, r := range runs {
		if _, err := store.Insert(d, r); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	d.Close()
}

func TestGetLastRunNoFilter(t *testing.T) {
	path := tempDB(t)
	seedReplayRuns(t, t, path)
	db, _ := store.Open(path)
	defer db.Close()

	run, err := store.GetLastRun(db, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.Command != "backup.sh" || run.Args != "--full" {
		t.Errorf("expected last backup.sh --full, got %s %s", run.Command, run.Args)
	}
}

func TestGetLastRunFilterByCommand(t *testing.T) {
	path := tempDB(t)
	seedReplayRuns(t, t, path)
	db, _ := store.Open(path)
	defer db.Close()

	run, err := store.GetLastRun(db, "sync.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if run.Command != "sync.sh" {
		t.Errorf("expected sync.sh, got %s", run.Command)
	}
}

func TestGetLastRunMissing(t *testing.T) {
	path := tempDB(t)
	db, _ := store.Open(path)
	defer db.Close()

	_, err := store.GetLastRun(db, "")
	if err == nil {
		t.Error("expected error for empty db")
	}
}

func TestGetReplayCommands(t *testing.T) {
	path := tempDB(t)
	seedReplayRuns(t, t, path)
	db, _ := store.Open(path)
	defer db.Close()

	cmds, err := store.GetReplayCommands(db, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cmds) != 2 {
		t.Errorf("expected 2 distinct commands, got %d", len(cmds))
	}
}
