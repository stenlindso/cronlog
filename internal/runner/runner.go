package runner

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

// Result holds the outcome of a single job execution.
type Result struct {
	Command  string
	Output   string
	ExitCode int
	Started  time.Time
	Duration time.Duration
}

// Run executes the given shell command and captures its combined output,
// exit code, and wall-clock duration.
func Run(ctx context.Context, command string) Result {
	started := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	duration := time.Since(started)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit error (e.g. command not found); treat as exit code 1.
			exitCode = 1
		}
	}

	return Result{
		Command:  command,
		Output:   buf.String(),
		ExitCode: exitCode,
		Started:  started,
		Duration: duration,
	}
}
