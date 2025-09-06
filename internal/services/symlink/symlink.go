package symlink

import (
	"fmt"
	"os"
	"path/filepath"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/utils"
)

// Service handles symlink operations for the Strategic Claude Basic CLI
type Service struct {
	fsValidator *utils.FileSystemValidator
}

// New creates a new symlink service instance
func New() *Service {
	return &Service{
		fsValidator: utils.NewFileSystemValidator(),
	}
}

// CreateSymlinks creates all required symlinks from .claude subdirectories to strategic-claude-basic core
func (s *Service) CreateSymlinks(targetDir string) error {
	if targetDir == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()

	// Ensure .claude directory exists
	if err := s.ensureClaudeDirectoryStructure(claudeDir); err != nil {
		return fmt.Errorf("failed to ensure .claude directory structure: %w", err)
	}

	// Create each required symlink
	for symlinkPath, target := range requiredSymlinks {
		if err := s.createRelativeSymlink(claudeDir, symlinkPath, target); err != nil {
			return fmt.Errorf("failed to create symlink %s: %w", symlinkPath, err)
		}
	}

	return nil
}

// RemoveSymlinks removes all Strategic Claude Basic symlinks from the .claude directory
func (s *Service) RemoveSymlinks(targetDir string) error {
	if targetDir == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()

	for symlinkPath := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Check if symlink exists
		if _, err := os.Lstat(fullSymlinkPath); os.IsNotExist(err) {
			continue // Skip if doesn't exist
		}

		// Remove the symlink
		if err := os.Remove(fullSymlinkPath); err != nil {
			if os.IsPermission(err) {
				return models.NewFileSystemError(models.ErrorCodePermissionDenied, fullSymlinkPath, err)
			}
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, fullSymlinkPath, err)
		}
	}

	return nil
}

// ValidateSymlinks checks all required symlinks and returns their status
func (s *Service) ValidateSymlinks(targetDir string) ([]models.SymlinkStatus, error) {
	if targetDir == "" {
		return nil, models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()
	var statuses []models.SymlinkStatus

	for symlinkPath, expectedTarget := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		status, err := s.fsValidator.ValidateSymlink(fullSymlinkPath, expectedTarget)
		if err != nil {
			// Create a basic status entry for the error
			statuses = append(statuses, models.SymlinkStatus{
				Name:   filepath.Base(symlinkPath),
				Path:   fullSymlinkPath,
				Valid:  false,
				Target: "",
				Exists: false,
				Error:  fmt.Sprintf("Failed to validate symlink: %v", err),
			})
			continue
		}

		if status != nil {
			statuses = append(statuses, *status)
		}
	}

	return statuses, nil
}

// UpdateSymlinks updates existing symlinks to ensure they point to the correct targets
func (s *Service) UpdateSymlinks(targetDir string) error {
	if targetDir == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	// Remove existing symlinks
	if err := s.RemoveSymlinks(targetDir); err != nil {
		return fmt.Errorf("failed to remove existing symlinks: %w", err)
	}

	// Create new symlinks
	if err := s.CreateSymlinks(targetDir); err != nil {
		return fmt.Errorf("failed to create updated symlinks: %w", err)
	}

	return nil
}

// RepairSymlinks fixes any broken or invalid symlinks
func (s *Service) RepairSymlinks(targetDir string) ([]string, error) {
	if targetDir == "" {
		return nil, models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	// Validate current symlinks
	statuses, err := s.ValidateSymlinks(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to validate symlinks: %w", err)
	}

	var repairedSymlinks []string
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()

	// Repair invalid symlinks
	for _, status := range statuses {
		if !status.Valid {
			// Find the corresponding required symlink
			var targetPath string
			var symlinkRelPath string

			for symPath, target := range requiredSymlinks {
				if filepath.Join(claudeDir, symPath) == status.Path {
					targetPath = target
					symlinkRelPath = symPath
					break
				}
			}

			if targetPath != "" {
				// Remove broken symlink
				if status.Exists {
					if err := os.Remove(status.Path); err != nil {
						return repairedSymlinks, models.NewFileSystemError(
							models.ErrorCodeFileSystemError,
							status.Path,
							err,
						)
					}
				}

				// Create new symlink
				if err := s.createRelativeSymlink(claudeDir, symlinkRelPath, targetPath); err != nil {
					return repairedSymlinks, fmt.Errorf("failed to repair symlink %s: %w", symlinkRelPath, err)
				}

				repairedSymlinks = append(repairedSymlinks, symlinkRelPath)
			}
		}
	}

	return repairedSymlinks, nil
}

// Helper methods

// ensureClaudeDirectoryStructure creates the .claude directory and its subdirectories if they don't exist
func (s *Service) ensureClaudeDirectoryStructure(claudeDir string) error {
	// Create main .claude directory
	if err := os.MkdirAll(claudeDir, config.DirPermissions); err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, claudeDir, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, claudeDir, err)
	}

	// Create required subdirectories
	subdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(claudeDir, subdir)
		if err := os.MkdirAll(subdirPath, config.DirPermissions); err != nil {
			if os.IsPermission(err) {
				return models.NewFileSystemError(models.ErrorCodePermissionDenied, subdirPath, err)
			}
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, subdirPath, err)
		}
	}

	return nil
}

// createRelativeSymlink creates a single symlink with proper error handling
func (s *Service) createRelativeSymlink(claudeDir, symlinkPath, target string) error {
	fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

	// Ensure parent directory exists
	parentDir := filepath.Dir(fullSymlinkPath)
	if err := os.MkdirAll(parentDir, config.DirPermissions); err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, parentDir, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, parentDir, err)
	}

	// Remove existing symlink if it exists
	if _, err := os.Lstat(fullSymlinkPath); err == nil {
		if err := os.Remove(fullSymlinkPath); err != nil {
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, fullSymlinkPath, err)
		}
	}

	// Create the symlink
	if err := os.Symlink(target, fullSymlinkPath); err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, fullSymlinkPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeSymlinkCreationFailed, fullSymlinkPath, err)
	}

	return nil
}

// GetSymlinkInfo returns information about a specific symlink
func (s *Service) GetSymlinkInfo(symlinkPath string) (*models.SymlinkStatus, error) {
	// Use Lstat to get symlink info without following the link
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &models.SymlinkStatus{
				Name:   filepath.Base(symlinkPath),
				Path:   symlinkPath,
				Valid:  false,
				Target: "",
				Exists: false,
				Error:  "Symlink does not exist",
			}, nil
		}
		return nil, models.NewFileSystemError(models.ErrorCodeFileSystemError, symlinkPath, err)
	}

	// Check if it's actually a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		return &models.SymlinkStatus{
			Name:   filepath.Base(symlinkPath),
			Path:   symlinkPath,
			Valid:  false,
			Target: "",
			Exists: true,
			Error:  "Path exists but is not a symlink",
		}, nil
	}

	// Get the target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		return &models.SymlinkStatus{
			Name:   filepath.Base(symlinkPath),
			Path:   symlinkPath,
			Valid:  false,
			Target: "",
			Exists: true,
			Error:  fmt.Sprintf("Failed to read symlink target: %v", err),
		}, nil
	}

	// Check if target exists
	targetPath := target
	if !filepath.IsAbs(target) {
		// Resolve relative path
		targetPath = filepath.Join(filepath.Dir(symlinkPath), target)
	}

	targetExists := true
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		targetExists = false
	}

	return &models.SymlinkStatus{
		Name:   filepath.Base(symlinkPath),
		Path:   symlinkPath,
		Valid:  targetExists,
		Target: target,
		Exists: true,
		Error:  "",
	}, nil
}
