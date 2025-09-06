package main

import (
	"os"
	"path/filepath"
	"testing"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/services/cleaner"
	"strategic-claude-basic-cli/internal/services/filesystem"
	"strategic-claude-basic-cli/internal/services/symlink"
)

func TestCleanCommand_NoInstallation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "clean-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original state and restore
	origTargetDir := targetDir
	origCleanForce := cleanForce
	defer func() {
		targetDir = origTargetDir
		cleanForce = origCleanForce
	}()

	// Set test parameters
	targetDir = tmpDir
	cleanForce = true // Skip confirmation

	// Run clean command
	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("Clean command failed with no installation: %v", err)
	}
}

func TestCleanCommand_WithInstallation(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "clean-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Set up a complete installation
	setupTestInstallation(t, tmpDir)

	// Save original state and restore
	origTargetDir := targetDir
	origCleanForce := cleanForce
	defer func() {
		targetDir = origTargetDir
		cleanForce = origCleanForce
	}()

	// Set test parameters
	targetDir = tmpDir
	cleanForce = true // Skip confirmation

	// Run clean command
	err = cleanCmd.RunE(cleanCmd, []string{})
	if err != nil {
		t.Errorf("Clean command failed: %v", err)
	}

	// Verify cleanup
	strategicDir := filepath.Join(tmpDir, config.StrategicClaudeBasicDir)
	if _, err := os.Stat(strategicDir); !os.IsNotExist(err) {
		t.Error("Strategic Claude directory should be removed after cleanup")
	}
}

func TestDisplayCleanupResults(t *testing.T) {
	// Create a mock result
	result := &cleaner.CleanupResult{
		Success:            true,
		RemovedDirectory:   true,
		RemovedSymlinks:    []string{"agents/strategic", "commands/strategic"},
		PreservedFiles:     []string{"user-file.txt"},
		CleanedDirectories: []string{},
		Warnings:           []string{"test warning"},
		Errors:             []string{},
	}

	// Test display (this mainly tests that it doesn't crash)
	// We can't easily test the output, but we can test that it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayCleanupResults panicked: %v", r)
		}
	}()

	displayCleanupResults(result, false)
	displayCleanupResults(result, true)
}

// setupTestInstallation creates a test installation
func setupTestInstallation(t *testing.T, tmpDir string) {
	fsService := filesystem.New()
	symlinkService := symlink.New()

	// Create directory structure
	if err := fsService.EnsureDirectoryStructure(tmpDir); err != nil {
		t.Fatalf("Failed to create directory structure: %v", err)
	}

	// Add some content
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
