package store_test

import (
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func seedSearch(t *testing.T) *store.DB {
	t.Helper()
	db := tempDB(t)
	now := time.Now()
	runs := []store.Run{
		{Command: "backup.sh", StartedAt: now.Add(-3 * time.Hour), DurationMs: 500, ExitCode: 0, Output: "done"},
		{Command: "backup.sh", StartedAt: now.Add(-2 * time.Hour), DurationMs: 600, ExitCode: 1, Output: "error"},
		{Command: "cleanup.sh", StartedAt: now.Add(-1 * time.Hour), DurationMs: 200, ExitCode: 0, Output: "ok"},
		{Command: "report.sh", StartedAt: now, DurationMs: 900, ExitCode: 0, Output: "sent"},
	}
	for _, r := range runs {
		if err := store.Insert(db.DB, r); err != nil {
			t.Fatalf("seed insert: %v", err)
		}
	}
	return db
}

func TestSearchByCommand(t *testing.T) {
	db := seedSearch(t)
	runs, err := store.Search(db.DB, store.SearchOptions{Command: "backup"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
}

func TestSearchByExitCode(t *testing.T) {
	db := seedSearch(t)
	exitOne := 1
	runs, err := store.Search(db.DB, store.SearchOptions{ExitCode: &exitOne})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].Command != "backup.sh" {
		t.Errorf("expected backup.sh, got %s", runs[0].Command)
	}
}

func TestSearchCombined(t *testing.T) {
	db := seedSearch(t)
	exitZero := 0
	runs, err := store.Search(db.DB, store.SearchOptions{Command: "backup", ExitCode: &exitZero})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
}

func TestSearchNoFiltersReturnsAll(t *testing.T) {
	db := seedSearch(t)
	runs, err := store.Search(db.DB, store.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(runs) != 4 {
		t.Fatalf("expected 4 runs, got %d", len(runs))
	}
}

func TestSearchOrderedMostRecentFirst(t *testing.T) {
	db := seedSearch(t)
	runs, err := store.Search(db.DB, store.SearchOptions{Limit: 10})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	for i := 1; i < len(runs); i++ {
		if runs[i].StartedAt.After(runs[i-1].StartedAt) {
			t.Errorf("results not ordered most recent first at index %d", i)
		}
	}
}
