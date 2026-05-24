package store_test

import (
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func seedStats(t *testing.T) *store.DB {
	t.Helper()
	db := tempDB(t)
	now := time.Now()
	runs := []store.Run{
		{Command: "backup.sh", ExitCode: 0, DurationMs: 1200, StartedAt: now.Add(-3 * time.Hour), Output: ""},
		{Command: "backup.sh", ExitCode: 0, DurationMs: 800, StartedAt: now.Add(-2 * time.Hour), Output: ""},
		{Command: "backup.sh", ExitCode: 1, DurationMs: 500, StartedAt: now.Add(-1 * time.Hour), Output: "error"},
		{Command: "cleanup.sh", ExitCode: 0, DurationMs: 300, StartedAt: now.Add(-30 * time.Minute), Output: ""},
	}
	for _, r := range runs {
		if err := store.Insert(db.DB, r); err != nil {
			t.Fatalf("seed insert: %v", err)
		}
	}
	return db
}

func TestStatsAllCommands(t *testing.T) {
	db := seedStats(t)
	stats, err := store.Stats(db.DB, "")
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 command groups, got %d", len(stats))
	}
}

func TestStatsFilterByCommand(t *testing.T) {
	db := seedStats(t)
	stats, err := store.Stats(db.DB, "backup.sh")
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 result, got %d", len(stats))
	}
	s := stats[0]
	if s.TotalRuns != 3 {
		t.Errorf("expected 3 total runs, got %d", s.TotalRuns)
	}
	if s.SuccessCount != 2 {
		t.Errorf("expected 2 successes, got %d", s.SuccessCount)
	}
	if s.FailureCount != 1 {
		t.Errorf("expected 1 failure, got %d", s.FailureCount)
	}
	expectedAvg := (1200.0 + 800.0 + 500.0) / 3.0
	if s.AvgDurationMs < expectedAvg-1 || s.AvgDurationMs > expectedAvg+1 {
		t.Errorf("expected avg ~%.2f, got %.2f", expectedAvg, s.AvgDurationMs)
	}
}

func TestStatsEmptyDB(t *testing.T) {
	db := tempDB(t)
	stats, err := store.Stats(db.DB, "")
	if err != nil {
		t.Fatalf("Stats error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected empty stats, got %d rows", len(stats))
	}
}
