// Package format provides display helpers for cronlog output.
package format

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/user/cronlog/internal/store"
)

const (
	colID      = 6
	colCommand = 30
	colStarted = 20
	colDur     = 10
	colExit    = 5
)

// PrintTable writes a formatted table of runs to w.
func PrintTable(w io.Writer, runs []store.Run) {
	header := fmt.Sprintf("%-*s  %-*s  %-*s  %-*s  %s",
		colID, "ID",
		colCommand, "COMMAND",
		colStarted, "STARTED",
		colDur, "DURATION",
		"EXIT",
	)
	fmt.Fprintln(w, header)
	fmt.Fprintln(w, strings.Repeat("-", len(header)+4))

	for _, r := range runs {
		cmd := r.Command
		if len(cmd) > colCommand {
			cmd = cmd[:colCommand-1] + "…"
		}
		fmt.Fprintf(w, "%-*d  %-*s  %-*s  %-*s  %d\n",
			colID, r.ID,
			colCommand, cmd,
			colStarted, r.StartedAt.Local().Format(time.DateTime),
			colDur, FormatDuration(r.Duration),
			r.ExitCode,
		)
	}
}

// PrintDetail writes full details of a single run to w.
func PrintDetail(w io.Writer, r *store.Run) {
	fmt.Fprintf(w, "ID:       %d\n", r.ID)
	fmt.Fprintf(w, "Command:  %s\n", r.Command)
	fmt.Fprintf(w, "Started:  %s\n", r.StartedAt.Local().Format(time.RFC1123))
	fmt.Fprintf(w, "Duration: %s\n", FormatDuration(r.Duration))
	fmt.Fprintf(w, "Exit:     %d\n", r.ExitCode)
	if r.Stdout != "" {
		fmt.Fprintf(w, "--- stdout ---\n%s\n", r.Stdout)
	}
	if r.Stderr != "" {
		fmt.Fprintf(w, "--- stderr ---\n%s\n", r.Stderr)
	}
}

// FormatDuration renders a duration in a compact human-readable form.
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}
