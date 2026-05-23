package format

import (
	"strings"
	"testing"
	"time"

	"github.com/user/cronlog/internal/store"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.50s"},
		{90 * time.Second, "1m30s"},
		{2*time.Minute + 5*time.Second, "2m5s"},
	}
	for _, tc := range cases {
		got := FormatDuration(tc.d)
		if got != tc.want {
			t.Errorf("FormatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestPrintTableContainsHeaders(t *testing.T) {
	var sb strings.Builder
	runs := []store.Run{
		{
			ID:        1,
			Command:   "echo hi",
			StartedAt: time.Now(),
			Duration:  42 * time.Millisecond,
			ExitCode:  0,
		},
	}
	PrintTable(&sb, runs)
	out := sb.String()

	for _, want := range []string{"ID", "COMMAND", "STARTED", "DURATION", "EXIT", "echo hi"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestPrintDetailShowsAllFields(t *testing.T) {
	var sb strings.Builder
	r := &store.Run{
		ID:        7,
		Command:   "backup.sh",
		StartedAt: time.Now(),
		Duration:  3 * time.Second,
		ExitCode:  1,
		Stdout:    "done\n",
		Stderr:    "warning: low disk\n",
	}
	PrintDetail(&sb, r)
	out := sb.String()

	for _, want := range []string{"backup.sh", "3.00s", "Exit:     1", "done", "warning: low disk"} {
		if !strings.Contains(out, want) {
			t.Errorf("PrintDetail output missing %q\nfull output:\n%s", want, out)
		}
	}
}

func TestPrintTableTruncatesLongCommand(t *testing.T) {
	var sb strings.Builder
	long := strings.Repeat("a", 50)
	runs := []store.Run{{ID: 1, Command: long, StartedAt: time.Now()}}
	PrintTable(&sb, runs)
	out := sb.String()
	if strings.Contains(out, long) {
		t.Error("expected long command to be truncated in table output")
	}
}
