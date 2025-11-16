package cmd_test

import (
	"bytes"
	"testing"

	"github.com/kapetan-io/claude-md.go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunWithInvalidCommand(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cmd.Run([]string{"invalid-command"}, cmd.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "unknown command")
}

func TestRunWithHelpFlag(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cmd.Run([]string{"--help"}, cmd.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "claude-md")
	assert.Contains(t, stdout.String(), "CLI tool for managing CLAUDE.md files")
}

func TestRunWithNoArgs(t *testing.T) {
	var stdout, stderr bytes.Buffer

	exitCode := cmd.Run([]string{}, cmd.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "claude-md")
}

func TestRunWithDefaultWriters(t *testing.T) {
	exitCode := cmd.Run([]string{"--help"}, cmd.RunOptions{})

	require.Equal(t, 0, exitCode)
}
