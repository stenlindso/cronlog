package store_test

import (
	"testing"
	"time"
)

func seedBaselineRuns(t *testing.T, db interface{ Exec(string, ...any) (interface{}, error) }) {}

func TestUpsertAndGetBaseline(t *testing.T) {
	db := tempDB(t)

	// Insert successful runs for "backup.sh"
	for i, ms := range []int{2000, 4000, 6000} {
		_, err := db.Exec(`
			INSERT INTO runs (command, started_at, duration_ms, exit_code, output)
			VALUES (?, ?, ?, 0, '')
		`, "backup.sh", time.Now().Add(time.Duration(-i)*time.Minute).UTC().Format(time.RFC3339), ms)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	baseline, err := UpsertBaseline(db, "backup.sh", 10)
	if err != nil {
		t.Fatalf("UpsertBaseline: %v", err)
	}
	if baseline.SampleCount != 3 {
		t.Errorf("expected sample count 3, got %d", baseline.SampleCount)
	}
	// avg of 2000,4000,6000 ms = 4000 ms = 4.0 s
	if baseline.AvgDuration != 4.0 {
		t.Errorf("expected avg 4.0s, got %f", baseline.AvgDuration)
	}

	got, err := GetBaseline(db, "backup.sh")
	if err != nil {
		t.Fatalf("GetBaseline: %v", err)
	}
	if got == nil {
		t.Fatal("expected baseline, got nil")
	}
	if got.AvgDuration != 4.0 {
		t.Errorf("expected 4.0s, got %f", got.AvgDuration)
	}
}

func TestGetBaselineMissing(t *testing.T) {
	db := tempDB(t)
	got, err := GetBaseline(db, "nonexistent.sh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil baseline for unknown command")
	}
}

func TestUpsertBaselineNoRuns(t *testing.T) {
	db := tempDB(t)
	_, err := UpsertBaseline(db, "ghost.sh", 5)
	if err == nil {
		t.Error("expected error for command with no runs")
	}
}

func TestUpsertBaselineInvalidSampleSize(t *testing.T) {
	db := tempDB(t)
	_, err := UpsertBaseline(db, "backup.sh", 0)
	if err == nil {
		t.Error("expected error for sampleSize=0")
	}
}

func TestUpsertBaselineIgnoresFailedRuns(t *testing.T) {
	db := tempDB(t)
	// Insert a failed run — should be excluded
	_, err := db.Exec(`
		INSERT INTO runs (command, started_at, duration_ms, exit_code, output)
		VALUES ('check.sh', ?, 9000, 1, '')
	`, time.Now().UTC().Format(time.RFC3339))
	if err != nil {
		t.Fatalf("seed failed run: %v", err)
	}
	_, err = UpsertBaseline(db, "check.sh", 5)
	if err == nil {
		t.Error("expected error since only failed runs exist")
	}
}
