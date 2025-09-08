package main

import (
	"fmt"
	"path/filepath"

	"strategic-claude-basic-cli/internal/config"
	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/mcp"
	"strategic-claude-basic-cli/internal/ui"
	"strategic-claude-basic-cli/internal/utils"

	"github.com/spf13/cobra"
)

var installMcpCmd = &cobra.Command{
	Use:   "install-mcp",
	Short: "Install MCP servers from available templates",
	Long: `Install MCP (Model Context Protocol) servers from available templates.

This command will:
- Scan for available MCP server templates
- Present an interactive selection interface
- Merge selected servers into your project's .mcp.json file
- Backup existing .mcp.json if it exists

Prerequisites:
- Must be run after 'init' command
- Requires .strategic-claude-basic directory with MCP templates

Examples:
  strategic-claude-basic-cli install-mcp                    # Interactive selection
  strategic-claude-basic-cli install-mcp --target ./myapp  # Install in specific directory`,
	RunE: runInstallMCP,
}

func init() {
	rootCmd.AddCommand(installMcpCmd)
}

func runInstallMCP(cmd *cobra.Command, args []string) error {
	// Get target directory
	absTargetDir, err := filepath.Abs(targetDir)
	if err != nil {
		utils.DisplayError(fmt.Errorf("failed to resolve target directory: %w", err))
		return err
	}

	utils.VerbosePrintf(verbose, "Target directory: %s\n", absTargetDir)

	// Create MCP service
	mcpService := mcp.New()

	// Step 1: Scan for available MCP templates
	strategicDir := filepath.Join(absTargetDir, config.StrategicClaudeBasicDir)
	utils.VerbosePrintln(verbose, "Scanning for available MCP templates...")

	availableMCPs, err := mcpService.ScanAvailableMCPs(strategicDir)
	if err != nil {
		utils.DisplayError(err)
		return err
	}

	// Check if there are any MCPs available
	if len(availableMCPs) == 0 {
		utils.DisplayInfo("No MCP servers available for installation.")
		utils.DisplayInfo("Available MCP templates are stored in .strategic-claude-basic/templates/mcps/")
		return nil
	}

	utils.VerbosePrintf(verbose, "Found %d available MCP server(s)\n", len(availableMCPs))

	// Step 2: Interactive MCP selection
	utils.VerbosePrintln(verbose, "Starting interactive MCP selection...")

	selectedMCPs, err := ui.SelectMCPs(availableMCPs)
	if err != nil {
		utils.DisplayError(err)
		return err
	}

	utils.VerbosePrintf(verbose, "Selected %d MCP server(s)\n", len(selectedMCPs))

	// Step 3: Analyze installation plan
	utils.VerbosePrintln(verbose, "Analyzing installation plan...")

	plan, err := mcpService.AnalyzeInstallation(absTargetDir, selectedMCPs)
	if err != nil {
		utils.DisplayError(fmt.Errorf("failed to analyze installation: %w", err))
		return err
	}

	// Step 4: Display installation plan and get confirmation
	confirmed, err := getMCPInstallationConfirmation(plan)
	if err != nil {
		utils.DisplayError(fmt.Errorf("confirmation failed: %w", err))
		return err
	}
	if !confirmed {
		utils.DisplayInfo("MCP installation cancelled by user")
		return nil
	}

	// Step 5: Perform MCP installation
	utils.DisplayInfo("Installing selected MCP servers...")

	if err := mcpService.InstallMCPServers(plan); err != nil {
		utils.DisplayError(fmt.Errorf("MCP installation failed: %w", err))
		return err
	}

	// Step 6: Display success message
	utils.DisplaySuccess("MCP servers installed successfully!")
	displayMCPPostInstallInfo(plan)

	return nil
}

// getMCPInstallationConfirmation displays the installation plan and asks for user confirmation
func getMCPInstallationConfirmation(plan *models.MCPInstallationPlan) (bool, error) {
	fmt.Println() // Empty line for readability
	fmt.Printf("Target directory: %s\n", plan.TargetDir)

	if plan.HasExistingMCP {
		fmt.Printf("Existing .mcp.json: %s (will be backed up)\n", plan.ExistingMCPPath)
		fmt.Printf("Backup location: %s\n", plan.BackupPath)
	} else {
		fmt.Println("No existing .mcp.json found - new file will be created")
	}

	fmt.Println()
	fmt.Printf("MCP servers to install (%d):\n", len(plan.SelectedMCPs))
	for _, mcp := range plan.SelectedMCPs {
		fmt.Printf("  • %s (%s %v)\n", mcp.Name, mcp.Server.Command, mcp.Server.Args)
	}
	fmt.Println()

	// Warning about MCP installation
	utils.DisplayWarning("This will modify your project's .mcp.json configuration file.")
	if plan.HasExistingMCP {
		utils.DisplayWarning("Existing configuration will be backed up before modification.")
	}
	fmt.Println()

	// Ask for confirmation
	interactionService := utils.NewInteractionService()
	return interactionService.ConfirmPrompt("Do you want to proceed with MCP server installation?")
}

// displayMCPPostInstallInfo shows information after successful installation
func displayMCPPostInstallInfo(plan *models.MCPInstallationPlan) {
	fmt.Println()
	fmt.Println("MCP Installation Complete!")
	fmt.Printf("• Configuration file: %s\n", plan.ExistingMCPPath)
	fmt.Printf("• Installed %d MCP server(s)\n", len(plan.SelectedMCPs))

	if plan.HasExistingMCP {
		fmt.Printf("• Backup created: %s\n", plan.BackupPath)
	}

	fmt.Println()
	utils.DisplayInfo("MCP servers are now configured and ready to use with Claude Code.")
	fmt.Println("The servers will be available the next time you start Claude Code in this project.")
}
