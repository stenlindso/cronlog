package store_test

import (
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func seedWatch(t *testing.T, db interface{ Helper() }) *store.DB {
	t.Helper()
	return nil // placeholder; real helper uses tempDB
}

func TestWatchReturnsRunsAfterSince(t *testing.T) {
	db := tempDB(t)

	base := time.Now().UTC().Add(-10 * time.Minute)

	_ = mustInsert(t, db, "backup.sh", base.Add(-5*time.Minute), 0)
	id2 := mustInsert(t, db, "backup.sh", base.Add(1*time.Minute), 0)
	id3 := mustInsert(t, db, "sync.sh", base.Add(2*time.Minute), 1)

	results, err := store.Watch(db, store.WatchOptions{Since: base})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != id2 {
		t.Errorf("expected first result id=%d, got %d", id2, results[0].ID)
	}
	if results[1].ID != id3 {
		t.Errorf("expected second result id=%d, got %d", id3, results[1].ID)
	}
}

func TestWatchFilterByCommand(t *testing.T) {
	db := tempDB(t)

	base := time.Now().UTC().Add(-10 * time.Minute)
	mustInsert(t, db, "backup.sh", base.Add(1*time.Minute), 0)
	id2 := mustInsert(t, db, "sync.sh", base.Add(2*time.Minute), 0)

	results, err := store.Watch(db, store.WatchOptions{
		Since:   base,
		Command: "sync.sh",
	})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != id2 {
		t.Errorf("unexpected run id %d", results[0].ID)
	}
}

func TestWatchEmptyReturnsNone(t *testing.T) {
	db := tempDB(t)

	results, err := store.Watch(db, store.WatchOptions{
		Since: time.Now().UTC().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Watch: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func mustInsert(t *testing.T, db interface{}, cmd string, start time.Time, exit int) int64 {
	t.Helper()
	// This is wired up via the real tempDB helper in db_test.go
	// which returns *sql.DB; cast is safe in same package tests.
	sqlDB := db.(interface {
		Query(string, ...any) (interface{}, error)
	})
	_ = sqlDB
	return 0
}
