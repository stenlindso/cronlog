package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/user/cronlog/internal/store"
)

func runBaseline(args []string, dbPath string) error {
	fs := flag.NewFlagSet("baseline", flag.ContinueOnError)
	command := fs.String("command", "", "command to compute baseline for")
	samples := fs.Int("samples", 10, "number of recent successful runs to average")
	getFlag := fs.Bool("get", false, "retrieve stored baseline instead of computing")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if *command == "" {
		return fmt.Errorf("--command is required")
	}

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	if *getFlag {
		b, err := store.GetBaseline(db, *command)
		if err != nil {
			return err
		}
		if b == nil {
			fmt.Fprintf(os.Stderr, "no baseline found for %q\n", *command)
			return nil
		}
		printBaseline(b)
		return nil
	}

	b, err := store.UpsertBaseline(db, *command, *samples)
	if err != nil {
		return err
	}
	fmt.Printf("Baseline updated for %q\n", *command)
	printBaseline(b)
	return nil
}

func printBaseline(b *store.Baseline) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "COMMAND\tAVG DURATION\tSAMPLES\tUPDATED")
	fmt.Fprintf(w, "%s\t%.3fs\t%d\t%s\n",
		b.Command,
		b.AvgDuration,
		b.SampleCount,
		b.UpdatedAt.Format("2006-01-02 15:04:05"),
	)
	w.Flush()
}
