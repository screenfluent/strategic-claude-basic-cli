package symlink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
)

func TestNew(t *testing.T) {
	service := New()

	if service == nil {
		t.Fatal("New() returned nil")
	}

	if service.fsValidator == nil {
		t.Error("FileSystem validator not initialized")
	}
}

func TestCreateSymlinks(t *testing.T) {
	service := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "symlink-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the strategic-claude-basic directory structure
	strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
	coreDir := filepath.Join(strategicDir, config.CoreDir)

	// Create required subdirectories
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(coreDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}
	}

	// Test creating symlinks
	err = service.CreateSymlinks(tempDir)
	if err != nil {
		t.Fatalf("CreateSymlinks failed: %v", err)
	}

	// Verify symlinks were created
	claudeDir := filepath.Join(tempDir, config.ClaudeDir)
	requiredSymlinks := config.GetRequiredSymlinks()

	for symlinkPath, expectedTarget := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Check if symlink exists
		info, err := os.Lstat(fullSymlinkPath)
		if err != nil {
			t.Errorf("Symlink %s does not exist: %v", symlinkPath, err)
			continue
		}

		// Check if it's actually a symlink
		if info.Mode()&os.ModeSymlink == 0 {
			t.Errorf("Path %s exists but is not a symlink", symlinkPath)
			continue
		}

		// Check target
		target, err := os.Readlink(fullSymlinkPath)
		if err != nil {
			t.Errorf("Failed to read symlink target for %s: %v", symlinkPath, err)
			continue
		}

		if target != expectedTarget {
			t.Errorf("Symlink %s has wrong target: expected %s, got %s", symlinkPath, expectedTarget, target)
		}
	}
}

func TestCreateSymlinksEmptyTargetDir(t *testing.T) {
	service := New()

	err := service.CreateSymlinks("")
	if err == nil {
		t.Error("Expected error for empty target directory")
	}
}

func TestRemoveSymlinks(t *testing.T) {
	service := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "symlink-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create .claude directory structure
	claudeDir := filepath.Join(tempDir, config.ClaudeDir)
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(claudeDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}
	}

	// Create some test symlinks
	requiredSymlinks := config.GetRequiredSymlinks()
	for symlinkPath, target := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Create parent directory if needed
		parentDir := filepath.Dir(fullSymlinkPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			t.Fatalf("Failed to create parent directory %s: %v", parentDir, err)
		}

		// Create symlink
		if err := os.Symlink(target, fullSymlinkPath); err != nil {
			t.Fatalf("Failed to create test symlink %s: %v", symlinkPath, err)
		}
	}

	// Remove symlinks
	err = service.RemoveSymlinks(tempDir)
	if err != nil {
		t.Fatalf("RemoveSymlinks failed: %v", err)
	}

	// Verify symlinks were removed
	for symlinkPath := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		if _, err := os.Lstat(fullSymlinkPath); err == nil {
			t.Errorf("Symlink %s still exists after removal", symlinkPath)
		}
	}
}

func TestRemoveSymlinksEmptyTargetDir(t *testing.T) {
	service := New()

	err := service.RemoveSymlinks("")
	if err == nil {
		t.Error("Expected error for empty target directory")
	}
}

func TestValidateSymlinks(t *testing.T) {
	service := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "symlink-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the strategic-claude-basic directory structure
	strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
	coreDir := filepath.Join(strategicDir, config.CoreDir)

	// Create required subdirectories
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(coreDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}
	}

	// Create .claude directory structure
	claudeDir := filepath.Join(tempDir, config.ClaudeDir)
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(claudeDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}
	}

	// Create valid symlinks
	requiredSymlinks := config.GetRequiredSymlinks()
	validSymlinks := 0
	for symlinkPath, target := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Create parent directory if needed
		parentDir := filepath.Dir(fullSymlinkPath)
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			t.Fatalf("Failed to create parent directory %s: %v", parentDir, err)
		}

		// Create symlink
		if err := os.Symlink(target, fullSymlinkPath); err != nil {
			t.Fatalf("Failed to create test symlink %s: %v", symlinkPath, err)
		}
		validSymlinks++
	}

	// Validate symlinks
	statuses, err := service.ValidateSymlinks(tempDir)
	if err != nil {
		t.Fatalf("ValidateSymlinks failed: %v", err)
	}

	if len(statuses) != validSymlinks {
		t.Errorf("Expected %d symlink statuses, got %d", validSymlinks, len(statuses))
	}

	// Check that all symlinks are reported as valid
	for _, status := range statuses {
		if !status.Valid {
			t.Errorf("Symlink %s reported as invalid: %s", status.Name, status.Error)
		}

		if !status.Exists {
			t.Errorf("Symlink %s reported as not existing", status.Name)
		}
	}
}

func TestUpdateSymlinks(t *testing.T) {
	service := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "symlink-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create the strategic-claude-basic directory structure
	strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
	coreDir := filepath.Join(strategicDir, config.CoreDir)

	// Create required subdirectories
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(coreDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}
	}

	// Create .claude directory structure with old symlinks
	claudeDir := filepath.Join(tempDir, config.ClaudeDir)
	for _, subdir := range []string{config.AgentsDir, config.CommandsDir, config.HooksDir} {
		subdirPath := filepath.Join(claudeDir, subdir)
		if err := os.MkdirAll(subdirPath, 0755); err != nil {
			t.Fatalf("Failed to create subdir %s: %v", subdirPath, err)
		}

		// Create old symlink with wrong target
		oldSymlinkPath := filepath.Join(subdirPath, "strategic")
		if err := os.Symlink("../wrong/target", oldSymlinkPath); err != nil {
			t.Fatalf("Failed to create old symlink %s: %v", oldSymlinkPath, err)
		}
	}

	// Update symlinks
	err = service.UpdateSymlinks(tempDir)
	if err != nil {
		t.Fatalf("UpdateSymlinks failed: %v", err)
	}

	// Verify symlinks were updated correctly
	requiredSymlinks := config.GetRequiredSymlinks()
	for symlinkPath, expectedTarget := range requiredSymlinks {
		fullSymlinkPath := filepath.Join(claudeDir, symlinkPath)

		// Check target
		target, err := os.Readlink(fullSymlinkPath)
		if err != nil {
			t.Errorf("Failed to read symlink target for %s: %v", symlinkPath, err)
			continue
		}

		if target != expectedTarget {
			t.Errorf("Symlink %s has wrong target after update: expected %s, got %s", symlinkPath, expectedTarget, target)
		}
	}
}

func TestGetSymlinkInfo(t *testing.T) {
	service := New()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "symlink-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test non-existent symlink
	nonExistentPath := filepath.Join(tempDir, "nonexistent")
	status, err := service.GetSymlinkInfo(nonExistentPath)
	if err != nil {
		t.Errorf("GetSymlinkInfo failed for non-existent path: %v", err)
	}
	if status == nil {
		t.Error("Expected status for non-existent path")
	} else {
		if status.Exists {
			t.Error("Non-existent path reported as existing")
		}
		if status.Valid {
			t.Error("Non-existent path reported as valid")
		}
	}

	// Create a regular file (not a symlink)
	regularFile := filepath.Join(tempDir, "regular-file")
	if err := os.WriteFile(regularFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create regular file: %v", err)
	}

	status, err = service.GetSymlinkInfo(regularFile)
	if err != nil {
		t.Errorf("GetSymlinkInfo failed for regular file: %v", err)
	}
	if status == nil {
		t.Error("Expected status for regular file")
	} else {
		if !status.Exists {
			t.Error("Regular file reported as not existing")
		}
		if status.Valid {
			t.Error("Regular file reported as valid symlink")
		}
	}

	// Create a target directory
	targetDir := filepath.Join(tempDir, "target")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Create a valid symlink
	symlinkPath := filepath.Join(tempDir, "valid-symlink")
	relativeTarget := "./target"
	if err := os.Symlink(relativeTarget, symlinkPath); err != nil {
		t.Fatalf("Failed to create valid symlink: %v", err)
	}

	status, err = service.GetSymlinkInfo(symlinkPath)
	if err != nil {
		t.Errorf("GetSymlinkInfo failed for valid symlink: %v", err)
	}
	if status == nil {
		t.Error("Expected status for valid symlink")
	} else {
		if !status.Exists {
			t.Error("Valid symlink reported as not existing")
		}
		if !status.Valid {
			t.Error("Valid symlink reported as invalid")
		}
		if status.Target != relativeTarget {
			t.Errorf("Wrong target reported: expected %s, got %s", relativeTarget, status.Target)
		}
	}

	// Create a broken symlink
	brokenSymlinkPath := filepath.Join(tempDir, "broken-symlink")
	if err := os.Symlink("./nonexistent-target", brokenSymlinkPath); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	status, err = service.GetSymlinkInfo(brokenSymlinkPath)
	if err != nil {
		t.Errorf("GetSymlinkInfo failed for broken symlink: %v", err)
	}
	if status == nil {
		t.Error("Expected status for broken symlink")
	} else {
		if !status.Exists {
			t.Error("Broken symlink reported as not existing")
		}
		if status.Valid {
			t.Error("Broken symlink reported as valid")
		}
	}
}
