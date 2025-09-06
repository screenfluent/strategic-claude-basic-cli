package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/installer"
	"strategic-claude-basic-cli/internal/services/status"
)

// Test constants
const (
	testDirPrefix = "strategic-claude-test-"
)

// TestInitCommand_NewInstallation tests installing in a clean directory
func TestInitCommand_NewInstallation(t *testing.T) {
	// Create temporary directory
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// Run init command
	err := runInitTest(tempDir, false, false, false)
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify installation
	verifyInstallation(t, tempDir, true)
}

// TestInitCommand_ForceCore tests selective core updates
func TestInitCommand_ForceCore(t *testing.T) {
	// Create temporary directory and install initial version
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// First installation
	err := runInitTest(tempDir, false, false, false)
	if err != nil {
		t.Fatalf("Initial installation failed: %v", err)
	}

	// Create user content in preserved directories
	userContent := createUserContent(t, tempDir)

	// Run force-core update
	err = runInitTest(tempDir, false, true, false)
	if err != nil {
		t.Fatalf("Force-core update failed: %v", err)
	}

	// Verify installation is still valid
	verifyInstallation(t, tempDir, true)

	// Verify user content was preserved
	verifyUserContent(t, userContent)
}

// TestInitCommand_Force tests full overwrite installation
func TestInitCommand_Force(t *testing.T) {
	// Create temporary directory and install initial version
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// First installation
	err := runInitTest(tempDir, false, false, false)
	if err != nil {
		t.Fatalf("Initial installation failed: %v", err)
	}

	// Create user content (should be overwritten)
	createUserContent(t, tempDir)

	// Run force installation
	err = runInitTest(tempDir, true, false, false)
	if err != nil {
		t.Fatalf("Force installation failed: %v", err)
	}

	// Verify installation is valid
	verifyInstallation(t, tempDir, true)

	// Note: User content verification not applicable for force install
	// as it should have been overwritten
}

// TestInitCommand_DryRun tests dry-run functionality
func TestInitCommand_DryRun(t *testing.T) {
	// Create temporary directory
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// Run dry-run (should not modify anything)
	err := runInitTest(tempDir, false, false, true)
	if err != nil {
		t.Fatalf("Dry-run failed: %v", err)
	}

	// Verify nothing was installed
	verifyInstallation(t, tempDir, false)
}

// TestInitCommand_InvalidFlags tests invalid flag combinations
func TestInitCommand_InvalidFlags(t *testing.T) {
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// Test force + force-core combination (should fail)
	err := runInitTest(tempDir, true, true, false)
	if err == nil {
		t.Fatal("Expected error for invalid flag combination, but got none")
	}

	// Verify error is about invalid configuration
	if !strings.Contains(err.Error(), "cannot specify both --force and --force-core") {
		t.Fatalf("Expected invalid configuration error, got: %v", err)
	}
}

// TestInitCommand_NoGit tests behavior when git is not available
func TestInitCommand_NoGit(t *testing.T) {
	// This test is challenging to implement without actually removing git
	// For now, we'll skip it in unit tests
	t.Skip("Git availability test requires environment manipulation")
}

// TestInitCommand_PermissionDenied tests behavior with permission issues
func TestInitCommand_PermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test permission denied as root user")
	}

	// Create a directory with restricted permissions
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// Create a subdirectory and make it read-only
	restrictedDir := filepath.Join(tempDir, "restricted")
	err := os.Mkdir(restrictedDir, 0555) // read + execute only
	if err != nil {
		t.Fatalf("Failed to create restricted directory: %v", err)
	}

	// Try to install in the restricted directory (should fail)
	err = runInitTest(restrictedDir, false, false, false)
	if err == nil {
		t.Fatal("Expected permission error, but installation succeeded")
	}

	// The error should be related to permissions
	// Note: The exact error depends on when the permission check occurs
}

// TestInitCommand_StatusValidation tests that installations pass status checks
func TestInitCommand_StatusValidation(t *testing.T) {
	tempDir, cleanup := createTempDir(t)
	defer cleanup()

	// Install
	err := runInitTest(tempDir, false, false, false)
	if err != nil {
		t.Fatalf("Installation failed: %v", err)
	}

	// Check status
	statusService := status.NewService()
	statusInfo, err := statusService.CheckInstallation(tempDir)
	if err != nil {
		t.Fatalf("Status check failed: %v", err)
	}

	// Verify status
	if !statusInfo.IsInstalled {
		t.Error("Status check reports not installed after successful installation")
	}

	if statusInfo.HasIssues() {
		t.Errorf("Status check found issues after successful installation: %v", statusInfo.Issues)
	}

	if statusInfo.ValidSymlinks() == 0 {
		t.Error("Status check found no valid symlinks after successful installation")
	}
}

// Helper functions

// runInitTest executes the init command with specified parameters
func runInitTest(targetDir string, force, forceCore, dryRun bool) error {
	// Create install configuration - always skip confirmation for tests
	installConfig := models.InstallConfig{
		TargetDir:   targetDir,
		Force:       force,
		ForceCore:   forceCore,
		SkipConfirm: true,  // Always skip confirmation in tests
		NoBackup:    false, // Don't skip backups in tests (unless it causes issues)
		DryRun:      dryRun,
		Verbose:     false, // Keep quiet for tests
	}

	// Validate configuration
	if err := installConfig.Validate(); err != nil {
		return err
	}

	if dryRun {
		// For dry-run, just validate the analysis
		installerService := installer.New()
		plan, err := installerService.AnalyzeInstallation(installConfig)
		if err != nil {
			return fmt.Errorf("dry-run analysis failed: %w", err)
		}
		// Dry-run is successful if we can create a plan
		_ = plan
		return nil
	}

	// Run the actual installation
	installerService := installer.New()
	return installerService.Install(installConfig)
}

// createTempDir creates a temporary directory for testing
func createTempDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", testDirPrefix)
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	cleanup := func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Failed to cleanup temp directory %s: %v", tempDir, err)
		}
	}

	return tempDir, cleanup
}

// verifyInstallation checks if the installation is correct
func verifyInstallation(t *testing.T, targetDir string, shouldExist bool) {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)

	// Check directories exist/don't exist as expected
	checkDirExists(t, strategicDir, shouldExist, ".strategic-claude-basic")
	checkDirExists(t, claudeDir, shouldExist, ".claude")

	if !shouldExist {
		return // No further checks needed
	}

	// Check required framework directories
	frameworkDirs := config.GetFrameworkDirectories()
	for _, dir := range frameworkDirs {
		dirPath := filepath.Join(strategicDir, dir)
		checkDirExists(t, dirPath, true, fmt.Sprintf("framework directory %s", dir))
	}

	// Check core subdirectories
	coreDir := filepath.Join(strategicDir, config.CoreDir)
	coreSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range coreSubdirs {
		subdirPath := filepath.Join(coreDir, subdir)
		checkDirExists(t, subdirPath, true, fmt.Sprintf("core subdirectory %s", subdir))
	}

	// Check .claude subdirectories
	claudeSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range claudeSubdirs {
		subdirPath := filepath.Join(claudeDir, subdir)
		checkDirExists(t, subdirPath, true, fmt.Sprintf("claude subdirectory %s", subdir))
	}

	// Check symlinks
	requiredSymlinks := config.GetRequiredSymlinks()
	for symlinkPath := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)
		checkSymlinkExists(t, fullSymlinkPath, fmt.Sprintf("symlink %s", symlinkPath))
	}
}

// createUserContent creates test user content in preserved directories
func createUserContent(t *testing.T, targetDir string) map[string]string {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	userContent := make(map[string]string)

	// Create content in user directories that should be preserved
	preservedDirs := config.GetUserPreservedDirectories()
	for _, dir := range preservedDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if err := os.MkdirAll(dirPath, config.DirPermissions); err != nil {
			t.Fatalf("Failed to create user directory %s: %v", dir, err)
		}

		filename := fmt.Sprintf("user-content-%s.txt", dir)
		filePath := filepath.Join(dirPath, filename)
		content := fmt.Sprintf("Test user content in %s directory", dir)

		if err := os.WriteFile(filePath, []byte(content), config.FilePermissions); err != nil {
			t.Fatalf("Failed to create user content file %s: %v", filePath, err)
		}

		userContent[filePath] = content
	}

	return userContent
}

// verifyUserContent checks that user content was preserved
func verifyUserContent(t *testing.T, expectedContent map[string]string) {
	for filePath, expectedData := range expectedContent {
		data, err := os.ReadFile(filePath)
		if err != nil {
			t.Errorf("User content file %s was not preserved: %v", filePath, err)
			continue
		}

		if string(data) != expectedData {
			t.Errorf("User content file %s has wrong content.\nExpected: %s\nGot: %s",
				filePath, expectedData, string(data))
		}
	}
}

// checkDirExists verifies directory existence matches expectation
func checkDirExists(t *testing.T, dirPath string, shouldExist bool, description string) {
	_, err := os.Stat(dirPath)
	exists := !os.IsNotExist(err)

	if shouldExist && !exists {
		t.Errorf("%s should exist but doesn't: %s", description, dirPath)
	} else if !shouldExist && exists {
		t.Errorf("%s should not exist but does: %s", description, dirPath)
	}
}

// checkSymlinkExists verifies symlink exists and is valid
func checkSymlinkExists(t *testing.T, symlinkPath, description string) {
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Errorf("%s should exist but doesn't: %s (%v)", description, symlinkPath, err)
		return
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("%s exists but is not a symlink: %s", description, symlinkPath)
		return
	}

	// Check if symlink target exists (following the symlink)
	_, err = os.Stat(symlinkPath)
	if err != nil {
		t.Errorf("%s exists but target is invalid: %s (%v)", description, symlinkPath, err)
	}
}
