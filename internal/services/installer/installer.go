package installer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/codexconfig"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/filesystem"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/git"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/script"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/settings"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/status"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/symlink"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/templates"
)

// Service provides installation functionality for the Strategic Claude Basic framework
type Service struct {
	gitService        *git.Service
	filesystemService *filesystem.Service
	statusService     *status.Service
	symlinkService    *symlink.Service
	settingsService   *settings.Service
	codexConfigService *codexconfig.Service
	scriptService     *script.Service
}

// New creates a new installer service instance
func New() *Service {
	return &Service{
		gitService:        git.New(),
		filesystemService: filesystem.New(),
		statusService:     status.NewService(),
		symlinkService:    symlink.New(),
		settingsService:   settings.New(),
		codexConfigService: codexconfig.New(),
		scriptService:     script.New(),
	}
}

// AnalyzeInstallation examines the target directory and determines what type of installation is needed
func (s *Service) AnalyzeInstallation(installConfig models.InstallConfig) (*models.InstallationPlan, error) {
	// Validate target directory exists
	absTarget, err := filepath.Abs(installConfig.TargetDir)
	if err != nil {
		return nil, models.NewAppError(
			models.ErrorCodeInvalidPath,
			fmt.Sprintf("Failed to resolve target directory: %s", installConfig.TargetDir),
			err,
		)
	}

	if _, err := os.Stat(absTarget); os.IsNotExist(err) {
		return nil, models.NewAppError(
			models.ErrorCodeDirectoryNotFound,
			fmt.Sprintf("Target directory does not exist: %s", absTarget),
			err,
		)
	}

	// Check current installation status
	currentStatus, err := s.statusService.CheckInstallation(absTarget)
	if err != nil {
		return nil, fmt.Errorf("failed to check installation status: %w", err)
	}

	// Get template configuration
	template, err := installConfig.GetTemplate()
	if err != nil {
		return nil, fmt.Errorf("failed to get template configuration: %w", err)
	}

	// Determine installation type
	installType := s.determineInstallationType(currentStatus, installConfig)
	plan := models.NewInstallationPlan(absTarget, installType, template)

	// Analyze what will be done based on installation type
	s.analyzeFileOperations(plan, currentStatus)

	// Determine if backup is needed
	plan.BackupRequired = s.needsBackup(plan, installConfig)
	if plan.BackupRequired && !installConfig.NoBackup {
		plan.BackupDir = s.filesystemService.GetBackupPath(absTarget)
	}

	// Set up directory operations
	s.analyzeDirectoryOperations(plan, currentStatus)

	// Set up symlink operations
	s.analyzeSymlinkOperations(plan, currentStatus)

	// Check for installation scripts
	s.analyzeScriptOperations(plan)

	return plan, nil
}

// Install performs the complete installation process
func (s *Service) Install(installConfig models.InstallConfig) error {
	// Analyze what needs to be done
	plan, err := s.AnalyzeInstallation(installConfig)
	if err != nil {
		return fmt.Errorf("installation analysis failed: %w", err)
	}

	// Validate the plan
	if !plan.IsValid() {
		return models.NewAppError(
			models.ErrorCodeInstallationFailed,
			fmt.Sprintf("Installation plan has errors: %v", plan.Errors),
			nil,
		)
	}

	// Create backup if needed
	if plan.BackupRequired && !installConfig.NoBackup {
		if err := s.CreateBackup(plan.TargetDir, plan.BackupDir); err != nil {
			return fmt.Errorf("backup creation failed: %w", err)
		}
	}

	// Get template configuration for cloning
	template, err := installConfig.GetTemplate()
	if err != nil {
		return fmt.Errorf("failed to get template configuration: %w", err)
	}

	// Clone repository to temporary location using template configuration
	tempDir, err := s.gitService.CloneRepositoryWithBranch(template.RepoURL, template.Branch, template.Commit)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	defer func() {
		if cleanupErr := s.gitService.CleanupTempDir(tempDir); cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup temporary directory: %v\n", cleanupErr)
		}
	}()

	// Update plan with actual script detection
	plan.HasPreInstallScript = s.scriptService.ScriptExists(tempDir, config.PreInstallScript)
	plan.HasPostInstallScript = s.scriptService.ScriptExists(tempDir, config.PostInstallScript)

	// Execute pre-install script if it exists
	if plan.HasPreInstallScript {
		if err := s.executePreInstallScript(tempDir, plan.TargetDir); err != nil {
			return fmt.Errorf("pre-install script failed: %w", err)
		}
	}

	// Perform the installation based on type
	switch plan.InstallationType {
	case models.InstallationTypeNew:
		err = s.installNew(tempDir, plan.TargetDir)
	case models.InstallationTypeUpdate:
		err = s.InstallCore(tempDir, plan.TargetDir)
	case models.InstallationTypeOverwrite:
		err = s.installOverwrite(tempDir, plan.TargetDir)
	default:
		err = models.NewAppError(
			models.ErrorCodeInstallationFailed,
			fmt.Sprintf("Unknown installation type: %s", plan.InstallationType),
			nil,
		)
	}

	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Create .claude directory structure if needed
	if err := s.ensureClaudeDirectory(plan.TargetDir); err != nil {
		return fmt.Errorf("failed to create .claude directory structure: %w", err)
	}

	// Create symlinks
	if err := s.symlinkService.CreateSymlinks(plan.TargetDir); err != nil {
		return fmt.Errorf("failed to create symlinks: %w", err)
	}

	// Create Codex symlinks
	if err := s.symlinkService.CreateCodexSymlinks(plan.TargetDir); err != nil {
		return fmt.Errorf("failed to create codex symlinks: %w", err)
	}

	// Process settings.json (merge template with existing user settings)
	if err := s.settingsService.ProcessSettings(plan.TargetDir); err != nil {
		return fmt.Errorf("failed to process settings: %w", err)
	}

	// Process Codex config.toml (copy template if it exists)
	if err := s.codexConfigService.ProcessCodexConfig(plan.TargetDir); err != nil {
		return fmt.Errorf("failed to process codex config: %w", err)
	}

	// Execute post-install script if it exists
	if plan.HasPostInstallScript {
		if err := s.executePostInstallScript(tempDir, plan.TargetDir); err != nil {
			return fmt.Errorf("post-install script failed: %w", err)
		}
	}

	// Apply gitignore templates based on mode
	if err := s.applyGitignoreTemplates(tempDir, plan.TargetDir, installConfig.GitignoreMode); err != nil {
		return fmt.Errorf("failed to apply gitignore templates: %w", err)
	}

	// Save template metadata
	if err := s.saveTemplateInfo(plan.TargetDir, template); err != nil {
		return fmt.Errorf("failed to save template metadata: %w", err)
	}

	// Validate installation
	if err := s.ValidateInstallation(plan.TargetDir); err != nil {
		return fmt.Errorf("installation validation failed: %w", err)
	}

	return nil
}

// InstallCore performs selective core updates (--force-core flag)
func (s *Service) InstallCore(sourceDir, targetDir string) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Ensure target directory exists
	if err := s.filesystemService.CreateDirectory(strategicDir); err != nil {
		return err
	}

	// Copy only framework directories (core, guides, templates)
	sourceStrategicDir := filepath.Join(sourceDir, config.StrategicClaudeBasicDir)
	if err := s.filesystemService.CopyFrameworkFiles(sourceStrategicDir, strategicDir); err != nil {
		return fmt.Errorf("failed to copy framework files: %w", err)
	}

	// Ensure user directories exist (but don't overwrite them)
	if err := s.filesystemService.PreserveUserContent(targetDir); err != nil {
		return fmt.Errorf("failed to preserve user content: %w", err)
	}

	// Process settings.json (merge updated template with existing user settings)
	if err := s.settingsService.ProcessSettings(targetDir); err != nil {
		return fmt.Errorf("failed to process settings during core update: %w", err)
	}

	// Process Codex config.toml (update template if it exists)
	if err := s.codexConfigService.ProcessCodexConfig(targetDir); err != nil {
		return fmt.Errorf("failed to process codex config during core update: %w", err)
	}

	return nil
}

// CreateBackup creates a backup of the existing installation
func (s *Service) CreateBackup(targetDir, backupPath string) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Check if strategic-claude-basic directory exists
	if _, err := os.Stat(strategicDir); os.IsNotExist(err) {
		return nil // Nothing to backup
	}

	// Create backup
	if err := s.filesystemService.BackupDirectory(strategicDir, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	return nil
}

// ValidateInstallation verifies that the installation was successful
func (s *Service) ValidateInstallation(targetDir string) error {
	// Check installation status
	status, err := s.statusService.CheckInstallation(targetDir)
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}

	// Verify installation is complete and valid
	if !status.IsInstalled {
		return models.NewAppError(
			models.ErrorCodeInstallationFailed,
			"Installation validation failed: framework not properly installed",
			nil,
		)
	}

	if status.HasIssues() {
		return models.NewAppError(
			models.ErrorCodeInstallationFailed,
			fmt.Sprintf("Installation validation failed with issues: %v", status.Issues),
			nil,
		)
	}

	return nil
}

// Helper methods

func (s *Service) determineInstallationType(status *models.StatusInfo, installConfig models.InstallConfig) models.InstallationType {
	// If force is set, always do full overwrite
	if installConfig.Force {
		return models.InstallationTypeOverwrite
	}

	// If force-core is set, do selective update
	if installConfig.ForceCore {
		return models.InstallationTypeUpdate
	}

	// If not installed, do new installation
	if !status.IsInstalled {
		return models.InstallationTypeNew
	}

	// If installed and no force flags, this would be an error case
	// The caller should handle this scenario
	return models.InstallationTypeOverwrite
}

func (s *Service) analyzeFileOperations(plan *models.InstallationPlan, status *models.StatusInfo) {
	strategicDir := filepath.Join(plan.TargetDir, config.StrategicClaudeBasicDir)

	switch plan.InstallationType {
	case models.InstallationTypeNew:
		plan.WillCreate = append(plan.WillCreate, config.StrategicClaudeBasicDir)
	case models.InstallationTypeUpdate:
		// Will replace only framework directories
		frameworkDirs := config.GetCoreDirectories()
		for _, dir := range frameworkDirs {
			dirPath := filepath.Join(strategicDir, dir)
			if _, err := os.Stat(dirPath); err == nil {
				plan.WillReplace = append(plan.WillReplace, filepath.Join(config.StrategicClaudeBasicDir, dir))
			} else {
				plan.WillCreate = append(plan.WillCreate, filepath.Join(config.StrategicClaudeBasicDir, dir))
			}
		}
		// Will preserve user directories
		userDirs := config.GetUserPreservedDirectories()
		for _, dir := range userDirs {
			plan.WillPreserve = append(plan.WillPreserve, filepath.Join(config.StrategicClaudeBasicDir, dir))
		}
	case models.InstallationTypeOverwrite:
		if status.StrategicClaudeDir {
			plan.WillReplace = append(plan.WillReplace, config.StrategicClaudeBasicDir)
		} else {
			plan.WillCreate = append(plan.WillCreate, config.StrategicClaudeBasicDir)
		}
	}
}

func (s *Service) needsBackup(plan *models.InstallationPlan, installConfig models.InstallConfig) bool {
	// No backup if explicitly disabled
	if installConfig.NoBackup {
		return false
	}

	// Backup if we're replacing existing files
	return len(plan.WillReplace) > 0
}

func (s *Service) analyzeDirectoryOperations(plan *models.InstallationPlan, status *models.StatusInfo) {
	// Always ensure .claude directory structure
	if !status.ClaudeDir {
		plan.DirectoriesToCreate = append(plan.DirectoriesToCreate, config.ClaudeDir)
		plan.DirectoriesToCreate = append(plan.DirectoriesToCreate,
			filepath.Join(config.ClaudeDir, config.AgentsDir))
		plan.DirectoriesToCreate = append(plan.DirectoriesToCreate,
			filepath.Join(config.ClaudeDir, config.CommandsDir))
		plan.DirectoriesToCreate = append(plan.DirectoriesToCreate,
			filepath.Join(config.ClaudeDir, config.HooksDir))
	}
}

func (s *Service) analyzeSymlinkOperations(plan *models.InstallationPlan, status *models.StatusInfo) {
	requiredSymlinks := config.GetRequiredSymlinks()

	for symlinkPath := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(status.ClaudeDirPath, symlinkPath)

		if _, err := os.Lstat(fullSymlinkPath); os.IsNotExist(err) {
			plan.SymlinksToCreate = append(plan.SymlinksToCreate, symlinkPath)
		} else {
			plan.SymlinksToUpdate = append(plan.SymlinksToUpdate, symlinkPath)
		}
	}
}

func (s *Service) installNew(sourceDir, targetDir string) error {
	sourceStrategicDir := filepath.Join(sourceDir, config.StrategicClaudeBasicDir)
	targetStrategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)

	// Copy entire .strategic-claude-basic directory
	return s.filesystemService.CopyDirectory(sourceStrategicDir, targetStrategicDir)
}

func (s *Service) installOverwrite(sourceDir, targetDir string) error {
	// Remove existing installation
	if err := s.filesystemService.RemoveStrategicClaudeBasic(targetDir); err != nil {
		return err
	}

	// Install fresh copy
	return s.installNew(sourceDir, targetDir)
}

func (s *Service) ensureClaudeDirectory(targetDir string) error {
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)

	// Create main .claude directory
	if err := s.filesystemService.CreateDirectory(claudeDir); err != nil {
		return err
	}

	// Create required subdirectories
	subdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range subdirs {
		subdirPath := filepath.Join(claudeDir, subdir)
		if err := s.filesystemService.CreateDirectory(subdirPath); err != nil {
			return err
		}
	}

	return nil
}

// saveTemplateInfo saves template metadata to the installation directory
func (s *Service) saveTemplateInfo(targetDir string, template templates.Template) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	templateInfoPath := filepath.Join(strategicDir, config.TemplateInfoFile)

	// Create template info
	templateInfo := templates.TemplateInfo{
		Template:        template,
		InstalledAt:     time.Now().Format(time.RFC3339),
		InstalledCommit: template.Commit,
		Metadata:        make(map[string]string),
	}

	// Add additional metadata
	templateInfo.Metadata["cli_version"] = "0.1.0" // TODO: Get from build info
	templateInfo.Metadata["installation_type"] = "cli"

	// Marshal to JSON
	data, err := json.MarshalIndent(templateInfo, "", "  ")
	if err != nil {
		return models.NewAppError(
			models.ErrorCodeFileSystemError,
			"Failed to marshal template info",
			err,
		)
	}

	// Write to file
	if err := os.WriteFile(templateInfoPath, data, config.FilePermissions); err != nil {
		return models.NewAppError(
			models.ErrorCodeFileSystemError,
			fmt.Sprintf("Failed to write template info to %s", templateInfoPath),
			err,
		)
	}

	return nil
}

// analyzeScriptOperations checks if installation scripts exist in the template
func (s *Service) analyzeScriptOperations(plan *models.InstallationPlan) {
	// This will be set after the repository is cloned, but we can initialize it here
	// The actual check will happen in the Install method once we have the temporary directory
	plan.HasPreInstallScript = false
	plan.HasPostInstallScript = false
}

// executePreInstallScript copies and executes the pre-install script
func (s *Service) executePreInstallScript(sourceDir, targetDir string) error {
	// Copy script to target directory
	if err := s.scriptService.CopyScript(sourceDir, targetDir, config.PreInstallScript); err != nil {
		return fmt.Errorf("failed to copy pre-install script: %w", err)
	}

	// Execute the script
	if err := s.scriptService.ExecuteScript(targetDir, config.PreInstallScript); err != nil {
		return fmt.Errorf("failed to execute pre-install script: %w", err)
	}

	// Clean up script after execution
	if err := s.scriptService.RemoveScript(targetDir, config.PreInstallScript); err != nil {
		// Log warning but don't fail installation
		fmt.Printf("Warning: Failed to remove pre-install script: %v\n", err)
	}

	return nil
}

// executePostInstallScript copies and executes the post-install script
func (s *Service) executePostInstallScript(sourceDir, targetDir string) error {
	// Copy script to target directory
	if err := s.scriptService.CopyScript(sourceDir, targetDir, config.PostInstallScript); err != nil {
		return fmt.Errorf("failed to copy post-install script: %w", err)
	}

	// Execute the script
	if err := s.scriptService.ExecuteScript(targetDir, config.PostInstallScript); err != nil {
		return fmt.Errorf("failed to execute post-install script: %w", err)
	}

	// Clean up script after execution
	if err := s.scriptService.RemoveScript(targetDir, config.PostInstallScript); err != nil {
		// Log warning but don't fail installation
		fmt.Printf("Warning: Failed to remove post-install script: %v\n", err)
	}

	return nil
}

// applyGitignoreTemplates applies gitignore templates based on the selected mode
func (s *Service) applyGitignoreTemplates(sourceDir, targetDir, gitignoreMode string) error {
	if gitignoreMode == "track" {
		// Track mode - don't apply any gitignore templates
		return nil
	}

	// Define template mappings based on mode
	var templateMappings map[string]string

	switch gitignoreMode {
	case "all":
		templateMappings = map[string]string{
			"dot_claude-strategic-ignore.template":           ".claude/.gitignore",
			"dot_strategic-claude-basic-ignore-all.template": ".strategic-claude-basic/.gitignore",
		}
	case "non-user":
		templateMappings = map[string]string{
			"dot_claude-strategic-ignore.template":                     ".claude/.gitignore",
			"dot_strategic-claude-basic-ignore-non-user-dirs.template": ".strategic-claude-basic/.gitignore",
		}
	default:
		return fmt.Errorf("unsupported gitignore mode: %s", gitignoreMode)
	}

	// Apply each template
	for templateFile, targetFile := range templateMappings {
		templatePath := filepath.Join(sourceDir, config.StrategicClaudeBasicDir, "templates", "ignore", templateFile)
		targetPath := filepath.Join(targetDir, targetFile)

		if err := s.filesystemService.ApplyGitignoreTemplate(templatePath, targetPath); err != nil {
			return fmt.Errorf("failed to apply template %s: %w", templateFile, err)
		}

		fmt.Printf("Applied gitignore template: %s -> %s\n", templateFile, targetFile)
	}

	return nil
}
