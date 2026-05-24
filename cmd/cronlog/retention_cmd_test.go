package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/example/cronlog/internal/store"
)

func tempRetentionDB(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "retention_test.db")
	db, err := store.Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.Close()
	return path
}

func TestRunRetentionList(t *testing.T) {
	path := tempRetentionDB(t)
	t.Setenv("CRONLOG_DB", path)

	_ = runRetention([]string{"--add", "--command", "job.sh", "--keep-last", "5"})

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := runRetention([]string{})
	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if !strings.Contains(buf.String(), "job.sh") {
		t.Errorf("expected job.sh in output, got: %s", buf.String())
	}
}

func TestRunRetentionAddMissingFlags(t *testing.T) {
	path := tempRetentionDB(t)
	t.Setenv("CRONLOG_DB", path)

	err := runRetention([]string{"--add", "--command", "job.sh"})
	if err == nil {
		t.Error("expected error when no age/count provided")
	}
}

func TestRunRetentionDeleteMissing(t *testing.T) {
	path := tempRetentionDB(t)
	t.Setenv("CRONLOG_DB", path)

	err := runRetention([]string{"--delete", "--command", "nonexistent"})
	if err == nil {
		t.Error("expected error deleting nonexistent policy")
	}
}

func TestRunRetentionApply(t *testing.T) {
	path := tempRetentionDB(t)
	t.Setenv("CRONLOG_DB", path)

	_ = runRetention([]string{"--add", "--max-age-days", "30"})

	err := runRetention([]string{"--apply"})
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
}

func TestRunRetentionListEmpty(t *testing.T) {
	path := tempRetentionDB(t)
	t.Setenv("CRONLOG_DB", path)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := runRetention([]string{})
	w.Close()
	os.Stdout = old

	var buf strings.Builder
	_, _ = buf.ReadFrom(r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "no retention policies") {
		t.Errorf("expected empty message, got: %s", buf.String())
	}
}
