package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

// Service provides settings management functionality
type Service struct{}

// New creates a new settings service instance
func New() *Service {
	return &Service{}
}

// ProcessSettings is the main entry point for managing .claude/settings.json
func (s *Service) ProcessSettings(targetDir string) error {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	settingsPath := filepath.Join(claudeDir, config.ClaudeSettingsFile)
	templatePath := filepath.Join(strategicDir, config.SettingsTemplateFile)

	// Check if template exists
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		// Template doesn't exist, nothing to do
		return nil
	}

	// Load template settings
	templateSettings, err := s.loadTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to load settings template: %w", err)
	}

	// Handle existing settings
	var existingSettings *models.ClaudeSettings
	if _, err := os.Stat(settingsPath); err == nil {
		// Backup existing settings
		if err := s.backupExistingSettings(settingsPath); err != nil {
			return fmt.Errorf("failed to backup existing settings: %w", err)
		}

		// Load existing settings
		existingSettings, err = s.loadExistingSettings(settingsPath)
		if err != nil {
			return fmt.Errorf("failed to load existing settings: %w", err)
		}
	}

	// Merge settings
	mergedSettings := s.mergeSettings(templateSettings, existingSettings)

	// Update hook paths to point to strategic directory
	s.updateStrategicHookPaths(mergedSettings)

	// Write merged settings
	if err := s.writeSettings(settingsPath, mergedSettings); err != nil {
		return fmt.Errorf("failed to write settings: %w", err)
	}

	return nil
}

// backupExistingSettings creates a timestamped backup of existing settings
func (s *Service) backupExistingSettings(settingsPath string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(
		filepath.Dir(settingsPath),
		config.SettingsBackupPrefix+timestamp+".json",
	)

	// Read existing file
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return err
	}

	// Write backup
	return os.WriteFile(backupPath, data, config.FilePermissions)
}

// loadTemplate loads the settings template from the framework
func (s *Service) loadTemplate(templatePath string) (*models.ClaudeSettings, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	var settings models.ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// loadExistingSettings loads the user's current settings
func (s *Service) loadExistingSettings(settingsPath string) (*models.ClaudeSettings, error) {
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, err
	}

	var settings models.ClaudeSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

// mergeSettings intelligently merges template settings with existing user settings
func (s *Service) mergeSettings(template *models.ClaudeSettings, existing *models.ClaudeSettings) *models.ClaudeSettings {
	result := &models.ClaudeSettings{}

	// Merge hooks section
	if template.Hooks != nil || (existing != nil && existing.Hooks != nil) {
		result.Hooks = s.mergeHooks(template.Hooks, existing)
	}

	// Preserve permissions section from existing settings
	if existing != nil && existing.Permissions != nil {
		result.Permissions = existing.Permissions
	}

	return result
}

// mergeHooks merges hook configurations by hook type and matcher
func (s *Service) mergeHooks(templateHooks *models.HooksSection, existing *models.ClaudeSettings) *models.HooksSection {
	result := &models.HooksSection{}

	if templateHooks == nil {
		templateHooks = &models.HooksSection{}
	}

	var existingHooks *models.HooksSection
	if existing != nil && existing.Hooks != nil {
		existingHooks = existing.Hooks
	} else {
		existingHooks = &models.HooksSection{}
	}

	// Merge each hook type
	result.PreToolUse = s.mergeHookType(templateHooks.PreToolUse, existingHooks.PreToolUse)
	result.PostToolUse = s.mergeHookType(templateHooks.PostToolUse, existingHooks.PostToolUse)
	result.Stop = s.mergeHookType(templateHooks.Stop, existingHooks.Stop)
	result.PreCompact = s.mergeHookType(templateHooks.PreCompact, existingHooks.PreCompact)
	result.Notification = s.mergeHookType(templateHooks.Notification, existingHooks.Notification)

	return result
}

// mergeHookType merges hooks for a specific hook type (PreToolUse, PostToolUse, etc.)
func (s *Service) mergeHookType(templateMatchers []models.HookMatcher, existingMatchers []models.HookMatcher) []models.HookMatcher {
	matcherMap := make(map[string][]models.HookEntry)

	// Add existing hooks first to preserve user customizations
	for _, matcher := range existingMatchers {
		matcherMap[matcher.Matcher] = append(matcherMap[matcher.Matcher], matcher.Hooks...)
	}

	// Add template hooks, avoiding duplicates
	for _, templateMatcher := range templateMatchers {
		existing := matcherMap[templateMatcher.Matcher]

		// Add template hooks that don't already exist
		for _, templateHook := range templateMatcher.Hooks {
			if !s.hookExists(existing, templateHook) {
				existing = append(existing, templateHook)
			}
		}
		matcherMap[templateMatcher.Matcher] = existing
	}

	// Convert back to slice format
	var result []models.HookMatcher
	for matcher, hooks := range matcherMap {
		if len(hooks) > 0 {
			result = append(result, models.HookMatcher{
				Matcher: matcher,
				Hooks:   hooks,
			})
		}
	}

	return result
}

// hookExists checks if a hook entry already exists in the list
func (s *Service) hookExists(hooks []models.HookEntry, target models.HookEntry) bool {
	for _, hook := range hooks {
		// Consider hooks equal if they have the same command (ignoring minor path differences)
		if s.normalizeHookCommand(hook.Command) == s.normalizeHookCommand(target.Command) {
			return true
		}
	}
	return false
}

// normalizeHookCommand normalizes hook commands for comparison
func (s *Service) normalizeHookCommand(command string) string {
	// Remove common variations and focus on the script name
	command = strings.TrimSpace(command)

	// Extract script name for strategic hooks
	if models.IsStrategicHook(command) {
		parts := strings.Split(command, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1] // Return just the script name
		}
	}

	return command
}

// updateStrategicHookPaths updates paths for strategic hooks to use the symlinked directory
func (s *Service) updateStrategicHookPaths(settings *models.ClaudeSettings) {
	if settings.Hooks == nil {
		return
	}

	s.updateHookTypePaths(settings.Hooks.PreToolUse)
	s.updateHookTypePaths(settings.Hooks.PostToolUse)
	s.updateHookTypePaths(settings.Hooks.Stop)
	s.updateHookTypePaths(settings.Hooks.PreCompact)
	s.updateHookTypePaths(settings.Hooks.Notification)
}

// updateHookTypePaths updates paths for a specific hook type
func (s *Service) updateHookTypePaths(matchers []models.HookMatcher) {
	for i := range matchers {
		for j := range matchers[i].Hooks {
			hook := &matchers[i].Hooks[j]
			if models.IsStrategicHook(hook.Command) {
				// Extract script name
				parts := strings.Split(hook.Command, "/")
				scriptName := parts[len(parts)-1]

				// Update to use symlinked strategic directory
				hook.Command = fmt.Sprintf("/usr/bin/python3 $CLAUDE_PROJECT_DIR/.claude/hooks/strategic/%s", scriptName)
			}
		}
	}
}

// writeSettings writes the merged settings to the settings file
func (s *Service) writeSettings(settingsPath string, settings *models.ClaudeSettings) error {
	// Pretty print JSON
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(settingsPath), config.DirPermissions); err != nil {
		return err
	}

	// Write settings file
	return os.WriteFile(settingsPath, data, config.FilePermissions)
}

// CleanSettings removes strategic hooks from settings.json while preserving user customizations
func (s *Service) CleanSettings(targetDir string) error {
	claudeDir := filepath.Join(targetDir, config.ClaudeDir)
	settingsPath := filepath.Join(claudeDir, config.ClaudeSettingsFile)

	// Check if settings file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		return nil // Nothing to clean
	}

	// Backup existing settings
	if err := s.backupExistingSettings(settingsPath); err != nil {
		return fmt.Errorf("failed to backup settings: %w", err)
	}

	// Load current settings
	currentSettings, err := s.loadExistingSettings(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to load settings: %w", err)
	}

	// Remove strategic hooks
	cleanedSettings := s.removeStrategicHooks(currentSettings)

	// If settings are now empty, remove the file
	if s.isEmptySettings(cleanedSettings) {
		return os.Remove(settingsPath)
	}

	// Write cleaned settings
	return s.writeSettings(settingsPath, cleanedSettings)
}

// removeStrategicHooks removes all strategic hooks from settings while preserving user content
func (s *Service) removeStrategicHooks(settings *models.ClaudeSettings) *models.ClaudeSettings {
	if settings == nil {
		return nil
	}

	result := &models.ClaudeSettings{
		Permissions: settings.Permissions, // Preserve all permissions
	}

	// Only process hooks if they exist
	if settings.Hooks != nil {
		result.Hooks = &models.HooksSection{
			PreToolUse:   s.filterNonStrategicHooks(settings.Hooks.PreToolUse),
			PostToolUse:  s.filterNonStrategicHooks(settings.Hooks.PostToolUse),
			Stop:         s.filterNonStrategicHooks(settings.Hooks.Stop),
			PreCompact:   s.filterNonStrategicHooks(settings.Hooks.PreCompact),
			Notification: s.filterNonStrategicHooks(settings.Hooks.Notification),
		}
	}

	return result
}

// filterNonStrategicHooks returns only non-strategic hooks from a list of hook matchers
func (s *Service) filterNonStrategicHooks(matchers []models.HookMatcher) []models.HookMatcher {
	if matchers == nil {
		return nil
	}

	var result []models.HookMatcher

	for _, matcher := range matchers {
		var nonStrategicHooks []models.HookEntry

		for _, hook := range matcher.Hooks {
			if !models.IsStrategicHook(hook.Command) {
				nonStrategicHooks = append(nonStrategicHooks, hook)
			}
		}

		// Only include matchers that have at least one non-strategic hook
		if len(nonStrategicHooks) > 0 {
			result = append(result, models.HookMatcher{
				Matcher: matcher.Matcher,
				Hooks:   nonStrategicHooks,
			})
		}
	}

	return result
}

// isEmptySettings checks if settings only contains empty structures
func (s *Service) isEmptySettings(settings *models.ClaudeSettings) bool {
	if settings == nil {
		return true
	}

	// Check if permissions exist
	if settings.Permissions != nil &&
		(len(settings.Permissions.Allow) > 0 ||
			len(settings.Permissions.AdditionalDirectories) > 0) {
		return false
	}

	// Check if any hooks exist
	if settings.Hooks != nil {
		if len(settings.Hooks.PreToolUse) > 0 ||
			len(settings.Hooks.PostToolUse) > 0 ||
			len(settings.Hooks.Stop) > 0 ||
			len(settings.Hooks.PreCompact) > 0 ||
			len(settings.Hooks.Notification) > 0 {
			return false
		}
	}

	return true
}
