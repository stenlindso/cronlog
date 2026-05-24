package store_test

import (
	"testing"
	"time"
)

func seedSnapshotRuns(t *testing.T, db interface{ Exec(string, ...interface{}) (interface{}, error) }) {
	t.Helper()
}

func TestTakeAndListSnapshots(t *testing.T) {
	db := tempDB(t)

	// Insert some runs for the command.
	now := time.Now().UTC()
	for i := 0; i < 5; i++ {
		exitCode := 0
		if i == 4 {
			exitCode = 1
		}
		_, err := db.Exec(`
			INSERT INTO runs (command, started_at, duration_ms, exit_code, output)
			VALUES (?, ?, ?, ?, '')
		`, "/bin/check", now.Add(time.Duration(i)*time.Minute).Unix(), 1200+i*100, exitCode)
		if err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	snap, err := TakeSnapshot(db, "/bin/check")
	if err != nil {
		t.Fatalf("TakeSnapshot: %v", err)
	}
	if snap.TotalRuns != 5 {
		t.Errorf("expected 5 total runs, got %d", snap.TotalRuns)
	}
	if snap.SuccessRate != 80.0 {
		t.Errorf("expected 80%% success rate, got %.2f", snap.SuccessRate)
	}
	if snap.LastExitCode != 0 && snap.LastExitCode != 1 {
		t.Errorf("unexpected last exit code: %d", snap.LastExitCode)
	}
	if snap.ID == 0 {
		t.Error("expected non-zero snapshot ID")
	}
}

func TestListSnapshotsOrdered(t *testing.T) {
	db := tempDB(t)

	now := time.Now().UTC()
	_, err := db.Exec(`INSERT INTO runs (command, started_at, duration_ms, exit_code, output) VALUES (?, ?, ?, ?, '')`,
		"/bin/job", now.Unix(), 500, 0)
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}

	for i := 0; i < 3; i++ {
		_, err := TakeSnapshot(db, "/bin/job")
		if err != nil {
			t.Fatalf("TakeSnapshot %d: %v", i, err)
		}
	}

	snaps, err := ListSnapshots(db, "/bin/job", 10)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 3 {
		t.Fatalf("expected 3 snapshots, got %d", len(snaps))
	}
	for i := 1; i < len(snaps); i++ {
		if snaps[i].CapturedAt.After(snaps[i-1].CapturedAt) {
			t.Errorf("snapshots not ordered most-recent-first at index %d", i)
		}
	}
}

func TestListSnapshotsLimit(t *testing.T) {
	db := tempDB(t)

	now := time.Now().UTC()
	_, _ = db.Exec(`INSERT INTO runs (command, started_at, duration_ms, exit_code, output) VALUES (?, ?, ?, ?, '')`,
		"/bin/limited", now.Unix(), 300, 0)

	for i := 0; i < 5; i++ {
		_, _ = TakeSnapshot(db, "/bin/limited")
	}

	snaps, err := ListSnapshots(db, "/bin/limited", 2)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 2 {
		t.Errorf("expected 2 snapshots with limit=2, got %d", len(snaps))
	}
}
