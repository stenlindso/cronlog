package store_test

import (
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func seedRuns(t *testing.T, db interface{ Exec(string, ...any) (interface{ RowsAffected() (int64, error) }, error) }) {
	t.Helper()
}

func TestPruneByAge(t *testing.T) {
	db := tempDB(t)

	now := time.Now()
	old := now.Add(-10 * 24 * time.Hour)

	_ = store.Insert(db, store.Run{Command: "old-cmd", StartedAt: old, DurationMs: 100, ExitCode: 0, Output: ""})
	_ = store.Insert(db, store.Run{Command: "new-cmd", StartedAt: now, DurationMs: 200, ExitCode: 0, Output: ""})

	deleted, err := store.Prune(db, store.PruneOptions{OlderThanDays: 5})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if deleted != 1 {
		t.Errorf("expected 1 deleted, got %d", deleted)
	}

	runs, err := store.List(db, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 1 || runs[0].Command != "new-cmd" {
		t.Errorf("expected only new-cmd to remain, got %+v", runs)
	}
}

func TestPruneByCount(t *testing.T) {
	db := tempDB(t)

	base := time.Now()
	for i := 0; i < 5; i++ {
		_ = store.Insert(db, store.Run{
			Command:    "cmd",
			StartedAt:  base.Add(time.Duration(i) * time.Minute),
			DurationMs: int64(i * 10),
			ExitCode:   0,
			Output:     "",
		})
	}

	deleted, err := store.Prune(db, store.PruneOptions{KeepLast: 3})
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if deleted != 2 {
		t.Errorf("expected 2 deleted, got %d", deleted)
	}

	runs, err := store.List(db, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs remaining, got %d", len(runs))
	}
}

func TestPruneNoOptions(t *testing.T) {
	db := tempDB(t)
	_ = store.Insert(db, store.Run{Command: "x", StartedAt: time.Now(), DurationMs: 1, ExitCode: 0, Output: ""})

	deleted, err := store.Prune(db, store.PruneOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}
