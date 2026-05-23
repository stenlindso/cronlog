package runner_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronlog/internal/runner"
)

func TestRunSuccess(t *testing.T) {
	res := runner.Run(context.Background(), "echo hello")

	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Output, "hello") {
		t.Fatalf("expected output to contain 'hello', got %q", res.Output)
	}
	if res.Duration <= 0 {
		t.Fatal("expected positive duration")
	}
	if res.Started.IsZero() {
		t.Fatal("expected non-zero start time")
	}
	if res.Command != "echo hello" {
		t.Fatalf("expected command 'echo hello', got %q", res.Command)
	}
}

func TestRunFailure(t *testing.T) {
	res := runner.Run(context.Background(), "exit 42")

	if res.ExitCode != 42 {
		t.Fatalf("expected exit code 42, got %d", res.ExitCode)
	}
}

func TestRunCapturesStderr(t *testing.T) {
	res := runner.Run(context.Background(), "echo err >&2")

	if !strings.Contains(res.Output, "err") {
		t.Fatalf("expected stderr in output, got %q", res.Output)
	}
}

func TestRunCancelledContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := runner.Run(ctx, "sleep 10")

	if res.ExitCode == 0 {
		t.Fatal("expected non-zero exit code for cancelled command")
	}
}
