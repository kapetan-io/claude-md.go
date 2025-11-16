package files

import (
	"os"
	"strings"

	"github.com/kapetan-io/claude-md.go/internal/storage"
)

// StoredFile represents a file in storage
type StoredFile struct {
	StorageFilename  string // Filename in storage (e.g., "source~go~api~CLAUDE.md")
	RepoRelativePath string // Path to restore in repo (e.g., "source/go/api/CLAUDE.md")
	StoragePath      string // Full path to stored file
}

// FindStoredFiles finds all stored CLAUDE.md files for a repository
func FindStoredFiles(repoStorageDir string, converter *storage.PathConverter) ([]StoredFile, error) {
	entries, err := os.ReadDir(repoStorageDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []StoredFile{}, nil
		}
		return nil, err
	}

	var storedFiles []StoredFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()

		// Only include files that match CLAUDE.md (case-insensitive)
		// They should end with CLAUDE.md or be exactly CLAUDE.md
		if !strings.EqualFold(filename, "CLAUDE.md") && !strings.HasSuffix(strings.ToLower(filename), "~claude.md") {
			continue
		}

		repoPath := converter.ConvertToRepoPath(filename)

		// Skip files with invalid paths (e.g., containing "..")
		if repoPath == "" {
			continue
		}

		storagePath := repoStorageDir + string(os.PathSeparator) + filename

		storedFiles = append(storedFiles, StoredFile{
			StorageFilename:  filename,
			RepoRelativePath: repoPath,
			StoragePath:      storagePath,
		})
	}

	return storedFiles, nil
}
