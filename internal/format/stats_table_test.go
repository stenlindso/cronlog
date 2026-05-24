package format_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/user/cronlog/internal/format"
	"github.com/user/cronlog/internal/store"
)

func TestPrintStatsTableHeaders(t *testing.T) {
	var buf bytes.Buffer
	stats := []store.CommandStats{
		{Command: "backup.sh", TotalRuns: 5, SuccessCount: 4, FailureCount: 1, AvgDurationMs: 1000, LastRunAt: "2024-01-01 12:00:00"},
	}
	format.PrintStatsTable(&buf, stats)
	out := buf.String()
	for _, hdr := range []string{"COMMAND", "TOTAL", "SUCCESS", "FAILURE", "AVG DURATION", "LAST RUN"} {
		if !strings.Contains(out, hdr) {
			t.Errorf("expected header %q in output", hdr)
		}
	}
}

func TestPrintStatsTableValues(t *testing.T) {
	var buf bytes.Buffer
	stats := []store.CommandStats{
		{Command: "sync.sh", TotalRuns: 10, SuccessCount: 8, FailureCount: 2, AvgDurationMs: 2500, LastRunAt: "2024-06-15 08:30:00"},
	}
	format.PrintStatsTable(&buf, stats)
	out := buf.String()
	if !strings.Contains(out, "sync.sh") {
		t.Error("expected command name in output")
	}
	if !strings.Contains(out, "80.0%") {
		t.Errorf("expected success rate 80.0%% in output, got:\n%s", out)
	}
}

func TestPrintStatsTableEmpty(t *testing.T) {
	var buf bytes.Buffer
	format.PrintStatsTable(&buf, nil)
	if !strings.Contains(buf.String(), "No stats available") {
		t.Error("expected empty message")
	}
}

func TestPrintStatsTableTruncatesLongCommand(t *testing.T) {
	var buf bytes.Buffer
	long := strings.Repeat("a", 50)
	stats := []store.CommandStats{
		{Command: long, TotalRuns: 1, SuccessCount: 1, AvgDurationMs: 100, LastRunAt: "2024-01-01"},
	}
	format.PrintStatsTable(&buf, stats)
	if strings.Contains(buf.String(), long) {
		t.Error("expected long command to be truncated")
	}
}
