package codexconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
)

func TestProcessCodexConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "codex-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := New()

	// Test case 1: No template exists - should not create config
	err = service.ProcessCodexConfig(tempDir)
	if err != nil {
		t.Errorf("Expected no error when template doesn't exist, got: %v", err)
	}

	// Check that no config was created
	codexDir := filepath.Join(tempDir, config.CodexDir)
	configPath := filepath.Join(codexDir, config.CodexConfigFile)
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should not exist when template doesn't exist")
	}

	// Test case 2: Template exists - should create config
	strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
	templateDir := filepath.Join(strategicDir, "templates", "hooks")
	err = os.MkdirAll(templateDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create template directory: %v", err)
	}

	templatePath := filepath.Join(strategicDir, config.CodexConfigTemplateFile)
	templateContent := `# Codex Configuration
[general]
enabled = true
`
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create template file: %v", err)
	}

	// Process config
	err = service.ProcessCodexConfig(tempDir)
	if err != nil {
		t.Errorf("ProcessCodexConfig failed: %v", err)
	}

	// Check that config was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file should exist after processing template")
	}

	// Check content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	if string(content) != templateContent {
		t.Errorf("Config content mismatch. Expected: %s, Got: %s", templateContent, string(content))
	}
}

func TestRemoveCodexConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "codex-config-remove-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := New()

	// Create .codex directory and config file
	codexDir := filepath.Join(tempDir, config.CodexDir)
	err = os.MkdirAll(codexDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create codex directory: %v", err)
	}

	configPath := filepath.Join(codexDir, config.CodexConfigFile)
	err = os.WriteFile(configPath, []byte("test config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Remove config
	err = service.RemoveCodexConfig(tempDir)
	if err != nil {
		t.Errorf("RemoveCodexConfig failed: %v", err)
	}

	// Check that config was removed
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		t.Error("Config file should be removed")
	}
}

func TestValidateCodexConfig(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "codex-config-validate-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	service := New()

	// Test case 1: Config doesn't exist
	err = service.ValidateCodexConfig(tempDir)
	if err == nil {
		t.Error("Expected error when config doesn't exist")
	}

	// Test case 2: Config exists and is valid
	codexDir := filepath.Join(tempDir, config.CodexDir)
	err = os.MkdirAll(codexDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create codex directory: %v", err)
	}

	configPath := filepath.Join(codexDir, config.CodexConfigFile)
	err = os.WriteFile(configPath, []byte("test config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	err = service.ValidateCodexConfig(tempDir)
	if err != nil {
		t.Errorf("ValidateCodexConfig failed for valid config: %v", err)
	}
}