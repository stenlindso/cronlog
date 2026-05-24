package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/example/cronlog/internal/store"
)

func runSummary(db *store.DB, args []string) error {
	fs := flag.NewFlagSet("summary", flag.ContinueOnError)
	days := fs.Int("days", 7, "number of past days to summarise")
	limit := fs.Int("limit", 30, "maximum number of day rows to return")
	cmd := fs.String("command", "", "filter by command")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *days <= 0 {
		return fmt.Errorf("--days must be a positive integer")
	}

	since := time.Now().UTC().AddDate(0, 0, -*days)

	summaries, err := store.DailySummaries(db.DB, store.DailySummaryOptions{
		Since:   since,
		Command: *cmd,
		Limit:   *limit,
	})
	if err != nil {
		return fmt.Errorf("summary query: %w", err)
	}

	if len(summaries) == 0 {
		fmt.Println("no runs found for the specified period")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "DATE\tTOTAL\tSUCCESS\tFAILURE\tAVG_DUR\tCOMMANDS")
	for _, s := range summaries {
		avg := fmt.Sprintf("%.0fms", s.AvgDurMs)
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%s\t%d\n",
			s.Date, s.TotalRuns, s.Successes, s.Failures, avg, s.Commands)
	}
	return w.Flush()
}
