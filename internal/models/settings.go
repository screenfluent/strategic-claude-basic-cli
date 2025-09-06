package models

// ClaudeSettings represents the structure of Claude Code settings.json
type ClaudeSettings struct {
	Hooks       *HooksSection       `json:"hooks,omitempty"`
	Permissions *PermissionsSection `json:"permissions,omitempty"`
}

// HooksSection contains all hook type configurations
type HooksSection struct {
	PreToolUse   []HookMatcher `json:"PreToolUse,omitempty"`
	PostToolUse  []HookMatcher `json:"PostToolUse,omitempty"`
	Stop         []HookMatcher `json:"Stop,omitempty"`
	PreCompact   []HookMatcher `json:"PreCompact,omitempty"`
	Notification []HookMatcher `json:"Notification,omitempty"`
}

// HookMatcher represents a matcher pattern with associated hooks
type HookMatcher struct {
	Matcher string      `json:"matcher"`
	Hooks   []HookEntry `json:"hooks"`
}

// HookEntry represents an individual hook configuration
type HookEntry struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// PermissionsSection contains Claude Code permissions
type PermissionsSection struct {
	Allow                 []string `json:"allow,omitempty"`
	AdditionalDirectories []string `json:"additionalDirectories,omitempty"`
}

// GetHookTypesInOrder returns hook types in the order they should be processed
func GetHookTypesInOrder() []string {
	return []string{
		"PreToolUse",
		"PostToolUse",
		"Stop",
		"PreCompact",
		"Notification",
	}
}

// IsStrategicHook checks if a hook command is one of our strategic hooks
func IsStrategicHook(command string) bool {
	strategicHooks := []string{
		"block-skip-hooks.py",
		"block-config-writes.py",
		"stop-session-notify.py",
		"precompact-notify.py",
		"notification-hook.py",
	}

	for _, hook := range strategicHooks {
		if contains(command, hook) {
			return true
		}
	}
	return false
}

// contains is a helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr)-1:len(s)-len(substr)] == "/" && s[len(s)-len(substr):] == substr
}
