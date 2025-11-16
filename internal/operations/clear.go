package operations

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/kapetan-io/claude-md.go/internal/files"
	"github.com/kapetan-io/claude-md.go/internal/storage"
)

// ClearResult represents the result of clearing symlinks
type ClearResult struct {
	RepoRelativePath string
	Success          bool
	Skipped          bool
	SkipReason       string
	Error            error
}

// ClearOptions contains options for clear operation
type ClearOptions struct {
	RepoRoot      string
	PathConverter *storage.PathConverter
}

// ClearSymlinks removes all CLAUDE.md symlinks from repository
func ClearSymlinks(opts ClearOptions) []ClearResult {
	var results []ClearResult

	// Find all CLAUDE.md files in repository
	claudeFiles, err := files.FindClaudeFiles(opts.RepoRoot)
	if err != nil {
		return results
	}

	// Get storage directory for validation
	storageDir := opts.PathConverter.GetRepoStorageDir()

	// Filter to only symlinks
	for _, file := range claudeFiles {
		if !file.IsSymlink {
			continue
		}

		result := ClearResult{
			RepoRelativePath: file.RepoRelativePath,
		}

		// Read symlink target
		target, err := os.Readlink(file.AbsolutePath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "failed to read symlink"
			result.Error = err
			results = append(results, result)
			continue
		}

		// Verify symlink points to our storage directory
		// Convert to absolute path for comparison
		absTarget, err := filepath.Abs(target)
		if err != nil {
			// If it's already absolute, use it as-is
			absTarget = target
		}

		// Check if target is within our storage directory
		if !strings.HasPrefix(absTarget, storageDir) {
			result.Skipped = true
			result.SkipReason = "symlink points outside storage"
			results = append(results, result)
			continue
		}

		// Remove the symlink
		if err := os.Remove(file.AbsolutePath); err != nil {
			result.Error = err
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return results
}
