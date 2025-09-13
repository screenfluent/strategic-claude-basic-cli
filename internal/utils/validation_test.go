package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

func TestPathValidator_ValidateDirectory(t *testing.T) {
	validator := NewPathValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		shouldErr bool
		errCode   models.ErrorCode
	}{
		{
			name: "empty path",
			setup: func() string {
				return ""
			},
			shouldErr: true,
			errCode:   models.ErrorCodeValidationFailed,
		},
		{
			name: "valid directory",
			setup: func() string {
				return tempDir
			},
			shouldErr: false,
		},
		{
			name: "nonexistent directory",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			shouldErr: true,
			errCode:   models.ErrorCodeDirectoryNotFound,
		},
		{
			name: "file instead of directory",
			setup: func() string {
				filePath := filepath.Join(tempDir, "testfile")
				_ = os.WriteFile(filePath, []byte("test"), 0644)
				return filePath
			},
			shouldErr: true,
			errCode:   models.ErrorCodeValidationFailed,
		},
		{
			name: "relative path",
			setup: func() string {
				return "."
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			err := validator.ValidateDirectory(path)

			switch {
			case tt.shouldErr:
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errCode != "" && !models.IsErrorCode(err, tt.errCode) {
					t.Errorf("Expected error code %s, got %v", tt.errCode, err)
				}
			case err != nil:
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPathValidator_ValidateDirectoryWritable(t *testing.T) {
	validator := NewPathValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		shouldErr bool
	}{
		{
			name: "writable directory",
			setup: func() string {
				return tempDir
			},
			shouldErr: false,
		},
		{
			name: "nonexistent directory",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			err := validator.ValidateDirectoryWritable(path)

			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPathValidator_ValidateDirectoryEmpty(t *testing.T) {
	validator := NewPathValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() string
		shouldErr bool
		errCode   models.ErrorCode
	}{
		{
			name: "empty directory",
			setup: func() string {
				emptyDir := filepath.Join(tempDir, "empty")
				_ = os.MkdirAll(emptyDir, 0755)
				return emptyDir
			},
			shouldErr: false,
		},
		{
			name: "directory with files",
			setup: func() string {
				dirWithFiles := filepath.Join(tempDir, "with-files")
				_ = os.MkdirAll(dirWithFiles, 0755)
				_ = os.WriteFile(filepath.Join(dirWithFiles, "test.txt"), []byte("test"), 0644)
				return dirWithFiles
			},
			shouldErr: true,
			errCode:   models.ErrorCodeDirectoryNotEmpty,
		},
		{
			name: "nonexistent directory",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			shouldErr: true,
			errCode:   models.ErrorCodeDirectoryNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup()

			err := validator.ValidateDirectoryEmpty(path)

			switch {
			case tt.shouldErr:
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if tt.errCode != "" && !models.IsErrorCode(err, tt.errCode) {
					t.Errorf("Expected error code %s, got %v", tt.errCode, err)
				}
			case err != nil:
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestPathValidator_ResolvePath(t *testing.T) {
	validator := NewPathValidator()

	tests := []struct {
		name      string
		path      string
		shouldErr bool
	}{
		{
			name:      "empty path",
			path:      "",
			shouldErr: true,
		},
		{
			name:      "relative path",
			path:      ".",
			shouldErr: false,
		},
		{
			name:      "absolute path",
			path:      "/tmp",
			shouldErr: false,
		},
		{
			name:      "relative nested path",
			path:      "./test/path",
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ResolvePath(tt.path)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if !filepath.IsAbs(result) {
					t.Errorf("Expected absolute path, got %s", result)
				}
			}
		})
	}
}

func TestInputValidator_ValidateInstallConfig(t *testing.T) {
	validator := NewInputValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		setup     func() *models.InstallConfig
		shouldErr bool
	}{
		{
			name: "nil config",
			setup: func() *models.InstallConfig {
				return nil
			},
			shouldErr: true,
		},
		{
			name: "empty target directory",
			setup: func() *models.InstallConfig {
				return &models.InstallConfig{
					TargetDir: "",
				}
			},
			shouldErr: true,
		},
		{
			name: "valid config",
			setup: func() *models.InstallConfig {
				return &models.InstallConfig{
					TargetDir: tempDir,
					Force:     false,
					ForceCore: false,
					NoBackup:  false,
				}
			},
			shouldErr: false,
		},
		{
			name: "conflicting flags - force and force-core",
			setup: func() *models.InstallConfig {
				return &models.InstallConfig{
					TargetDir: tempDir,
					Force:     true,
					ForceCore: true,
				}
			},
			shouldErr: true,
		},
		{
			name: "conflicting flags - no-backup and backup-dir",
			setup: func() *models.InstallConfig {
				return &models.InstallConfig{
					TargetDir: tempDir,
					NoBackup:  true,
					BackupDir: "/tmp/backup",
				}
			},
			shouldErr: true,
		},
		{
			name: "invalid backup directory parent",
			setup: func() *models.InstallConfig {
				return &models.InstallConfig{
					TargetDir: tempDir,
					BackupDir: "/nonexistent/parent/backup",
				}
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setup()

			err := validator.ValidateInstallConfig(config)

			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestInputValidator_ValidateCleanConfig(t *testing.T) {
	validator := NewInputValidator()

	tests := []struct {
		name      string
		setup     func() *models.CleanConfig
		shouldErr bool
	}{
		{
			name: "nil config",
			setup: func() *models.CleanConfig {
				return nil
			},
			shouldErr: true,
		},
		{
			name: "empty target directory",
			setup: func() *models.CleanConfig {
				return &models.CleanConfig{
					TargetDir: "",
				}
			},
			shouldErr: true,
		},
		{
			name: "valid config",
			setup: func() *models.CleanConfig {
				return &models.CleanConfig{
					TargetDir: "/tmp/test",
					Force:     false,
				}
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setup()

			err := validator.ValidateCleanConfig(config)

			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestFileSystemValidator_ValidateSymlink(t *testing.T) {
	validator := NewFileSystemValidator()
	tempDir := t.TempDir()

	tests := []struct {
		name         string
		setup        func() (symlinkPath, expectedTarget string)
		expectValid  bool
		expectExists bool
		expectError  bool
	}{
		{
			name: "nonexistent symlink",
			setup: func() (string, string) {
				symlinkPath := filepath.Join(tempDir, "nonexistent-link")
				expectedTarget := filepath.Join(tempDir, "target")
				return symlinkPath, expectedTarget
			},
			expectValid:  false,
			expectExists: false,
			expectError:  false,
		},
		{
			name: "valid symlink",
			setup: func() (string, string) {
				targetDir := filepath.Join(tempDir, "valid-target")
				_ = os.MkdirAll(targetDir, 0755)
				symlinkPath := filepath.Join(tempDir, "valid-link")
				_ = os.Symlink(targetDir, symlinkPath)
				return symlinkPath, targetDir
			},
			expectValid:  true,
			expectExists: true,
			expectError:  false,
		},
		{
			name: "symlink with wrong target",
			setup: func() (string, string) {
				actualTarget := filepath.Join(tempDir, "actual-target")
				expectedTarget := filepath.Join(tempDir, "expected-target")
				_ = os.MkdirAll(actualTarget, 0755)
				_ = os.MkdirAll(expectedTarget, 0755)
				symlinkPath := filepath.Join(tempDir, "wrong-link")
				_ = os.Symlink(actualTarget, symlinkPath)
				return symlinkPath, expectedTarget
			},
			expectValid:  false,
			expectExists: true,
			expectError:  false,
		},
		{
			name: "regular file instead of symlink",
			setup: func() (string, string) {
				filePath := filepath.Join(tempDir, "regular-file")
				_ = os.WriteFile(filePath, []byte("test"), 0644)
				expectedTarget := filepath.Join(tempDir, "target")
				return filePath, expectedTarget
			},
			expectValid:  false,
			expectExists: true,
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			symlinkPath, expectedTarget := tt.setup()

			status, err := validator.ValidateSymlink(symlinkPath, expectedTarget)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if status.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, status.Valid)
			}

			if status.Exists != tt.expectExists {
				t.Errorf("Expected exists=%v, got %v", tt.expectExists, status.Exists)
			}

			if status.Name != filepath.Base(symlinkPath) {
				t.Errorf("Expected name=%s, got %s", filepath.Base(symlinkPath), status.Name)
			}

			if status.Path != symlinkPath {
				t.Errorf("Expected path=%s, got %s", symlinkPath, status.Path)
			}
		})
	}
}

func TestValidateDirectoryName(t *testing.T) {
	tests := []struct {
		name      string
		dirName   string
		shouldErr bool
	}{
		{
			name:      "empty name",
			dirName:   "",
			shouldErr: true,
		},
		{
			name:      "valid name",
			dirName:   "valid-directory",
			shouldErr: false,
		},
		{
			name:      "name with spaces",
			dirName:   "directory with spaces",
			shouldErr: false,
		},
		{
			name:      "name with invalid characters",
			dirName:   "dir<>name",
			shouldErr: true,
		},
		{
			name:      "reserved name CON",
			dirName:   "CON",
			shouldErr: true,
		},
		{
			name:      "reserved name con (lowercase)",
			dirName:   "con",
			shouldErr: true,
		},
		{
			name:      "reserved name PRN",
			dirName:   "PRN",
			shouldErr: true,
		},
		{
			name:      "name with colon",
			dirName:   "dir:name",
			shouldErr: true,
		},
		{
			name:      "name with question mark",
			dirName:   "dir?name",
			shouldErr: true,
		},
		{
			name:      "name with asterisk",
			dirName:   "dir*name",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDirectoryName(tt.dirName)

			if tt.shouldErr && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestCheckGitAvailable(t *testing.T) {
	// This test checks if the function works, but the result depends on system state
	err := CheckGitAvailable()

	// We can't assert a specific result since it depends on whether git is installed
	// But we can check that it returns either nil or a specific error
	if err != nil {
		if !models.IsErrorCode(err, models.ErrorCodeGitNotInstalled) {
			t.Errorf("Expected ErrorCodeGitNotInstalled or nil, got %v", err)
		}
	}
}

// Test helpers for creating common test scenarios
func createTestSymlink(t *testing.T, tempDir, name, target string) string {
	t.Helper()

	// Create target directory
	_ = os.MkdirAll(target, 0755)

	// Create symlink
	symlinkPath := filepath.Join(tempDir, name)
	err := os.Symlink(target, symlinkPath)
	if err != nil {
		t.Fatalf("Failed to create test symlink: %v", err)
	}

	return symlinkPath
}

func TestSymlinkHelpers(t *testing.T) {
	tempDir := t.TempDir()

	// Test the helper function itself
	target := filepath.Join(tempDir, "target")
	symlinkPath := createTestSymlink(t, tempDir, "test-link", target)

	// Verify the symlink was created correctly
	if _, err := os.Lstat(symlinkPath); err != nil {
		t.Errorf("Test symlink was not created properly: %v", err)
	}

	// Verify it's actually a symlink
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to stat symlink: %v", err)
	}

	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Created path is not a symlink")
	}
}

// Benchmark tests for performance-critical validation functions
func BenchmarkPathValidator_ValidateDirectory(b *testing.B) {
	validator := NewPathValidator()
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateDirectory(tempDir)
	}
}

func BenchmarkValidateDirectoryName(b *testing.B) {
	testName := "valid-directory-name"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateDirectoryName(testName)
	}
}
