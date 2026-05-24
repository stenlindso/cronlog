package store

import (
	"database/sql"
	"fmt"
	"time"
)

// Annotation holds a user-supplied note attached to a run.
type Annotation struct {
	ID        int64
	RunID     int64
	Note      string
	CreatedAt time.Time
}

// AddAnnotation attaches a note to the run identified by runID.
// Returns an error if the run does not exist.
func AddAnnotation(db *sql.DB, runID int64, note string) error {
	var exists int
	err := db.QueryRow(`SELECT COUNT(*) FROM runs WHERE id = ?`, runID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("annotate: check run: %w", err)
	}
	if exists == 0 {
		return fmt.Errorf("annotate: run %d not found", runID)
	}

	_, err = db.Exec(
		`INSERT INTO annotations (run_id, note, created_at) VALUES (?, ?, ?)`,
		runID, note, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("annotate: insert: %w", err)
	}
	return nil
}

// GetAnnotations returns all annotations for the given run, oldest first.
func GetAnnotations(db *sql.DB, runID int64) ([]Annotation, error) {
	rows, err := db.Query(
		`SELECT id, run_id, note, created_at FROM annotations WHERE run_id = ? ORDER BY created_at ASC`,
		runID,
	)
	if err != nil {
		return nil, fmt.Errorf("annotate: query: %w", err)
	}
	defer rows.Close()

	var annotations []Annotation
	for rows.Next() {
		var a Annotation
		if err := rows.Scan(&a.ID, &a.RunID, &a.Note, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("annotate: scan: %w", err)
		}
		annotations = append(annotations, a)
	}
	return annotations, rows.Err()
}
