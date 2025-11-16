package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveCommand(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	storageDir := filepath.Join(home, ".claude", "claude-md", "test", "repo.git")
	os.RemoveAll(storageDir)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cli.Run([]string{"save"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Saved: CLAUDE.md")
	assert.Contains(t, stdout.String(), "Summary: 1 saved, 0 skipped, 0 errors")

	info, err := os.Lstat(claudeFile)
	require.NoError(t, err)
	assert.NotEqual(t, 0, info.Mode()&os.ModeSymlink)
}

func TestSaveCommandNoFiles(t *testing.T) {
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
	exitCode = cli.Run([]string{"save"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "No CLAUDE.md files found in repository")
}

func TestSaveCommandAlreadySymlink(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cli.Run([]string{"save"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	stdout.Reset()
	exitCode = cli.Run([]string{"save"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Summary: 0 saved, 1 skipped, 0 errors")
}

func TestSaveCommandErrorCases(t *testing.T) {
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := cli.Run([]string{"save"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Error")
}
