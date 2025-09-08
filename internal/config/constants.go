package config

import (
	"strings"
	"time"
)

const (
	// Repository configuration (deprecated - use templates package instead)
	// Kept for backward compatibility during migration
	DefaultRepoURL = "https://github.com/Fomo-Driven-Development/strategic-claude-basic-template.git"
	FixedCommit    = "23bcdc7088bff19c6e51660337ed00ccd3f28098" // Pinned commit hash, updated with releases
	Branch         = "main"

	// Directory configuration
	DefaultTargetDir        = "."
	TempDirPrefix           = "strategic-claude-basic-template-"
	StrategicClaudeBasicDir = ".strategic-claude-basic"
	ClaudeDir               = ".claude"
	BackupDirPrefix         = "strategic-claude-basic-backup-"

	// Framework directory structure within .strategic-claude-basic/
	CoreDir      = "core"
	GuidesDir    = "guides"
	TemplatesDir = "templates"
	ConfigDir    = "config"

	// User directories that are preserved during updates
	ArchivesDir   = "archives"
	IssuesDir     = "issues"
	PlanDir       = "plan"
	ProductDir    = "product"
	ResearchDir   = "research"
	SummaryDir    = "summary"
	ToolsDir      = "tools"
	ValidationDir = "validation"

	// Core subdirectories
	AgentsDir   = "agents"
	CommandsDir = "commands"
	HooksDir    = "hooks"

	// Symlink targets within .claude/
	ClaudeCommandsDir = "commands"
	ClaudeConfigFile  = "CLAUDE.md"

	// Settings files
	SettingsTemplateFile = "templates/hooks/settings.template.json"
	ClaudeSettingsFile   = "settings.json"
	SettingsBackupPrefix = "settings-backup-"

	// Directories that are replaced during updates
	ReplacedDirs = "core/,guides/,templates/"

	// User directories preserved during updates
	UserPreservedDirs = "archives/,issues/,plan/,product/,research/,summary/,tools/,validation/"

	// Default timeout values
	DefaultGitTimeout     = 30 * time.Second
	DefaultNetworkTimeout = 30 * time.Second

	// Validation constants
	MaxPathLength       = 260 // Windows compatibility
	MaxDirectoryNameLen = 255
	MinDirectoryNameLen = 1

	// Application metadata
	AppName        = "strategic-claude-basic-cli"
	AppDescription = "CLI tool for managing Strategic Claude Basic framework installations"
	ConfigFileName = "strategic-claude-basic.json"

	// Template metadata file
	TemplateInfoFile = ".template-info"

	// Exit codes
	ExitSuccess           = 0
	ExitGeneralError      = 1
	ExitValidationError   = 2
	ExitPermissionError   = 3
	ExitNetworkError      = 4
	ExitUserCancellation  = 5
	ExitInstallationError = 6
	ExitAlreadyInstalled  = 7
	ExitNotInstalled      = 8

	// File permissions
	DirPermissions  = 0755
	FilePermissions = 0644

	// Backup configuration
	MaxBackupAge = 30 * 24 * time.Hour // 30 days
	MaxBackups   = 10                  // Maximum number of backups to keep
)

// GetFrameworkDirectories returns the list of framework directories
func GetFrameworkDirectories() []string {
	return []string{
		CoreDir,
		GuidesDir,
		TemplatesDir,
	}
}

// GetCoreDirectories returns directories that are replaced during updates
func GetCoreDirectories() []string {
	return []string{
		CoreDir,
		GuidesDir,
		TemplatesDir,
	}
}

// GetUserPreservedDirectories returns directories that should be preserved during selective updates
func GetUserPreservedDirectories() []string {
	return []string{
		ArchivesDir,
		IssuesDir,
		PlanDir,
		ProductDir,
		ResearchDir,
		SummaryDir,
		ToolsDir,
		ValidationDir,
	}
}

// GetRequiredSymlinks returns the symlinks that should be created
func GetRequiredSymlinks() map[string]string {
	return map[string]string{
		"agents/strategic":   "../../" + StrategicClaudeBasicDir + "/core/agents",
		"commands/strategic": "../../" + StrategicClaudeBasicDir + "/core/commands",
		"hooks/strategic":    "../../" + StrategicClaudeBasicDir + "/core/hooks",
	}
}

// GetBackupDirName generates a backup directory name with timestamp
func GetBackupDirName() string {
	return BackupDirPrefix + time.Now().Format("20060102-150405")
}

// IsUserPreservedPath checks if a path should be preserved during selective updates
func IsUserPreservedPath(path string) bool {
	for _, preserved := range GetUserPreservedDirectories() {
		if path == preserved || strings.HasPrefix(path, preserved+"/") {
			return true
		}
	}
	return false
}

// IsCoreFile checks if a file is part of directories that get replaced during updates
func IsCoreFile(path string) bool {
	for _, coreDir := range GetCoreDirectories() {
		if path == coreDir || strings.HasPrefix(path, coreDir+"/") {
			return true
		}
	}
	return false
}
