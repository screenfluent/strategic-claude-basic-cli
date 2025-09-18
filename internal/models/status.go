package models

import (
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/templates"
)

// InstallationType represents the type of installation operation
type InstallationType string

const (
	InstallationTypeNew       InstallationType = "New Installation"
	InstallationTypeUpdate    InstallationType = "Update Core Only"
	InstallationTypeOverwrite InstallationType = "Full Overwrite"
)

// StatusInfo represents the overall installation status
type StatusInfo struct {
	// Basic installation status
	IsInstalled        bool `json:"is_installed"`
	StrategicClaudeDir bool `json:"strategic_claude_dir_exists"`
	ClaudeDir          bool `json:"claude_dir_exists"`
	CodexDir           bool `json:"codex_dir_exists"`

	// Template information
	InstalledTemplate *templates.TemplateInfo `json:"installed_template,omitempty"`

	// Script detection
	HasPreInstallScript  bool `json:"has_pre_install_script"`
	HasPostInstallScript bool `json:"has_post_install_script"`

	// Detailed component status
	Symlinks      []SymlinkStatus `json:"symlinks"`
	CodexSymlinks []SymlinkStatus `json:"codex_symlinks"`
	Issues        []string        `json:"issues"`

	// Installation metadata (deprecated - use InstalledTemplate instead)
	InstallationDate *time.Time `json:"installation_date,omitempty"`
	Version          string     `json:"version,omitempty"`
	CommitHash       string     `json:"commit_hash,omitempty"`

	// Directory paths
	TargetDir              string `json:"target_dir"`
	StrategicClaudeDirPath string `json:"strategic_claude_dir_path"`
	ClaudeDirPath          string `json:"claude_dir_path"`
	CodexDirPath           string `json:"codex_dir_path"`
}

// SymlinkStatus represents the status of an individual symlink
type SymlinkStatus struct {
	Name   string `json:"name"`            // Name of the symlink (e.g., "core", "guides")
	Path   string `json:"path"`            // Full path to the symlink
	Valid  bool   `json:"valid"`           // Whether the symlink is valid and points to the right target
	Target string `json:"target"`          // Target path the symlink points to
	Exists bool   `json:"exists"`          // Whether the symlink file exists
	Error  string `json:"error,omitempty"` // Error message if validation failed
}

// InstallationPlan represents what will happen during an installation
type InstallationPlan struct {
	// Basic information
	TargetDir        string           `json:"target_dir"`
	InstallationType InstallationType `json:"installation_type"`

	// Template information
	Template templates.Template `json:"template"`

	// Script information
	HasPreInstallScript  bool `json:"has_pre_install_script"`
	HasPostInstallScript bool `json:"has_post_install_script"`

	// File operations
	ExistingFiles []string `json:"existing_files"` // Files that already exist
	WillReplace   []string `json:"will_replace"`   // Files that will be replaced
	WillPreserve  []string `json:"will_preserve"`  // Files that will be preserved
	WillCreate    []string `json:"will_create"`    // New files that will be created

	// Directory operations
	DirectoriesToCreate []string `json:"directories_to_create"`
	SymlinksToCreate    []string `json:"symlinks_to_create"`
	SymlinksToUpdate    []string `json:"symlinks_to_update"`

	// Backup information
	BackupRequired bool   `json:"backup_required"`
	BackupDir      string `json:"backup_dir,omitempty"`

	// Validation results
	HasConflicts bool     `json:"has_conflicts"`
	Warnings     []string `json:"warnings,omitempty"`
	Errors       []string `json:"errors,omitempty"`
}

// NewStatusInfo creates a new StatusInfo for the given target directory
func NewStatusInfo(targetDir string) *StatusInfo {
	return &StatusInfo{
		IsInstalled:            false,
		StrategicClaudeDir:     false,
		ClaudeDir:              false,
		CodexDir:               false,
		Symlinks:               make([]SymlinkStatus, 0),
		CodexSymlinks:          make([]SymlinkStatus, 0),
		Issues:                 make([]string, 0),
		TargetDir:              targetDir,
		StrategicClaudeDirPath: "",
		ClaudeDirPath:          "",
		CodexDirPath:           "",
	}
}

// NewInstallationPlan creates a new InstallationPlan for the given target directory and template
func NewInstallationPlan(targetDir string, installType InstallationType, template templates.Template) *InstallationPlan {
	return &InstallationPlan{
		TargetDir:           targetDir,
		InstallationType:    installType,
		Template:            template,
		ExistingFiles:       make([]string, 0),
		WillReplace:         make([]string, 0),
		WillPreserve:        make([]string, 0),
		WillCreate:          make([]string, 0),
		DirectoriesToCreate: make([]string, 0),
		SymlinksToCreate:    make([]string, 0),
		SymlinksToUpdate:    make([]string, 0),
		BackupRequired:      false,
		BackupDir:           "",
		HasConflicts:        false,
		Warnings:            make([]string, 0),
		Errors:              make([]string, 0),
	}
}

// AddIssue adds an issue to the status info
func (s *StatusInfo) AddIssue(issue string) {
	s.Issues = append(s.Issues, issue)
}

// AddSymlink adds a symlink status to the status info
func (s *StatusInfo) AddSymlink(symlink SymlinkStatus) {
	s.Symlinks = append(s.Symlinks, symlink)
}

// AddCodexSymlink adds a codex symlink status to the status info
func (s *StatusInfo) AddCodexSymlink(symlink SymlinkStatus) {
	s.CodexSymlinks = append(s.CodexSymlinks, symlink)
}

// HasIssues returns true if there are any issues
func (s *StatusInfo) HasIssues() bool {
	return len(s.Issues) > 0
}

// ValidSymlinks returns the number of valid symlinks
func (s *StatusInfo) ValidSymlinks() int {
	count := 0
	for _, symlink := range s.Symlinks {
		if symlink.Valid {
			count++
		}
	}
	return count
}

// ValidCodexSymlinks returns the number of valid Codex symlinks
func (s *StatusInfo) ValidCodexSymlinks() int {
	count := 0
	for _, symlink := range s.CodexSymlinks {
		if symlink.Valid {
			count++
		}
	}
	return count
}

// AddWarning adds a warning to the installation plan
func (p *InstallationPlan) AddWarning(warning string) {
	p.Warnings = append(p.Warnings, warning)
}

// AddError adds an error to the installation plan
func (p *InstallationPlan) AddError(err string) {
	p.Errors = append(p.Errors, err)
	p.HasConflicts = true
}

// IsValid returns true if the installation plan has no errors
func (p *InstallationPlan) IsValid() bool {
	return !p.HasConflicts && len(p.Errors) == 0
}

// RequiresConfirmation returns true if the plan requires user confirmation
func (p *InstallationPlan) RequiresConfirmation() bool {
	return len(p.WillReplace) > 0 || p.HasConflicts || len(p.Warnings) > 0
}
