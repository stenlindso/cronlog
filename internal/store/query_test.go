package store

import (
	"testing"
	"time"
)

func TestInsertAndGetByID(t *testing.T) {
	db := tempDB(t)

	now := time.Now().UTC().Truncate(time.Millisecond)
	r := Run{
		Command:   "echo hello",
		StartedAt: now,
		Duration:  250 * time.Millisecond,
		ExitCode:  0,
		Stdout:    "hello\n",
		Stderr:    "",
	}

	id, err := Insert(db, r)
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive id, got %d", id)
	}

	got, err := GetByID(db, id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got == nil {
		t.Fatal("expected run, got nil")
	}
	if got.Command != r.Command {
		t.Errorf("Command: want %q, got %q", r.Command, got.Command)
	}
	if got.ExitCode != r.ExitCode {
		t.Errorf("ExitCode: want %d, got %d", r.ExitCode, got.ExitCode)
	}
	if got.Duration != r.Duration {
		t.Errorf("Duration: want %v, got %v", r.Duration, got.Duration)
	}
	if got.Stdout != r.Stdout {
		t.Errorf("Stdout: want %q, got %q", r.Stdout, got.Stdout)
	}
}

func TestGetByIDMissing(t *testing.T) {
	db := tempDB(t)
	got, err := GetByID(db, 9999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for missing id, got %+v", got)
	}
}

func TestListLimit(t *testing.T) {
	db := tempDB(t)
	base := time.Now().UTC()

	for i := 0; i < 5; i++ {
		_, err := Insert(db, Run{
			Command:   "true",
			StartedAt: base.Add(time.Duration(i) * time.Second),
			Duration:  10 * time.Millisecond,
			ExitCode:  0,
		})
		if err != nil {
			t.Fatalf("Insert %d: %v", i, err)
		}
	}

	runs, err := List(db, 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(runs) != 3 {
		t.Fatalf("expected 3 runs, got %d", len(runs))
	}
	// Verify descending order
	if !runs[0].StartedAt.After(runs[1].StartedAt) {
		t.Errorf("expected descending order: runs[0]=%v runs[1]=%v", runs[0].StartedAt, runs[1].StartedAt)
	}
}
