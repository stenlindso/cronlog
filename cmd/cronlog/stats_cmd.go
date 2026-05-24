package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/user/cronlog/internal/format"
	"github.com/user/cronlog/internal/store"
)

// runStats handles the "stats" subcommand.
// Usage: cronlog stats [--command <cmd>]
//
// Prints aggregate statistics for all logged cron jobs, optionally filtered
// to a specific command string. Output includes total runs, success/failure
// counts, and average duration.
func runStats(args []string, dbPath string) int {
	fs := flag.NewFlagSet("stats", flag.ContinueOnError)
	command := fs.String("command", "", "filter stats to a specific command")

	if err := fs.Parse(args); err != nil {
		fmt.Fprintf(os.Stderr, "stats: %v\n", err)
		return 2
	}

	if fs.NArg() > 0 {
		fmt.Fprintf(os.Stderr, "stats: unexpected arguments: %v\n", fs.Args())
		return 2
	}

	db, err := store.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open db: %v\n", err)
		return 1
	}
	defer db.Close()

	stats, err := store.Stats(db.DB, *command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "stats query: %v\n", err)
		return 1
	}

	format.PrintStatsTable(os.Stdout, stats)
	return 0
}
