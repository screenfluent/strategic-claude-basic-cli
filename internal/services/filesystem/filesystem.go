package filesystem

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/utils"
)

// Service handles file system operations for the Strategic Claude Basic CLI
type Service struct {
	pathValidator *utils.PathValidator
}

// New creates a new filesystem service instance
func New() *Service {
	return &Service{
		pathValidator: utils.NewPathValidator(),
	}
}

// DirectoryOperations provides directory manipulation functions

// CreateDirectory creates a directory with proper permissions, including parent directories
func (s *Service) CreateDirectory(path string) error {
	if path == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Directory path cannot be empty",
			nil,
		)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, path, err)
	}

	// Check if directory already exists
	if info, err := os.Stat(absPath); err == nil {
		if info.IsDir() {
			return nil // Already exists and is a directory
		}
		return models.NewFileSystemError(
			models.ErrorCodeFileAlreadyExists,
			absPath,
			fmt.Errorf("path exists but is not a directory"),
		)
	}

	// Create directory with proper permissions
	err = os.MkdirAll(absPath, config.DirPermissions)
	if err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, absPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, absPath, err)
	}

	return nil
}

// RemoveStrategicClaudeBasic removes only the .strategic-claude-basic directory
func (s *Service) RemoveStrategicClaudeBasic(targetDir string) error {
	if targetDir == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory cannot be empty",
			nil,
		)
	}

	// Build the exact path to .strategic-claude-basic
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Resolve to absolute path for validation
	absPath, err := filepath.Abs(strategicDir)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, strategicDir, err)
	}

	// Validate that we're removing what we expect
	if !strings.HasSuffix(absPath, config.StrategicClaudeBasicDir) {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Path does not end with expected directory name: %s", absPath),
			nil,
		)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	// Remove the strategic-claude-basic directory
	err = os.RemoveAll(absPath)
	if err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, absPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, absPath, err)
	}

	return nil
}

// RemoveSymlinks removes only the known Strategic Claude Basic symlinks
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
		err := os.Remove(fullSymlinkPath)
		if err != nil {
			if os.IsPermission(err) {
				return models.NewFileSystemError(models.ErrorCodePermissionDenied, fullSymlinkPath, err)
			}
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, fullSymlinkPath, err)
		}
	}

	return nil
}

// RemoveBackup removes a specific backup directory with validation
func (s *Service) RemoveBackup(targetDir, backupName string) error {
	if targetDir == "" || backupName == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory and backup name cannot be empty",
			nil,
		)
	}

	// Validate backup name follows expected pattern
	if !strings.HasPrefix(backupName, config.BackupDirPrefix) {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Backup name must start with %s, got: %s", config.BackupDirPrefix, backupName),
			nil,
		)
	}

	// Build path to backup directory
	backupPath := filepath.Join(targetDir, backupName)

	// Resolve to absolute path for validation
	absPath, err := filepath.Abs(backupPath)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, backupPath, err)
	}

	// Additional safety check - ensure we're in the expected target directory
	expectedParent, err := filepath.Abs(targetDir)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, targetDir, err)
	}

	if filepath.Dir(absPath) != expectedParent {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Backup path is not in expected parent directory: %s", absPath),
			nil,
		)
	}

	// Check if directory exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil // Already doesn't exist
	}

	// Remove the backup directory
	err = os.RemoveAll(absPath)
	if err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, absPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, absPath, err)
	}

	return nil
}

// BackupDirectory creates a backup of an existing directory
func (s *Service) BackupDirectory(sourcePath, backupPath string) error {
	if sourcePath == "" || backupPath == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Source and backup paths cannot be empty",
			nil,
		)
	}

	// Resolve paths
	sourceAbs, err := filepath.Abs(sourcePath)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, sourcePath, err)
	}

	backupAbs, err := filepath.Abs(backupPath)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeInvalidPath, backupPath, err)
	}

	// Check if source exists and is directory
	sourceInfo, err := os.Stat(sourceAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return models.NewFileSystemError(models.ErrorCodeDirectoryNotFound, sourceAbs, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourceAbs, err)
	}

	if !sourceInfo.IsDir() {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Source path is not a directory: %s", sourceAbs),
			nil,
		)
	}

	// Check if backup path already exists
	if _, err := os.Stat(backupAbs); err == nil {
		return models.NewFileSystemError(
			models.ErrorCodeFileAlreadyExists,
			backupAbs,
			fmt.Errorf("backup directory already exists"),
		)
	}

	// Copy directory to backup location
	return s.CopyDirectory(sourceAbs, backupAbs)
}

// EnsureDirectoryStructure creates the Strategic Claude Basic directory structure
func (s *Service) EnsureDirectoryStructure(targetDir string) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Create main directory
	if err := s.CreateDirectory(strategicDir); err != nil {
		return err
	}

	// Create framework directories
	frameworkDirs := config.GetFrameworkDirectories()
	for _, dir := range frameworkDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if err := s.CreateDirectory(dirPath); err != nil {
			return err
		}
	}

	// Create user preserved directories
	userDirs := config.GetUserPreservedDirectories()
	for _, dir := range userDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if err := s.CreateDirectory(dirPath); err != nil {
			return err
		}
	}

	// Create core subdirectories
	coreDir := filepath.Join(strategicDir, config.CoreDir)
	coreSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range coreSubdirs {
		subdirPath := filepath.Join(coreDir, subdir)
		if err := s.CreateDirectory(subdirPath); err != nil {
			return err
		}
	}

	return nil
}

// File Operations

// CopyFile copies a single file with permission preservation
func (s *Service) CopyFile(sourcePath, destPath string) error {
	if sourcePath == "" || destPath == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Source and destination paths cannot be empty",
			nil,
		)
	}

	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.NewFileSystemError(models.ErrorCodeDirectoryNotFound, sourcePath, err)
		}
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, sourcePath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourcePath, err)
	}
	defer sourceFile.Close()

	// Get source file info for permissions
	sourceInfo, err := sourceFile.Stat()
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourcePath, err)
	}

	// Create destination directory if it doesn't exist
	destDir := filepath.Dir(destPath)
	if err := s.CreateDirectory(destDir); err != nil {
		return err
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		if os.IsPermission(err) {
			return models.NewFileSystemError(models.ErrorCodePermissionDenied, destPath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, destPath, err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, destPath, err)
	}

	// Set permissions to match source
	err = os.Chmod(destPath, sourceInfo.Mode())
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodePermissionDenied, destPath, err)
	}

	return nil
}

// CopyDirectory copies an entire directory tree
func (s *Service) CopyDirectory(sourcePath, destPath string) error {
	if sourcePath == "" || destPath == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Source and destination paths cannot be empty",
			nil,
		)
	}

	// Get source directory info
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.NewFileSystemError(models.ErrorCodeDirectoryNotFound, sourcePath, err)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourcePath, err)
	}

	if !sourceInfo.IsDir() {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			fmt.Sprintf("Source path is not a directory: %s", sourcePath),
			nil,
		)
	}

	// Create destination directory
	if err := s.CreateDirectory(destPath); err != nil {
		return err
	}

	// Set permissions to match source
	err = os.Chmod(destPath, sourceInfo.Mode())
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodePermissionDenied, destPath, err)
	}

	// Walk through source directory
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip root directory (already created)
		if path == sourcePath {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, path, err)
		}

		destItemPath := filepath.Join(destPath, relPath)

		switch {
		case info.IsDir():
			// Create directory
			err = os.MkdirAll(destItemPath, info.Mode())
			if err != nil {
				return models.NewFileSystemError(models.ErrorCodeFileSystemError, destItemPath, err)
			}
		case info.Mode()&os.ModeSymlink != 0:
			// Handle symlinks
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return models.NewFileSystemError(models.ErrorCodeFileSystemError, path, err)
			}
			err = os.Symlink(linkTarget, destItemPath)
			if err != nil {
				return models.NewFileSystemError(models.ErrorCodeSymlinkCreationFailed, destItemPath, err)
			}
		default:
			// Copy regular file
			if err := s.CopyFile(path, destItemPath); err != nil {
				return err
			}
		}

		return nil
	})
}

// CopyFrameworkFiles copies only the framework directories (core, guides, templates)
func (s *Service) CopyFrameworkFiles(sourceDir, destDir string) error {
	frameworkDirs := config.GetCoreDirectories()

	for _, dir := range frameworkDirs {
		sourcePath := filepath.Join(sourceDir, dir)
		destPath := filepath.Join(destDir, dir)

		// Check if source directory exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue // Skip if source doesn't have this directory
		}

		// Remove existing framework directory if it exists
		if _, err := os.Stat(destPath); err == nil {
			// Only remove if it's one of the expected framework directories
			expectedFrameworkDir := filepath.Base(destPath)
			isFrameworkDir := false
			for _, fDir := range frameworkDirs {
				if expectedFrameworkDir == fDir {
					isFrameworkDir = true
					break
				}
			}

			if !isFrameworkDir {
				return models.NewAppError(
					models.ErrorCodeValidationFailed,
					fmt.Sprintf("Refusing to remove unexpected directory: %s", destPath),
					nil,
				)
			}

			// Safe to remove framework directory
			err = os.RemoveAll(destPath)
			if err != nil {
				if os.IsPermission(err) {
					return models.NewFileSystemError(models.ErrorCodePermissionDenied, destPath, err)
				}
				return models.NewFileSystemError(models.ErrorCodeFileSystemError, destPath, err)
			}
		}

		// Copy the directory
		if err := s.CopyDirectory(sourcePath, destPath); err != nil {
			return err
		}
	}

	return nil
}

// PreserveUserContent ensures user directories are not overwritten
func (s *Service) PreserveUserContent(targetDir string) error {
	userDirs := config.GetUserPreservedDirectories()
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	for _, dir := range userDirs {
		dirPath := filepath.Join(strategicDir, dir)

		// Create directory if it doesn't exist (but don't overwrite)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			if err := s.CreateDirectory(dirPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Path Operations

// IsSubPath checks if childPath is within parentPath (prevents directory traversal)
func (s *Service) IsSubPath(parentPath, childPath string) (bool, error) {
	parentAbs, err := filepath.Abs(parentPath)
	if err != nil {
		return false, models.NewFileSystemError(models.ErrorCodeInvalidPath, parentPath, err)
	}

	childAbs, err := filepath.Abs(childPath)
	if err != nil {
		return false, models.NewFileSystemError(models.ErrorCodeInvalidPath, childPath, err)
	}

	// Clean paths to handle . and .. properly
	parentClean := filepath.Clean(parentAbs)
	childClean := filepath.Clean(childAbs)

	// Check if child path starts with parent path
	return strings.HasPrefix(childClean, parentClean+string(os.PathSeparator)) || childClean == parentClean, nil
}

// GetRelativePath gets relative path from base to target (for symlinks)
func (s *Service) GetRelativePath(basePath, targetPath string) (string, error) {
	baseAbs, err := filepath.Abs(basePath)
	if err != nil {
		return "", models.NewFileSystemError(models.ErrorCodeInvalidPath, basePath, err)
	}

	targetAbs, err := filepath.Abs(targetPath)
	if err != nil {
		return "", models.NewFileSystemError(models.ErrorCodeInvalidPath, targetPath, err)
	}

	relPath, err := filepath.Rel(baseAbs, targetAbs)
	if err != nil {
		return "", models.NewFileSystemError(models.ErrorCodeFileSystemError, targetPath, err)
	}

	return relPath, nil
}

// Permission Management

// SetFilePermissions sets proper file permissions
func (s *Service) SetFilePermissions(path string) error {
	return os.Chmod(path, config.FilePermissions)
}

// SetDirectoryPermissions sets proper directory permissions
func (s *Service) SetDirectoryPermissions(path string) error {
	return os.Chmod(path, config.DirPermissions)
}

// CheckWritePermission checks if we have write permission to a directory
func (s *Service) CheckWritePermission(path string) error {
	return s.pathValidator.ValidateDirectoryWritable(path)
}

// Helper functions

// GetBackupPath generates a backup path with timestamp
func (s *Service) GetBackupPath(targetDir string) string {
	timestamp := time.Now().Format("20060102-150405")
	backupName := fmt.Sprintf("%s%s", config.BackupDirPrefix, timestamp)
	return filepath.Join(targetDir, backupName)
}

// ApplyGitignoreTemplate applies a gitignore template to a target location
func (s *Service) ApplyGitignoreTemplate(templatePath, targetPath string) error {
	if templatePath == "" || targetPath == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Template and target paths cannot be empty",
			nil,
		)
	}

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		utils.DisplayWarning(fmt.Sprintf("Gitignore template %s not found, skipping", templatePath))
		return nil
	}

	// Read template content
	templateContent, err := s.readFileLines(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read gitignore template: %w", err)
	}

	// Check if target .gitignore exists
	if _, err := os.Stat(targetPath); err == nil {
		// File exists, merge content
		return s.mergeGitignoreContent(targetPath, templateContent)
	}

	// File doesn't exist, create new one
	return s.writeGitignoreContent(targetPath, templateContent)
}

// readFileLines reads a file and returns its lines
func (s *Service) readFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// mergeGitignoreContent merges template content with existing .gitignore
func (s *Service) mergeGitignoreContent(targetPath string, templateLines []string) error {
	// Read existing content
	existingLines, err := s.readFileLines(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read existing .gitignore: %w", err)
	}

	// Create backup of existing .gitignore
	backupPath := targetPath + ".backup"
	if err := s.CopyFile(targetPath, backupPath); err != nil {
		utils.DisplayWarning(fmt.Sprintf("Failed to create backup of .gitignore: %v", err))
	}

	// Merge content with deduplication
	mergedLines := s.deduplicateGitignoreLines(existingLines, templateLines)

	// Write merged content
	return s.writeGitignoreContent(targetPath, mergedLines)
}

// writeGitignoreContent writes gitignore content to target file
func (s *Service) writeGitignoreContent(targetPath string, lines []string) error {
	// Ensure target directory exists
	targetDir := filepath.Dir(targetPath)
	if err := s.CreateDirectory(targetDir); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}
	defer file.Close()

	// Write Strategic Claude Basic header
	if _, err := file.WriteString("# Strategic Claude Basic entries\n"); err != nil {
		return fmt.Errorf("failed to write header to .gitignore: %w", err)
	}

	// Write content
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return fmt.Errorf("failed to write line to .gitignore: %w", err)
		}
	}

	return nil
}

// deduplicateGitignoreLines merges and deduplicates gitignore lines
func (s *Service) deduplicateGitignoreLines(existing, template []string) []string {
	seen := make(map[string]bool)
	var result []string

	// Add existing lines first (preserve user content order)
	for _, line := range existing {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !seen[trimmed] && !strings.HasPrefix(trimmed, "# Strategic Claude Basic") {
			result = append(result, line)
			seen[trimmed] = true
		}
	}

	// Add template lines that aren't duplicates
	for _, line := range template {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !seen[trimmed] {
			result = append(result, line)
			seen[trimmed] = true
		}
	}

	return result
}
