package cli

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testCaseRun struct {
	name string
	args []string
}

var dispatchBinary = filepath.Join("../build", runtime.GOOS, runtime.GOARCH, "dispatch")

func TestRunCommand(t *testing.T) {
	// Cobra unit tests
	tcs := []struct {
		in  testCaseRun
		out expectedOutput
	}{
		{
			in: testCaseRun{
				name: "Run without arguments",
			},
			out: expectedOutput{
				stderr: "Error: requires at least 1 arg(s), only received 0\n",
			},
		},
		{
			in: testCaseRun{
				name: "Run properly",
				args: []string{"--", "echo", "42"},
			},
			out: expectedOutput{
				stderr: "Error: command 'echo 42' exited unexpectedly\n",
			},
		},
		{
			in: testCaseRun{
				name: "Run with invalid command",
				args: []string{"--", "invalid-command"},
			},
			out: expectedOutput{
				stderr: "Error: failed to start invalid-command: exec: \"invalid-command\": executable file not found in $PATH\n",
			},
		},
		{
			in: testCaseRun{
				name: "Run with invalid flag",
				args: []string{"--invalid-flag", "--", "echo", "42"},
			},
			out: expectedOutput{
				stderr: "Error: unknown flag: --invalid-flag\n",
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.in.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			cmd := runCommand()
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			cmd.SetArgs(tc.in.args)

			if err := cmd.Execute(); err != nil {
				t.Logf("Received unexpected error: %v", err)
			}

			assert.Equal(t, tc.out.stdout, stdout.String())
			assert.Equal(t, tc.out.stderr, stderr.String())
		})
	}

	// Integration tests
	t.Run("Run with non-existent env file", func(t *testing.T) {
		t.Parallel()

		// Create a context with a timeout to ensure the process doesn't run indefinitely
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Set up the command
		cmd := exec.CommandContext(ctx, dispatchBinary, "run", "--env-file", "non-existent.env", "--", "echo", "hello")

		// Capture the standard error
		var errBuf bytes.Buffer
		cmd.Stderr = &errBuf

		// Start the command
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start command: %v", err)
		}

		// Wait for the command to finish or for the context to timeout
		if err := cmd.Wait(); err != nil {
			// Check if the error is due to context timeout (command running too long)
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Command timed out")
			}
		}

		assert.Regexp(t, "Error: failed to load env file from .+: open .+: no such file or directory\n", errBuf.String())
	})
	t.Run("Run with env file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		envFile := filepath.Join(tempDir, "test.env")
		err := os.WriteFile(envFile, []byte("DISPATCH_API_URL=test"), 0600)
		if err != nil {
			t.Fatalf("Failed to write env file: %v", err)
		}

		// Create a context with a timeout to ensure the process doesn't run indefinitely
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		// Set up the command
		cmd := exec.CommandContext(ctx, dispatchBinary, "run", "--env-file", envFile, "--", "printenv", "DISPATCH_API_URL")

		// Capture the standard output and error
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf

		// Start the command
		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start command: %v", err)
		}

		// Wait for the command to finish or for the context to timeout
		if err := cmd.Wait(); err != nil {
			// Check if the error is due to context timeout (command running too long)
			if ctx.Err() == context.DeadlineExceeded {
				t.Fatalf("Command timed out")
			}
		}

		found := false
		// Split the log into lines
		lines := strings.Split(errBuf.String(), "\n")
		// Iterate over each line and check for the condition
		for _, line := range lines {
			if strings.Contains(line, "printenv | test") {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected 'printenv | test' in the output")
	})
}
