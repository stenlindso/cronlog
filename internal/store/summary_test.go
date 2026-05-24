package store_test

import (
	"testing"
	"time"
)

func seedSummaryRuns(t *testing.T) *tempDB {
	t.Helper()
	td := mustTempDB(t)
	now := time.Now().UTC()

	runs := []struct {
		cmd      string
		exitCode int
		durMs    int64
		daysAgo  int
	}{
		{"backup.sh", 0, 1200, 0},
		{"backup.sh", 1, 800, 0},
		{"sync.sh", 0, 500, 0},
		{"backup.sh", 0, 900, 1},
		{"sync.sh", 0, 600, 2},
	}

	for _, r := range runs {
		started := now.AddDate(0, 0, -r.daysAgo)
		_, err := Insert(td.db, r.cmd, r.exitCode, "", "", r.durMs, started)
		if err != nil {
			t.Fatalf("seed insert: %v", err)
		}
	}
	return td
}

func TestDailySummariesAllDays(t *testing.T) {
	td := seedSummaryRuns(t)
	defer td.cleanup()

	summaries, err := DailySummaries(td.db, DailySummaryOptions{
		Since: time.Now().UTC().AddDate(0, 0, -7),
	})
	if err != nil {
		t.Fatalf("DailySummaries: %v", err)
	}
	if len(summaries) != 3 {
		t.Fatalf("expected 3 days, got %d", len(summaries))
	}
	// most recent day first
	if summaries[0].TotalRuns != 3 {
		t.Errorf("day 0 total: want 3, got %d", summaries[0].TotalRuns)
	}
	if summaries[0].Failures != 1 {
		t.Errorf("day 0 failures: want 1, got %d", summaries[0].Failures)
	}
	if summaries[0].Successes != 2 {
		t.Errorf("day 0 successes: want 2, got %d", summaries[0].Successes)
	}
}

func TestDailySummariesFilterByCommand(t *testing.T) {
	td := seedSummaryRuns(t)
	defer td.cleanup()

	summaries, err := DailySummaries(td.db, DailySummaryOptions{
		Since:   time.Now().UTC().AddDate(0, 0, -7),
		Command: "backup.sh",
	})
	if err != nil {
		t.Fatalf("DailySummaries: %v", err)
	}
	for _, s := range summaries {
		if s.Commands != 1 {
			t.Errorf("expected 1 distinct command per day, got %d", s.Commands)
		}
	}
}

func TestDailySummariesEmptyDB(t *testing.T) {
	td := mustTempDB(t)
	defer td.cleanup()

	summaries, err := DailySummaries(td.db, DailySummaryOptions{
		Since: time.Now().UTC().AddDate(0, 0, -7),
	})
	if err != nil {
		t.Fatalf("DailySummaries: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("expected empty result, got %d rows", len(summaries))
	}
}

func TestDailySummariesLimit(t *testing.T) {
	td := seedSummaryRuns(t)
	defer td.cleanup()

	summaries, err := DailySummaries(td.db, DailySummaryOptions{
		Since: time.Now().UTC().AddDate(0, 0, -30),
		Limit: 2,
	})
	if err != nil {
		t.Fatalf("DailySummaries: %v", err)
	}
	if len(summaries) > 2 {
		t.Errorf("expected at most 2 rows, got %d", len(summaries))
	}
}
