package models

import "time"

// InstallConfig holds configuration options for installation operations
type InstallConfig struct {
	// Target directory for installation
	TargetDir string

	// Installation behavior flags
	Force       bool // Force installation, overwriting existing files
	ForceCore   bool // Update only core framework files, preserving user content
	SkipConfirm bool // Skip confirmation prompts (--yes flag)
	NoBackup    bool // Skip creating backups of existing files
	DryRun      bool // Show what would be done without making changes
	Verbose     bool // Enable verbose output

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
		TargetDir:   targetDir,
		Force:       false,
		ForceCore:   false,
		SkipConfirm: false,
		NoBackup:    false,
		DryRun:      false,
		Verbose:     false,
		BackupDir:   "",
		GitTimeout:  30 * time.Second,
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
