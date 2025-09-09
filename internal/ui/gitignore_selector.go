package ui

import (
	"fmt"
	"strconv"
	"strings"

	"strategic-claude-basic-cli/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
)

// Key constants for the interactive UI
const (
	keyCtrlC = "ctrl+c"
	keyEnter = "enter"
	keyDown  = "down"
	keyEsc   = "esc"
	keyQ     = "q"
)

// GitignoreModeOption represents a gitignore mode choice
type GitignoreModeOption struct {
	ID          string
	Name        string
	Description string
}

// GitignoreModeSelectorModel represents the state of the gitignore mode selector
type GitignoreModeSelectorModel struct {
	options  []GitignoreModeOption
	cursor   int
	selected string
	quitting bool
}

// getGitignoreModeOptions returns available gitignore mode options
func getGitignoreModeOptions() []GitignoreModeOption {
	return []GitignoreModeOption{
		{
			ID:          "track",
			Name:        "Track all files (default)",
			Description: "Don't add any .gitignore files - track all Strategic Claude Basic files",
		},
		{
			ID:          "all",
			Name:        "Ignore entire framework",
			Description: "Ignore all Strategic Claude Basic files and directories",
		},
		{
			ID:          "non-user",
			Name:        "Ignore framework, keep user content",
			Description: "Ignore framework directories (core, guides, templates) but track user content",
		},
	}
}

// NewGitignoreModeSelectorModel creates a new gitignore mode selector model
func NewGitignoreModeSelectorModel() GitignoreModeSelectorModel {
	options := getGitignoreModeOptions()

	// Set cursor to track mode (first option) by default
	cursor := 0
	for i, option := range options {
		if option.ID == "track" {
			cursor = i
			break
		}
	}

	return GitignoreModeSelectorModel{
		options: options,
		cursor:  cursor,
	}
}

// Init is called when the program starts
func (m GitignoreModeSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles input events and updates the model state
func (m GitignoreModeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case keyCtrlC, keyQ, keyEsc:
			m.quitting = true
			return m, tea.Quit
		case keyEnter, "tab":
			if len(m.options) > 0 {
				m.selected = m.options[m.cursor].ID
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case keyDown, "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the gitignore mode selector UI
func (m GitignoreModeSelectorModel) View() string {
	if m.quitting {
		if m.selected == "" {
			return quitTextStyle.Render("Selection cancelled.\n")
		}
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Select Gitignore Mode"))
	s.WriteString("\n\n")

	// Mode options list
	for i, option := range m.options {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Mode name
		line := fmt.Sprintf("%s %s", cursor, option.Name)

		if i == m.cursor {
			s.WriteString(selectedItemStyle.Render(line))
		} else {
			s.WriteString(itemStyle.Render(line))
		}
		s.WriteString("\n")

		// Description
		if option.Description != "" {
			var desc string
			if i == m.cursor {
				desc = selectedDescriptionStyle.Render(option.Description)
			} else {
				desc = descriptionStyle.Render(option.Description)
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

// GetSelectedMode returns the selected gitignore mode ID
func (m GitignoreModeSelectorModel) GetSelectedMode() string {
	return m.selected
}

// IsQuitting returns whether the user cancelled the selection
func (m GitignoreModeSelectorModel) IsQuitting() bool {
	return m.quitting && m.selected == ""
}

// fallbackSelectGitignoreMode provides a simple prompt-based selector when TTY isn't available
func fallbackSelectGitignoreMode(availableOptions []GitignoreModeOption) (string, error) {
	// Display mode options
	fmt.Println()
	fmt.Println("Available gitignore modes:")
	for i, option := range availableOptions {
		fmt.Printf("  %d. %s\n", i+1, option.Name)
		if option.Description != "" {
			fmt.Printf("     %s\n", option.Description)
		}
	}
	fmt.Println()

	// Get user selection
	interactionService := utils.NewInteractionService()
	for {
		input, err := interactionService.PromptWithDefault(fmt.Sprintf("Select gitignore mode (1-%d)", len(availableOptions)), "1")
		if err != nil {
			return "", fmt.Errorf("failed to get user input: %w", err)
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(availableOptions) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(availableOptions))
			continue
		}

		selectedOption := availableOptions[choice-1]
		fmt.Printf("Selected: %s\n", selectedOption.Name)
		return selectedOption.ID, nil
	}
}

// SelectGitignoreMode runs the interactive gitignore mode selector and returns the selected mode ID
func SelectGitignoreMode() (string, error) {
	availableOptions := getGitignoreModeOptions()
	if len(availableOptions) == 0 {
		return "", fmt.Errorf("no gitignore modes available")
	}

	// Check if we have a TTY for interactive mode
	if !isTTY() {
		// Fallback to simple prompts
		return fallbackSelectGitignoreMode(availableOptions)
	}

	// Run interactive Bubble Tea selector
	m := NewGitignoreModeSelectorModel()
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		// If Bubble Tea fails, fallback to simple prompts
		fmt.Printf("Interactive mode failed (%v), falling back to simple mode...\n", err)
		return fallbackSelectGitignoreMode(availableOptions)
	}

	model := finalModel.(GitignoreModeSelectorModel)
	if model.IsQuitting() {
		return "", fmt.Errorf("gitignore mode selection cancelled by user")
	}

	selectedID := model.GetSelectedMode()
	if selectedID == "" {
		return "", fmt.Errorf("no gitignore mode selected")
	}

	// Display confirmation
	var selectedOption GitignoreModeOption
	for _, option := range availableOptions {
		if option.ID == selectedID {
			selectedOption = option
			break
		}
	}

	fmt.Printf("\nSelected: %s\n", selectedOption.Name)
	return selectedID, nil
}
