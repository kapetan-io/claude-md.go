package operations

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kapetan-io/claude-md.go/internal/files"
)

// RestoreResult represents the result of restoring a file
type RestoreResult struct {
	RepoRelativePath string
	Success          bool
	Skipped          bool
	SkipReason       string // "already exists", "parent dir missing", etc.
	Warning          string // For parent directory missing case
	StoragePath      string // Path to stored file (for warnings)
	Error            error
}

// RestoreOptions contains options for restore operation
type RestoreOptions struct {
	RepoRoot string
}

// RestoreFiles creates symlinks for stored files
func RestoreFiles(storedFiles []files.StoredFile, opts RestoreOptions) []RestoreResult {
	var results []RestoreResult

	for _, stored := range storedFiles {
		result := RestoreResult{
			RepoRelativePath: stored.RepoRelativePath,
			StoragePath:      stored.StoragePath,
		}

		// Construct target path in repo
		targetPath := filepath.Join(opts.RepoRoot, stored.RepoRelativePath)

		// Validate target path is within repository bounds (prevent path traversal)
		absRepoRoot, err := filepath.Abs(opts.RepoRoot)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "failed to get absolute repo path"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: %v", stored.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		absTargetPath, err := filepath.Abs(targetPath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "failed to get absolute target path"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: %v", stored.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		// Check if target path escapes repository
		relPath, err := filepath.Rel(absRepoRoot, absTargetPath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			result.Skipped = true
			result.SkipReason = "path escapes repository"
			result.Warning = fmt.Sprintf("Skipping %s: path would escape repository bounds (storage: %s)",
				stored.RepoRelativePath, stored.StoragePath)
			results = append(results, result)
			continue
		}

		// Check if parent directory exists
		parentDir := filepath.Dir(targetPath)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			result.Skipped = true
			result.SkipReason = "parent dir missing"
			result.Warning = fmt.Sprintf("Skipping %s: parent directory does not exist (storage: %s)",
				stored.RepoRelativePath, stored.StoragePath)
			results = append(results, result)
			continue
		}

		// Check if file already exists at target location
		if info, err := os.Lstat(targetPath); err == nil {
			// File exists - check if it's a symlink
			if info.Mode()&os.ModeSymlink != 0 {
				// It's a symlink - check if it points to the correct location
				currentTarget, err := os.Readlink(targetPath)
				if err != nil {
					result.Skipped = true
					result.SkipReason = "symlink read failed"
					result.Error = err
					result.Warning = fmt.Sprintf("Skipping %s: failed to read symlink: %v",
						stored.RepoRelativePath, err)
					results = append(results, result)
					continue
				}

				// Get absolute path to storage for comparison
				absStoragePath, err := filepath.Abs(stored.StoragePath)
				if err != nil {
					result.Skipped = true
					result.SkipReason = "absolute path failed"
					result.Error = err
					result.Warning = fmt.Sprintf("Skipping %s: failed to get absolute path: %v",
						stored.RepoRelativePath, err)
					results = append(results, result)
					continue
				}

				if currentTarget == absStoragePath {
					// Already points to correct location
					result.Skipped = true
					result.SkipReason = "already correct"
					results = append(results, result)
					continue
				}

				// Points to wrong location
				result.Skipped = true
				result.SkipReason = "wrong target"
				result.Warning = fmt.Sprintf("Skipping %s: symlink exists but points to %s (storage: %s)",
					stored.RepoRelativePath, currentTarget, stored.StoragePath)
				results = append(results, result)
				continue
			}

			// It's a regular file
			result.Skipped = true
			result.SkipReason = "file exists"
			result.Warning = fmt.Sprintf("Skipping %s: regular file already exists (storage: %s)",
				stored.RepoRelativePath, stored.StoragePath)
			results = append(results, result)
			continue
		}

		// Get absolute storage path for symlink
		absStoragePath, err := filepath.Abs(stored.StoragePath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "absolute path failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to get absolute path: %v",
				stored.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		// Create symlink
		if err := os.Symlink(absStoragePath, targetPath); err != nil {
			result.Skipped = true
			result.SkipReason = "symlink creation failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to create symlink: %v",
				stored.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		result.Success = true
		results = append(results, result)
	}

	return results
}
