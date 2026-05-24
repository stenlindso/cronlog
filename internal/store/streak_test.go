package store

import (
	"testing"
	"time"
)

func seedStreakRuns(t *testing.T, db interface{ Exec(string, ...interface{}) (interface{}, error) }) {}

func TestGetStreakSuccess(t *testing.T) {
	db := tempDB(t)
	now := time.Now()

	for i := 0; i < 3; i++ {
		_, err := db.Exec(`INSERT INTO runs (command, output, exit_code, started_at, duration_ms) VALUES (?, ?, ?, ?, ?)`,
			"backup.sh", "", 0, now.Add(-time.Duration(i)*time.Minute), 100)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	s, err := GetStreak(db, "backup.sh")
	if err != nil {
		t.Fatalf("GetStreak: %v", err)
	}
	if s.StreakType != "success" {
		t.Errorf("expected success streak, got %q", s.StreakType)
	}
	if s.CurrentStreak != 3 {
		t.Errorf("expected current streak 3, got %d", s.CurrentStreak)
	}
	if s.LongestSuccess != 3 {
		t.Errorf("expected longest success 3, got %d", s.LongestSuccess)
	}
}

func TestGetStreakFailureBreaksSuccess(t *testing.T) {
	db := tempDB(t)
	now := time.Now()

	// Most recent: 2 failures, then 4 successes
	for i := 0; i < 2; i++ {
		_, err := db.Exec(`INSERT INTO runs (command, output, exit_code, started_at, duration_ms) VALUES (?, ?, ?, ?, ?)`,
			"sync.sh", "", 1, now.Add(-time.Duration(i)*time.Minute), 50)
		if err != nil {
			t.Fatalf("seed failure: %v", err)
		}
	}
	for i := 2; i < 6; i++ {
		_, err := db.Exec(`INSERT INTO runs (command, output, exit_code, started_at, duration_ms) VALUES (?, ?, ?, ?, ?)`,
			"sync.sh", "", 0, now.Add(-time.Duration(i)*time.Minute), 50)
		if err != nil {
			t.Fatalf("seed success: %v", err)
		}
	}

	s, err := GetStreak(db, "sync.sh")
	if err != nil {
		t.Fatalf("GetStreak: %v", err)
	}
	if s.StreakType != "failure" {
		t.Errorf("expected failure streak, got %q", s.StreakType)
	}
	if s.CurrentStreak != 2 {
		t.Errorf("expected current streak 2, got %d", s.CurrentStreak)
	}
	if s.LongestSuccess != 4 {
		t.Errorf("expected longest success 4, got %d", s.LongestSuccess)
	}
	if s.LongestFailure != 2 {
		t.Errorf("expected longest failure 2, got %d", s.LongestFailure)
	}
}

func TestGetStreakMissingCommand(t *testing.T) {
	db := tempDB(t)
	_, err := GetStreak(db, "nonexistent.sh")
	if err == nil {
		t.Error("expected error for missing command, got nil")
	}
}
