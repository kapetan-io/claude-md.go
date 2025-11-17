package git_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractRepoName(t *testing.T) {
	for _, test := range []struct {
		name    string
		url     string
		want    string
		wantErr string
	}{
		{
			name: "SSHWithGitExtension",
			url:  "git@github.com:user/repo.git",
			want: "repo.git",
		},
		{
			name: "SSHWithoutGitExtension",
			url:  "git@github.com:user/repo",
			want: "repo",
		},
		{
			name: "HTTPSWithGitExtension",
			url:  "https://github.com/user/repo.git",
			want: "repo.git",
		},
		{
			name: "HTTPSWithoutGitExtension",
			url:  "https://github.com/user/repo",
			want: "repo",
		},
		{
			name: "HTTPWithGitExtension",
			url:  "http://github.com/user/repo.git",
			want: "repo.git",
		},
		{
			name:    "EmptyURL",
			url:     "",
			wantErr: "remote URL is empty",
		},
		{
			name:    "InvalidSSHFormat",
			url:     "git@github.com",
			wantErr: "invalid SSH URL format",
		},
		{
			name:    "InvalidHTTPSFormat",
			url:     "https://github.com",
			wantErr: "invalid HTTPS URL format",
		},
		{
			name:    "UnsupportedFormat",
			url:     "file:///path/to/repo",
			wantErr: "unsupported URL format",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := git.ExtractRepoName(test.url)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestExtractUserFromEmail(t *testing.T) {
	for _, test := range []struct {
		name    string
		email   string
		want    string
		wantErr string
	}{
		{
			name:  "StandardEmail",
			email: "john.doe@example.com",
			want:  "john.doe",
		},
		{
			name:  "SimpleEmail",
			email: "user@domain.com",
			want:  "user",
		},
		{
			name:  "ComplexUsername",
			email: "first.middle.last@company.co.uk",
			want:  "first.middle.last",
		},
		{
			name:    "EmptyEmail",
			email:   "",
			wantErr: "email is empty",
		},
		{
			name:    "NoAtSign",
			email:   "notanemail",
			wantErr: "email does not contain @",
		},
		{
			name:    "EmptyUsername",
			email:   "@example.com",
			wantErr: "username part of email is empty",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := git.ExtractUserFromEmail(test.email)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestFindRepositoryWorktree(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with regular .git directory
	t.Run("RegularGitDirectory", func(t *testing.T) {
		repoDir := filepath.Join(tmpDir, "regular-repo")
		require.NoError(t, os.MkdirAll(filepath.Join(repoDir, ".git"), 0755))

		// Change to repo directory
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()

		require.NoError(t, os.Chdir(repoDir))

		repo, err := git.FindRepository()
		require.NoError(t, err)

		// Use EvalSymlinks to handle /var vs /private/var on macOS
		expectedPath, err := filepath.EvalSymlinks(repoDir)
		require.NoError(t, err)
		actualPath, err := filepath.EvalSymlinks(repo.RootPath)
		require.NoError(t, err)
		assert.Equal(t, expectedPath, actualPath)
	})

	// Test with .git file (worktree)
	t.Run("GitWorktreeFile", func(t *testing.T) {
		worktreeDir := filepath.Join(tmpDir, "worktree-repo")
		require.NoError(t, os.MkdirAll(worktreeDir, 0755))

		// Create .git file (simulating worktree)
		gitFile := filepath.Join(worktreeDir, ".git")
		require.NoError(t, os.WriteFile(gitFile, []byte("gitdir: /some/path/.git/worktrees/branch"), 0644))

		// Change to worktree directory
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }()

		require.NoError(t, os.Chdir(worktreeDir))

		repo, err := git.FindRepository()
		require.NoError(t, err)

		// Use EvalSymlinks to handle /var vs /private/var on macOS
		expectedPath, err := filepath.EvalSymlinks(worktreeDir)
		require.NoError(t, err)
		actualPath, err := filepath.EvalSymlinks(repo.RootPath)
		require.NoError(t, err)
		assert.Equal(t, expectedPath, actualPath)
	})
}
