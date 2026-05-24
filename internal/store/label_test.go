package store_test

import (
	"database/sql"
	"testing"
)

func seedLabelRun(t *testing.T) int64 {
	t.Helper()
	db := tempDB(t)
	id, err := Insert(db, Run{
		Command:  "backup.sh",
		ExitCode: 0,
		Duration: 1,
		Output:   "",
	})
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	return id
}

func TestSetAndGetLabels(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})

	err := SetLabels(db, id, []string{"env=prod", "region=us-east"})
	if err != nil {
		t.Fatalf("SetLabels: %v", err)
	}

	labels, err := GetLabels(db, id)
	if err != nil {
		t.Fatalf("GetLabels: %v", err)
	}
	if len(labels) != 2 {
		t.Fatalf("expected 2 labels, got %d", len(labels))
	}
	if labels[0].Key != "env" || labels[0].Value != "prod" {
		t.Errorf("unexpected label[0]: %+v", labels[0])
	}
	if labels[1].Key != "region" || labels[1].Value != "us-east" {
		t.Errorf("unexpected label[1]: %+v", labels[1])
	}
}

func TestSetLabelsUpserts(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})

	_ = SetLabels(db, id, []string{"env=staging"})
	_ = SetLabels(db, id, []string{"env=prod"})

	labels, _ := GetLabels(db, id)
	if len(labels) != 1 || labels[0].Value != "prod" {
		t.Errorf("expected upserted value 'prod', got %+v", labels)
	}
}

func TestSetLabelsInvalidFormat(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})

	err := SetLabels(db, id, []string{"noequalssign"})
	if err == nil {
		t.Error("expected error for invalid label format")
	}
}

func TestSetLabelsEmpty(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})

	if err := SetLabels(db, id, nil); err != nil {
		t.Errorf("unexpected error for empty labels: %v", err)
	}
}

func TestDeleteLabel(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})
	_ = SetLabels(db, id, []string{"env=prod", "tier=web"})

	if err := DeleteLabel(db, id, "env"); err != nil {
		t.Fatalf("DeleteLabel: %v", err)
	}

	labels, _ := GetLabels(db, id)
	if len(labels) != 1 || labels[0].Key != "tier" {
		t.Errorf("expected only 'tier' label remaining, got %+v", labels)
	}
}

func TestDeleteLabelMissing(t *testing.T) {
	db := tempDB(t)
	id, _ := Insert(db, Run{Command: "job.sh", ExitCode: 0})

	err := DeleteLabel(db, id, "nonexistent")
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}
