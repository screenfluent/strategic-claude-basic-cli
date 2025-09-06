package cleaner

import (
	"fmt"
	"os"
	"path/filepath"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/filesystem"
	"strategic-claude-basic-cli/internal/services/status"
	"strategic-claude-basic-cli/internal/services/symlink"
)

// Service handles cleanup operations for Strategic Claude Basic installations
type Service struct {
	filesystemService *filesystem.Service
	symlinkService    *symlink.Service
	statusService     *status.Service
}

// New creates a new cleaner service instance
func New() *Service {
	return &Service{
		filesystemService: filesystem.New(),
		symlinkService:    symlink.New(),
		statusService:     status.NewService(),
	}
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	// What was removed
	RemovedDirectory bool     `json:"removed_directory"`
	RemovedSymlinks  []string `json:"removed_symlinks"`

	// What was preserved
	PreservedFiles []string `json:"preserved_files"`

	// Empty directories cleaned up
	CleanedDirectories []string `json:"cleaned_directories"`

	// Issues encountered
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`

	// Status
	Success bool `json:"success"`
}

// RemoveInstallation performs a complete cleanup of Strategic Claude Basic installation
func (s *Service) RemoveInstallation(targetDir string) (*CleanupResult, error) {
	if targetDir == "" {
		return nil, models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	result := &CleanupResult{
		RemovedSymlinks:    make([]string, 0),
		PreservedFiles:     make([]string, 0),
		CleanedDirectories: make([]string, 0),
		Warnings:           make([]string, 0),
		Errors:             make([]string, 0),
		Success:            false,
	}

	// Get current installation status
	statusInfo, err := s.statusService.CheckInstallation(targetDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get installation status: %v", err))
		return result, err
	}

	// If nothing is installed, return early with success
	if !statusInfo.IsInstalled && !statusInfo.StrategicClaudeDir && !statusInfo.ClaudeDir {
		result.Success = true
		result.Warnings = append(result.Warnings, "No Strategic Claude Basic installation found")
		return result, nil
	}

	// Step 1: Remove symlinks
	if err := s.removeSymlinks(targetDir, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove symlinks: %v", err))
		// Continue with cleanup even if symlinks fail
	}

	// Step 2: Remove Strategic Claude Basic directory
	if err := s.removeStrategicDirectory(targetDir, result); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove Strategic Claude directory: %v", err))
		return result, err
	}

	// Step 3: Clean up empty directories (but preserve user content)
	if err := s.cleanupEmptyDirectories(targetDir, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Warning during directory cleanup: %v", err))
		// Non-fatal error, continue
	}

	// Step 4: Validate cleanup
	if err := s.validateCleanup(targetDir, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Cleanup validation warning: %v", err))
	}

	// Determine overall success
	result.Success = len(result.Errors) == 0

	return result, nil
}

// removeSymlinks removes Strategic Claude Basic symlinks
func (s *Service) removeSymlinks(targetDir string, result *CleanupResult) error {
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()

	for symlinkPath := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Check if symlink exists
		if info, err := os.Lstat(fullSymlinkPath); os.IsNotExist(err) {
			continue // Skip if doesn't exist
		} else if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Could not check symlink %s: %v", fullSymlinkPath, err))
			continue
		} else if info.Mode()&os.ModeSymlink == 0 {
			// Path exists but is not a symlink - preserve it
			result.PreservedFiles = append(result.PreservedFiles, fullSymlinkPath)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Preserving non-symlink file: %s", fullSymlinkPath))
			continue
		}

		// Validate it's a Strategic Claude symlink before removing
		if isStrategicSymlink, err := s.isStrategicClaudeSymlink(fullSymlinkPath); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Could not validate symlink %s: %v", fullSymlinkPath, err))
			continue
		} else if !isStrategicSymlink {
			// Not our symlink - preserve it
			result.PreservedFiles = append(result.PreservedFiles, fullSymlinkPath)
			result.Warnings = append(result.Warnings, fmt.Sprintf("Preserving non-Strategic Claude symlink: %s", fullSymlinkPath))
			continue
		}

		// Remove the Strategic Claude symlink
		if err := os.Remove(fullSymlinkPath); err != nil {
			if os.IsPermission(err) {
				return models.NewFileSystemError(models.ErrorCodePermissionDenied, fullSymlinkPath, err)
			}
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, fullSymlinkPath, err)
		}

		result.RemovedSymlinks = append(result.RemovedSymlinks, symlinkPath)
	}

	return nil
}

// removeStrategicDirectory removes the .strategic-claude-basic directory
func (s *Service) removeStrategicDirectory(targetDir string, result *CleanupResult) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Check if directory exists
	if _, err := os.Stat(strategicDir); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	// Use filesystem service for safe removal
	if err := s.filesystemService.RemoveStrategicClaudeBasic(targetDir); err != nil {
		return err
	}

	result.RemovedDirectory = true
	return nil
}

// cleanupEmptyDirectories removes empty .claude subdirectories if they contain no user content
func (s *Service) cleanupEmptyDirectories(targetDir string, result *CleanupResult) error {
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)

	// Check if .claude directory exists
	if _, err := os.Stat(claudeDir); os.IsNotExist(err) {
		return nil // Nothing to clean
	}

	// Check each subdirectory (agents, commands, hooks)
	subdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(claudeDir, subdir)

		if err := s.cleanupEmptySubdirectory(subdirPath, result); err != nil {
			return err
		}
	}

	// Check if .claude directory itself is now empty
	if err := s.cleanupEmptySubdirectory(claudeDir, result); err != nil {
		return err
	}

	return nil
}

// cleanupEmptySubdirectory removes a subdirectory if it's empty
func (s *Service) cleanupEmptySubdirectory(dirPath string, result *CleanupResult) error {
	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return nil // Nothing to clean
	}

	// Read directory contents
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, dirPath, err)
	}

	// If directory is empty, remove it
	if len(entries) == 0 {
		if err := os.Remove(dirPath); err != nil {
			if os.IsPermission(err) {
				return models.NewFileSystemError(models.ErrorCodePermissionDenied, dirPath, err)
			}
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, dirPath, err)
		}
		result.CleanedDirectories = append(result.CleanedDirectories, dirPath)
	} else {
		// Directory has content - preserve it and log what's there
		for _, entry := range entries {
			preservedPath := filepath.Join(dirPath, entry.Name())
			result.PreservedFiles = append(result.PreservedFiles, preservedPath)
		}
	}

	return nil
}

// validateCleanup verifies that the cleanup was successful
func (s *Service) validateCleanup(targetDir string, result *CleanupResult) error {
	// Get post-cleanup status
	statusInfo, err := s.statusService.CheckInstallation(targetDir)
	if err != nil {
		return fmt.Errorf("failed to validate cleanup: %w", err)
	}

	// Check that Strategic Claude directory is gone
	if statusInfo.StrategicClaudeDir {
		result.Warnings = append(result.Warnings, "Strategic Claude directory still exists after cleanup")
	}

	// Check that our symlinks are gone
	for _, symlink := range statusInfo.Symlinks {
		if symlink.Valid && symlink.Exists {
			result.Warnings = append(result.Warnings, fmt.Sprintf("Strategic Claude symlink still exists: %s", symlink.Path))
		}
	}

	return nil
}

// isStrategicClaudeSymlink checks if a symlink points to a Strategic Claude target
func (s *Service) isStrategicClaudeSymlink(symlinkPath string) (bool, error) {
	// Read the symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return false, models.NewFileSystemError(models.ErrorCodeFileSystemError, symlinkPath, err)
	}

	// Check if target contains strategic-claude-basic path components
	expectedTargets := config.GetRequiredSymlinks()
	for _, expectedTarget := range expectedTargets {
		if target == expectedTarget {
			return true, nil
		}
	}

	return false, nil
}

// HandlePartialInstallation specifically handles cleanup of broken or incomplete installations
func (s *Service) HandlePartialInstallation(targetDir string) (*CleanupResult, error) {
	result := &CleanupResult{
		RemovedSymlinks:    make([]string, 0),
		PreservedFiles:     make([]string, 0),
		CleanedDirectories: make([]string, 0),
		Warnings:           make([]string, 0),
		Errors:             make([]string, 0),
		Success:            false,
	}

	// Get status to understand what's broken
	statusInfo, err := s.statusService.CheckInstallation(targetDir)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get status: %v", err))
		return result, err
	}

	result.Warnings = append(result.Warnings, "Handling partial installation cleanup")

	// Remove any broken or invalid symlinks
	for _, symlink := range statusInfo.Symlinks {
		if symlink.Exists && !symlink.Valid {
			if err := os.Remove(symlink.Path); err != nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Could not remove broken symlink %s: %v", symlink.Path, err))
			} else {
				result.RemovedSymlinks = append(result.RemovedSymlinks, symlink.Name)
			}
		}
	}

	// Remove Strategic Claude directory if it exists
	if statusInfo.StrategicClaudeDir {
		if err := s.removeStrategicDirectory(targetDir, result); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to remove Strategic Claude directory: %v", err))
			return result, err
		}
	}

	// Clean up empty directories
	if err := s.cleanupEmptyDirectories(targetDir, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Warning during directory cleanup: %v", err))
	}

	result.Success = len(result.Errors) == 0
	return result, nil
}
