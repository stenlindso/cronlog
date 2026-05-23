// cronlog is a lightweight wrapper for cron jobs that captures output,
// duration, and exit codes to a local SQLite store.
//
// Usage:
//
//	cronlog [flags] -- <command> [args...]
//
// Flags:
//
//	-db string
//		Path to the SQLite database file (default: ~/.cronlog/cronlog.db)
//	-name string
//		Human-readable label for this job (defaults to the command)
//	-list
//		List recent job runs and exit
//	-n int
//		Number of recent runs to show when using -list (default 20)
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/user/cronlog/internal/runner"
	"github.com/user/cronlog/internal/store"
)

func defaultDBPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "cronlog.db"
	}
	return filepath.Join(home, ".cronlog", "cronlog.db")
}

func main() {
	dbPath := flag.String("db", defaultDBPath(), "path to SQLite database file")
	jobName := flag.String("name", "", "human-readable label for this job")
	listMode := flag.Bool("list", false, "list recent job runs and exit")
	listN := flag.Int("n", 20, "number of recent runs to show with -list")

	flag.Parse()

	// Ensure the directory for the database exists.
	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "cronlog: failed to create db directory: %v\n", err)
		os.Exit(1)
	}

	db, err := store.Open(*dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cronlog: failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if *listMode {
		runList(db, *listN)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "cronlog: no command specified")
		fmt.Fprintln(os.Stderr, "usage: cronlog [flags] -- <command> [args...]")
		os.Exit(1)
	}

	name := *jobName
	if name == "" {
		name = strings.Join(args, " ")
	}

	// Propagate OS signals so we can cancel the child process cleanly.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	result := runner.Run(ctx, args[0], args[1:]...)

	if err := db.Insert(store.Run{
		Name:      name,
		Command:   strings.Join(args, " "),
		StartedAt: result.StartedAt,
		Duration:  result.Duration,
		ExitCode:  result.ExitCode,
		Stdout:    result.Stdout,
		Stderr:    result.Stderr,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "cronlog: failed to record run: %v\n", err)
	}

	// Mirror the child's output so cron mailers still work.
	if result.Stdout != "" {
		fmt.Print(result.Stdout)
	}
	if result.Stderr != "" {
		fmt.Fprint(os.Stderr, result.Stderr)
	}

	os.Exit(result.ExitCode)
}

// runList prints the N most recent job runs in a human-readable table.
func runList(db *store.DB, n int) {
	runs, err := db.List(n)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cronlog: failed to list runs: %v\n", err)
		os.Exit(1)
	}

	if len(runs) == 0 {
		fmt.Println("No runs recorded yet.")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTARTED\tDURATION\tEXIT\tNAME")
	for _, r := range runs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%d\t%s\n",
			r.ID,
			r.StartedAt.Format(time.DateTime),
			r.Duration.Round(time.Millisecond),
			r.ExitCode,
			r.Name,
		)
	}
	w.Flush()
}
