package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/user/cronlog/internal/store"
)

func runSnapshot(args []string, dbPath string) error {
	fs := flag.NewFlagSet("snapshot", flag.ContinueOnError)
	command := fs.String("command", "", "command to snapshot or list snapshots for (required)")
	take := fs.Bool("take", false, "take a new snapshot now")
	limit := fs.Int("limit", 10, "max snapshots to list")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *command == "" {
		fs.Usage()
		return fmt.Errorf("--command is required")
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if *take {
		snap, err := store.TakeSnapshot(db, *command)
		if err != nil {
			return fmt.Errorf("take snapshot: %w", err)
		}
		fmt.Printf("Snapshot #%d recorded at %s\n",
			snap.ID, snap.CapturedAt.Format(time.RFC3339))
		fmt.Printf("  Total runs   : %d\n", snap.TotalRuns)
		fmt.Printf("  Success rate : %.1f%%\n", snap.SuccessRate)
		fmt.Printf("  Avg duration : %.3fs\n", snap.AvgDuration)
		fmt.Printf("  Last exit    : %d\n", snap.LastExitCode)
		return nil
	}

	snaps, err := store.ListSnapshots(db, *command, *limit)
	if err != nil {
		return fmt.Errorf("list snapshots: %w", err)
	}
	if len(snaps) == 0 {
		fmt.Println("No snapshots found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCAPTURED AT\tTOTAL\tSUCCESS%\tAVG DUR\tLAST EXIT")
	for _, s := range snaps {
		fmt.Fprintf(w, "%d\t%s\t%d\t%.1f\t%.3fs\t%d\n",
			s.ID,
			s.CapturedAt.Format("2006-01-02 15:04:05"),
			s.TotalRuns,
			s.SuccessRate,
			s.AvgDuration,
			s.LastExitCode,
		)
	}
	return w.Flush()
}
