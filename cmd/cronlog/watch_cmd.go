package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/example/cronlog/internal/format"
	"github.com/example/cronlog/internal/store"
)

// runWatch polls the database for new runs and prints them as they appear.
// Usage: cronlog watch [--command CMD] [--interval N]
func runWatch(db *store.DB, args []string) error {
	fs := flag.NewFlagSet("watch", flag.ContinueOnError)
	command := fs.String("command", "", "filter by command name")
	interval := fs.Int("interval", 5, "polling interval in seconds")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *interval < 1 {
		return fmt.Errorf("--interval must be >= 1")
	}

	fmt.Fprintf(os.Stderr, "Watching for new runs (interval: %ds, Ctrl-C to stop)...\n", *interval)

	since := time.Now().UTC()
	tick := time.NewTicker(time.Duration(*interval) * time.Second)
	defer tick.Stop()

	for range tick.C {
		results, err := store.Watch(db, store.WatchOptions{
			Command: *command,
			Since:   since,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
			continue
		}
		if len(results) == 0 {
			continue
		}
		for _, r := range results {
			format.PrintWatchLine(r)
		}
		// Advance cursor so we don't re-print the same runs.
		since = results[len(results)-1].FinishedAt
	}
	return nil
}
