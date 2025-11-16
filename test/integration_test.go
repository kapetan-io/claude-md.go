package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullWorkflow(t *testing.T) {
	// Create temporary directory for test repo
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "test-repo")

	// Initialize git repository
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	// Initialize git
	gitCmd := exec.Command("git", "init")
	gitCmd.Dir = repoDir
	require.NoError(t, gitCmd.Run())

	// Configure git user
	gitCmd = exec.Command("git", "config", "user.email", "test@example.com")
	gitCmd.Dir = repoDir
	require.NoError(t, gitCmd.Run())

	gitCmd = exec.Command("git", "config", "user.name", "Test User")
	gitCmd.Dir = repoDir
	require.NoError(t, gitCmd.Run())

	// Add origin remote
	gitCmd = exec.Command("git", "remote", "add", "origin", "https://github.com/test/repo.git")
	gitCmd.Dir = repoDir
	require.NoError(t, gitCmd.Run())

	// Create test CLAUDE.md files
	rootClaudeFile := filepath.Join(repoDir, "CLAUDE.md")
	nestedClaudeFile := filepath.Join(repoDir, "docs", "CLAUDE.md")

	require.NoError(t, os.WriteFile(rootClaudeFile, []byte("root claude content"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "docs"), 0755))
	require.NoError(t, os.WriteFile(nestedClaudeFile, []byte("nested claude content"), 0644))

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(oldDir)

	err = os.Chdir(repoDir)
	require.NoError(t, err)

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	storageDir := filepath.Join(home, ".claude", "claude-md", "test", "repo.git")
	os.RemoveAll(storageDir)

	// Test init command
	t.Run("Init", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		exitCode := cmd.Run([]string{"init"}, cmd.RunOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})
		require.Equal(t, 0, exitCode, "init failed - stdout: %s, stderr: %s", stdout.String(), stderr.String())
		assert.Contains(t, stdout.String(), "User: test")
	})

	// Test save command
	t.Run("Save", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		exitCode := cmd.Run([]string{"save"}, cmd.RunOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})
		require.Equal(t, 0, exitCode, "save failed - stdout: %s, stderr: %s", stdout.String(), stderr.String())

		// Verify files are now symlinks
		info, err := os.Lstat(rootClaudeFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)

		info, err = os.Lstat(nestedClaudeFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)

		// Verify content is preserved
		content, err := os.ReadFile(rootClaudeFile)
		require.NoError(t, err)
		assert.Equal(t, "root claude content", string(content))

		content, err = os.ReadFile(nestedClaudeFile)
		require.NoError(t, err)
		assert.Equal(t, "nested claude content", string(content))
	})

	// Test clear command
	t.Run("Clear", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		exitCode := cmd.Run([]string{"clear"}, cmd.RunOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})
		require.Equal(t, 0, exitCode, "clear failed - stdout: %s, stderr: %s", stdout.String(), stderr.String())

		// Verify symlinks are removed
		_, err = os.Lstat(rootClaudeFile)
		assert.True(t, os.IsNotExist(err))

		_, err = os.Lstat(nestedClaudeFile)
		assert.True(t, os.IsNotExist(err))

		// Verify storage files still exist
		assert.DirExists(t, storageDir)
	})

	// Test restore command
	t.Run("Restore", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		exitCode := cmd.Run([]string{"restore"}, cmd.RunOptions{
			Stdout: &stdout,
			Stderr: &stderr,
		})
		require.Equal(t, 0, exitCode, "restore failed - stdout: %s, stderr: %s", stdout.String(), stderr.String())

		// Verify files are symlinks again
		info, err := os.Lstat(rootClaudeFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)

		info, err = os.Lstat(nestedClaudeFile)
		require.NoError(t, err)
		assert.True(t, info.Mode()&os.ModeSymlink != 0)

		// Verify content is still correct
		content, err := os.ReadFile(rootClaudeFile)
		require.NoError(t, err)
		assert.Equal(t, "root claude content", string(content))

		content, err = os.ReadFile(nestedClaudeFile)
		require.NoError(t, err)
		assert.Equal(t, "nested claude content", string(content))
	})

	// Cleanup storage
	os.RemoveAll(storageDir)
}
