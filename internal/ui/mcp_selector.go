package ui

import (
	"fmt"
	"strings"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styles for the MCP selector
var (
	mcpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	mcpItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	mcpSelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(2).
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true)

	mcpDescriptionStyle = lipgloss.NewStyle().
				PaddingLeft(4).
				Foreground(lipgloss.Color("#626262")).
				Italic(true)

	mcpSelectedDescriptionStyle = lipgloss.NewStyle().
					PaddingLeft(4).
					Foreground(lipgloss.Color("#AD58B4")).
					Italic(true)

	mcpHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)

	mcpQuitTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF5F87"))
)

// MCPSelectorModel represents the state of the MCP selector
type MCPSelectorModel struct {
	mcps      []models.MCPTemplate
	cursor    int
	selected  map[int]bool // Track which MCPs are selected
	confirmed bool         // Whether user confirmed selection
	quitting  bool
}

// NewMCPSelectorModel creates a new MCP selector model
func NewMCPSelectorModel(mcps []models.MCPTemplate) MCPSelectorModel {
	return MCPSelectorModel{
		mcps:     mcps,
		cursor:   0,
		selected: make(map[int]bool),
	}
}

// Init is called when the program starts
func (m MCPSelectorModel) Init() tea.Cmd {
	return nil
}

// Update handles input events and updates the model state
func (m MCPSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			m.confirmed = true
			return m, tea.Quit
		case " ":
			// Toggle selection
			m.selected[m.cursor] = !m.selected[m.cursor]
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.mcps)-1 {
				m.cursor++
			}
		}
	}

	return m, nil
}

// View renders the MCP selector UI
func (m MCPSelectorModel) View() string {
	if m.quitting {
		if !m.confirmed {
			return mcpQuitTextStyle.Render("MCP installation cancelled.\n")
		}
		return ""
	}

	var s strings.Builder

	// Title
	s.WriteString(mcpTitleStyle.Render("Select MCP Servers to Install"))
	s.WriteString("\n\n")

	// MCP list
	for i, mcp := range m.mcps {
		cursor := " "
		if i == m.cursor {
			cursor = ">"
		}

		// Checkbox
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[x]"
		}

		// MCP name
		line := fmt.Sprintf("%s %s %s", cursor, checkbox, mcp.Name)

		if i == m.cursor {
			s.WriteString(mcpSelectedItemStyle.Render(line))
		} else {
			s.WriteString(mcpItemStyle.Render(line))
		}
		s.WriteString("\n")

		// Command info as description
		cmdInfo := fmt.Sprintf("Command: %s %s", mcp.Server.Command, strings.Join(mcp.Server.Args, " "))
		var desc string
		if i == m.cursor {
			desc = mcpSelectedDescriptionStyle.Render(cmdInfo)
		} else {
			desc = mcpDescriptionStyle.Render(cmdInfo)
		}
		s.WriteString(desc)
		s.WriteString("\n\n")
	}

	// Help text
	s.WriteString(mcpHelpStyle.Render("↑/↓: navigate • space: toggle selection • enter: confirm • q: quit"))
	s.WriteString("\n")

	return s.String()
}

// GetSelectedMCPs returns the list of selected MCP templates
func (m MCPSelectorModel) GetSelectedMCPs() []models.MCPTemplate {
	var selectedMCPs []models.MCPTemplate
	for i, isSelected := range m.selected {
		if isSelected && i < len(m.mcps) {
			selectedMCPs = append(selectedMCPs, m.mcps[i])
		}
	}
	return selectedMCPs
}

// IsConfirmed returns whether the user confirmed their selection
func (m MCPSelectorModel) IsConfirmed() bool {
	return m.confirmed
}

// IsQuitting returns whether the user cancelled the selection
func (m MCPSelectorModel) IsQuitting() bool {
	return m.quitting && !m.confirmed
}

// SelectMCPs runs the interactive MCP selector and returns the selected MCPs
func SelectMCPs(availableMCPs []models.MCPTemplate) ([]models.MCPTemplate, error) {
	if len(availableMCPs) == 0 {
		return nil, fmt.Errorf("no MCP servers available for installation")
	}

	// Run interactive selector
	m := NewMCPSelectorModel(availableMCPs)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run MCP selector: %w", err)
	}

	model := finalModel.(MCPSelectorModel)
	if model.IsQuitting() {
		return nil, fmt.Errorf("MCP selection cancelled by user")
	}

	selectedMCPs := model.GetSelectedMCPs()
	if len(selectedMCPs) == 0 {
		return nil, fmt.Errorf("no MCP servers selected")
	}

	return selectedMCPs, nil
}
