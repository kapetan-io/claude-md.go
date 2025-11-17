package storage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kapetan-io/claude-md.go/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToStorageName(t *testing.T) {
	pc, err := storage.NewPathConverter("testuser", "testrepo.git")
	require.NoError(t, err)

	for _, test := range []struct {
		name    string
		path    string
		want    string
		wantErr string
	}{
		{
			name: "RootLevelFile",
			path: "CLAUDE.md",
			want: "CLAUDE.md",
		},
		{
			name: "NestedPath",
			path: "source/go/api/CLAUDE.md",
			want: "source~go~api~CLAUDE.md",
		},
		{
			name: "SingleDirectory",
			path: "docs/CLAUDE.md",
			want: "docs~CLAUDE.md",
		},
		{
			name: "DeeplyNestedPath",
			path: "a/b/c/d/e/CLAUDE.md",
			want: "a~b~c~d~e~CLAUDE.md",
		},
		{
			name:    "PathWithTilde",
			path:    "some~dir/CLAUDE.md",
			wantErr: "path contains ~ character",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := pc.ConvertToStorageName(test.path)
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

func TestConvertToRepoPath(t *testing.T) {
	pc, err := storage.NewPathConverter("testuser", "testrepo.git")
	require.NoError(t, err)

	for _, test := range []struct {
		name        string
		storageName string
		want        string
	}{
		{
			name:        "RootLevelFile",
			storageName: "CLAUDE.md",
			want:        "CLAUDE.md",
		},
		{
			name:        "NestedPath",
			storageName: "source~go~api~CLAUDE.md",
			want:        filepath.Join("source", "go", "api", "CLAUDE.md"),
		},
		{
			name:        "SingleDirectory",
			storageName: "docs~CLAUDE.md",
			want:        filepath.Join("docs", "CLAUDE.md"),
		},
		{
			name:        "DeeplyNestedPath",
			storageName: "a~b~c~d~e~CLAUDE.md",
			want:        filepath.Join("a", "b", "c", "d", "e", "CLAUDE.md"),
		},
		{
			name:        "PathTraversalAttempt",
			storageName: "..~..~..~etc~CLAUDE.md",
			want:        "", // Should return empty string for invalid path
		},
		{
			name:        "PathTraversalInMiddle",
			storageName: "docs~..~api~CLAUDE.md",
			want:        "", // Should return empty string for invalid path
		},
		{
			name:        "PathTraversalAtRoot",
			storageName: "..~CLAUDE.md",
			want:        "", // Should return empty string for invalid path
		},
		{
			name:        "DotsInFilename",
			storageName: "CLAUDE..md",
			want:        "", // Should return empty string for invalid path
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := pc.ConvertToRepoPath(test.storageName)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestValidatePath(t *testing.T) {
	for _, test := range []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "ValidPath",
			path:    "source/go/api/CLAUDE.md",
			wantErr: false,
		},
		{
			name:    "InvalidPathWithTilde",
			path:    "some~dir/CLAUDE.md",
			wantErr: true,
		},
		{
			name:    "RootPath",
			path:    "CLAUDE.md",
			wantErr: false,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := storage.ValidatePath(test.path)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetStoragePath(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	pc, err := storage.NewPathConverter("testuser", "testrepo.git")
	require.NoError(t, err)

	for _, test := range []struct {
		name string
		path string
		want string
	}{
		{
			name: "RootLevelFile",
			path: "CLAUDE.md",
			want: filepath.Join(home, ".claude", "claude-md", "testuser", "testrepo.git", "CLAUDE.md"),
		},
		{
			name: "NestedPath",
			path: "source/go/api/CLAUDE.md",
			want: filepath.Join(home, ".claude", "claude-md", "testuser", "testrepo.git", "source~go~api~CLAUDE.md"),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := pc.GetStoragePath(test.path)
			require.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func TestEnsureStorageDirPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	testUser := "permtest"
	testRepo := "testrepo.git"

	pc := &storage.PathConverter{
		StorageRoot: filepath.Join(tmpDir, ".claude", "claude-md", testUser),
		RepoName:    testRepo,
	}

	err := pc.EnsureStorageDir()
	require.NoError(t, err)

	storageDir := pc.GetRepoStorageDir()
	info, err := os.Stat(storageDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify permissions are 0700 (owner-only access)
	perms := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0700), perms, "Storage directory should have 0700 permissions for security")
}
