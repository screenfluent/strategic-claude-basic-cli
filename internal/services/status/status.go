package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/templates"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/utils"
)

// Service provides status checking functionality
type Service struct {
	pathValidator  *utils.PathValidator
	fsValidator    *utils.FileSystemValidator
	inputValidator *utils.InputValidator
}

// NewService creates a new status service
func NewService() *Service {
	return &Service{
		pathValidator:  utils.NewPathValidator(),
		fsValidator:    utils.NewFileSystemValidator(),
		inputValidator: utils.NewInputValidator(),
	}
}

// CheckInstallation performs comprehensive status checking for a target directory
func (s *Service) CheckInstallation(targetDir string) (*models.StatusInfo, error) {
	// Resolve target directory to absolute path
	absTarget, err := s.pathValidator.ResolvePath(targetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target directory: %w", err)
	}

	// Validate target directory exists
	if err := s.pathValidator.ValidateDirectory(absTarget); err != nil {
		return nil, fmt.Errorf("invalid target directory: %w", err)
	}

	// Initialize status info
	status := models.NewStatusInfo(absTarget)
	status.StrategicClaudeDirPath = filepath.Join(absTarget, config.StrategicClaudeBasicDir)
	status.ClaudeDirPath = filepath.Join(absTarget, config.ClaudeDir)
	status.CodexDirPath = filepath.Join(absTarget, config.CodexDir)

	// Check .strategic-claude-basic directory
	if err := s.detectStrategicClaudeBasic(status); err != nil {
		return nil, fmt.Errorf("failed to check strategic-claude-basic directory: %w", err)
	}

	// Check .claude directory structure
	if err := s.verifyClaudeDirectory(status); err != nil {
		return nil, fmt.Errorf("failed to verify claude directory: %w", err)
	}

	// Check .codex directory structure
	if err := s.verifyCodexDirectory(status); err != nil {
		return nil, fmt.Errorf("failed to verify codex directory: %w", err)
	}

	// Load template information if installation exists
	if status.StrategicClaudeDir {
		templateInfo, err := s.loadTemplateInfo(absTarget)
		if err != nil {
			status.AddIssue(fmt.Sprintf("Failed to load template information: %v", err))
		} else {
			status.InstalledTemplate = templateInfo
		}
	}

	// Validate symlinks
	s.validateSymlinks(status)
	s.validateCodexSymlinks(status)

	// Identify any issues
	s.identifyIssues(status)

	// Determine overall installation status
	status.IsInstalled = status.StrategicClaudeDir && (status.ClaudeDir || status.CodexDir) && (status.ValidSymlinks() > 0 || status.ValidCodexSymlinks() > 0)

	return status, nil
}

// detectStrategicClaudeBasic checks if the .strategic-claude-basic directory exists and is properly structured
func (s *Service) detectStrategicClaudeBasic(status *models.StatusInfo) error {
	strategicDir := status.StrategicClaudeDirPath

	// Check if directory exists
	info, err := os.Stat(strategicDir)
	if err != nil {
		if os.IsNotExist(err) {
			status.StrategicClaudeDir = false
			status.AddIssue(".strategic-claude-basic directory does not exist")
			return nil
		}
		return fmt.Errorf("failed to stat strategic-claude-basic directory: %w", err)
	}

	if !info.IsDir() {
		status.StrategicClaudeDir = false
		status.AddIssue(".strategic-claude-basic exists but is not a directory")
		return nil
	}

	status.StrategicClaudeDir = true

	// Check for required framework directories
	requiredDirs := config.GetFrameworkDirectories()
	for _, dir := range requiredDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			status.AddIssue(fmt.Sprintf("Missing framework directory: %s", dir))
		}
	}

	// Check for core subdirectories that will be symlinked
	coreDir := filepath.Join(strategicDir, config.CoreDir)
	requiredCoreSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}

	for _, subdir := range requiredCoreSubdirs {
		subdirPath := filepath.Join(coreDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			status.AddIssue(fmt.Sprintf("Missing core subdirectory: core/%s", subdir))
		}
	}

	return nil
}

// verifyClaudeDirectory checks if the .claude directory exists and has the correct structure
func (s *Service) verifyClaudeDirectory(status *models.StatusInfo) error {
	claudeDir := status.ClaudeDirPath

	// Check if directory exists
	info, err := os.Stat(claudeDir)
	if err != nil {
		if os.IsNotExist(err) {
			status.ClaudeDir = false
			status.AddIssue(".claude directory does not exist")
			return nil
		}
		return fmt.Errorf("failed to stat claude directory: %w", err)
	}

	if !info.IsDir() {
		status.ClaudeDir = false
		status.AddIssue(".claude exists but is not a directory")
		return nil
	}

	status.ClaudeDir = true

	// Check for required subdirectories
	requiredSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range requiredSubdirs {
		subdirPath := filepath.Join(claudeDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			status.AddIssue(fmt.Sprintf("Missing .claude subdirectory: %s", subdir))
		}
	}

	return nil
}

// validateSymlinks checks all required symlinks and their targets
func (s *Service) validateSymlinks(status *models.StatusInfo) {
	requiredSymlinks := config.GetRequiredSymlinks()

	for symlinkPath, expectedTarget := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(status.ClaudeDirPath, symlinkPath)
		// Use the relative target as-is - the ValidateSymlink function will handle path resolution

		symlinkStatus, err := s.fsValidator.ValidateSymlink(fullSymlinkPath, expectedTarget)
		if err != nil {
			// Log error but continue checking other symlinks
			status.AddIssue(fmt.Sprintf("Failed to check symlink %s: %v", symlinkPath, err))
		}

		if symlinkStatus != nil {
			status.AddSymlink(*symlinkStatus)
		}
	}
}

// identifyIssues performs additional issue identification based on the gathered information
func (s *Service) identifyIssues(status *models.StatusInfo) {
	// Check for permission issues
	if status.StrategicClaudeDir {
		if err := s.pathValidator.ValidateDirectoryWritable(status.StrategicClaudeDirPath); err != nil {
			status.AddIssue(fmt.Sprintf("Strategic Claude Basic directory is not writable: %v", err))
		}
	}

	if status.ClaudeDir {
		if err := s.pathValidator.ValidateDirectoryWritable(status.ClaudeDirPath); err != nil {
			status.AddIssue(fmt.Sprintf("Claude directory is not writable: %v", err))
		}
	}

	// Check for partial installation
	if status.StrategicClaudeDir && !status.ClaudeDir {
		status.AddIssue("Partial installation detected: .strategic-claude-basic exists but .claude directory is missing")
	}

	if !status.StrategicClaudeDir && status.ClaudeDir {
		status.AddIssue("Partial installation detected: .claude directory exists but .strategic-claude-basic is missing")
	}

	// Check for symlink integrity
	validSymlinks := status.ValidSymlinks()
	totalSymlinks := len(status.Symlinks)

	if totalSymlinks > 0 && validSymlinks < totalSymlinks {
		status.AddIssue(fmt.Sprintf("Some symlinks are broken or invalid (%d/%d valid)", validSymlinks, totalSymlinks))
	}

	if status.StrategicClaudeDir && status.ClaudeDir && totalSymlinks == 0 {
		status.AddIssue("Installation directories exist but no strategic symlinks were found")
	}
}

// verifyCodexDirectory checks if the .codex directory exists and has the correct structure
func (s *Service) verifyCodexDirectory(status *models.StatusInfo) error {
	codexDir := status.CodexDirPath

	// Check if directory exists
	info, err := os.Stat(codexDir)
	if err != nil {
		if os.IsNotExist(err) {
			status.CodexDir = false
			// Only report as issue if strategic-claude-basic is installed (indicating expectation for integration dirs)
			if status.StrategicClaudeDir {
				status.AddIssue(".codex directory does not exist")
			}
			return nil
		}
		return fmt.Errorf("failed to stat codex directory: %w", err)
	}

	if !info.IsDir() {
		status.CodexDir = false
		status.AddIssue(".codex exists but is not a directory")
		return nil
	}

	status.CodexDir = true

	// Check for required subdirectories
	requiredSubdirs := []string{config.PromptsDir, config.HooksDir}
	for _, subdir := range requiredSubdirs {
		subdirPath := filepath.Join(codexDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			status.AddIssue(fmt.Sprintf("Missing codex subdirectory: %s", subdir))
		}
	}

	return nil
}

// validateCodexSymlinks validates all Codex symlinks and populates status
func (s *Service) validateCodexSymlinks(status *models.StatusInfo) {
	if !status.CodexDir {
		return // Skip if .codex directory doesn't exist
	}

	codexDir := status.CodexDirPath
	requiredSymlinks := config.GetCodexRequiredSymlinks()

	for symlinkPath, expectedTarget := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(codexDir, symlinkPath)

		symlinkStatus, err := s.fsValidator.ValidateSymlink(fullSymlinkPath, expectedTarget)
		if err != nil {
			// Create a basic status entry for the error
			status.AddCodexSymlink(models.SymlinkStatus{
				Name:   filepath.Base(symlinkPath),
				Path:   fullSymlinkPath,
				Valid:  false,
				Target: "",
				Exists: false,
				Error:  fmt.Sprintf("Failed to validate codex symlink: %v", err),
			})
			continue
		}

		if symlinkStatus != nil {
			status.AddCodexSymlink(*symlinkStatus)
		}
	}
}

// GetStatusSummary returns a human-readable summary of the installation status
func (s *Service) GetStatusSummary(status *models.StatusInfo) string {
	if !status.IsInstalled {
		return "Strategic Claude Basic is not installed"
	}

	if status.HasIssues() {
		return fmt.Sprintf("Strategic Claude Basic is installed but has %d issue(s)", len(status.Issues))
	}

	return "Strategic Claude Basic is installed and configured correctly"
}

// loadTemplateInfo loads template metadata from the installation directory
func (s *Service) loadTemplateInfo(targetDir string) (*templates.TemplateInfo, error) {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	templateInfoPath := filepath.Join(strategicDir, config.TemplateInfoFile)

	// Check if file exists
	if _, err := os.Stat(templateInfoPath); os.IsNotExist(err) {
		return nil, nil // No template info found
	}

	// Read file
	data, err := os.ReadFile(templateInfoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template info from %s: %w", templateInfoPath, err)
	}

	// Unmarshal JSON
	var templateInfo templates.TemplateInfo
	if err := json.Unmarshal(data, &templateInfo); err != nil {
		return nil, fmt.Errorf("failed to parse template info JSON: %w", err)
	}

	return &templateInfo, nil
}
