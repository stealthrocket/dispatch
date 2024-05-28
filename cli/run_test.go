package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCaseRun struct {
	name string
	args []string
}

func TestMainCommand(t *testing.T) {
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
		// {
		// 	in: testCaseRun{
		// 		name: "Run with non-existent env file",
		// 		args: []string{"--env-file", "non-existent.env", "--", "echo", "hello"},
		// 	},
		// 	out: expectedOutput{
		// 		stderr: "what\n",
		// 	},
		// },
		// {
		// 	in: testCaseRun{
		// 		name: "Run with env file",
		// 		args: []string{"--env-file", "test.env", "--", "printenv", "DISPATCH_API_URL"},
		// 	},
		// 	out: expectedOutput{
		// 		stdout: "test\n",
		// 	},
		// },
	}

	runCmd, _, err := mainCommand().Find([]string{"run"})
	if err != nil {
		t.Fatalf("Received unexpected error: %v", err)
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.in.name, func(t *testing.T) {
			t.Parallel()

			// runCmd := runCommand()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			runCmd.SetOut(stdout)
			runCmd.SetErr(stderr)
			runCmd.SetArgs(tc.in.args)

			if err := runCmd.Execute(); err != nil {
				t.Logf("Received unexpected error: %v", err)
			}

			assert.Equal(t, tc.out.stdout, stdout.String())
			assert.Equal(t, tc.out.stderr, stderr.String())
		})
	}
}
