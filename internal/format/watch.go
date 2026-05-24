package format

import (
	"fmt"
	"strings"
	"time"

	"github.com/example/cronlog/internal/store"
)

const (
	watchCmdWidth = 30
	statusOK      = "OK  "
	statusFail    = "FAIL"
)

// PrintWatchLine prints a single watch result in a compact, human-readable
// format suitable for tailing output.
//
// Example:
//
//	2024-05-01 14:32:01  OK    backup.sh           0.42s
//	2024-05-01 14:33:10  FAIL  sync.sh             exit=2  1.10s
func PrintWatchLine(r store.WatchResult) {
	status := statOK
	if r.ExitCode != 0 {
		status = statusFail
	}

	cmd := r.Command
	if len(cmd) > watchCmdWidth {
		cmd = cmd[:watchCmdWidth-1] + "…"
	}

	dur := r.FinishedAt.Sub(r.StartedAt)
	ts := r.FinishedAt.UTC().Format(time.DateTime)

	var extra string
	if r.ExitCode != 0 {
		extra = fmt.Sprintf(" exit=%-3d", r.ExitCode)
	}

	tagStr := ""
	if len(r.Tags) > 0 {
		tagStr = " [" + strings.Join(r.Tags, ",") + "]"
	}

	fmt.Printf("%s  %-4s  %-*s%s  %s%s\n",
		ts, status, watchCmdWidth, cmd, extra, FormatDuration(dur), tagStr)
}

const statOK = "OK  "
