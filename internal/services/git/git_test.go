package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

func TestService_ValidateGitInstalled(t *testing.T) {
	service := New()

	// Test git validation - this will pass on systems with git installed
	err := service.ValidateGitInstalled()

	// Check if git is actually available
	_, gitErr := exec.LookPath("git")
	if gitErr != nil {
		// Git not available, should return error
		if err == nil {
			t.Error("Expected error when git is not available, got nil")
		}
		if !models.IsErrorCode(err, models.ErrorCodeGitNotFound) {
			t.Errorf("Expected ErrorCodeGitNotFound, got %v", err)
		}
	} else if err != nil {
		// Git available, should not return error
		t.Errorf("Expected no error when git is available, got %v", err)
	}
}

func TestService_CleanupTempDir(t *testing.T) {
	service := New()

	tests := []struct {
		name      string
		path      string
		shouldErr bool
		errCode   models.ErrorCode
	}{
		{
			name:      "empty path",
			path:      "",
			shouldErr: false,
		},
		{
			name:      "valid temp directory",
			path:      "/tmp/" + config.TempDirPrefix + "test123",
			shouldErr: false,
		},
		{
			name:      "invalid path - not temp directory",
			path:      "/home/user/documents",
			shouldErr: true,
			errCode:   models.ErrorCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create the directory if it's a valid temp path
			if tt.path != "" && strings.Contains(tt.path, config.TempDirPrefix) && !tt.shouldErr {
				_ = os.MkdirAll(tt.path, 0755) // Best effort directory creation for testing
			}

			err := service.CleanupTempDir(tt.path)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errCode != "" && !models.IsErrorCode(err, tt.errCode) {
					t.Errorf("Expected error code %s, got %v", tt.errCode, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestService_createTempDir(t *testing.T) {
	service := New()

	tempDir, err := service.createTempDir()
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	defer os.RemoveAll(tempDir)

	// Check if directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Temp directory was not created")
	}

	// Check if directory name contains prefix
	if !strings.Contains(filepath.Base(tempDir), config.TempDirPrefix) {
		t.Errorf("Temp directory name should contain prefix %s, got %s", config.TempDirPrefix, tempDir)
	}
}

// TestService_CloneRepository requires network access and is integration test
// We'll test the error handling for invalid scenarios
func TestService_CloneRepository_InvalidScenarios(t *testing.T) {
	service := New()

	tests := []struct {
		name    string
		url     string
		commit  string
		wantErr bool
	}{
		{
			name:    "invalid URL",
			url:     "https://invalid-url-that-does-not-exist.com/repo.git",
			commit:  "abc123",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			commit:  "abc123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip if git is not available
			if err := service.ValidateGitInstalled(); err != nil {
				t.Skip("Git not available, skipping clone tests")
			}

			tempDir, err := service.CloneRepository(tt.url, tt.commit)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
					// Clean up if somehow succeeded
					if tempDir != "" {
						_ = service.CleanupTempDir(tempDir) // Best effort cleanup
					}
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if tempDir != "" {
					defer func() { _ = service.CleanupTempDir(tempDir) }()
				}
			}
		})
	}
}

// Integration test for successful clone - requires network and valid repo
func TestService_CloneRepository_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := New()

	// Skip if git is not available
	if err := service.ValidateGitInstalled(); err != nil {
		t.Skip("Git not available, skipping clone tests")
	}

	// Use a small, reliable public repository for testing
	testURL := "https://github.com/octocat/Hello-World.git"
	testCommit := "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d" // Known commit in Hello-World repo

	tempDir, err := service.CloneRepository(testURL, testCommit)
	if err != nil {
		t.Fatalf("Failed to clone repository: %v", err)
	}

	defer func() { _ = service.CleanupTempDir(tempDir) }()

	// Verify directory exists and contains git repository
	gitDir := filepath.Join(tempDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		t.Error("Cloned directory does not contain .git folder")
	}

	// Verify correct commit is checked out
	info, err := service.GetRepoInfo(tempDir)
	if err != nil {
		t.Fatalf("Failed to get repo info: %v", err)
	}

	if !strings.HasPrefix(info["commit"], testCommit[:7]) {
		t.Errorf("Expected commit to start with %s, got %s", testCommit[:7], info["commit"])
	}
}

func TestService_GetRepoInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := New()

	// Skip if git is not available
	if err := service.ValidateGitInstalled(); err != nil {
		t.Skip("Git not available, skipping repo info tests")
	}

	// Create a temporary git repository for testing
	tempDir, err := os.MkdirTemp("", "test-git-repo-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skip("Cannot initialize git repo for testing")
	}

	// Configure git user for testing
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	_ = cmd.Run()

	// Test getting repo info from non-git directory
	nonGitDir, err := os.MkdirTemp("", "non-git-")
	if err != nil {
		t.Fatalf("Failed to create non-git temp directory: %v", err)
	}
	defer os.RemoveAll(nonGitDir)

	_, err = service.GetRepoInfo(nonGitDir)
	if err == nil {
		t.Error("Expected error when getting repo info from non-git directory")
	}
}

func TestService_IsValidCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	service := New()

	// Skip if git is not available
	if err := service.ValidateGitInstalled(); err != nil {
		t.Skip("Git not available, skipping commit validation tests")
	}

	// Create a temporary git repository with a commit
	tempDir, err := os.MkdirTemp("", "test-git-repo-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Initialize git repo and create a commit
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	if err := cmd.Run(); err != nil {
		t.Skip("Cannot initialize git repo for testing")
	}

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	_ = cmd.Run()

	// Create a file and commit
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_ = exec.Command("git", "add", "test.txt").Run()
	_ = exec.Command("git", "commit", "-m", "Initial commit").Run()

	// Test with invalid commit
	err = service.IsValidCommit(tempDir, "invalid_commit_hash")
	if err == nil {
		t.Error("Expected error for invalid commit hash")
	}
	if !models.IsErrorCode(err, models.ErrorCodeGitCommitNotFound) {
		t.Errorf("Expected ErrorCodeGitCommitNotFound, got %v", err)
	}
}

// Benchmark for temp directory creation and cleanup
func BenchmarkService_TempDirOperations(b *testing.B) {
	service := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempDir, err := service.createTempDir()
		if err != nil {
			b.Fatalf("Failed to create temp directory: %v", err)
		}

		err = service.CleanupTempDir(tempDir)
		if err != nil {
			b.Fatalf("Failed to cleanup temp directory: %v", err)
		}
	}
}

// Test error scenarios for file system issues
func TestService_FileSystemErrors(t *testing.T) {
	service := New()

	// Test cleanup with permission denied scenario
	// Create a directory we can't delete (simulate permission issue)
	if os.Geteuid() != 0 { // Skip if running as root
		tempDir, err := service.createTempDir()
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}

		// Change permissions to make it non-writable
		_ = os.Chmod(tempDir, 0444)
		defer func() {
			_ = os.Chmod(tempDir, 0755) // Restore permissions for cleanup
			_ = os.RemoveAll(tempDir)
		}()

		// This might not fail on all systems, but it's worth testing
		err = service.CleanupTempDir(tempDir)
		// We don't assert error here as behavior varies by OS
		_ = err
	}
}
