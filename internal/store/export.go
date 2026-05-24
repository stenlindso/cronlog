package store

import (
	"database/sql"
	"encoding/json"
	"io"
)

// ExportJSON writes all runs (up to limit, 0 = all) as a JSON array to w.
func ExportJSON(db *sql.DB, w io.Writer, limit int) error {
	runs, err := List(db, limit)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(runs)
}

// ExportJSONByCommand writes runs filtered by command as a JSON array to w.
func ExportJSONByCommand(db *sql.DB, w io.Writer, command string, limit int) error {
	allRuns, err := List(db, 0)
	if err != nil {
		return err
	}

	var filtered []Run
	for _, r := range allRuns {
		if r.Command == command {
			filtered = append(filtered, r)
			if limit > 0 && len(filtered) >= limit {
				break
			}
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(filtered)
}
