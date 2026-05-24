package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/example/cronlog/internal/store"
)

func tempDigestDB(t *testing.T) *store.DB {
	t.Helper()
	tmp := t.TempDir()
	db, err := store.Open(fmt.Sprintf("%s/digest_test.db", tmp))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func seedDigestDB(t *testing.T, db *store.DB) {
	t.Helper()
	now := time.Now().UTC()
	for i := 0; i < 3; i++ {
		_, err := store.Insert(db, store.Run{
			Command:    "backup.sh",
			StartedAt:  now.AddDate(0, 0, -i),
			DurationMs: int64(1000 + i*200),
			ExitCode:   0,
			Output:     "ok",
		})
		if err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
	_, err := store.Insert(db, store.Run{
		Command:    "backup.sh",
		StartedAt:  now.AddDate(0, 0, -1),
		DurationMs: 500,
		ExitCode:   1,
		Output:     "failed",
	})
	if err != nil {
		t.Fatalf("insert failure run: %v", err)
	}
}

func TestRunDigestAllCommands(t *testing.T) {
	db := tempDigestDB(t)
	seedDigestDB(t, db)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDigest([]string{"--days", "30"}, db)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "backup.sh") {
		t.Errorf("expected command in output, got:\n%s", out)
	}
	if !contains(out, "RUNS") {
		t.Errorf("expected header RUNS in output, got:\n%s", out)
	}
}

func TestRunDigestFilterByCommand(t *testing.T) {
	db := tempDigestDB(t)
	seedDigestDB(t, db)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDigest([]string{"--command", "backup.sh", "--days", "30"}, db)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "backup.sh") {
		t.Errorf("expected backup.sh in output, got:\n%s", out)
	}
}

func TestRunDigestEmptyDB(t *testing.T) {
	db := tempDigestDB(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runDigest([]string{"--days", "7"}, db)

	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "no data") {
		t.Errorf("expected 'no data' message, got:\n%s", out)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
