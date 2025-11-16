package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathConverter handles conversion between repository paths and storage paths
type PathConverter struct {
	StorageRoot string // ~/.claude/claude-md/<user>
	RepoName    string // Repository name as extracted from origin URL (e.g., "kapetan.git" or "kapetan")
}

// NewPathConverter creates a new path converter
// Returns error if home directory cannot be determined
func NewPathConverter(user, repoName string) (*PathConverter, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	storageRoot := filepath.Join(home, ".claude", "claude-md", user)

	return &PathConverter{
		StorageRoot: storageRoot,
		RepoName:    repoName,
	}, nil
}

// ConvertToStorageName converts a repository relative path to storage filename
// Example: "source/go/api/CLAUDE.md" → "source~go~api~CLAUDE.md"
// Example: "CLAUDE.md" → "CLAUDE.md"
func (pc *PathConverter) ConvertToStorageName(repoRelativePath string) (string, error) {
	// Validate path doesn't contain ~
	if err := ValidatePath(repoRelativePath); err != nil {
		return "", err
	}

	// Clean the path
	cleaned := filepath.Clean(repoRelativePath)

	// Handle root case
	if cleaned == "CLAUDE.md" || cleaned == filepath.Base(repoRelativePath) {
		return filepath.Base(repoRelativePath), nil
	}

	// Split path and join with ~
	parts := strings.Split(cleaned, string(filepath.Separator))
	return strings.Join(parts, "~"), nil
}

// ConvertToRepoPath converts a storage filename back to repository relative path
// Example: "source~go~api~CLAUDE.md" → "source/go/api/CLAUDE.md"
// Example: "CLAUDE.md" → "CLAUDE.md"
// Returns empty string if the path contains invalid sequences like ".."
func (pc *PathConverter) ConvertToRepoPath(storageFilename string) string {
	// Handle root case
	if !strings.Contains(storageFilename, "~") {
		// Validate even the root case
		if strings.Contains(storageFilename, "..") {
			return ""
		}
		return storageFilename
	}

	// Split by ~ and validate each part doesn't contain ".."
	parts := strings.Split(storageFilename, "~")
	for _, part := range parts {
		if part == ".." || strings.Contains(part, "..") {
			return ""
		}
	}

	return filepath.Join(parts...)
}

// GetStoragePath returns full path to stored file
func (pc *PathConverter) GetStoragePath(repoRelativePath string) (string, error) {
	storageName, err := pc.ConvertToStorageName(repoRelativePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(pc.GetRepoStorageDir(), storageName), nil
}

// GetRepoStorageDir returns the directory where this repo's files are stored
// Returns: ~/.claude/claude-md/<user>/<repo-name>/
func (pc *PathConverter) GetRepoStorageDir() string {
	return filepath.Join(pc.StorageRoot, pc.RepoName)
}

// ValidatePath checks if path contains invalid characters (like ~)
func ValidatePath(path string) error {
	if strings.Contains(path, "~") {
		return errors.New("path contains ~ character which is not supported")
	}
	return nil
}

// EnsureStorageDir creates storage directory if it doesn't exist
func (pc *PathConverter) EnsureStorageDir() error {
	dir := pc.GetRepoStorageDir()
	// Use 0700 to restrict access to owner only (CLAUDE.md may contain sensitive info)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}
	return nil
}
