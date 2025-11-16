package operations_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/files"
	"github.com/kapetan-io/claude-md.go/internal/operations"
	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClearSymlinksValidation(t *testing.T) {
	tmpDir := t.TempDir()
	repoDir := filepath.Join(tmpDir, "repo")
	require.NoError(t, os.MkdirAll(repoDir, 0755))

	storageDir := filepath.Join(tmpDir, "storage")
	require.NoError(t, os.MkdirAll(storageDir, 0700))

	pc := &storage.PathConverter{
		StorageRoot: filepath.Join(tmpDir, "storage"),
		RepoName:    "test.git",
	}

	t.Run("RemovesSymlinksPointingToStorage", func(t *testing.T) {
		// Create storage file
		storageFile := filepath.Join(storageDir, "test.git", "CLAUDE.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(storageFile), 0700))
		require.NoError(t, os.WriteFile(storageFile, []byte("content"), 0644))

		// Create symlink in repo pointing to storage
		repoFile := filepath.Join(repoDir, "CLAUDE.md")
		require.NoError(t, os.Symlink(storageFile, repoFile))

		// Verify it's a symlink
		isSymlink, err := files.IsSymlink(repoFile)
		require.NoError(t, err)
		assert.True(t, isSymlink)

		// Run clear operation
		results := operations.ClearSymlinks(operations.ClearOptions{
			RepoRoot:      repoDir,
			PathConverter: pc,
		})

		// Should remove the symlink
		require.Len(t, results, 1)
		assert.True(t, results[0].Success)
		assert.Equal(t, "CLAUDE.md", results[0].RepoRelativePath)

		// Verify symlink is removed
		_, err = os.Lstat(repoFile)
		assert.True(t, os.IsNotExist(err))

		// Verify storage file still exists
		_, err = os.Stat(storageFile)
		assert.NoError(t, err)
	})

	t.Run("SkipsSymlinksPointingOutsideStorage", func(t *testing.T) {
		// Create a file outside storage
		externalFile := filepath.Join(tmpDir, "external", "file.md")
		require.NoError(t, os.MkdirAll(filepath.Dir(externalFile), 0755))
		require.NoError(t, os.WriteFile(externalFile, []byte("external content"), 0644))

		// Create symlink in repo pointing to external file (named CLAUDE.md so it gets picked up)
		claudeFile := filepath.Join(repoDir, "CLAUDE.md")
		require.NoError(t, os.Symlink(externalFile, claudeFile))

		// Run clear operation
		results := operations.ClearSymlinks(operations.ClearOptions{
			RepoRoot:      repoDir,
			PathConverter: pc,
		})

		// Should skip the symlink
		require.Len(t, results, 1)
		assert.True(t, results[0].Skipped)
		assert.Equal(t, "symlink points outside storage", results[0].SkipReason)

		// Verify symlink still exists (not removed)
		_, err := os.Lstat(claudeFile)
		assert.NoError(t, err)
	})

	t.Run("SkipsRegularFiles", func(t *testing.T) {
		// Create regular file (not a symlink)
		regularFile := filepath.Join(repoDir, "regular-CLAUDE.md")
		require.NoError(t, os.WriteFile(regularFile, []byte("regular file"), 0644))

		// Run clear operation
		results := operations.ClearSymlinks(operations.ClearOptions{
			RepoRoot:      repoDir,
			PathConverter: pc,
		})

		// Should not process regular files at all
		for _, result := range results {
			assert.NotEqual(t, "regular-CLAUDE.md", result.RepoRelativePath)
		}

		// Verify file still exists
		_, err := os.Stat(regularFile)
		assert.NoError(t, err)
	})
}
