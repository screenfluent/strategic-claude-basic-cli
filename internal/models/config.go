package models

import (
	"time"

	"strategic-claude-basic-cli/internal/templates"
)

// InstallConfig holds configuration options for installation operations
type InstallConfig struct {
	// Target directory for installation
	TargetDir string

	// Template selection
	TemplateID string // ID of the template to install

	// Installation behavior flags
	Force         bool   // Force installation, overwriting existing files
	ForceCore     bool   // Update only core framework files, preserving user content
	SkipConfirm   bool   // Skip confirmation prompts (--yes flag)
	NoBackup      bool   // Skip creating backups of existing files
	DryRun        bool   // Show what would be done without making changes
	Verbose       bool   // Enable verbose output
	GitignoreMode string // Gitignore behavior: "track", "all", or "non-user"

	// Optional custom backup directory
	BackupDir string

	// Timeout for git operations
	GitTimeout time.Duration
}

// CleanConfig holds configuration options for cleanup operations
type CleanConfig struct {
	// Target directory for cleanup
	TargetDir string

	// Cleanup behavior flags
	Force   bool // Force cleanup without confirmation
	Verbose bool // Enable verbose output
	DryRun  bool // Show what would be done without making changes

	// Preserve user content during cleanup
	PreserveUserContent bool
}

// NewInstallConfig creates a new InstallConfig with default values
func NewInstallConfig(targetDir string) *InstallConfig {
	return &InstallConfig{
		TargetDir:     targetDir,
		TemplateID:    templates.DefaultTemplateID,
		Force:         false,
		ForceCore:     false,
		SkipConfirm:   false,
		NoBackup:      false,
		DryRun:        false,
		Verbose:       false,
		GitignoreMode: "track",
		BackupDir:     "",
		GitTimeout:    30 * time.Second,
	}
}

// NewCleanConfig creates a new CleanConfig with default values
func NewCleanConfig(targetDir string) *CleanConfig {
	return &CleanConfig{
		TargetDir:           targetDir,
		Force:               false,
		Verbose:             false,
		DryRun:              false,
		PreserveUserContent: true,
	}
}

// IsSelectiveUpdate returns true if this is a selective core-only update
func (c *InstallConfig) IsSelectiveUpdate() bool {
	return c.ForceCore && !c.Force
}

// ShouldCreateBackup returns true if backups should be created
func (c *InstallConfig) ShouldCreateBackup() bool {
	return !c.NoBackup && !c.DryRun
}

// ShouldPromptUser returns true if user prompts should be shown
func (c *InstallConfig) ShouldPromptUser() bool {
	return !c.SkipConfirm && !c.DryRun
}

// Validate checks that the configuration is valid
func (c *InstallConfig) Validate() error {
	if c.TargetDir == "" {
		return NewAppError(ErrorCodeInvalidPath, "target directory cannot be empty", nil)
	}

	// Validate template ID
	if c.TemplateID == "" {
		return NewAppError(ErrorCodeInvalidConfiguration, "template ID cannot be empty", nil)
	}

	if err := templates.ValidateTemplateID(c.TemplateID); err != nil {
		return NewAppError(ErrorCodeInvalidConfiguration, "invalid template ID: "+c.TemplateID, err)
	}

	// Both force and force-core cannot be true at the same time
	if c.Force && c.ForceCore {
		return NewAppError(ErrorCodeInvalidConfiguration, "cannot specify both --force and --force-core flags", nil)
	}

	// Validate gitignore mode
	validModes := []string{"track", "all", "non-user"}
	validMode := false
	for _, mode := range validModes {
		if c.GitignoreMode == mode {
			validMode = true
			break
		}
	}
	if !validMode {
		return NewAppError(ErrorCodeInvalidConfiguration, "invalid gitignore mode: "+c.GitignoreMode, nil)
	}

	return nil
}

// GetTemplate returns the template configuration for this install
func (c *InstallConfig) GetTemplate() (templates.Template, error) {
	return templates.GetTemplate(c.TemplateID)
}
