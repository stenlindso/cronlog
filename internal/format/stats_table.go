package format

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/user/cronlog/internal/store"
)

// PrintStatsTable writes a formatted statistics table to w.
func PrintStatsTable(w io.Writer, stats []store.CommandStats) {
	if len(stats) == 0 {
		fmt.Fprintln(w, "No stats available.")
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "COMMAND\tTOTAL\tSUCCESS\tFAILURE\tSUCCESS%\tAVG DURATION\tLAST RUN")
	fmt.Fprintln(tw, "-------\t-----\t-------\t-------\t--------\t------------\t--------")

	for _, s := range stats {
		successPct := 0.0
		if s.TotalRuns > 0 {
			successPct = float64(s.SuccessCount) / float64(s.TotalRuns) * 100
		}
		avgDur := FormatDuration(int64(s.AvgDurationMs))
		cmd := truncate(s.Command, 40)
		fmt.Fprintf(tw, "%s\t%d\t%d\t%d\t%.1f%%\t%s\t%s\n",
			cmd,
			s.TotalRuns,
			s.SuccessCount,
			s.FailureCount,
			successPct,
			avgDur,
			s.LastRunAt,
		)
	}
	tw.Flush()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
