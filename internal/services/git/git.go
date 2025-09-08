package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
)

// Service handles git operations for the Strategic Claude Basic CLI
type Service struct {
	timeout time.Duration
}

// New creates a new git service instance
func New() *Service {
	return &Service{
		timeout: config.DefaultGitTimeout,
	}
}

// ValidateGitInstalled checks if git is available in the system
func (s *Service) ValidateGitInstalled() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return models.NewAppError(
			models.ErrorCodeGitNotFound,
			"Git is not installed or not available in PATH",
			err,
		)
	}
	return nil
}

// CloneRepository clones a git repository to a temporary directory and checks out a specific commit
func (s *Service) CloneRepository(url, commit string) (string, error) {
	return s.CloneRepositoryWithBranch(url, "", commit)
}

// CloneRepositoryWithBranch clones a git repository with optional branch specification and checks out a specific commit
func (s *Service) CloneRepositoryWithBranch(url, branch, commit string) (string, error) {
	if err := s.ValidateGitInstalled(); err != nil {
		return "", err
	}

	tempDir, err := s.createTempDir()
	if err != nil {
		return "", models.NewAppError(
			models.ErrorCodeFileSystemError,
			"Failed to create temporary directory",
			err,
		)
	}

	// Attempt clone with retries for network issues
	var cloneErr error
	for attempt := 1; attempt <= 3; attempt++ {
		cloneErr = s.cloneWithRetry(url, branch, tempDir, attempt)
		if cloneErr == nil {
			break
		}

		if attempt < 3 {
			time.Sleep(time.Second * time.Duration(attempt))
		}
	}

	if cloneErr != nil {
		_ = s.CleanupTempDir(tempDir) // Best effort cleanup
		return "", cloneErr
	}

	// Checkout specific commit
	if err := s.checkoutCommit(tempDir, commit); err != nil {
		_ = s.CleanupTempDir(tempDir) // Best effort cleanup
		return "", err
	}

	return tempDir, nil
}

// CleanupTempDir removes the temporary directory and its contents
func (s *Service) CleanupTempDir(path string) error {
	if path == "" {
		return nil
	}

	// Safety check to ensure we're only deleting temp directories
	if !strings.Contains(path, config.TempDirPrefix) {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Refusing to delete directory that doesn't appear to be a temp directory: %s", path),
			nil,
		)
	}

	err := os.RemoveAll(path)
	if err != nil {
		return models.NewAppError(
			models.ErrorCodeFileSystemError,
			fmt.Sprintf("Failed to clean up temporary directory: %s", path),
			err,
		)
	}

	return nil
}

// createTempDir creates a temporary directory for git operations
func (s *Service) createTempDir() (string, error) {
	tempDir, err := os.MkdirTemp("", config.TempDirPrefix)
	if err != nil {
		return "", err
	}
	return tempDir, nil
}

// cloneWithRetry performs a git clone operation with error handling
func (s *Service) cloneWithRetry(url, branch, tempDir string, attempt int) error {
	var cmd *exec.Cmd

	if branch != "" {
		// Clone specific branch
		cmd = exec.Command("git", "clone", "-b", branch, url, tempDir)
	} else {
		// Clone default branch
		cmd = exec.Command("git", "clone", url, tempDir)
	}

	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		if attempt == 3 { // Last attempt, return detailed error
			branchInfo := ""
			if branch != "" {
				branchInfo = fmt.Sprintf(" (branch: %s)", branch)
			}
			return models.NewAppError(
				models.ErrorCodeGitCloneError,
				fmt.Sprintf("Failed to clone repository %s%s after %d attempts", url, branchInfo, attempt),
				err,
			)
		}
		return err // Will be retried
	}

	return nil
}

// checkoutCommit checks out a specific commit in the cloned repository
func (s *Service) checkoutCommit(repoPath, commit string) error {
	cmd := exec.Command("git", "checkout", commit)
	cmd.Dir = repoPath
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return models.NewAppError(
			models.ErrorCodeGitCheckoutError,
			fmt.Sprintf("Failed to checkout commit %s", commit),
			err,
		)
	}

	return nil
}

// GetRepoInfo returns information about the repository state
func (s *Service) GetRepoInfo(repoPath string) (map[string]string, error) {
	info := make(map[string]string)

	// Get current commit hash
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return nil, models.NewAppError(
			models.ErrorCodeGitError,
			"Failed to get current commit hash",
			err,
		)
	}
	info["commit"] = strings.TrimSpace(string(output))

	// Get remote URL
	cmd = exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoPath
	output, err = cmd.Output()
	if err != nil {
		return nil, models.NewAppError(
			models.ErrorCodeGitError,
			"Failed to get remote URL",
			err,
		)
	}
	info["remote_url"] = strings.TrimSpace(string(output))

	return info, nil
}

// IsValidCommit checks if a commit hash exists in the repository
func (s *Service) IsValidCommit(repoPath, commit string) error {
	cmd := exec.Command("git", "cat-file", "-e", commit)
	cmd.Dir = repoPath

	err := cmd.Run()
	if err != nil {
		return models.NewAppError(
			models.ErrorCodeGitCommitNotFound,
			fmt.Sprintf("Commit %s not found in repository", commit),
			err,
		)
	}

	return nil
}
