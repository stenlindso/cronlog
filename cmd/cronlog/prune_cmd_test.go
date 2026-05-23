package main

import (
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func TestRunPruneKeepLast(t *testing.T) {
	db, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	base := time.Now()
	for i := 0; i < 4; i++ {
		_ = store.Insert(db, store.Run{
			Command:   "job",
			StartedAt: base.Add(time.Duration(i) * time.Second),
		})
	}

	if err := runPrune(db, []string{"--keep-last", "2"}); err != nil {
		t.Fatalf("runPrune: %v", err)
	}

	runs, _ := store.List(db, 10)
	if len(runs) != 2 {
		t.Errorf("expected 2 runs, got %d", len(runs))
	}
}

func TestRunPruneMissingValue(t *testing.T) {
	db, _ := store.Open(":memory:")
	if err := runPrune(db, []string{"--keep-last"}); err == nil {
		t.Error("expected error for missing value")
	}
}

func TestRunPruneUnknownFlag(t *testing.T) {
	db, _ := store.Open(":memory:")
	if err := runPrune(db, []string{"--bogus", "5"}); err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestRunPruneNoOptions(t *testing.T) {
	db, _ := store.Open(":memory:")
	// Should not error — just print a notice.
	if err := runPrune(db, []string{}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
