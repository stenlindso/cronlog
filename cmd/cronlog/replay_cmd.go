package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/user/cronlog/internal/runner"
	"github.com/user/cronlog/internal/store"
)

// runReplay re-executes the most recent matching run's command.
func runReplay(db *sql.DB, args []string) error {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	command := fs.String("command", "", "filter by command name (optional)")
	dryRun := fs.Bool("dry-run", false, "print command without executing")

	if err := fs.Parse(args); err != nil {
		return err
	}

	run, err := store.GetLastRun(db, *command)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}

	full := run.Command
	if run.Args != "" {
		full = run.Command + " " + run.Args
	}

	if *dryRun {
		fmt.Fprintf(os.Stdout, "would replay: %s\n", full)
		return nil
	}

	fmt.Fprintf(os.Stdout, "replaying: %s\n", full)

	parts := strings.Fields(full)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	result, err := runner.Run(context.Background(), parts[0], parts[1:])
	if err != nil {
		return fmt.Errorf("run failed: %w", err)
	}

	if _, err := store.Insert(db, store.Run{
		Command:    result.Command,
		Args:       strings.Join(result.Args, " "),
		ExitCode:   result.ExitCode,
		Stdout:     result.Stdout,
		Stderr:     result.Stderr,
		StartedAt:  result.StartedAt,
		DurationMs: result.DurationMs,
	}); err != nil {
		return fmt.Errorf("store replay result: %w", err)
	}

	fmt.Fprintf(os.Stdout, "exit code: %d\n", result.ExitCode)
	if result.Stdout != "" {
		fmt.Fprint(os.Stdout, result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}
	return nil
}
