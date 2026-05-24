package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/yourusername/cronlog/internal/store"
)

func tempQuotaDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "quota_test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestRunQuotaList(t *testing.T) {
	db := tempQuotaDB(t)
	_ = store.UpsertQuota(db, "backup.sh", 3, 3600)
	if err := runQuota(db, []string{"--list"}); err != nil {
		t.Fatalf("runQuota --list: %v", err)
	}
}

func TestRunQuotaListEmpty(t *testing.T) {
	db := tempQuotaDB(t)
	if err := runQuota(db, []string{"--list"}); err != nil {
		t.Fatalf("runQuota --list empty: %v", err)
	}
}

func TestRunQuotaAdd(t *testing.T) {
	db := tempQuotaDB(t)
	err := runQuota(db, []string{"--add", "--command", "sync.sh", "--max-runs", "5", "--window", "600"})
	if err != nil {
		t.Fatalf("runQuota --add: %v", err)
	}
	quotas, _ := store.ListQuotas(db)
	if len(quotas) != 1 || quotas[0].Command != "sync.sh" {
		t.Errorf("quota not stored: %+v", quotas)
	}
}

func TestRunQuotaAddMissingFlags(t *testing.T) {
	db := tempQuotaDB(t)
	if err := runQuota(db, []string{"--add", "--command", "x.sh"}); err == nil {
		t.Error("expected error when max-runs/window missing")
	}
}

func TestRunQuotaDelete(t *testing.T) {
	db := tempQuotaDB(t)
	_ = store.UpsertQuota(db, "clean.sh", 1, 60)
	if err := runQuota(db, []string{"--delete", "--command", "clean.sh"}); err != nil {
		t.Fatalf("runQuota --delete: %v", err)
	}
	quotas, _ := store.ListQuotas(db)
	if len(quotas) != 0 {
		t.Error("expected quota to be deleted")
	}
}

func TestRunQuotaDeleteMissing(t *testing.T) {
	db := tempQuotaDB(t)
	if err := runQuota(db, []string{"--delete", "--command", "ghost.sh"}); err == nil {
		t.Error("expected error deleting non-existent quota")
	}
}

func TestRunQuotaNoOptions(t *testing.T) {
	db := tempQuotaDB(t)
	if err := runQuota(db, []string{}); err == nil {
		t.Error("expected error with no options")
	}
}
