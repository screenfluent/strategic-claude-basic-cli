package script

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

// Service handles script operations for the Strategic Claude Basic CLI
type Service struct{}

// New creates a new script service instance
func New() *Service {
	return &Service{}
}

// CopyScript copies a script from the template source directory to the target directory
func (s *Service) CopyScript(sourceDir, targetDir, scriptName string) error {
	if sourceDir == "" || targetDir == "" || scriptName == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Source directory, target directory, and script name cannot be empty",
			nil,
		)
	}

	// Source script path
	sourcePath := filepath.Join(sourceDir, scriptName)

	// Check if source script exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil // Script doesn't exist, not an error
	} else if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourcePath, err)
	}

	// Target script path
	targetPath := filepath.Join(targetDir, scriptName)

	// Copy the script file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, sourcePath, err)
	}
	defer sourceFile.Close()

	targetFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, targetPath, err)
	}
	defer targetFile.Close()

	// Copy file contents
	if _, err := targetFile.ReadFrom(sourceFile); err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, targetPath, err)
	}

	return nil
}

// ExecuteScript executes a script in the target directory
func (s *Service) ExecuteScript(targetDir, scriptName string) error {
	if targetDir == "" || scriptName == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory and script name cannot be empty",
			nil,
		)
	}

	scriptPath := filepath.Join(targetDir, scriptName)

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil // Script doesn't exist, not an error
	} else if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, scriptPath, err)
	}

	// Make sure script is executable
	if err := os.Chmod(scriptPath, 0755); err != nil {
		return models.NewFileSystemError(models.ErrorCodePermissionDenied, scriptPath, err)
	}

	// Execute the script in the target directory
	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = targetDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return models.NewAppError(
			models.ErrorCodeInstallationFailed,
			fmt.Sprintf("Script execution failed: %s", scriptName),
			err,
		)
	}

	return nil
}

// RemoveScript removes a script from the target directory
func (s *Service) RemoveScript(targetDir, scriptName string) error {
	if targetDir == "" || scriptName == "" {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"Target directory and script name cannot be empty",
			nil,
		)
	}

	scriptPath := filepath.Join(targetDir, scriptName)

	// Check if script exists
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil // Script doesn't exist, nothing to remove
	} else if err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, scriptPath, err)
	}

	// Remove the script
	if err := os.Remove(scriptPath); err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, scriptPath, err)
	}

	return nil
}

// ScriptExists checks if a script exists in the source directory
func (s *Service) ScriptExists(sourceDir, scriptName string) bool {
	if sourceDir == "" || scriptName == "" {
		return false
	}

	scriptPath := filepath.Join(sourceDir, scriptName)
	_, err := os.Stat(scriptPath)
	return err == nil
}

// GetScriptPath returns the full path to a script in the target directory
func (s *Service) GetScriptPath(targetDir, scriptName string) string {
	if targetDir == "" || scriptName == "" {
		return ""
	}
	return filepath.Join(targetDir, scriptName)
}
