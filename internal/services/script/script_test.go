package script

import (
	"os"
	"path/filepath"
	"testing"
)

func TestService_ScriptExists(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "script-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := New()

	tests := []struct {
		name     string
		setup    func() error
		script   string
		expected bool
	}{
		{
			name: "script exists",
			setup: func() error {
				scriptPath := filepath.Join(tempDir, "test-script.sh")
				return os.WriteFile(scriptPath, []byte("#!/bin/bash\necho test"), 0755)
			},
			script:   "test-script.sh",
			expected: true,
		},
		{
			name:     "script does not exist",
			setup:    func() error { return nil },
			script:   "nonexistent.sh",
			expected: false,
		},
		{
			name:     "empty script name",
			setup:    func() error { return nil },
			script:   "",
			expected: false,
		},
		{
			name:     "empty source dir",
			setup:    func() error { return nil },
			script:   "test.sh",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.setup(); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			sourceDir := tempDir
			if tt.name == "empty source dir" {
				sourceDir = ""
			}

			got := service.ScriptExists(sourceDir, tt.script)
			if got != tt.expected {
				t.Errorf("ScriptExists() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestService_CopyScript(t *testing.T) {
	// Create source directory
	sourceDir, err := os.MkdirTemp("", "script-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	// Create target directory
	targetDir, err := os.MkdirTemp("", "script-target-")
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	defer os.RemoveAll(targetDir)

	service := New()

	// Create a test script
	scriptContent := "#!/bin/bash\necho 'Hello World'\n"
	scriptPath := filepath.Join(sourceDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Test copying the script
	err = service.CopyScript(sourceDir, targetDir, "test.sh")
	if err != nil {
		t.Fatalf("CopyScript() error = %v", err)
	}

	// Verify the script was copied
	targetScriptPath := filepath.Join(targetDir, "test.sh")
	copiedContent, err := os.ReadFile(targetScriptPath)
	if err != nil {
		t.Fatalf("Failed to read copied script: %v", err)
	}

	if string(copiedContent) != scriptContent {
		t.Errorf("Copied script content = %q, want %q", string(copiedContent), scriptContent)
	}

	// Verify permissions (should be executable)
	info, err := os.Stat(targetScriptPath)
	if err != nil {
		t.Fatalf("Failed to stat copied script: %v", err)
	}

	if info.Mode().Perm() != 0755 {
		t.Errorf("Copied script permissions = %v, want %v", info.Mode().Perm(), 0755)
	}
}

func TestService_CopyScript_NonexistentScript(t *testing.T) {
	sourceDir, err := os.MkdirTemp("", "script-source-")
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}
	defer os.RemoveAll(sourceDir)

	targetDir, err := os.MkdirTemp("", "script-target-")
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	defer os.RemoveAll(targetDir)

	service := New()

	// Try to copy a script that doesn't exist - should not return error
	err = service.CopyScript(sourceDir, targetDir, "nonexistent.sh")
	if err != nil {
		t.Errorf("CopyScript() with nonexistent script should not error, got: %v", err)
	}

	// Verify no file was created
	targetScriptPath := filepath.Join(targetDir, "nonexistent.sh")
	if _, err := os.Stat(targetScriptPath); err == nil {
		t.Error("Expected nonexistent script not to be copied, but file exists")
	}
}

func TestService_RemoveScript(t *testing.T) {
	targetDir, err := os.MkdirTemp("", "script-target-")
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	defer os.RemoveAll(targetDir)

	service := New()

	// Create a test script
	scriptContent := "#!/bin/bash\necho 'test'\n"
	scriptPath := filepath.Join(targetDir, "test.sh")
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Remove the script
	err = service.RemoveScript(targetDir, "test.sh")
	if err != nil {
		t.Fatalf("RemoveScript() error = %v", err)
	}

	// Verify the script was removed
	if _, err := os.Stat(scriptPath); err == nil {
		t.Error("Expected script to be removed, but file still exists")
	}
}

func TestService_RemoveScript_NonexistentScript(t *testing.T) {
	targetDir, err := os.MkdirTemp("", "script-target-")
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}
	defer os.RemoveAll(targetDir)

	service := New()

	// Try to remove a script that doesn't exist - should not return error
	err = service.RemoveScript(targetDir, "nonexistent.sh")
	if err != nil {
		t.Errorf("RemoveScript() with nonexistent script should not error, got: %v", err)
	}
}

func TestService_GetScriptPath(t *testing.T) {
	service := New()

	tests := []struct {
		name       string
		targetDir  string
		scriptName string
		expected   string
	}{
		{
			name:       "valid paths",
			targetDir:  "/test/dir",
			scriptName: "script.sh",
			expected:   "/test/dir/script.sh",
		},
		{
			name:       "empty target dir",
			targetDir:  "",
			scriptName: "script.sh",
			expected:   "",
		},
		{
			name:       "empty script name",
			targetDir:  "/test/dir",
			scriptName: "",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.GetScriptPath(tt.targetDir, tt.scriptName)
			if got != tt.expected {
				t.Errorf("GetScriptPath() = %v, want %v", got, tt.expected)
			}
		})
	}
}
