package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRestoreCommand(t *testing.T) {
	repoDir := cmd.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cmd.Run([]string{"save"}, cmd.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	err = os.Remove(claudeFile)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cmd.Run([]string{"restore"}, cmd.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Restored: CLAUDE.md")
	assert.Contains(t, stdout.String(), "Summary: 1 restored, 0 skipped")

	info, err := os.Lstat(claudeFile)
	require.NoError(t, err)
	assert.NotEqual(t, 0, info.Mode()&os.ModeSymlink)
}

func TestRestoreCommandNoStoredFiles(t *testing.T) {
	repoDir := cmd.SetupTestGitRepo(t)

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
	exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	stdout.Reset()
	exitCode = cmd.Run([]string{"restore"}, cmd.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "No stored CLAUDE.md files found for this repository")
}

func TestRestoreCommandAlreadyExists(t *testing.T) {
	repoDir := cmd.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cmd.Run([]string{"save"}, cmd.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	stdout.Reset()
	exitCode = cmd.Run([]string{"restore"}, cmd.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Summary: 0 restored, 1 skipped")
}

func TestRestoreCommandErrorCases(t *testing.T) {
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := cmd.Run([]string{"restore"}, cmd.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Error")
}
