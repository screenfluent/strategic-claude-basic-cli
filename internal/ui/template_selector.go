package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/templates"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the template selector
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	itemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	selectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Foreground(lipgloss.Color("#626262")).
				Italic(true)

	selectedDescriptionStyle = lipgloss.NewStyle().
					PaddingLeft(4).
					Foreground(lipgloss.Color("#AD58B4")).
					Italic(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	quitTextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))
)

// TemplateSelectorModel represents the state of the template selector
type TemplateSelectorModel struct {
	templates []templates.Template
	cursor    int
	selected  string
	quitting  bool
}

// NewTemplateSelectorModel creates a new template selector model
func NewTemplateSelectorModel() TemplateSelectorModel {
	templateList := templates.ListActiveTemplates()

	// Set cursor to main template by default
	cursor := 0
	for i, template := range templateList {
		if template.ID == "main" {
			cursor = i
			break
		}
	}

	return TemplateSelectorModel{
		templates: templateList,
		cursor:    cursor,
	}
}

// Init is called when the program starts
func (m TemplateSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles input events and updates the model state
func (m TemplateSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter", "tab":
			if len(m.templates) > 0 {
				m.selected = m.templates[m.cursor].ID
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.templates)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the template selector UI
func (m TemplateSelectorModel) View() string {
	if m.quitting {
		if m.selected == "" {
			return quitTextStyle.Render("Selection cancelled.\n")
		}
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Select Template"))
	s.WriteString("\n\n")

	// Template list
	for i, template := range m.templates {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Template name and ID
		line := fmt.Sprintf("%s %s (%s)", cursor, template.DisplayName(), template.ID)

		if i == m.cursor {
			s.WriteString(selectedItemStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")

		// Description
		if template.Description != "" {
			var desc string
			if i == m.cursor {
				desc = selectedDescriptionStyle.Render(template.Description)
			} else {
				desc = descriptionStyle.Render(template.Description)
			}
			s.WriteString(desc)
			s.WriteString("\n")
		}
		s.WriteString("\n")
	}

	// Help text
	s.WriteString(helpStyle.Render("↑/↓: navigate • enter: select • q: quit"))
	s.WriteString("\n")

	return s.String()
}

// GetSelectedTemplate returns the selected template ID
func (m TemplateSelectorModel) GetSelectedTemplate() string {
	return m.selected
}

// IsQuitting returns whether the user cancelled the selection
func (m TemplateSelectorModel) IsQuitting() bool {
	return m.quitting && m.selected == ""
}

// isTTY checks if we're running in an interactive terminal
func isTTY() bool {
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}

// fallbackSelectTemplate provides a simple prompt-based selector when TTY isn't available
func fallbackSelectTemplate(availableTemplates []templates.Template) (string, error) {
	// Display template options
	fmt.Println()
	fmt.Println("Available templates:")
	for i, template := range availableTemplates {
		fmt.Printf("  %d. %s (%s)\n", i+1, template.DisplayName(), template.ID)
		if template.Description != "" {
			fmt.Printf("     %s\n", template.Description)
		}
	}
	fmt.Println()

	// Get user selection
	interactionService := utils.NewInteractionService()
	for {
		input, err := interactionService.PromptWithDefault(fmt.Sprintf("Select template (1-%d)", len(availableTemplates)), "")
		if err != nil {
			return "", fmt.Errorf("failed to get user input: %w", err)
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(availableTemplates) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(availableTemplates))
			continue
		}

		selectedTemplate := availableTemplates[choice-1]
		fmt.Printf("Selected: %s (%s)\n", selectedTemplate.DisplayName(), selectedTemplate.ID)
		return selectedTemplate.ID, nil
	}
}

// SelectTemplate runs the interactive template selector and returns the selected template ID
func SelectTemplate() (string, error) {
	availableTemplates := templates.ListActiveTemplates()
	if len(availableTemplates) == 0 {
		return "", fmt.Errorf("no templates available")
	}

	// If only one template, use it automatically
	if len(availableTemplates) == 1 {
		template := availableTemplates[0]
		fmt.Printf("Using template: %s (%s)\n", template.DisplayName(), template.ID)
		return template.ID, nil
	}

	// Check if we have a TTY for interactive mode
	if !isTTY() {
		// Fallback to simple prompts
		return fallbackSelectTemplate(availableTemplates)
	}

	// Run interactive Bubble Tea selector
	m := NewTemplateSelectorModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		// If Bubble Tea fails, fallback to simple prompts
		fmt.Printf("Interactive mode failed (%v), falling back to simple mode...\n", err)
		return fallbackSelectTemplate(availableTemplates)
	}

	model := finalModel.(TemplateSelectorModel)
	if model.IsQuitting() {
		return "", fmt.Errorf("template selection cancelled by user")
	}

	selectedID := model.GetSelectedTemplate()
	if selectedID == "" {
		return "", fmt.Errorf("no template selected")
	}

	// Display confirmation
	selectedTemplate, err := templates.GetTemplate(selectedID)
	if err != nil {
		return "", fmt.Errorf("failed to get selected template: %w", err)
	}

	fmt.Printf("\nSelected: %s (%s)\n", selectedTemplate.DisplayName(), selectedTemplate.ID)
	return selectedID, nil
}
