package templates

import (
	"testing"
)

func TestTemplate_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		template Template
		wantErr  bool
	}{
		{
			name: "valid template",
			template: Template{
				ID:      "test",
				Name:    "Test Template",
				RepoURL: "https://example.com/repo.git",
				Branch:  "main",
				Commit:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: false,
		},
		{
			name: "empty ID",
			template: Template{
				Name:    "Test Template",
				RepoURL: "https://example.com/repo.git",
				Branch:  "main",
				Commit:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: true,
		},
		{
			name: "empty name",
			template: Template{
				ID:      "test",
				RepoURL: "https://example.com/repo.git",
				Branch:  "main",
				Commit:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: true,
		},
		{
			name: "empty repo URL",
			template: Template{
				ID:     "test",
				Name:   "Test Template",
				Branch: "main",
				Commit: "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: true,
		},
		{
			name: "empty branch",
			template: Template{
				ID:      "test",
				Name:    "Test Template",
				RepoURL: "https://example.com/repo.git",
				Commit:  "1234567890abcdef1234567890abcdef12345678",
			},
			wantErr: true,
		},
		{
			name: "empty commit",
			template: Template{
				ID:      "test",
				Name:    "Test Template",
				RepoURL: "https://example.com/repo.git",
				Branch:  "main",
			},
			wantErr: true,
		},
		{
			name: "invalid commit hash",
			template: Template{
				ID:      "test",
				Name:    "Test Template",
				RepoURL: "https://example.com/repo.git",
				Branch:  "main",
				Commit:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.template.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("Template.IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplate_HasTag(t *testing.T) {
	template := Template{
		Tags: []string{"web", "api", "golang"},
	}

	tests := []struct {
		tag  string
		want bool
	}{
		{"web", true},
		{"API", true}, // Case insensitive
		{"golang", true},
		{"python", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			if got := template.HasTag(tt.tag); got != tt.want {
				t.Errorf("Template.HasTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate_DisplayName(t *testing.T) {
	tests := []struct {
		name       string
		template   Template
		wantResult string
	}{
		{
			name: "normal template",
			template: Template{
				Name:       "Test Template",
				Deprecated: false,
			},
			wantResult: "Test Template",
		},
		{
			name: "deprecated template",
			template: Template{
				Name:       "Old Template",
				Deprecated: true,
			},
			wantResult: "Old Template (deprecated)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.template.DisplayName(); got != tt.wantResult {
				t.Errorf("Template.DisplayName() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestTemplate_ShortDescription(t *testing.T) {
	template := Template{
		Description: "This is a very long description that should be truncated when requested",
	}

	tests := []struct {
		name      string
		maxLength int
		want      string
	}{
		{
			name:      "shorter than max",
			maxLength: 100,
			want:      "This is a very long description that should be truncated when requested",
		},
		{
			name:      "truncated",
			maxLength: 20,
			want:      "This is a very lo...",
		},
		{
			name:      "exact length",
			maxLength: 81,
			want:      "This is a very long description that should be truncated when requested",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := template.ShortDescription(tt.maxLength); got != tt.want {
				t.Errorf("Template.ShortDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}
