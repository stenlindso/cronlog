package store_test

import (
	"testing"
	"time"
)

func TestUpsertAndListRetentionPolicies(t *testing.T) {
	db := tempDB(t)

	err := UpsertRetentionPolicy(db, RetentionPolicy{Command: "", MaxAgeDays: 30})
	if err != nil {
		t.Fatalf("upsert global policy: %v", err)
	}
	err = UpsertRetentionPolicy(db, RetentionPolicy{Command: "backup.sh", MaxCount: 10})
	if err != nil {
		t.Fatalf("upsert command policy: %v", err)
	}

	policies, err := ListRetentionPolicies(db)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(policies))
	}
}

func TestUpsertRetentionPolicyUpdates(t *testing.T) {
	db := tempDB(t)

	_ = UpsertRetentionPolicy(db, RetentionPolicy{Command: "job.sh", MaxAgeDays: 7})
	_ = UpsertRetentionPolicy(db, RetentionPolicy{Command: "job.sh", MaxAgeDays: 14, MaxCount: 5})

	policies, _ := ListRetentionPolicies(db)
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy after upsert, got %d", len(policies))
	}
	if policies[0].MaxAgeDays != 14 {
		t.Errorf("expected MaxAgeDays=14, got %d", policies[0].MaxAgeDays)
	}
}

func TestRetentionPolicyRequiresField(t *testing.T) {
	db := tempDB(t)
	err := UpsertRetentionPolicy(db, RetentionPolicy{Command: "job.sh"})
	if err == nil {
		t.Error("expected error for empty policy, got nil")
	}
}

func TestDeleteRetentionPolicy(t *testing.T) {
	db := tempDB(t)
	_ = UpsertRetentionPolicy(db, RetentionPolicy{Command: "cleanup.sh", MaxCount: 5})

	if err := DeleteRetentionPolicy(db, "cleanup.sh"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	policies, _ := ListRetentionPolicies(db)
	if len(policies) != 0 {
		t.Errorf("expected 0 policies after delete, got %d", len(policies))
	}
}

func TestDeleteRetentionPolicyMissing(t *testing.T) {
	db := tempDB(t)
	err := DeleteRetentionPolicy(db, "nonexistent")
	if err == nil {
		t.Error("expected error deleting missing policy")
	}
}

func TestApplyRetentionPolicies(t *testing.T) {
	db := tempDB(t)

	// seed some old runs
	for i := 0; i < 5; i++ {
		_, _ = Insert(db, Run{
			Command:   "old-job.sh",
			StartedAt: time.Now().Add(-40 * 24 * time.Hour),
			Duration:  1,
			ExitCode:  0,
		})
	}

	_ = UpsertRetentionPolicy(db, RetentionPolicy{Command: "", MaxAgeDays: 30})

	n, err := ApplyRetentionPolicies(db)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if n != 5 {
		t.Errorf("expected 5 pruned, got %d", n)
	}
}
