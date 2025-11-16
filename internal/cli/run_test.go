package cli_test

import (
	"bytes"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWithInvalidCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cli.Run([]string{"invalid-command"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "unknown command")
}

func TestRunWithHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cli.Run([]string{"--help"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "claude-md")
	assert.Contains(t, stdout.String(), "CLI tool for managing CLAUDE.md files")
}

func TestRunWithNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cli.Run([]string{}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "claude-md")
}

func TestRunWithDefaultWriters(t *testing.T) {
	exitCode := cli.Run([]string{"--help"}, cli.RunOptions{})

	require.Equal(t, 0, exitCode)
}
