package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/cronlog/internal/store"
)

func runExport(args []string, dbPath string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	command := fs.String("command", "", "filter by command string")
	limit := fs.Int("limit", 0, "maximum number of runs to export (0 = all)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if *command != "" {
		return store.ExportJSONByCommand(db, os.Stdout, *command, *limit)
	}
	return store.ExportJSON(db, os.Stdout, *limit)
}
