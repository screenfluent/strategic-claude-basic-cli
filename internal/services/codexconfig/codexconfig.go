package codexconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

// Service provides Codex configuration management functionality
type Service struct{}

// New creates a new codex config service instance
func New() *Service {
	return &Service{}
}

// ProcessCodexConfig is the main entry point for managing .codex/config.toml
func (s *Service) ProcessCodexConfig(targetDir string) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	codexDir := filepath.Join(targetDir, config.CodexDir)
	configPath := filepath.Join(codexDir, config.CodexConfigFile)
	templatePath := filepath.Join(strategicDir, config.CodexConfigTemplateFile)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Template doesn't exist, nothing to do
		return nil
	}

	// Ensure .codex directory exists
	if err := os.MkdirAll(codexDir, config.DirPermissions); err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, codexDir, err)
	}

	// Handle existing config
	if _, err := os.Stat(configPath); err == nil {
		// Backup existing config
		if err := s.backupExistingConfig(configPath); err != nil {
			return fmt.Errorf("failed to backup existing config: %w", err)
		}
	}

	// Copy template to config.toml
	if err := s.copyTemplate(templatePath, configPath); err != nil {
		return fmt.Errorf("failed to copy config template: %w", err)
	}

	return nil
}

// backupExistingConfig creates a timestamped backup of existing config.toml
func (s *Service) backupExistingConfig(configPath string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(
		filepath.Dir(configPath),
		config.CodexConfigBackupPrefix+timestamp+".toml",
	)

	// Read existing file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	// Write backup
	return os.WriteFile(backupPath, data, config.FilePermissions)
}

// copyTemplate copies the template file to the config location
func (s *Service) copyTemplate(templatePath, configPath string) error {
	// Read template file
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Write to config location
	return os.WriteFile(configPath, data, config.FilePermissions)
}

// RemoveCodexConfig removes the config.toml file and backups
func (s *Service) RemoveCodexConfig(targetDir string) error {
	codexDir := filepath.Join(targetDir, config.CodexDir)
	configPath := filepath.Join(codexDir, config.CodexConfigFile)

	// Remove main config file
	if _, err := os.Stat(configPath); err == nil {
		if err := os.Remove(configPath); err != nil {
			return models.NewFileSystemError(models.ErrorCodeFileSystemError, configPath, err)
		}
	}

	// Remove backup files
	pattern := filepath.Join(codexDir, config.CodexConfigBackupPrefix+"*.toml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	for _, backupFile := range matches {
		if err := os.Remove(backupFile); err != nil {
			// Log warning but continue
			fmt.Printf("Warning: Failed to remove backup file %s: %v\n", backupFile, err)
		}
	}

	return nil
}

// ValidateCodexConfig checks if the config.toml file exists and is readable
func (s *Service) ValidateCodexConfig(targetDir string) error {
	codexDir := filepath.Join(targetDir, config.CodexDir)
	configPath := filepath.Join(codexDir, config.CodexConfigFile)

	// Check if file exists
	info, err := os.Stat(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.NewAppError(
				models.ErrorCodeValidationFailed,
				"Codex config.toml file does not exist",
				nil,
			)
		}
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, configPath, err)
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return models.NewAppError(
			models.ErrorCodeValidationFailed,
			"config.toml exists but is not a regular file",
			nil,
		)
	}

	// Try to read the file to ensure it's accessible
	if _, err := os.ReadFile(configPath); err != nil {
		return models.NewFileSystemError(models.ErrorCodeFileSystemError, configPath, err)
	}

	return nil
}