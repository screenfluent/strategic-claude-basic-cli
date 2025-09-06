package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/filesystem"
	"strategic-claude-basic-cli/internal/services/git"
	"strategic-claude-basic-cli/internal/services/status"
	"strategic-claude-basic-cli/internal/services/symlink"
)

// Service provides installation functionality for the Strategic Claude Basic framework
type Service struct {
	gitService        *git.Service
	filesystemService *filesystem.Service
	statusService     *status.Service
	symlinkService    *symlink.Service
}

// New creates a new installer service instance
func New() *Service {
	return &Service{
		gitService:        git.New(),
		filesystemService: filesystem.New(),
		statusService:     status.NewService(),
		symlinkService:    symlink.New(),
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

	// Determine installation type
	installType := s.determineInstallationType(currentStatus, installConfig)
	plan := models.NewInstallationPlan(absTarget, installType)

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

	// Clone repository to temporary location
	tempDir, err := s.gitService.CloneRepository(config.DefaultRepoURL, config.FixedCommit)
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	defer func() {
		if cleanupErr := s.gitService.CleanupTempDir(tempDir); cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup temporary directory: %v\n", cleanupErr)
		}
	}()

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
