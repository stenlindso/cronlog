package main

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func tempAlertDB(t *testing.T) *sql.DB {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "alert-*.db")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	db, err := store.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedAlertEntry(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	runID, err := store.Insert(db, store.Run{
		Command:   "nightly.sh",
		StartedAt: time.Now().UTC(),
		Duration:  5,
		ExitCode:  1,
		Output:    "err",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.RecordAlert(db, runID, "nightly.sh", "on_failure", "slack"); err != nil {
		t.Fatal(err)
	}
	return runID
}

func TestRunAlertList(t *testing.T) {
	db := tempAlertDB(t)
	seedAlertEntry(t, db)

	if err := runAlert(db, []string{"list"}); err != nil {
		t.Errorf("runAlert list: %v", err)
	}
}

func TestRunAlertListFilterByCommand(t *testing.T) {
	db := tempAlertDB(t)
	seedAlertEntry(t, db)

	if err := runAlert(db, []string{"list", "--command", "nightly.sh"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunAlertDeleteMissingID(t *testing.T) {
	db := tempAlertDB(t)
	if err := runAlert(db, []string{"delete"}); err == nil {
		t.Error("expected error for missing id")
	}
}

func TestRunAlertUnknownSubcommand(t *testing.T) {
	db := tempAlertDB(t)
	if err := runAlert(db, []string{"purge"}); err == nil {
		t.Error("expected error for unknown subcommand")
	}
}

func TestRunAlertNoSubcommand(t *testing.T) {
	db := tempAlertDB(t)
	if err := runAlert(db, []string{}); err == nil {
		t.Error("expected error when no subcommand given")
	}
}
