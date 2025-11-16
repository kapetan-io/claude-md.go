package cli_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitCommand(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	storageDir := home + "/.claude/claude-md/test/repo.git"
	os.RemoveAll(storageDir)

	var stdout, stderr bytes.Buffer

	exitCode := cli.Run([]string{"init"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 0, exitCode)
	assert.Empty(t, stderr.String())
	assert.Contains(t, stdout.String(), "Created storage directory")
	assert.Contains(t, stdout.String(), "User: test")
}

func TestInitCommandAlreadyExists(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	stdout.Reset()
	exitCode = cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Storage directory already exists")
	assert.Contains(t, stdout.String(), "User: test")
}

func TestInitCommandErrorCases(t *testing.T) {
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer

	exitCode := cli.Run([]string{"init"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Error")
}
