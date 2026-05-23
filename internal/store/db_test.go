package store_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func tempDB(t *testing.T) *store.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "cronlog_test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenCreatesSchema(t *testing.T) {
	// Simply opening a DB should not error — schema is applied automatically.
	_ = tempDB(t)
}

func TestInsertAndList(t *testing.T) {
	db := tempDB(t)

	now := time.Now().Truncate(time.Millisecond)
	run := &store.JobRun{
		Name:      "backup",
		Command:   "/usr/local/bin/backup.sh",
		StartedAt: now,
		Duration:  3*time.Second + 250*time.Millisecond,
		ExitCode:  0,
		Output:    "backup complete",
	}

	if err := db.Insert(run); err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if run.ID == 0 {
		t.Fatal("expected non-zero ID after insert")
	}

	runs, err := db.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	got := runs[0]
	if got.Name != run.Name {
		t.Errorf("Name: want %q, got %q", run.Name, got.Name)
	}
	if got.ExitCode != run.ExitCode {
		t.Errorf("ExitCode: want %d, got %d", run.ExitCode, got.ExitCode)
	}
	if got.Duration != run.Duration {
		t.Errorf("Duration: want %v, got %v", run.Duration, got.Duration)
	}
	if !got.StartedAt.Equal(run.StartedAt) {
		t.Errorf("StartedAt: want %v, got %v", run.StartedAt, got.StartedAt)
	}
}

func TestListOrderedMostRecentFirst(t *testing.T) {
	db := tempDB(t)
	base := time.Now().Truncate(time.Millisecond)

	for i, name := range []string{"first", "second", "third"} {
		if err := db.Insert(&store.JobRun{
			Name:      name,
			Command:   "echo " + name,
			StartedAt: base.Add(time.Duration(i) * time.Minute),
			Duration:  time.Second,
			ExitCode:  0,
			Output:    "",
		}); err != nil {
			t.Fatalf("Insert %s: %v", name, err)
		}
	}

	runs, err := db.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if runs[0].Name != "third" {
		t.Errorf("expected most recent first, got %q", runs[0].Name)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
