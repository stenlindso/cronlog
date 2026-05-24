package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/cronlog/internal/store"
)

func tempNotifyDB(t *testing.T) interface{ Close() error } {
	t.Helper()
	dir := t.TempDir()
	db, err := store.Open(filepath.Join(dir, "cronlog.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRunNotifyAddAndList(t *testing.T) {
	dir := t.TempDir()
	db, _ := store.Open(filepath.Join(dir, "cronlog.db"))
	defer db.Close()

	out := captureStdout(t, func() {
		if err := runNotify(db, []string{"--add", "/bin/backup.sh", "--on-failure"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("Added notify rule")) {
		t.Errorf("expected 'Added notify rule' in output, got: %s", out)
	}

	out = captureStdout(t, func() {
		if err := runNotify(db, []string{"--list"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("/bin/backup.sh")) {
		t.Errorf("expected command in list output, got: %s", out)
	}
}

func TestRunNotifyNoOptions(t *testing.T) {
	dir := t.TempDir()
	db, _ := store.Open(filepath.Join(dir, "cronlog.db"))
	defer db.Close()

	err := runNotify(db, []string{})
	if err == nil {
		t.Fatal("expected error with no options")
	}
}

func TestRunNotifyAddMissingCondition(t *testing.T) {
	dir := t.TempDir()
	db, _ := store.Open(filepath.Join(dir, "cronlog.db"))
	defer db.Close()

	err := runNotify(db, []string{"--add", "/bin/job.sh"})
	if err == nil {
		t.Fatal("expected error when neither --on-failure nor --on-success specified")
	}
}

func TestRunNotifyListEmpty(t *testing.T) {
	dir := t.TempDir()
	db, _ := store.Open(filepath.Join(dir, "cronlog.db"))
	defer db.Close()

	out := captureStdout(t, func() {
		if err := runNotify(db, []string{"--list"}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if !bytes.Contains([]byte(out), []byte("No notification rules")) {
		t.Errorf("expected empty message, got: %s", out)
	}
}

// captureStdout redirects os.Stdout and returns captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	old := os.Stdout
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}
