package operations

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kapetan-io/claude-md.go/internal/files"
	"github.com/kapetan-io/claude-md.go/internal/storage"
)

// SaveResult represents the result of saving a file
type SaveResult struct {
	RepoRelativePath string
	StoragePath      string // Path to storage file (for warnings/info)
	Success          bool
	Skipped          bool
	SkipReason       string // "already symlink", "invalid path", "storage file exists", etc.
	Warning          string // Warning message for user
	Error            error
}

// SaveOptions contains options for save operation
type SaveOptions struct {
	RepoRoot      string
	PathConverter *storage.PathConverter
}

// SaveFiles converts CLAUDE.md files to symlinks
func SaveFiles(claudeFiles []files.ClaudeFile, opts SaveOptions) []SaveResult {
	var results []SaveResult

	for _, file := range claudeFiles {
		result := SaveResult{
			RepoRelativePath: file.RepoRelativePath,
		}

		// Validate path
		if err := storage.ValidatePath(file.RepoRelativePath); err != nil {
			result.Skipped = true
			result.SkipReason = "invalid path"
			result.Warning = fmt.Sprintf("Skipping %s: %v", file.RepoRelativePath, err)
			result.Error = err
			results = append(results, result)
			continue
		}

		// Skip if already a symlink
		if file.IsSymlink {
			result.Skipped = true
			result.SkipReason = "already symlink"
			results = append(results, result)
			continue
		}

		// Get storage path
		storagePath, err := opts.PathConverter.GetStoragePath(file.RepoRelativePath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "path conversion error"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: %v", file.RepoRelativePath, err)
			results = append(results, result)
			continue
		}
		result.StoragePath = storagePath

		// Ensure storage directory exists
		if err := opts.PathConverter.EnsureStorageDir(); err != nil {
			result.Skipped = true
			result.SkipReason = "storage directory creation failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to create storage directory: %v",
				file.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		// Read file content before any modifications
		content, err := os.ReadFile(file.AbsolutePath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "read failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to read file: %v",
				file.RepoRelativePath, err)
			results = append(results, result)
			continue
		}

		// Atomically create storage file with O_EXCL to prevent race conditions
		storageFile, err := os.OpenFile(storagePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			if os.IsExist(err) {
				// File already exists - skip with warning
				result.Skipped = true
				result.SkipReason = "storage file exists"
				result.Warning = fmt.Sprintf("Skipping %s: storage file already exists at %s",
					file.RepoRelativePath, storagePath)
			} else {
				// Other error
				result.Skipped = true
				result.SkipReason = "storage file creation failed"
				result.Error = err
				result.Warning = fmt.Sprintf("Skipping %s: failed to create storage file: %v",
					file.RepoRelativePath, err)
			}
			results = append(results, result)
			continue
		}

		// Write content to storage file
		if _, err := storageFile.Write(content); err != nil {
			_ = storageFile.Close()
			_ = os.Remove(storagePath) // Clean up partial file
			result.Skipped = true
			result.SkipReason = "write failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to write to storage: %v",
				file.RepoRelativePath, err)
			results = append(results, result)
			continue
		}
		_ = storageFile.Close()

		// Remove original file
		if err := os.Remove(file.AbsolutePath); err != nil {
			result.Skipped = true
			result.SkipReason = "remove failed"
			result.Error = err
			result.Warning = fmt.Sprintf("Skipping %s: failed to remove original file: %v",
				file.RepoRelativePath, err)
			// Clean up storage file
			_ = os.Remove(storagePath)
			results = append(results, result)
			continue
		}

		// Get absolute storage path for symlink
		absStoragePath, err := filepath.Abs(storagePath)
		if err != nil {
			result.Skipped = true
			result.SkipReason = "absolute path failed"
			result.Error = err
			// Try to restore original file
			if restoreErr := os.WriteFile(file.AbsolutePath, content, 0644); restoreErr != nil {
				// CRITICAL: Failed to restore file
				result.Error = fmt.Errorf("CRITICAL: failed to restore file after error (data is in storage): %w (original error: %v)",
					restoreErr, err)
				result.Warning = fmt.Sprintf("CRITICAL: %s - original file deleted but restore failed. Content saved in %s",
					file.RepoRelativePath, storagePath)
			} else {
				// Successfully restored, clean up storage
				_ = os.Remove(storagePath)
				result.Warning = fmt.Sprintf("Skipping %s: failed to get absolute path: %v",
					file.RepoRelativePath, err)
			}
			results = append(results, result)
			continue
		}

		// Create symlink
		if err := os.Symlink(absStoragePath, file.AbsolutePath); err != nil {
			result.Skipped = true
			result.SkipReason = "symlink creation failed"
			result.Error = err
			// Try to restore original file
			if restoreErr := os.WriteFile(file.AbsolutePath, content, 0644); restoreErr != nil {
				// CRITICAL: Failed to restore file
				result.Error = fmt.Errorf("CRITICAL: failed to restore file after error (data is in storage): %w (original error: %v)",
					restoreErr, err)
				result.Warning = fmt.Sprintf("CRITICAL: %s - original file deleted but restore failed. Content saved in %s",
					file.RepoRelativePath, storagePath)
			} else {
				// Successfully restored, clean up storage
				_ = os.Remove(storagePath)
				result.Warning = fmt.Sprintf("Skipping %s: failed to create symlink: %v",
					file.RepoRelativePath, err)
			}
			results = append(results, result)
			continue
		}

		result.Success = true
		results = append(results, result)
	}

	return results
}
