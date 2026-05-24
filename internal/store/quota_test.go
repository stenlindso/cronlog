package store_test

import (
	"testing"
	"time"

	"github.com/yourusername/cronlog/internal/store"
)

func TestUpsertAndListQuotas(t *testing.T) {
	db := tempDB(t)
	if err := store.UpsertQuota(db, "backup.sh", 5, 3600); err != nil {
		t.Fatalf("UpsertQuota: %v", err)
	}
	quotas, err := store.ListQuotas(db)
	if err != nil {
		t.Fatalf("ListQuotas: %v", err)
	}
	if len(quotas) != 1 {
		t.Fatalf("expected 1 quota, got %d", len(quotas))
	}
	if quotas[0].Command != "backup.sh" || quotas[0].MaxRuns != 5 || quotas[0].WindowSec != 3600 {
		t.Errorf("unexpected quota: %+v", quotas[0])
	}
}

func TestUpsertQuotaUpdates(t *testing.T) {
	db := tempDB(t)
	_ = store.UpsertQuota(db, "sync.sh", 3, 600)
	_ = store.UpsertQuota(db, "sync.sh", 10, 1200)
	quotas, _ := store.ListQuotas(db)
	if len(quotas) != 1 {
		t.Fatalf("expected 1 quota after upsert, got %d", len(quotas))
	}
	if quotas[0].MaxRuns != 10 || quotas[0].WindowSec != 1200 {
		t.Errorf("quota not updated: %+v", quotas[0])
	}
}

func TestUpsertQuotaValidation(t *testing.T) {
	db := tempDB(t)
	if err := store.UpsertQuota(db, "", 5, 60); err == nil {
		t.Error("expected error for empty command")
	}
	if err := store.UpsertQuota(db, "cmd", 0, 60); err == nil {
		t.Error("expected error for zero max_runs")
	}
	if err := store.UpsertQuota(db, "cmd", 5, 0); err == nil {
		t.Error("expected error for zero window_sec")
	}
}

func TestDeleteQuota(t *testing.T) {
	db := tempDB(t)
	_ = store.UpsertQuota(db, "clean.sh", 2, 300)
	if err := store.DeleteQuota(db, "clean.sh"); err != nil {
		t.Fatalf("DeleteQuota: %v", err)
	}
	quotas, _ := store.ListQuotas(db)
	if len(quotas) != 0 {
		t.Errorf("expected 0 quotas after delete, got %d", len(quotas))
	}
}

func TestDeleteQuotaMissing(t *testing.T) {
	db := tempDB(t)
	if err := store.DeleteQuota(db, "nonexistent"); err == nil {
		t.Error("expected error deleting missing quota")
	}
}

func TestCheckQuotaNoRule(t *testing.T) {
	db := tempDB(t)
	v, err := store.CheckQuota(db, "any.sh")
	if err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if v != nil {
		t.Errorf("expected nil violation with no rule, got %v", v)
	}
}

func TestCheckQuotaWithinLimit(t *testing.T) {
	db := tempDB(t)
	_ = store.UpsertQuota(db, "job.sh", 5, 3600)
	v, err := store.CheckQuota(db, "job.sh")
	if err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if v != nil {
		t.Errorf("expected no violation, got %v", v)
	}
}

func TestCheckQuotaExceeded(t *testing.T) {
	db := tempDB(t)
	_ = store.UpsertQuota(db, "busy.sh", 2, 3600)
	for i := 0; i < 3; i++ {
		_, _ = store.Insert(db, store.Run{
			Command:   "busy.sh",
			StartedAt: time.Now().UTC(),
			Duration:  100,
			ExitCode:  0,
		})
	}
	v, err := store.CheckQuota(db, "busy.sh")
	if err != nil {
		t.Fatalf("CheckQuota: %v", err)
	}
	if v == nil {
		t.Fatal("expected quota violation")
	}
	if v.Actual != 3 || v.MaxRuns != 2 {
		t.Errorf("unexpected violation values: %+v", v)
	}
}
