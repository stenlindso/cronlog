package main

import (
	"database/sql"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/user/cronlog/internal/format"
	"github.com/user/cronlog/internal/store"
)

// runShow handles the "show <id>" subcommand, printing detail for a single run.
func runShow(db *sql.DB, args []string, out io.Writer) error {
	fs := flag.NewFlagSet("show", flag.ContinueOnError)
	fs.SetOutput(out)

	if err := fs.Parse(args); err != nil {
		return err
	}

	if fs.NArg() == 0 {
		fmt.Fprintln(out, "usage: cronlog show <id>")
		return fmt.Errorf("missing run ID")
	}

	idStr := fs.Arg(0)
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID %q: %w", idStr, err)
	}

	run, err := store.GetByID(db, id)
	if err != nil {
		return fmt.Errorf("get run %d: %w", id, err)
	}

	format.PrintDetail(out, run)
	return nil
}

// registerShowCmd wires the show subcommand into the main dispatch.
// Called from main() after flag parsing.
func registerShowCmd(db *sql.DB, args []string) {
	if err := runShow(db, args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "cronlog show: %v\n", err)
		os.Exit(1)
	}
}
