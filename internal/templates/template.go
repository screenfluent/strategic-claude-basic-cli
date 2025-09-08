package templates

import (
	"fmt"
	"strings"
)

// Template represents a Strategic Claude Basic template variant
type Template struct {
	// Unique identifier for the template
	ID string `json:"id"`

	// Display name for the template
	Name string `json:"name"`

	// Description of what this template is for
	Description string `json:"description"`

	// Repository URL (can be same repo with different branches)
	RepoURL string `json:"repo_url"`

	// Git branch to use
	Branch string `json:"branch"`

	// Specific commit hash to checkout (pinned for stability)
	Commit string `json:"commit"`

	// Optional metadata for filtering/categorization
	Language string   `json:"language,omitempty"` // e.g., "go", "python", "typescript"
	Tags     []string `json:"tags,omitempty"`     // e.g., ["web", "cli", "api"]

	// Whether this template is deprecated
	Deprecated bool `json:"deprecated,omitempty"`
}

// TemplateInfo represents metadata about an installed template
type TemplateInfo struct {
	// Template that was installed
	Template Template `json:"template"`

	// When it was installed
	InstalledAt string `json:"installed_at"`

	// Version or commit at time of installation
	InstalledCommit string `json:"installed_commit"`

	// Any additional installation metadata
	Metadata map[string]string `json:"metadata,omitempty"`
}

// IsValid checks if the template configuration is valid
func (t *Template) IsValid() error {
	if t.ID == "" {
		return fmt.Errorf("template ID cannot be empty")
	}

	if t.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if t.RepoURL == "" {
		return fmt.Errorf("template repository URL cannot be empty")
	}

	if t.Branch == "" {
		return fmt.Errorf("template branch cannot be empty")
	}

	if t.Commit == "" {
		return fmt.Errorf("template commit cannot be empty")
	}

	// Validate commit hash format (basic check)
	if len(t.Commit) != 40 || !isHexString(t.Commit) {
		return fmt.Errorf("template commit must be a valid 40-character hex string")
	}

	return nil
}

// DisplayName returns a formatted display name for UI
func (t *Template) DisplayName() string {
	if t.Deprecated {
		return fmt.Sprintf("%s (deprecated)", t.Name)
	}
	return t.Name
}

// ShortDescription returns a truncated description for compact display
func (t *Template) ShortDescription(maxLength int) string {
	if len(t.Description) <= maxLength {
		return t.Description
	}
	return t.Description[:maxLength-3] + "..."
}

// HasTag checks if the template has a specific tag
func (t *Template) HasTag(tag string) bool {
	for _, t := range t.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// isHexString checks if a string contains only hexadecimal characters
func isHexString(s string) bool {
	for _, c := range s {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
