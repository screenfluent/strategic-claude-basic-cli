package main

import (
	"fmt"
	"path/filepath"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/models"
	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/services/status"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status [directory]",
	Short: "Check Strategic Claude Basic installation status",
	Long: `Check the installation status of Strategic Claude Basic framework in the specified directory.

This command will:
- Check for .strategic-claude-basic directory
- Verify .claude directory structure
- Check symlink integrity
- Report any configuration issues
- Display detailed installation information

Examples:
  strategic-claude-basic-cli status                 # Check current directory
  strategic-claude-basic-cli status ./my-project   # Check specific directory
  strategic-claude-basic-cli status --verbose      # Show detailed information`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Determine target directory
		target := targetDir
		if len(args) > 0 {
			target = args[0]
		}

		// Convert to absolute path
		absTarget, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("failed to resolve target directory: %w", err)
		}

		if verbose {
			fmt.Printf("Checking directory: %s\n", absTarget)
		}

		// Create status service and check installation
		statusService := status.NewService()
		statusInfo, err := statusService.CheckInstallation(absTarget)
		if err != nil {
			return fmt.Errorf("failed to check installation status: %w", err)
		}

		// Display status information
		displayStatus(statusInfo, statusService, verbose)

		return nil
	},
}

// displayStatus formats and displays the installation status information
func displayStatus(statusInfo *models.StatusInfo, statusService *status.Service, verbose bool) {
	// Display main status summary
	summary := statusService.GetStatusSummary(statusInfo)
	if statusInfo.IsInstalled {
		if statusInfo.HasIssues() {
			fmt.Printf("⚠️  %s\n", summary)
		} else {
			fmt.Printf("✅ %s\n", summary)
		}
	} else {
		fmt.Printf("❌ %s\n", summary)
	}

	// Display directory information
	fmt.Printf("\nDirectories:\n")
	if statusInfo.StrategicClaudeDir {
		fmt.Printf("  ✅ Strategic Claude Basic: %s\n", statusInfo.StrategicClaudeDirPath)
	} else {
		fmt.Printf("  ❌ Strategic Claude Basic: %s (not found)\n", statusInfo.StrategicClaudeDirPath)
	}

	if statusInfo.ClaudeDir {
		fmt.Printf("  ✅ Claude Integration: %s\n", statusInfo.ClaudeDirPath)
	} else {
		fmt.Printf("  ❌ Claude Integration: %s (not found)\n", statusInfo.ClaudeDirPath)
	}

	// Display template information
	if statusInfo.InstalledTemplate != nil {
		fmt.Printf("\nTemplate Information:\n")
		template := statusInfo.InstalledTemplate.Template
		fmt.Printf("  Name: %s\n", template.DisplayName())
		fmt.Printf("  ID: %s\n", template.ID)
		fmt.Printf("  Description: %s\n", template.Description)
		fmt.Printf("  Branch: %s\n", template.Branch)
		fmt.Printf("  Commit: %s\n", template.Commit)
		if statusInfo.InstalledTemplate.InstalledAt != "" {
			fmt.Printf("  Installed At: %s\n", statusInfo.InstalledTemplate.InstalledAt)
		}
		if template.Language != "" {
			fmt.Printf("  Language: %s\n", template.Language)
		}
		if len(template.Tags) > 0 {
			fmt.Printf("  Tags: %v\n", template.Tags)
		}
	}

	// Display symlink information
	if len(statusInfo.Symlinks) > 0 {
		fmt.Printf("\nSymlinks:\n")
		for _, symlink := range statusInfo.Symlinks {
			switch {
			case symlink.Valid:
				fmt.Printf("  ✅ %s → %s\n", symlink.Name, symlink.Target)
			case symlink.Exists:
				fmt.Printf("  ⚠️  %s → %s (%s)\n", symlink.Name, symlink.Target, symlink.Error)
			default:
				fmt.Printf("  ❌ %s (not found)\n", symlink.Name)
			}
		}
	}

	// Display issues
	if statusInfo.HasIssues() {
		fmt.Printf("\nIssues Found:\n")
		for _, issue := range statusInfo.Issues {
			fmt.Printf("  - %s\n", issue)
		}
	}

	// Verbose information
	if verbose {
		fmt.Printf("\nDetailed Information:\n")
		fmt.Printf("  Target Directory: %s\n", statusInfo.TargetDir)
		fmt.Printf("  Valid Symlinks: %d/%d\n", statusInfo.ValidSymlinks(), len(statusInfo.Symlinks))

		if statusInfo.InstallationDate != nil {
			fmt.Printf("  Installation Date: %s\n", statusInfo.InstallationDate.Format("2006-01-02 15:04:05"))
		}

		if statusInfo.Version != "" {
			fmt.Printf("  Version: %s\n", statusInfo.Version)
		}

		if statusInfo.CommitHash != "" {
			fmt.Printf("  Commit Hash: %s\n", statusInfo.CommitHash)
		}
	}

	// Add recommendation for next steps
	if !statusInfo.IsInstalled {
		fmt.Printf("\nTo install Strategic Claude Basic, run:\n")
		fmt.Printf("  %s init\n", "strategic-claude-basic-cli")
	} else if statusInfo.HasIssues() {
		fmt.Printf("\nTo fix issues, you may need to:\n")
		fmt.Printf("  - Run '%s clean' to remove the installation\n", "strategic-claude-basic-cli")
		fmt.Printf("  - Then run '%s init' to reinstall\n", "strategic-claude-basic-cli")
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Custom completion for directory argument
	statusCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{}, cobra.ShellCompDirectiveFilterDirs
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
}
