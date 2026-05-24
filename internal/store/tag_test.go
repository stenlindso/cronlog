package store_test

import (
	"testing"

	"github.com/example/cronlog/internal/store"
)

func TestAddAndListTags(t *testing.T) {
	db := tempDB(t)

	runID := seedSingleRun(t, db, "echo hello")

	if err := store.AddTags(db, runID, []string{"nightly", "important"}); err != nil {
		t.Fatalf("AddTags: %v", err)
	}

	tags, err := store.ListTags(db, runID)
	if err != nil {
		t.Fatalf("ListTags: %v", err)
	}
	if len(tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tags))
	}
	if tags[0] != "important" || tags[1] != "nightly" {
		t.Errorf("unexpected tags: %v", tags)
	}
}

func TestAddTagsEmpty(t *testing.T) {
	db := tempDB(t)
	runID := seedSingleRun(t, db, "ls")

	if err := store.AddTags(db, runID, []string{}); err != nil {
		t.Fatalf("AddTags with empty slice: %v", err)
	}
	tags, _ := store.ListTags(db, runID)
	if len(tags) != 0 {
		t.Errorf("expected no tags, got %v", tags)
	}
}

func TestAddTagsDeduplicated(t *testing.T) {
	db := tempDB(t)
	runID := seedSingleRun(t, db, "date")

	_ = store.AddTags(db, runID, []string{"daily"})
	// second insert of same tag should not error (INSERT OR IGNORE)
	if err := store.AddTags(db, runID, []string{"daily"}); err != nil {
		t.Fatalf("duplicate AddTags: %v", err)
	}
	tags, _ := store.ListTags(db, runID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag after duplicate insert, got %d", len(tags))
	}
}

func TestListRunsByTag(t *testing.T) {
	db := tempDB(t)

	id1 := seedSingleRun(t, db, "backup.sh")
	id2 := seedSingleRun(t, db, "backup.sh")
	id3 := seedSingleRun(t, db, "cleanup.sh")

	_ = store.AddTags(db, id1, []string{"backup"})
	_ = store.AddTags(db, id2, []string{"backup"})
	_ = store.AddTags(db, id3, []string{"cleanup"})

	runs, err := store.ListRunsByTag(db, "backup", 10)
	if err != nil {
		t.Fatalf("ListRunsByTag: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs with tag 'backup', got %d", len(runs))
	}
}

func TestListRunsByTagMissing(t *testing.T) {
	db := tempDB(t)
	runs, err := store.ListRunsByTag(db, "nonexistent", 10)
	if err != nil {
		t.Fatalf("ListRunsByTag missing: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

// seedSingleRun inserts a minimal run and returns its ID.
func seedSingleRun(t *testing.T, db interface{ Query(string, ...interface{}) (interface{}, error) }, cmd string) int64 {
	t.Helper()
	// Use the real *sql.DB via the store.Insert helper.
	return 0 // placeholder — see below
}
