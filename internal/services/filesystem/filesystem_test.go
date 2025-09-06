package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
)

func TestService_CreateDirectory(t *testing.T) {
	service := New()

	tests := []struct {
		name      string
		path      string
		shouldErr bool
		errCode   models.ErrorCode
	}{
		{
			name:      "empty path",
			path:      "",
			shouldErr: true,
			errCode:   models.ErrorCodeValidationFailed,
		},
		{
			name:      "valid path",
			path:      "test-dir",
			shouldErr: false,
		},
		{
			name:      "nested path",
			path:      "test-parent/test-child",
			shouldErr: false,
		},
	}

	// Use temp directory for tests
	tempDir := t.TempDir()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.path
			if testPath != "" {
				testPath = filepath.Join(tempDir, tt.path)
			}

			err := service.CreateDirectory(testPath)

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
			default:
				// Verify directory was created
				if _, err := os.Stat(testPath); os.IsNotExist(err) {
					t.Error("Directory was not created")
				}
			}
		})
	}
}

func TestService_CreateDirectory_ExistingFile(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Create a file first
	filePath := filepath.Join(tempDir, "existing-file")
	file, err := os.Create(filePath)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	file.Close()

	// Try to create directory with same name
	err = service.CreateDirectory(filePath)
	if err == nil {
		t.Error("Expected error when trying to create directory over existing file")
	}
	if !models.IsErrorCode(err, models.ErrorCodeFileAlreadyExists) {
		t.Errorf("Expected ErrorCodeFileAlreadyExists, got %v", err)
	}
}

func TestService_RemoveDirectory(t *testing.T) {
	service := New()
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
			name: "nonexistent directory",
			setup: func() string {
				return filepath.Join(tempDir, "nonexistent")
			},
			shouldErr: false, // Should succeed (already doesn't exist)
		},
		{
			name: "valid directory",
			setup: func() string {
				testDir := filepath.Join(tempDir, "test-remove")
				_ = os.MkdirAll(testDir, 0755)
				return testDir
			},
			shouldErr: false,
		},
		{
			name: "system path",
			setup: func() string {
				return "/bin" // System directory
			},
			shouldErr: true,
			errCode:   models.ErrorCodeValidationFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.setup()

			err := service.RemoveDirectory(testPath)

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
			case testPath != "" && !strings.HasPrefix(testPath, "/"):
				// Verify directory was removed (skip system paths)
				if _, err := os.Stat(testPath); !os.IsNotExist(err) {
					t.Error("Directory was not removed")
				}
			}
		})
	}
}

func TestService_BackupDirectory(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Create source directory with content
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Add some content
	testFile := filepath.Join(sourceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backupDir := filepath.Join(tempDir, "backup")

	// Test successful backup
	err := service.BackupDirectory(sourceDir, backupDir)
	if err != nil {
		t.Fatalf("Backup failed: %v", err)
	}

	// Verify backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		t.Error("Backup directory was not created")
	}

	// Verify backup content
	backupFile := filepath.Join(backupDir, "test.txt")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file was not created")
	}

	// Test backup of nonexistent directory
	err = service.BackupDirectory(filepath.Join(tempDir, "nonexistent"), filepath.Join(tempDir, "backup2"))
	if err == nil {
		t.Error("Expected error when backing up nonexistent directory")
	}
	if !models.IsErrorCode(err, models.ErrorCodeDirectoryNotFound) {
		t.Errorf("Expected ErrorCodeDirectoryNotFound, got %v", err)
	}
}

func TestService_EnsureDirectoryStructure(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	err := service.EnsureDirectoryStructure(tempDir)
	if err != nil {
		t.Fatalf("EnsureDirectoryStructure failed: %v", err)
	}

	// Verify main directory was created
	strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
	if _, err := os.Stat(strategicDir); os.IsNotExist(err) {
		t.Error("Strategic Claude Basic directory was not created")
	}

	// Verify framework directories
	frameworkDirs := config.GetFrameworkDirectories()
	for _, dir := range frameworkDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Framework directory %s was not created", dir)
		}
	}

	// Verify user directories
	userDirs := config.GetUserPreservedDirectories()
	for _, dir := range userDirs {
		dirPath := filepath.Join(strategicDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("User directory %s was not created", dir)
		}
	}

	// Verify core subdirectories
	coreDir := filepath.Join(strategicDir, config.CoreDir)
	coreSubdirs := []string{config.AgentsDir, config.CommandsDir, config.HooksDir}
	for _, subdir := range coreSubdirs {
		subdirPath := filepath.Join(coreDir, subdir)
		if _, err := os.Stat(subdirPath); os.IsNotExist(err) {
			t.Errorf("Core subdirectory %s was not created", subdir)
		}
	}
}

func TestService_CopyFile(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	testContent := "test file content"
	if err := os.WriteFile(sourceFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	destFile := filepath.Join(tempDir, "dest.txt")

	// Test successful copy
	err := service.CopyFile(sourceFile, destFile)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file exists
	if _, err := os.Stat(destFile); os.IsNotExist(err) {
		t.Error("Destination file was not created")
	}

	// Verify content
	destContent, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}
	if string(destContent) != testContent {
		t.Errorf("File content mismatch. Expected %q, got %q", testContent, string(destContent))
	}

	// Test copy of nonexistent file
	err = service.CopyFile(filepath.Join(tempDir, "nonexistent.txt"), filepath.Join(tempDir, "dest2.txt"))
	if err == nil {
		t.Error("Expected error when copying nonexistent file")
	}
	if !models.IsErrorCode(err, models.ErrorCodeDirectoryNotFound) {
		t.Errorf("Expected ErrorCodeDirectoryNotFound, got %v", err)
	}
}

func TestService_CopyDirectory(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Create source directory structure
	sourceDir := filepath.Join(tempDir, "source")
	if err := os.MkdirAll(filepath.Join(sourceDir, "subdir"), 0755); err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	// Add files
	sourceFile1 := filepath.Join(sourceDir, "file1.txt")
	sourceFile2 := filepath.Join(sourceDir, "subdir", "file2.txt")
	if err := os.WriteFile(sourceFile1, []byte("content1"), 0644); err != nil {
		t.Fatalf("Failed to create source file1: %v", err)
	}
	if err := os.WriteFile(sourceFile2, []byte("content2"), 0644); err != nil {
		t.Fatalf("Failed to create source file2: %v", err)
	}

	destDir := filepath.Join(tempDir, "dest")

	// Test successful directory copy
	err := service.CopyDirectory(sourceDir, destDir)
	if err != nil {
		t.Fatalf("CopyDirectory failed: %v", err)
	}

	// Verify destination structure
	destFile1 := filepath.Join(destDir, "file1.txt")
	destFile2 := filepath.Join(destDir, "subdir", "file2.txt")

	if _, err := os.Stat(destFile1); os.IsNotExist(err) {
		t.Error("Destination file1 was not created")
	}
	if _, err := os.Stat(destFile2); os.IsNotExist(err) {
		t.Error("Destination file2 was not created")
	}
	if _, err := os.Stat(filepath.Join(destDir, "subdir")); os.IsNotExist(err) {
		t.Error("Destination subdir was not created")
	}

	// Verify content
	content1, _ := os.ReadFile(destFile1)
	content2, _ := os.ReadFile(destFile2)
	if string(content1) != "content1" {
		t.Error("File1 content mismatch")
	}
	if string(content2) != "content2" {
		t.Error("File2 content mismatch")
	}
}

func TestService_CopyFrameworkFiles(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Create source directory with framework structure
	sourceDir := filepath.Join(tempDir, "source")
	frameworkDirs := config.GetCoreDirectories()

	for _, dir := range frameworkDirs {
		dirPath := filepath.Join(sourceDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create framework directory %s: %v", dir, err)
		}

		// Add a test file
		testFile := filepath.Join(dirPath, "test.txt")
		if err := os.WriteFile(testFile, []byte(dir+" content"), 0644); err != nil {
			t.Fatalf("Failed to create test file in %s: %v", dir, err)
		}
	}

	destDir := filepath.Join(tempDir, "dest")
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create dest directory: %v", err)
	}

	// Test framework files copy
	err := service.CopyFrameworkFiles(sourceDir, destDir)
	if err != nil {
		t.Fatalf("CopyFrameworkFiles failed: %v", err)
	}

	// Verify framework directories were copied
	for _, dir := range frameworkDirs {
		destDirPath := filepath.Join(destDir, dir)
		if _, err := os.Stat(destDirPath); os.IsNotExist(err) {
			t.Errorf("Framework directory %s was not copied", dir)
		}

		// Verify test file was copied
		testFile := filepath.Join(destDirPath, "test.txt")
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Errorf("Test file in %s was not copied", dir)
		}
	}
}

func TestService_IsSubPath(t *testing.T) {
	service := New()

	tests := []struct {
		name       string
		parentPath string
		childPath  string
		expected   bool
	}{
		{
			name:       "direct child",
			parentPath: "/parent",
			childPath:  "/parent/child",
			expected:   true,
		},
		{
			name:       "nested child",
			parentPath: "/parent",
			childPath:  "/parent/child/grandchild",
			expected:   true,
		},
		{
			name:       "same path",
			parentPath: "/parent",
			childPath:  "/parent",
			expected:   true,
		},
		{
			name:       "not a child",
			parentPath: "/parent",
			childPath:  "/other",
			expected:   false,
		},
		{
			name:       "similar prefix but not child",
			parentPath: "/parent",
			childPath:  "/parentother",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.IsSubPath(tt.parentPath, tt.childPath)
			if err != nil {
				t.Fatalf("IsSubPath failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("IsSubPath(%s, %s) = %v, expected %v", tt.parentPath, tt.childPath, result, tt.expected)
			}
		})
	}
}

func TestService_GetRelativePath(t *testing.T) {
	service := New()

	tests := []struct {
		name       string
		basePath   string
		targetPath string
		expected   string
	}{
		{
			name:       "sibling directories",
			basePath:   "/base/dir1",
			targetPath: "/base/dir2",
			expected:   "../dir2",
		},
		{
			name:       "parent to child",
			basePath:   "/base",
			targetPath: "/base/child",
			expected:   "child",
		},
		{
			name:       "child to parent",
			basePath:   "/base/child",
			targetPath: "/base",
			expected:   "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.GetRelativePath(tt.basePath, tt.targetPath)
			if err != nil {
				t.Fatalf("GetRelativePath failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("GetRelativePath(%s, %s) = %s, expected %s", tt.basePath, tt.targetPath, result, tt.expected)
			}
		})
	}
}

func TestService_CheckWritePermission(t *testing.T) {
	service := New()
	tempDir := t.TempDir()

	// Test writable directory
	err := service.CheckWritePermission(tempDir)
	if err != nil {
		t.Errorf("Expected writable temp directory to pass, got error: %v", err)
	}

	// Test nonexistent directory
	nonexistentDir := filepath.Join(tempDir, "nonexistent")
	err = service.CheckWritePermission(nonexistentDir)
	if err == nil {
		t.Error("Expected error for nonexistent directory")
	}
}

func TestService_GetBackupPath(t *testing.T) {
	service := New()
	targetDir := "/test/target"

	backupPath := service.GetBackupPath(targetDir)

	// Should be under target directory
	if !strings.HasPrefix(backupPath, targetDir) {
		t.Errorf("Backup path %s should be under target directory %s", backupPath, targetDir)
	}

	// Should contain backup prefix
	if !strings.Contains(backupPath, config.BackupDirPrefix) {
		t.Errorf("Backup path %s should contain backup prefix %s", backupPath, config.BackupDirPrefix)
	}

	// Should contain timestamp pattern (YYYYMMDD-HHMMSS)
	basename := filepath.Base(backupPath)
	if len(basename) != len(config.BackupDirPrefix)+15 { // prefix + timestamp
		t.Errorf("Backup path basename %s should have correct timestamp format", basename)
	}
}

// Benchmark tests
func BenchmarkService_CreateDirectory(b *testing.B) {
	service := New()
	tempDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dirPath := filepath.Join(tempDir, "bench", "test", fmt.Sprintf("dir%d", i))
		err := service.CreateDirectory(dirPath)
		if err != nil {
			b.Fatalf("CreateDirectory failed: %v", err)
		}
	}
}

func BenchmarkService_CopyFile(b *testing.B) {
	service := New()
	tempDir := b.TempDir()

	// Create source file
	sourceFile := filepath.Join(tempDir, "source.txt")
	testContent := strings.Repeat("test content ", 1000) // Larger content for meaningful benchmark
	if err := os.WriteFile(sourceFile, []byte(testContent), 0644); err != nil {
		b.Fatalf("Failed to create source file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		destFile := filepath.Join(tempDir, fmt.Sprintf("dest%d.txt", i))
		err := service.CopyFile(sourceFile, destFile)
		if err != nil {
			b.Fatalf("CopyFile failed: %v", err)
		}
	}
}
