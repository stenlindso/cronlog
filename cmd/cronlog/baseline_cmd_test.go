package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func tempBaselineDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "baseline_test.db")
	db, err := store.Open(p)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	return p
}

func seedBaselineDB(t *testing.T, dbPath string) {
	t.Helper()
	db, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	for i, ms := range []int{1000, 3000, 5000} {
		_, err := db.Exec(`
			INSERT INTO runs (command, started_at, duration_ms, exit_code, output)
			VALUES ('sync.sh', ?, ?, 0, '')
		`, time.Now().Add(time.Duration(-i)*time.Minute).UTC().Format(time.RFC3339), ms)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}
}

func TestRunBaselineCompute(t *testing.T) {
	dbPath := tempBaselineDB(t)
	seedBaselineDB(t, dbPath)

	err := runBaseline([]string{"--command", "sync.sh", "--samples", "5"}, dbPath)
	if err != nil {
		t.Fatalf("runBaseline: %v", err)
	}
}

func TestRunBaselineGet(t *testing.T) {
	dbPath := tempBaselineDB(t)
	seedBaselineDB(t, dbPath)

	// First compute
	if err := runBaseline([]string{"--command", "sync.sh"}, dbPath); err != nil {
		t.Fatalf("compute: %v", err)
	}
	// Then retrieve
	if err := runBaseline([]string{"--command", "sync.sh", "--get"}, dbPath); err != nil {
		t.Fatalf("get: %v", err)
	}
}

func TestRunBaselineMissingCommand(t *testing.T) {
	dbPath := tempBaselineDB(t)
	err := runBaseline([]string{}, dbPath)
	if err == nil {
		t.Error("expected error when --command is missing")
	}
}

func TestRunBaselineNoRuns(t *testing.T) {
	dbPath := tempBaselineDB(t)
	err := runBaseline([]string{"--command", "ghost.sh"}, dbPath)
	if err == nil {
		t.Error("expected error for command with no runs")
	}
}
