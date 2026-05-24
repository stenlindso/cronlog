package store_test

import (
	"testing"
	"time"
)

func seedAlertRun(t *testing.T) (*tempDBHandle, int64) {
	t.Helper()
	db := tempDB(t)
	id, err := Insert(db.DB, Run{
		Command:   "backup.sh",
		StartedAt: time.Now().UTC(),
		Duration:  2,
		ExitCode:  1,
		Output:    "failed",
	})
	if err != nil {
		t.Fatalf("seed insert: %v", err)
	}
	return db, id
}

func TestRecordAndListAlerts(t *testing.T) {
	db, runID := seedAlertRun(t)

	if err := RecordAlert(db.DB, runID, "backup.sh", "on_failure", "email"); err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}

	alerts, err := ListAlerts(db.DB, "", 10)
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RunID != runID {
		t.Errorf("expected run_id %d, got %d", runID, alerts[0].RunID)
	}
	if alerts[0].Channel != "email" {
		t.Errorf("expected channel email, got %s", alerts[0].Channel)
	}
}

func TestListAlertsFilterByCommand(t *testing.T) {
	db, runID := seedAlertRun(t)
	_ = RecordAlert(db.DB, runID, "backup.sh", "on_failure", "email")
	_ = RecordAlert(db.DB, runID, "other.sh", "on_failure", "slack")

	alerts, err := ListAlerts(db.DB, "backup.sh", 10)
	if err != nil {
		t.Fatalf("ListAlerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert for backup.sh, got %d", len(alerts))
	}
}

func TestDeleteAlert(t *testing.T) {
	db, runID := seedAlertRun(t)
	_ = RecordAlert(db.DB, runID, "backup.sh", "on_failure", "email")

	alerts, _ := ListAlerts(db.DB, "", 10)
	if len(alerts) == 0 {
		t.Fatal("expected alert to exist before delete")
	}

	if err := DeleteAlert(db.DB, alerts[0].ID); err != nil {
		t.Fatalf("DeleteAlert: %v", err)
	}

	alerts, _ = ListAlerts(db.DB, "", 10)
	if len(alerts) != 0 {
		t.Errorf("expected 0 alerts after delete, got %d", len(alerts))
	}
}

func TestDeleteAlertMissing(t *testing.T) {
	db := tempDB(t)
	if err := DeleteAlert(db.DB, 9999); err == nil {
		t.Error("expected error for missing alert, got nil")
	}
}
