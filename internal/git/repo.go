package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repository represents a git repository and its configuration
type Repository struct {
	RootPath  string
	RepoName  string
	UserEmail string
}

// FindRepository detects the git repository from the current or parent directories
func FindRepository() (*Repository, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Walk up from current directory to find .git
	for {
		gitPath := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// Accept both .git directory and .git file (worktree)
			if info.IsDir() || info.Mode().IsRegular() {
				return &Repository{RootPath: dir}, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, errors.New("not in a git repository")
		}
		dir = parent
	}
}

// GetOriginURL retrieves the origin remote URL
func (r *Repository) GetOriginURL() (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("origin remote not configured")
	}
	return strings.TrimSpace(string(output)), nil
}

// GetUserEmail retrieves user.email from git config
func (r *Repository) GetUserEmail() (string, error) {
	cmd := exec.Command("git", "config", "--get", "user.email")
	cmd.Dir = r.RootPath
	output, err := cmd.Output()
	if err != nil {
		return "", errors.New("user.email not configured in git")
	}
	return strings.TrimSpace(string(output)), nil
}

// ExtractRepoName extracts repository name from git remote URL
// Handles both SSH (git@github.com:user/repo.git) and HTTPS (https://github.com/user/repo.git)
// Returns repo name as-is from URL (e.g., "kapetan.git" if URL ends with "kapetan.git", "kapetan" if URL ends with "kapetan")
func ExtractRepoName(remoteURL string) (string, error) {
	if remoteURL == "" {
		return "", errors.New("remote URL is empty")
	}

	// Remove .git suffix if present for parsing, but we'll preserve the original
	url := remoteURL

	// Check for SSH prefix but invalid format
	if strings.HasPrefix(url, "git@") && !strings.Contains(url, ":") {
		return "", fmt.Errorf("invalid SSH URL format: %s", remoteURL)
	}

	// Handle SSH format: git@github.com:user/repo.git
	if strings.Contains(url, ":") && strings.HasPrefix(url, "git@") {
		parts := strings.Split(url, ":")
		if len(parts) < 2 || parts[1] == "" {
			return "", fmt.Errorf("invalid SSH URL format: %s", remoteURL)
		}
		path := parts[len(parts)-1]
		segments := strings.Split(path, "/")
		if len(segments) == 0 || segments[len(segments)-1] == "" {
			return "", fmt.Errorf("invalid SSH URL format: %s", remoteURL)
		}
		return segments[len(segments)-1], nil
	}

	// Handle HTTPS format: https://github.com/user/repo.git
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		// Remove protocol
		url = strings.TrimPrefix(url, "https://")
		url = strings.TrimPrefix(url, "http://")

		// Split by / and get last segment
		segments := strings.Split(url, "/")
		if len(segments) < 2 {
			return "", fmt.Errorf("invalid HTTPS URL format: %s", remoteURL)
		}
		return segments[len(segments)-1], nil
	}

	return "", fmt.Errorf("unsupported URL format: %s", remoteURL)
}

// ExtractUserFromEmail extracts username from email by trimming @ and everything after
// Example: "john.doe@example.com" → "john.doe"
// Returns error if email is empty or doesn't contain @
func ExtractUserFromEmail(email string) (string, error) {
	if email == "" {
		return "", errors.New("email is empty")
	}

	idx := strings.Index(email, "@")
	if idx == -1 {
		return "", errors.New("email does not contain @")
	}

	user := email[:idx]
	if user == "" {
		return "", errors.New("username part of email is empty")
	}

	return user, nil
}
