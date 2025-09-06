package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
)

func TestService_ProcessSettings(t *testing.T) {
	tests := []struct {
		name               string
		templateSettings   *models.ClaudeSettings
		existingSettings   *models.ClaudeSettings
		expectError        bool
		expectSettingsFile bool
		expectBackup       bool
	}{
		{
			name: "new installation with template",
			templateSettings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PreToolUse: []models.HookMatcher{
						{
							Matcher: "Bash",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/block-skip-hooks.py"},
							},
						},
					},
				},
			},
			existingSettings:   nil,
			expectError:        false,
			expectSettingsFile: true,
			expectBackup:       false,
		},
		{
			name: "update with existing user settings",
			templateSettings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PreToolUse: []models.HookMatcher{
						{
							Matcher: "Bash",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/block-skip-hooks.py"},
							},
						},
					},
				},
			},
			existingSettings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PostToolUse: []models.HookMatcher{
						{
							Matcher: "Write|Edit|MultiEdit",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
							},
						},
					},
				},
				Permissions: &models.PermissionsSection{
					Allow:                 []string{"Read(/home/user/**)", "Bash(go env:*)"},
					AdditionalDirectories: []string{"/home/user/projects"},
				},
			},
			expectError:        false,
			expectSettingsFile: true,
			expectBackup:       true,
		},
		{
			name:               "no template file",
			templateSettings:   nil,
			existingSettings:   nil,
			expectError:        false,
			expectSettingsFile: false,
			expectBackup:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tempDir := t.TempDir()
			strategicDir := filepath.Join(tempDir, config.StrategicClaudeBasicDir)
			claudeDir := filepath.Join(tempDir, config.ClaudeDir)
			templatesDir := filepath.Join(strategicDir, "templates", "hooks")

			if err := os.MkdirAll(templatesDir, 0755); err != nil {
				t.Fatalf("Failed to create templates dir: %v", err)
			}
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("Failed to create claude dir: %v", err)
			}

			// Create template file if provided
			templatePath := filepath.Join(templatesDir, "settings.template.json")
			if tt.templateSettings != nil {
				templateData, _ := json.MarshalIndent(tt.templateSettings, "", "  ")
				if err := os.WriteFile(templatePath, templateData, 0644); err != nil {
					t.Fatalf("Failed to write template file: %v", err)
				}
			}

			// Create existing settings file if provided
			settingsPath := filepath.Join(claudeDir, config.ClaudeSettingsFile)
			if tt.existingSettings != nil {
				existingData, _ := json.MarshalIndent(tt.existingSettings, "", "  ")
				if err := os.WriteFile(settingsPath, existingData, 0644); err != nil {
					t.Fatalf("Failed to write existing settings file: %v", err)
				}
			}

			// Run the test
			service := New()
			err := service.ProcessSettings(tempDir)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check if settings file was created
			_, settingsExists := os.Stat(settingsPath)
			if tt.expectSettingsFile && os.IsNotExist(settingsExists) {
				t.Error("Expected settings file to be created but it doesn't exist")
			}
			if !tt.expectSettingsFile && settingsExists == nil {
				t.Error("Expected no settings file but it exists")
			}

			// Check if backup was created
			if tt.expectBackup {
				files, err := os.ReadDir(claudeDir)
				if err != nil {
					t.Fatalf("Failed to read claude directory: %v", err)
				}

				hasBackup := false
				for _, file := range files {
					if strings.HasPrefix(file.Name(), config.SettingsBackupPrefix) {
						hasBackup = true
						break
					}
				}

				if !hasBackup {
					t.Error("Expected backup file but none found")
				}
			}

			// If settings file exists, validate the merge
			if tt.expectSettingsFile && settingsExists == nil {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("Failed to read settings file: %v", err)
				}

				var result models.ClaudeSettings
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("Failed to unmarshal settings: %v", err)
				}

				// Validate merge results
				if tt.existingSettings != nil && tt.existingSettings.Permissions != nil {
					if result.Permissions == nil {
						t.Error("Expected permissions to be preserved")
					} else if len(result.Permissions.Allow) != len(tt.existingSettings.Permissions.Allow) {
						t.Error("Permissions.Allow not preserved correctly")
					}
				}

				// Check that strategic hooks use correct paths
				if result.Hooks != nil {
					checkStrategicHookPaths(t, result.Hooks)
				}
			}
		})
	}
}

func TestService_mergeHooks(t *testing.T) {
	service := New()

	templateHooks := &models.HooksSection{
		PreToolUse: []models.HookMatcher{
			{
				Matcher: "Bash",
				Hooks: []models.HookEntry{
					{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/block-skip-hooks.py"},
				},
			},
		},
	}

	existing := &models.ClaudeSettings{
		Hooks: &models.HooksSection{
			PostToolUse: []models.HookMatcher{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []models.HookEntry{
						{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
					},
				},
			},
		},
		Permissions: &models.PermissionsSection{
			Allow: []string{"Read(/home/user/**)"},
		},
	}

	result := service.mergeHooks(templateHooks, existing)

	// Should have both PreToolUse and PostToolUse
	if len(result.PreToolUse) != 1 {
		t.Errorf("Expected 1 PreToolUse matcher, got %d", len(result.PreToolUse))
	}
	if len(result.PostToolUse) != 1 {
		t.Errorf("Expected 1 PostToolUse matcher, got %d", len(result.PostToolUse))
	}

	// Check that strategic hook is present
	found := false
	for _, matcher := range result.PreToolUse {
		if matcher.Matcher == "Bash" {
			for _, hook := range matcher.Hooks {
				if strings.Contains(hook.Command, "block-skip-hooks.py") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("Strategic hook not found in merged result")
	}

	// Check that user's custom hook is preserved
	found = false
	for _, matcher := range result.PostToolUse {
		if matcher.Matcher == "Write|Edit|MultiEdit" {
			for _, hook := range matcher.Hooks {
				if strings.Contains(hook.Command, "format-go-hook.py") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("User's custom hook not preserved in merged result")
	}
}

func TestService_hookExists(t *testing.T) {
	service := New()

	existingHooks := []models.HookEntry{
		{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
		{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/block-skip-hooks.py"},
	}

	tests := []struct {
		name     string
		target   models.HookEntry
		expected bool
	}{
		{
			name:     "exact match",
			target:   models.HookEntry{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
			expected: true,
		},
		{
			name:     "strategic hook variation",
			target:   models.HookEntry{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/block-skip-hooks.py"},
			expected: true,
		},
		{
			name:     "different hook",
			target:   models.HookEntry{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/other-hook.py"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.hookExists(existingHooks, tt.target)
			if result != tt.expected {
				t.Errorf("hookExists() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestModels_IsStrategicHook(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		expected bool
	}{
		{
			name:     "block-skip-hooks",
			command:  "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/block-skip-hooks.py",
			expected: true,
		},
		{
			name:     "notification-hook",
			command:  "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/notification-hook.py",
			expected: true,
		},
		{
			name:     "user custom hook",
			command:  "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py",
			expected: false,
		},
		{
			name:     "different path same script",
			command:  "/usr/bin/python3 /some/other/path/block-skip-hooks.py",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := models.IsStrategicHook(tt.command)
			if result != tt.expected {
				t.Errorf("IsStrategicHook(%s) = %v, want %v", tt.command, result, tt.expected)
			}
		})
	}
}

func TestService_CleanSettings(t *testing.T) {
	tests := []struct {
		name             string
		existingSettings *models.ClaudeSettings
		expectError      bool
		expectRemoved    bool
		expectBackup     bool
	}{
		{
			name:             "no settings file",
			existingSettings: nil,
			expectError:      false,
			expectRemoved:    false,
			expectBackup:     false,
		},
		{
			name: "only strategic hooks - file should be removed",
			existingSettings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PreToolUse: []models.HookMatcher{
						{
							Matcher: "Bash",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/block-skip-hooks.py"},
							},
						},
					},
				},
			},
			expectError:   false,
			expectRemoved: true,
			expectBackup:  true,
		},
		{
			name: "mixed hooks - keep user hooks, remove strategic",
			existingSettings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PreToolUse: []models.HookMatcher{
						{
							Matcher: "Bash",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/block-skip-hooks.py"},
							},
						},
					},
					PostToolUse: []models.HookMatcher{
						{
							Matcher: "Write|Edit|MultiEdit",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
							},
						},
					},
				},
				Permissions: &models.PermissionsSection{
					Allow: []string{"Read(/home/user/**)"},
				},
			},
			expectError:   false,
			expectRemoved: false,
			expectBackup:  true,
		},
		{
			name: "only permissions - keep file",
			existingSettings: &models.ClaudeSettings{
				Permissions: &models.PermissionsSection{
					Allow:                 []string{"Read(/home/user/**)"},
					AdditionalDirectories: []string{"/home/user/projects"},
				},
			},
			expectError:   false,
			expectRemoved: false,
			expectBackup:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory structure
			tempDir := t.TempDir()
			claudeDir := filepath.Join(tempDir, config.ClaudeDir)
			if err := os.MkdirAll(claudeDir, 0755); err != nil {
				t.Fatalf("Failed to create claude dir: %v", err)
			}

			settingsPath := filepath.Join(claudeDir, config.ClaudeSettingsFile)

			// Create existing settings file if provided
			if tt.existingSettings != nil {
				settingsData, _ := json.MarshalIndent(tt.existingSettings, "", "  ")
				if err := os.WriteFile(settingsPath, settingsData, 0644); err != nil {
					t.Fatalf("Failed to write existing settings: %v", err)
				}
			}

			// Run the cleanup
			service := New()
			err := service.CleanSettings(tempDir)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check if file was removed
			_, settingsExists := os.Stat(settingsPath)
			if tt.expectRemoved && settingsExists == nil {
				t.Error("Expected settings file to be removed but it still exists")
			}
			if !tt.expectRemoved && tt.existingSettings != nil && os.IsNotExist(settingsExists) {
				t.Error("Expected settings file to be kept but it was removed")
			}

			// Check if backup was created
			if tt.expectBackup {
				files, err := os.ReadDir(claudeDir)
				if err != nil {
					t.Fatalf("Failed to read claude directory: %v", err)
				}

				hasBackup := false
				for _, file := range files {
					if strings.HasPrefix(file.Name(), config.SettingsBackupPrefix) {
						hasBackup = true
						break
					}
				}

				if !hasBackup {
					t.Error("Expected backup file but none found")
				}
			}

			// If file still exists, verify cleanup worked correctly
			if !tt.expectRemoved && settingsExists == nil {
				data, err := os.ReadFile(settingsPath)
				if err != nil {
					t.Fatalf("Failed to read cleaned settings: %v", err)
				}

				var result models.ClaudeSettings
				if err := json.Unmarshal(data, &result); err != nil {
					t.Fatalf("Failed to unmarshal cleaned settings: %v", err)
				}

				// Check that no strategic hooks remain
				if result.Hooks != nil {
					checkNoStrategicHooks(t, result.Hooks)
				}

				// Check that permissions are preserved
				if tt.existingSettings != nil && tt.existingSettings.Permissions != nil {
					if result.Permissions == nil {
						t.Error("Expected permissions to be preserved")
					}
				}
			}
		})
	}
}

func TestService_removeStrategicHooks(t *testing.T) {
	service := New()

	input := &models.ClaudeSettings{
		Hooks: &models.HooksSection{
			PreToolUse: []models.HookMatcher{
				{
					Matcher: "Bash",
					Hooks: []models.HookEntry{
						{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/block-skip-hooks.py"},
						{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/user-hook.py"},
					},
				},
			},
			PostToolUse: []models.HookMatcher{
				{
					Matcher: "Write|Edit|MultiEdit",
					Hooks: []models.HookEntry{
						{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/format-go-hook.py"},
					},
				},
			},
		},
		Permissions: &models.PermissionsSection{
			Allow: []string{"Read(/home/user/**)"},
		},
	}

	result := service.removeStrategicHooks(input)

	// Should preserve permissions
	if result.Permissions == nil || len(result.Permissions.Allow) == 0 {
		t.Error("Expected permissions to be preserved")
	}

	// Should keep user hook
	found := false
	if result.Hooks != nil && result.Hooks.PreToolUse != nil {
		for _, matcher := range result.Hooks.PreToolUse {
			for _, hook := range matcher.Hooks {
				if strings.Contains(hook.Command, "user-hook.py") {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Error("Expected user hook to be preserved")
	}

	// Should remove strategic hook
	foundStrategic := false
	if result.Hooks != nil && result.Hooks.PreToolUse != nil {
		for _, matcher := range result.Hooks.PreToolUse {
			for _, hook := range matcher.Hooks {
				if models.IsStrategicHook(hook.Command) {
					foundStrategic = true
					break
				}
			}
		}
	}
	if foundStrategic {
		t.Error("Expected strategic hooks to be removed")
	}

	// Should keep PostToolUse (user hook)
	if result.Hooks == nil || len(result.Hooks.PostToolUse) == 0 {
		t.Error("Expected PostToolUse hooks to be preserved")
	}
}

func TestService_isEmptySettings(t *testing.T) {
	service := New()

	tests := []struct {
		name     string
		settings *models.ClaudeSettings
		expected bool
	}{
		{
			name:     "nil settings",
			settings: nil,
			expected: true,
		},
		{
			name:     "empty settings",
			settings: &models.ClaudeSettings{},
			expected: true,
		},
		{
			name: "settings with permissions",
			settings: &models.ClaudeSettings{
				Permissions: &models.PermissionsSection{
					Allow: []string{"Read(/home/user/**)"},
				},
			},
			expected: false,
		},
		{
			name: "settings with hooks",
			settings: &models.ClaudeSettings{
				Hooks: &models.HooksSection{
					PostToolUse: []models.HookMatcher{
						{
							Matcher: "Write|Edit|MultiEdit",
							Hooks: []models.HookEntry{
								{Type: "command", Command: "/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/user-hook.py"},
							},
						},
					},
				},
			},
			expected: false,
		},
		{
			name: "settings with empty hooks and permissions",
			settings: &models.ClaudeSettings{
				Hooks:       &models.HooksSection{},
				Permissions: &models.PermissionsSection{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.isEmptySettings(tt.settings)
			if result != tt.expected {
				t.Errorf("isEmptySettings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Helper function to check that no strategic hooks remain
func checkNoStrategicHooks(t *testing.T, hooks *models.HooksSection) {
	checkHookTypeStrategic := func(matchers []models.HookMatcher, hookType string) {
		for _, matcher := range matchers {
			for _, hook := range matcher.Hooks {
				if models.IsStrategicHook(hook.Command) {
					t.Errorf("Found strategic hook in %s after cleanup: %s", hookType, hook.Command)
				}
			}
		}
	}

	if hooks.PreToolUse != nil {
		checkHookTypeStrategic(hooks.PreToolUse, "PreToolUse")
	}
	if hooks.PostToolUse != nil {
		checkHookTypeStrategic(hooks.PostToolUse, "PostToolUse")
	}
	if hooks.Stop != nil {
		checkHookTypeStrategic(hooks.Stop, "Stop")
	}
	if hooks.PreCompact != nil {
		checkHookTypeStrategic(hooks.PreCompact, "PreCompact")
	}
	if hooks.Notification != nil {
		checkHookTypeStrategic(hooks.Notification, "Notification")
	}
}

// Helper function to check that strategic hooks use correct paths
func checkStrategicHookPaths(t *testing.T, hooks *models.HooksSection) {
	checkHookTypePaths := func(matchers []models.HookMatcher, hookType string) {
		for _, matcher := range matchers {
			for _, hook := range matcher.Hooks {
				if models.IsStrategicHook(hook.Command) && !strings.Contains(hook.Command, "/strategic/") {
					t.Errorf("Strategic hook in %s doesn't use strategic path: %s", hookType, hook.Command)
				}
			}
		}
	}

	if hooks.PreToolUse != nil {
		checkHookTypePaths(hooks.PreToolUse, "PreToolUse")
	}
	if hooks.PostToolUse != nil {
		checkHookTypePaths(hooks.PostToolUse, "PostToolUse")
	}
	if hooks.Stop != nil {
		checkHookTypePaths(hooks.Stop, "Stop")
	}
	if hooks.PreCompact != nil {
		checkHookTypePaths(hooks.PreCompact, "PreCompact")
	}
	if hooks.Notification != nil {
		checkHookTypePaths(hooks.Notification, "Notification")
	}
}
