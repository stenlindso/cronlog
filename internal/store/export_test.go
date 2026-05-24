package store_test

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func seedExportRuns(t *testing.T, db interface{ Exec(string, ...any) (interface{}, error) }) {
	t.Helper()
}

func TestExportJSONEmpty(t *testing.T) {
	db := tempDB(t)

	var buf bytes.Buffer
	if err := store.ExportJSON(db, &buf, 0); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	var runs []store.Run
	if err := json.Unmarshal(buf.Bytes(), &runs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestExportJSONAllRuns(t *testing.T) {
	db := tempDB(t)

	for i := 0; i < 3; i++ {
		_, err := store.Insert(db, store.Run{
			Command:   "echo hello",
			StartedAt: time.Now(),
			Duration:  100,
			ExitCode:  0,
			Output:    "hello",
		})
		if err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := store.ExportJSON(db, &buf, 0); err != nil {
		t.Fatalf("ExportJSON error: %v", err)
	}

	var runs []store.Run
	if err := json.Unmarshal(buf.Bytes(), &runs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(runs) != 3 {
		t.Errorf("expected 3 runs, got %d", len(runs))
	}
}

func TestExportJSONByCommand(t *testing.T) {
	db := tempDB(t)

	commands := []string{"echo hello", "echo hello", "ls -la"}
	for _, cmd := range commands {
		_, err := store.Insert(db, store.Run{
			Command:   cmd,
			StartedAt: time.Now(),
			Duration:  50,
			ExitCode:  0,
			Output:    "",
		})
		if err != nil {
			t.Fatalf("Insert: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := store.ExportJSONByCommand(db, &buf, "echo hello", 0); err != nil {
		t.Fatalf("ExportJSONByCommand error: %v", err)
	}

	var runs []store.Run
	if err := json.Unmarshal(buf.Bytes(), &runs); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(runs) != 2 {
		t.Errorf("expected 2 runs for 'echo hello', got %d", len(runs))
	}
	for _, r := range runs {
		if r.Command != "echo hello" {
			t.Errorf("unexpected command %q", r.Command)
		}
	}
}
