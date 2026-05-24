package store_test

import (
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func seedAnnotationRun(t *testing.T, db interface{ Exec(string, ...any) (interface{}, error) }) int64 {
	t.Helper()
	// Use the shared tempDB helper and Insert from query.go
	return 0
}

func TestAddAndGetAnnotations(t *testing.T) {
	db := tempDB(t)

	run := store.Run{
		Command:   "backup.sh",
		StartedAt: time.Now().UTC(),
		Duration:  2 * time.Second,
		ExitCode:  0,
		Output:    "done",
	}
	id, err := store.Insert(db, run)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	if err := store.AddAnnotation(db, id, "first note"); err != nil {
		t.Fatalf("AddAnnotation: %v", err)
	}
	if err := store.AddAnnotation(db, id, "second note"); err != nil {
		t.Fatalf("AddAnnotation: %v", err)
	}

	annotations, err := store.GetAnnotations(db, id)
	if err != nil {
		t.Fatalf("GetAnnotations: %v", err)
	}
	if len(annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(annotations))
	}
	if annotations[0].Note != "first note" {
		t.Errorf("expected 'first note', got %q", annotations[0].Note)
	}
	if annotations[1].Note != "second note" {
		t.Errorf("expected 'second note', got %q", annotations[1].Note)
	}
}

func TestAddAnnotationMissingRun(t *testing.T) {
	db := tempDB(t)

	err := store.AddAnnotation(db, 9999, "orphan note")
	if err == nil {
		t.Fatal("expected error for missing run, got nil")
	}
}

func TestGetAnnotationsEmpty(t *testing.T) {
	db := tempDB(t)

	run := store.Run{
		Command:   "noop.sh",
		StartedAt: time.Now().UTC(),
		Duration:  time.Second,
		ExitCode:  0,
		Output:    "",
	}
	id, err := store.Insert(db, run)
	if err != nil {
		t.Fatalf("insert: %v", err)
	}

	annotations, err := store.GetAnnotations(db, id)
	if err != nil {
		t.Fatalf("GetAnnotations: %v", err)
	}
	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}
}
