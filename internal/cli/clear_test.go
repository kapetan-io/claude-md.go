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

func TestClearCommand(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	storageDir := filepath.Join(home, ".claude", "claude-md", "test", "repo.git")
	_ = os.RemoveAll(storageDir)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	var saveStderr bytes.Buffer
	exitCode = cli.Run([]string{"save"}, cli.RunOptions{Stdout: &stdout, Stderr: &saveStderr})
	require.Equal(t, 0, exitCode, "save stdout: %s, stderr: %s", stdout.String(), saveStderr.String())

	info, err := os.Lstat(claudeFile)
	require.NoError(t, err)
	require.NotEqual(t, 0, info.Mode()&os.ModeSymlink, "file should be a symlink after save")

	stdout.Reset()
	exitCode = cli.Run([]string{"clear"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "Removed: CLAUDE.md")
	assert.Contains(t, stdout.String(), "Summary: 1 removed, 0 skipped, 0 errors")
	assert.Contains(t, stdout.String(), "Note: Stored files remain in storage")

	_, err = os.Lstat(claudeFile)
	assert.True(t, os.IsNotExist(err))
}

func TestClearCommandNoSymlinks(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	stdout.Reset()
	exitCode = cli.Run([]string{"clear"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "No CLAUDE.md symlinks found in repository")
}

func TestClearCommandSkipsRegularFiles(t *testing.T) {
	repoDir := cli.SetupTestGitRepo(t)

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	storageDir := filepath.Join(home, ".claude", "claude-md", "test", "repo.git")
	_ = os.RemoveAll(storageDir)

	var stdout bytes.Buffer
	exitCode := cli.Run([]string{"init"}, cli.RunOptions{Stdout: &stdout})
	require.Equal(t, 0, exitCode)

	claudeFile := filepath.Join(repoDir, "CLAUDE.md")
	err = os.WriteFile(claudeFile, []byte("test content"), 0644)
	require.NoError(t, err)

	stdout.Reset()
	exitCode = cli.Run([]string{"clear"}, cli.RunOptions{Stdout: &stdout})

	require.Equal(t, 0, exitCode)
	assert.Contains(t, stdout.String(), "No CLAUDE.md symlinks found in repository")

	_, err = os.Stat(claudeFile)
	assert.NoError(t, err)
}

func TestClearCommandErrorCases(t *testing.T) {
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(oldDir) }()

	tmpDir := t.TempDir()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	var stdout, stderr bytes.Buffer
	exitCode := cli.Run([]string{"clear"}, cli.RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	require.Equal(t, 1, exitCode)
	assert.Contains(t, stderr.String(), "Error")
}
