package store_test

import (
	"testing"

	"github.com/user/cronlog/internal/store"
)

func seedNotifyRun(t *testing.T, db interface{ Exec(string, ...interface{}) (interface{ LastInsertId() (int64, error) }, error) }) {}

func TestAddAndListNotifyRules(t *testing.T) {
	db := tempDB(t)

	id, err := store.AddNotifyRule(db, "/bin/backup.sh", true, false)
	if err != nil {
		t.Fatalf("AddNotifyRule: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}

	rules, err := store.ListNotifyRules(db, "")
	if err != nil {
		t.Fatalf("ListNotifyRules: %v", err)
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Command != "/bin/backup.sh" {
		t.Errorf("expected command /bin/backup.sh, got %s", rules[0].Command)
	}
	if !rules[0].OnFailure {
		t.Error("expected OnFailure=true")
	}
}

func TestListNotifyRulesFilterByCommand(t *testing.T) {
	db := tempDB(t)

	_, _ = store.AddNotifyRule(db, "/bin/a.sh", true, false)
	_, _ = store.AddNotifyRule(db, "/bin/b.sh", false, true)

	rules, err := store.ListNotifyRules(db, "/bin/a.sh")
	if err != nil {
		t.Fatalf("ListNotifyRules: %v", err)
	}
	if len(rules) != 1 || rules[0].Command != "/bin/a.sh" {
		t.Errorf("expected 1 rule for /bin/a.sh, got %d", len(rules))
	}
}

func TestDeleteNotifyRule(t *testing.T) {
	db := tempDB(t)

	id, _ := store.AddNotifyRule(db, "/bin/cleanup.sh", true, true)
	if err := store.DeleteNotifyRule(db, id); err != nil {
		t.Fatalf("DeleteNotifyRule: %v", err)
	}

	rules, _ := store.ListNotifyRules(db, "")
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after delete, got %d", len(rules))
	}
}

func TestMatchingRulesOnFailure(t *testing.T) {
	db := tempDB(t)

	_, _ = store.AddNotifyRule(db, "/bin/job.sh", true, false)

	matched, err := store.MatchingRules(db, "/bin/job.sh", 1)
	if err != nil {
		t.Fatalf("MatchingRules: %v", err)
	}
	if len(matched) != 1 {
		t.Errorf("expected 1 match on failure, got %d", len(matched))
	}

	matched, _ = store.MatchingRules(db, "/bin/job.sh", 0)
	if len(matched) != 0 {
		t.Errorf("expected 0 matches on success, got %d", len(matched))
	}
}

func TestMatchingRulesOnSuccess(t *testing.T) {
	db := tempDB(t)

	_, _ = store.AddNotifyRule(db, "/bin/job.sh", false, true)

	matched, err := store.MatchingRules(db, "/bin/job.sh", 0)
	if err != nil {
		t.Fatalf("MatchingRules: %v", err)
	}
	if len(matched) != 1 {
		t.Errorf("expected 1 match on success, got %d", len(matched))
	}
}
