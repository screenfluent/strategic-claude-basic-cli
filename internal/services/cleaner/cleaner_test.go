package cleaner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/filesystem"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/symlink"
)

func TestNew(t *testing.T) {
	service := New()
	if service == nil {
		t.Fatal("New() returned nil")
	}

	if service.filesystemService == nil {
		t.Error("filesystemService is nil")
	}

	if service.symlinkService == nil {
		t.Error("symlinkService is nil")
	}

	if service.statusService == nil {
		t.Error("statusService is nil")
	}
}

func TestRemoveInstallation_NoInstallation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	service := New()
	result, err := service.RemoveInstallation(tmpDir)

	if err != nil {
		t.Errorf("RemoveInstallation() error = %v", err)
	}

	if !result.Success {
		t.Error("Expected success when no installation exists")
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warning about no installation found")
	}

	if result.RemovedDirectory {
		t.Error("Should not report directory removal when none existed")
	}
}

func TestRemoveInstallation_EmptyDirectory(t *testing.T) {
	service := New()
	_, err := service.RemoveInstallation("")

	if err == nil {
		t.Error("Expected error for empty directory")
	}
}

func TestRemoveInstallation_CompleteInstallation(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up a complete installation
	setupCompleteInstallation(t, tmpDir)

	service := New()
	result, err := service.RemoveInstallation(tmpDir)

	if err != nil {
		t.Errorf("RemoveInstallation() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful removal, got errors: %v", result.Errors)
	}

	if !result.RemovedDirectory {
		t.Error("Expected directory to be removed")
	}

	if len(result.RemovedSymlinks) == 0 {
		t.Error("Expected symlinks to be removed")
	}

	// Verify strategic claude directory is gone
	strategicDir := filepath.Join(tmpDir, config.StrategicClaudeBasicDir)
	if _, err := os.Stat(strategicDir); !os.IsNotExist(err) {
		t.Error("Strategic Claude directory should be removed")
	}

	// Verify symlinks are gone
	claudeDir := filepath.Join(tmpDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()
	for symlinkPath := range requiredSymlinks {
		fullPath := filepath.Join(claudeDir, symlinkPath)
		if _, err := os.Lstat(fullPath); !os.IsNotExist(err) {
			t.Errorf("Symlink should be removed: %s", fullPath)
		}
	}
}

func TestRemoveInstallation_WithUserContent(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up installation with user content
	setupInstallationWithUserContent(t, tmpDir)

	service := New()
	result, err := service.RemoveInstallation(tmpDir)

	if err != nil {
		t.Errorf("RemoveInstallation() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful removal, got errors: %v", result.Errors)
	}

	// Check that user content was preserved
	userFile := filepath.Join(tmpDir, config.ClaudeDir, config.AgentsDir, "user-agent.md")
	if _, err := os.Stat(userFile); os.IsNotExist(err) {
		t.Error("User content should be preserved")
	}

	// Check that it was logged as preserved
	found := false
	for _, preserved := range result.PreservedFiles {
		if preserved == userFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("User content should be in preserved files list")
	}
}

func TestRemoveInstallation_PartialInstallation(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up a partial installation (directory but no symlinks)
	setupPartialInstallation(t, tmpDir)

	service := New()
	result, err := service.RemoveInstallation(tmpDir)

	if err != nil {
		t.Errorf("RemoveInstallation() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful removal of partial installation, got errors: %v", result.Errors)
	}

	if !result.RemovedDirectory {
		t.Error("Expected partial installation directory to be removed")
	}
}

func TestHandlePartialInstallation(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up broken symlinks
	setupBrokenInstallation(t, tmpDir)

	service := New()
	result, err := service.HandlePartialInstallation(tmpDir)

	if err != nil {
		t.Errorf("HandlePartialInstallation() error = %v", err)
	}

	if !result.Success {
		t.Errorf("Expected successful cleanup of partial installation, got errors: %v", result.Errors)
	}

	if len(result.Warnings) == 0 {
		t.Error("Expected warnings about partial installation")
	}
}

func TestIsStrategicClaudeSymlink(t *testing.T) {
	// Create temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "cleaner-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	service := New()

	// Create a Strategic Claude symlink
	symlinkPath := filepath.Join(tmpDir, "strategic")
	target := "../../.strategic-claude-basic/core/agents"
	err = os.Symlink(target, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	isStrategic, err := service.isStrategicClaudeSymlink(symlinkPath)
	if err != nil {
		t.Errorf("isStrategicClaudeSymlink() error = %v", err)
	}

	if !isStrategic {
		t.Error("Expected symlink to be identified as Strategic Claude symlink")
	}

	// Create a non-Strategic Claude symlink
	userSymlinkPath := filepath.Join(tmpDir, "user-symlink")
	userTarget := "../some-other-path"
	err = os.Symlink(userTarget, userSymlinkPath)
	if err != nil {
		t.Fatalf("Failed to create test user symlink: %v", err)
	}

	isStrategic, err = service.isStrategicClaudeSymlink(userSymlinkPath)
	if err != nil {
		t.Errorf("isStrategicClaudeSymlink() error = %v", err)
	}

	if isStrategic {
		t.Error("Expected user symlink to not be identified as Strategic Claude symlink")
	}
}

// Helper functions for setting up test scenarios

func setupCompleteInstallation(t *testing.T, tmpDir string) {
	fsService := filesystem.New()
	symlinkService := symlink.New()

	// Create Strategic Claude directory structure
	if err := fsService.EnsureDirectoryStructure(tmpDir); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create some content in strategic claude directories
	strategicDir := filepath.Join(tmpDir, config.StrategicClaudeBasicDir)
	coreDir := filepath.Join(strategicDir, config.CoreDir, config.AgentsDir)
	testFile := filepath.Join(coreDir, "test-agent.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create symlinks
	if err := symlinkService.CreateSymlinks(tmpDir); err != nil {
		t.Fatalf("Failed to create symlinks: %v", err)
	}
}

func setupInstallationWithUserContent(t *testing.T, tmpDir string) {
	setupCompleteInstallation(t, tmpDir)

	// Add user content to .claude directories
	claudeDir := filepath.Join(tmpDir, config.ClaudeDir)
	agentsDir := filepath.Join(claudeDir, config.AgentsDir)
	userFile := filepath.Join(agentsDir, "user-agent.md")

	if err := os.WriteFile(userFile, []byte("user content"), 0644); err != nil {
		t.Fatalf("Failed to create user content: %v", err)
	}
}

func setupPartialInstallation(t *testing.T, tmpDir string) {
	fsService := filesystem.New()

	// Create only the Strategic Claude directory without symlinks
	if err := fsService.EnsureDirectoryStructure(tmpDir); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Create some content
	strategicDir := filepath.Join(tmpDir, config.StrategicClaudeBasicDir)
	coreDir := filepath.Join(strategicDir, config.CoreDir, config.AgentsDir)
	testFile := filepath.Join(coreDir, "test-agent.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

func setupBrokenInstallation(t *testing.T, tmpDir string) {
	// Create .claude directory structure
	claudeDir := filepath.Join(tmpDir, config.ClaudeDir)
	agentsDir := filepath.Join(claudeDir, config.AgentsDir)
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatalf("Failed to create agents dir: %v", err)
	}

	// Create broken symlink (pointing to non-existent target)
	brokenSymlink := filepath.Join(agentsDir, "strategic")
	brokenTarget := "../../.strategic-claude-basic/core/agents"
	if err := os.Symlink(brokenTarget, brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Note: We don't create the strategic-claude-basic directory, making the symlink broken
}
