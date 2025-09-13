package mcp

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/config"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
)

// Service handles MCP server installation operations
type Service struct{}

// New creates a new MCP service
func New() *Service {
	return &Service{}
}

// ScanAvailableMCPs scans for available MCP templates in the templates directory
func (s *Service) ScanAvailableMCPs(strategicDir string) ([]models.MCPTemplate, error) {
	templatesDir := filepath.Join(strategicDir, config.TemplatesDir, "mcps")

	// Check if templates directory exists
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		return nil, models.NewAppError(
			models.ErrorCodeNotInstalled,
			"MCP templates directory not found - run 'init' command first",
			err,
		)
	}

	var templates []models.MCPTemplate

	// Read all .mcp.json files in the templates directory
	err := filepath.WalkDir(templatesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-json files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".mcp.json") {
			return nil
		}

		// Skip .gitkeep and template.mcp.json
		if d.Name() == ".gitkeep" || d.Name() == "template.mcp.json" {
			return nil
		}

		// Extract name from filename (remove .mcp.json extension)
		name := strings.TrimSuffix(d.Name(), ".mcp.json")

		// Read and parse the MCP template file
		server, err := s.readMCPTemplate(path)
		if err != nil {
			return fmt.Errorf("failed to read MCP template %s: %w", d.Name(), err)
		}

		templates = append(templates, models.MCPTemplate{
			Name:     name,
			FileName: d.Name(),
			Server:   server,
		})

		return nil
	})

	if err != nil {
		return nil, models.NewAppError(
			models.ErrorCodeFileSystemError,
			"failed to scan MCP templates",
			err,
		)
	}

	return templates, nil
}

// readMCPTemplate reads and parses a single MCP template file
func (s *Service) readMCPTemplate(filePath string) (models.MCPServer, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return models.MCPServer{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the template - it should contain a single server config
	var templateConfig map[string]models.MCPServer
	if err := json.Unmarshal(data, &templateConfig); err != nil {
		return models.MCPServer{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Template should have exactly one server config
	if len(templateConfig) != 1 {
		return models.MCPServer{}, fmt.Errorf("template should contain exactly one MCP server")
	}

	// Return the first (and only) server config
	for _, server := range templateConfig {
		return server, nil
	}

	return models.MCPServer{}, fmt.Errorf("no server configuration found")
}

// AnalyzeInstallation analyzes what will be done during MCP installation
func (s *Service) AnalyzeInstallation(targetDir string, selectedMCPs []models.MCPTemplate) (*models.MCPInstallationPlan, error) {
	strategicDir := filepath.Join(targetDir, config.StrategicClaudeBasicDir)
	templatesDir := filepath.Join(strategicDir, config.TemplatesDir, "mcps")
	mcpPath := filepath.Join(targetDir, ".mcp.json")

	// Check if .strategic-claude-basic directory exists
	if _, err := os.Stat(strategicDir); os.IsNotExist(err) {
		return nil, models.NewAppError(
			models.ErrorCodeNotInstalled,
			"Strategic Claude Basic not installed - run 'init' command first",
			err,
		)
	}

	// Check if existing .mcp.json file exists
	hasExisting := false
	if _, err := os.Stat(mcpPath); err == nil {
		hasExisting = true
	}

	// Generate backup path
	timestamp := time.Now().Format("20060102-150405")
	backupPath := filepath.Join(targetDir, fmt.Sprintf(".mcp-backup-%s.json", timestamp))

	plan := &models.MCPInstallationPlan{
		TargetDir:       targetDir,
		SelectedMCPs:    selectedMCPs,
		HasExistingMCP:  hasExisting,
		ExistingMCPPath: mcpPath,
		BackupPath:      backupPath,
		TemplatesDir:    templatesDir,
	}

	return plan, plan.Validate()
}

// InstallMCPServers performs the actual installation of selected MCP servers
func (s *Service) InstallMCPServers(plan *models.MCPInstallationPlan) error {
	// Backup existing .mcp.json if it exists
	if plan.HasExistingMCP {
		if err := s.backupExistingMCP(plan.ExistingMCPPath, plan.BackupPath); err != nil {
			return fmt.Errorf("failed to backup existing .mcp.json: %w", err)
		}
	}

	// Load or create base MCP configuration
	mcpConfig, err := s.loadOrCreateBaseMCPConfig(plan)
	if err != nil {
		return fmt.Errorf("failed to load/create base MCP config: %w", err)
	}

	// Merge selected MCP servers
	for _, template := range plan.SelectedMCPs {
		mcpConfig.MCPServers[template.Name] = template.Server
	}

	// Write the merged configuration
	if err := s.writeMCPConfig(plan.ExistingMCPPath, mcpConfig); err != nil {
		return fmt.Errorf("failed to write MCP configuration: %w", err)
	}

	return nil
}

// backupExistingMCP creates a timestamped backup of the existing .mcp.json file
func (s *Service) backupExistingMCP(mcpPath, backupPath string) error {
	data, err := os.ReadFile(mcpPath)
	if err != nil {
		return fmt.Errorf("failed to read existing .mcp.json: %w", err)
	}

	if err := os.WriteFile(backupPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// loadOrCreateBaseMCPConfig loads existing .mcp.json or creates a new one from template
func (s *Service) loadOrCreateBaseMCPConfig(plan *models.MCPInstallationPlan) (*models.MCPConfig, error) {
	if plan.HasExistingMCP {
		// Load existing configuration
		return s.loadMCPConfig(plan.ExistingMCPPath)
	}

	// Create new configuration from template
	templatePath := filepath.Join(plan.TemplatesDir, "template.mcp.json")
	return s.loadMCPConfig(templatePath)
}

// loadMCPConfig loads an MCP configuration from file
func (s *Service) loadMCPConfig(filePath string) (*models.MCPConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read MCP config: %w", err)
	}

	var config models.MCPConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse MCP config: %w", err)
	}

	// Ensure MCPServers map is initialized
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]models.MCPServer)
	}

	return &config, nil
}

// writeMCPConfig writes an MCP configuration to file
func (s *Service) writeMCPConfig(filePath string, config *models.MCPConfig) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal MCP config: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write MCP config: %w", err)
	}

	return nil
}
