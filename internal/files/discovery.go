package files

import (
	"os"
	"path/filepath"
	"strings"
)

// ClaudeFile represents a found CLAUDE.md file
type ClaudeFile struct {
	AbsolutePath     string // Full path to file
	RepoRelativePath string // Path relative to repo root
	IsSymlink        bool   // Whether it's already a symlink
}

// FindClaudeFiles finds all CLAUDE.md files in the repository
func FindClaudeFiles(repoRoot string) ([]ClaudeFile, error) {
	var claudeFiles []ClaudeFile

	err := filepath.Walk(repoRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Check if filename matches CLAUDE.md (case-insensitive)
		if !info.IsDir() && strings.EqualFold(filepath.Base(path), "CLAUDE.md") {
			relPath, err := filepath.Rel(repoRoot, path)
			if err != nil {
				return err
			}

			isSymlink, err := IsSymlink(path)
			if err != nil {
				return err
			}

			claudeFiles = append(claudeFiles, ClaudeFile{
				AbsolutePath:     path,
				RepoRelativePath: relPath,
				IsSymlink:        isSymlink,
			})
		}

		return nil
	})

	return claudeFiles, err
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	return info.Mode()&os.ModeSymlink != 0, nil
}
