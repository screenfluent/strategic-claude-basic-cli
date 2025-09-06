package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"strategic-claude-basic-cli/internal/models"
)

// PathValidator provides utilities for path validation
type PathValidator struct{}

// NewPathValidator creates a new path validator
func NewPathValidator() *PathValidator {
	return &PathValidator{}
}

// ValidateDirectory validates that a directory path is valid and accessible
func (p *PathValidator) ValidateDirectory(path string) error {
	if path == "" {
		return models.NewValidationError("directory", path, "path cannot be empty")
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, path, err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.NewFileSystemError(models.ErrorCodeDirectoryNotFound, absPath, err)
		}
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, absPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, absPath, err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return models.NewValidationError("directory", absPath, "path is not a directory")
	}

	return nil
}

// ValidateDirectoryWritable checks if a directory is writable
func (p *PathValidator) ValidateDirectoryWritable(path string) error {
	if err := p.ValidateDirectory(path); err != nil {
		return err
	}

	// Try to create a temporary file to test write permissions
	tempFile := filepath.Join(path, ".strategic-claude-basic-test-write")
	file, err := os.Create(tempFile)
	if err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, path, err)
		}
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, path, err)
	}
	file.Close()
	os.Remove(tempFile) // Clean up

	return nil
}

// ValidateDirectoryEmpty checks if a directory is empty (for new installations)
func (p *PathValidator) ValidateDirectoryEmpty(path string) error {
	if err := p.ValidateDirectory(path); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodePermissionDenied, path, err)
	}

	if len(entries) > 0 {
		return models.NewFileSystemError(models.ErrorCodeDirectoryNotEmpty, path, nil)
	}

	return nil
}

// ResolvePath resolves a path to its absolute form
func (p *PathValidator) ResolvePath(path string) (string, error) {
	if path == "" {
		return "", models.NewValidationError("path", path, "path cannot be empty")
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", models.NewFileSystemError(models.ErrorCodeInvalidPath, path, err)
	}

	return absPath, nil
}

// InputValidator provides utilities for input validation
type InputValidator struct{}

// NewInputValidator creates a new input validator
func NewInputValidator() *InputValidator {
	return &InputValidator{}
}

// ValidateInstallConfig validates an installation configuration
func (i *InputValidator) ValidateInstallConfig(config *models.InstallConfig) error {
	if config == nil {
		return models.NewValidationError("config", nil, "configuration cannot be nil")
	}

	if config.TargetDir == "" {
		return models.NewValidationError("target_dir", config.TargetDir, "target directory cannot be empty")
	}

	// Validate conflicting flags
	if config.Force && config.ForceCore {
		return models.NewValidationError("flags", "force+force-core", "cannot use both --force and --force-core flags together")
	}

	if config.NoBackup && config.BackupDir != "" {
		return models.NewValidationError("flags", "no-backup+backup-dir", "cannot specify backup directory when --no-backup is set")
	}

	// Validate backup directory if specified
	if config.BackupDir != "" {
		pathValidator := NewPathValidator()
		if err := pathValidator.ValidateDirectoryWritable(filepath.Dir(config.BackupDir)); err != nil {
			return models.NewValidationError("backup_dir", config.BackupDir, fmt.Sprintf("backup directory parent is not writable: %v", err))
		}
	}

	return nil
}

// ValidateCleanConfig validates a cleanup configuration
func (i *InputValidator) ValidateCleanConfig(config *models.CleanConfig) error {
	if config == nil {
		return models.NewValidationError("config", nil, "configuration cannot be nil")
	}

	if config.TargetDir == "" {
		return models.NewValidationError("target_dir", config.TargetDir, "target directory cannot be empty")
	}

	return nil
}

// FileSystemValidator provides utilities for file system validation
type FileSystemValidator struct{}

// NewFileSystemValidator creates a new file system validator
func NewFileSystemValidator() *FileSystemValidator {
	return &FileSystemValidator{}
}

// ValidateSymlink validates that a symlink exists and points to the correct target
func (f *FileSystemValidator) ValidateSymlink(symlinkPath, expectedTarget string) (*models.SymlinkStatus, error) {
	// Extract a more descriptive name that includes parent directory for Claude symlinks
	// For paths like "/path/to/.claude/agents/strategic", extract "agents/strategic"
	name := filepath.Base(symlinkPath)
	if strings.Contains(symlinkPath, ".claude") {
		parent := filepath.Base(filepath.Dir(symlinkPath))
		if parent != "." && parent != "/" && parent != filepath.Base(symlinkPath) {
			name = parent + "/" + name
		}
	}

	status := &models.SymlinkStatus{
		Name:   name,
		Path:   symlinkPath,
		Valid:  false,
		Target: "",
		Exists: false,
		Error:  "",
	}

	// Check if symlink exists
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		if os.IsNotExist(err) {
			status.Error = "symlink does not exist"
			return status, nil
		}
		status.Error = fmt.Sprintf("failed to stat symlink: %v", err)
		return status, err
	}

	status.Exists = true

	// Check if it's actually a symlink
	if info.Mode()&os.ModeSymlink == 0 {
		status.Error = "path exists but is not a symlink"
		return status, nil
	}

	// Read symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		status.Error = fmt.Sprintf("failed to read symlink target: %v", err)
		return status, err
	}

	status.Target = target

	// Resolve relative paths for comparison
	if !filepath.IsAbs(target) {
		target = filepath.Join(filepath.Dir(symlinkPath), target)
	}
	if !filepath.IsAbs(expectedTarget) {
		expectedTarget = filepath.Join(filepath.Dir(symlinkPath), expectedTarget)
	}

	// Check if target matches expected
	targetAbs, err := filepath.Abs(target)
	if err == nil {
		expectedAbs, err := filepath.Abs(expectedTarget)
		if err == nil && targetAbs == expectedAbs {
			status.Valid = true
		}
	}

	if !status.Valid {
		status.Error = fmt.Sprintf("symlink points to '%s', expected '%s'", target, expectedTarget)
	}

	return status, nil
}

// CheckGitAvailable checks if git is available in the system
func CheckGitAvailable() error {
	_, err := os.Stat("/usr/bin/git")
	if err == nil {
		return nil
	}

	// Try to find git in PATH
	paths := strings.Split(os.Getenv("PATH"), ":")
	for _, path := range paths {
		gitPath := filepath.Join(path, "git")
		if _, err := os.Stat(gitPath); err == nil {
			return nil
		}
	}

	return models.NewGitError(models.ErrorCodeGitNotInstalled, "check git availability", nil)
}

// ValidateDirectoryName validates a directory name for invalid characters
func ValidateDirectoryName(name string) error {
	if name == "" {
		return models.NewValidationError("directory_name", name, "directory name cannot be empty")
	}

	// Check for invalid characters in directory names
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*]`)
	if invalidChars.MatchString(name) {
		return models.NewValidationError("directory_name", name, "directory name contains invalid characters")
	}

	// Check for reserved names on Windows (even though we're on Linux, be safe)
	reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(name)
	for _, res := range reserved {
		if upperName == res {
			return models.NewValidationError("directory_name", name, fmt.Sprintf("'%s' is a reserved directory name", name))
		}
	}

	return nil
}
