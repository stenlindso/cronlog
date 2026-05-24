package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/example/cronlog/internal/store"
)

func runDigest(args []string, db *store.DB) error {
	fs := flag.NewFlagSet("digest", flag.ContinueOnError)
	command := fs.String("command", "", "filter digest by command")
	window := fs.Int("days", 7, "number of days to include in digest")

	if err := fs.Parse(args); err != nil {
		return err
	}

	since := time.Now().UTC().AddDate(0, 0, -*window)

	entries, err := store.GetDigest(db, *command, since)
	if err != nil {
		return fmt.Errorf("digest: %w", err)
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stdout, "no data in the requested window")
		return nil
	}

	fmt.Fprintf(os.Stdout, "%-40s %8s %8s %8s %10s %10s\n",
		"COMMAND", "RUNS", "SUCCESS", "FAILURE", "AVG_DUR", "LAST_RUN")
	fmt.Fprintf(os.Stdout, "%s\n", repeatChar('-', 90))

	for _, e := range entries {
		lastRun := e.LastRun.Format("2006-01-02")
		avgDur := time.Duration(e.AvgDurationMs) * time.Millisecond
		cmd := e.Command
		if len(cmd) > 38 {
			cmd = cmd[:35] + "..."
		}
		fmt.Fprintf(os.Stdout, "%-40s %8d %8d %8d %10s %10s\n",
			cmd, e.TotalRuns, e.SuccessCount, e.FailureCount,
			avgDur.Round(time.Millisecond).String(), lastRun)
	}

	return nil
}

func repeatChar(c rune, n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}
