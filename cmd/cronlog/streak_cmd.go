package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/user/cronlog/internal/store"
)

func runStreak(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("streak", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: cronlog streak --command <cmd>")
		fs.PrintDefaults()
	}

	var command string
	fs.StringVar(&command, "command", "", "command to query streak for (required)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if command == "" {
		fs.Usage()
		return fmt.Errorf("--command is required")
	}

	s, err := store.GetStreak(db, command)
	if err == sql.ErrNoRows {
		fmt.Printf("No runs found for command: %s\n", command)
		return nil
	}
	if err != nil {
		return fmt.Errorf("streak query failed: %w", err)
	}

	fmt.Printf("Command:         %s\n", s.Command)
	fmt.Printf("Current Streak:  %d %s\n", s.CurrentStreak, s.StreakType)
	fmt.Printf("Longest Success: %d\n", s.LongestSuccess)
	fmt.Printf("Longest Failure: %d\n", s.LongestFailure)
	fmt.Printf("Last Run At:     %s\n", s.LastRunAt.Format("2006-01-02 15:04:05"))
	return nil
}
