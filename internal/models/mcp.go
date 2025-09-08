package models

// MCPConfig represents the structure of an .mcp.json file
type MCPConfig struct {
	MCPServers map[string]MCPServer `json:"mcpServers"`
}

// MCPServer represents a single MCP server configuration
type MCPServer struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// MCPTemplate represents an available MCP template for installation
type MCPTemplate struct {
	Name     string    // Display name (extracted from filename)
	FileName string    // Full filename (e.g., "context7.mcp.json")
	Server   MCPServer // Server configuration from the template file
}

// MCPInstallationPlan represents what will be installed
type MCPInstallationPlan struct {
	TargetDir       string        // Directory where .mcp.json will be created/updated
	SelectedMCPs    []MCPTemplate // MCPs selected for installation
	HasExistingMCP  bool          // Whether .mcp.json already exists
	ExistingMCPPath string        // Path to existing .mcp.json
	BackupPath      string        // Path where backup will be created
	TemplatesDir    string        // Path to MCP templates directory
}

// Validate validates an MCP installation plan
func (p *MCPInstallationPlan) Validate() error {
	if p.TargetDir == "" {
		return NewAppError(ErrorCodeValidationFailed, "target directory is required", nil)
	}

	if len(p.SelectedMCPs) == 0 {
		return NewAppError(ErrorCodeValidationFailed, "at least one MCP must be selected", nil)
	}

	if p.TemplatesDir == "" {
		return NewAppError(ErrorCodeValidationFailed, "templates directory is required", nil)
	}

	return nil
}
