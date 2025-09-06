package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"strategic-claude-basic-cli/internal/services/cleaner"
	"strategic-claude-basic-cli/internal/services/status"
	"strategic-claude-basic-cli/internal/utils"
)

var (
	cleanForce bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [directory]",
	Short: "Remove Strategic Claude Basic framework installation",
	Long: `Remove Strategic Claude Basic framework files from the specified directory.

This command will:
- Remove the .strategic-claude-basic directory
- Clean up symlinks in .claude directory
- Remove framework-specific entries from CLAUDE.md
- Preserve user-created content and configurations

Safety features:
- Confirmation prompt (unless --force is used)
- Preserves user content in guides/ and templates/ directories
- Creates backup before removal (unless --no-backup was used during installation)

Examples:
  strategic-claude-basic-cli clean                  # Clean current directory
  strategic-claude-basic-cli clean ./my-project    # Clean specific directory
  strategic-claude-basic-cli clean --force         # Clean without confirmation`,
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
			fmt.Printf("Cleaning directory: %s\n", absTarget)
			fmt.Printf("Force: %v\n", cleanForce)
		}

		// Initialize services
		cleanerService := cleaner.New()
		statusService := status.NewService()
		interactionService := utils.NewInteractionService()

		// Check if there's anything to clean first
		statusInfo, err := statusService.CheckInstallation(absTarget)
		if err != nil {
			return fmt.Errorf("failed to check installation status: %w", err)
		}

		// Check if there's any Strategic Claude Basic content to clean
		hasValidSymlinks := false
		for _, symlink := range statusInfo.Symlinks {
			if symlink.Valid || symlink.Exists {
				hasValidSymlinks = true
				break
			}
		}

		hasStrategicContent := statusInfo.StrategicClaudeDir || // Has .strategic-claude-basic
			hasValidSymlinks || // Has valid or existing strategic symlinks
			statusInfo.IsInstalled // Fully installed

		if !hasStrategicContent {
			utils.DisplayWarning("No Strategic Claude Basic installation found")
			return nil
		}

		// Confirm cleanup operation unless --force is used
		if !cleanForce {
			confirmed, err := interactionService.ConfirmCleanup(absTarget)
			if err != nil {
				return fmt.Errorf("failed to get user confirmation: %w", err)
			}
			if !confirmed {
				fmt.Println("Cleanup cancelled by user")
				return nil
			}
		}

		// Perform cleanup
		result, err := cleanerService.RemoveInstallation(absTarget)
		if err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		// Display results
		displayCleanupResults(result, verbose)

		if !result.Success {
			return fmt.Errorf("cleanup completed with errors")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "force cleanup without confirmation")

	// Custom completion for directory argument
	cleanCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{}, cobra.ShellCompDirectiveFilterDirs
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
}

// displayCleanupResults shows the results of the cleanup operation
func displayCleanupResults(result *cleaner.CleanupResult, verbose bool) {
	fmt.Println()

	if result.Success {
		if result.RemovedDirectory {
			utils.DisplaySuccess("Removed .strategic-claude-basic directory")
		}

		if len(result.RemovedSymlinks) > 0 {
			utils.DisplaySuccess(fmt.Sprintf("Removed %d Strategic Claude symlink(s)", len(result.RemovedSymlinks)))
			if verbose {
				for _, symlink := range result.RemovedSymlinks {
					fmt.Printf("  • %s\n", symlink)
				}
			}
		}

		if len(result.CleanedDirectories) > 0 {
			utils.DisplaySuccess(fmt.Sprintf("Cleaned up %d empty director(ies)", len(result.CleanedDirectories)))
			if verbose {
				for _, dir := range result.CleanedDirectories {
					fmt.Printf("  • %s\n", dir)
				}
			}
		}

		if len(result.PreservedFiles) > 0 {
			utils.DisplayInfo(fmt.Sprintf("Preserved %d user file(s)", len(result.PreservedFiles)))
			if verbose {
				for _, file := range result.PreservedFiles {
					fmt.Printf("  • %s\n", file)
				}
			}
		}

		if len(result.RemovedSymlinks) == 0 && !result.RemovedDirectory && len(result.CleanedDirectories) == 0 {
			utils.DisplayInfo("No Strategic Claude Basic installation found to clean")
		} else {
			utils.DisplaySuccess("Strategic Claude Basic cleanup completed successfully")
		}
	} else {
		utils.DisplayError(fmt.Errorf("cleanup completed with errors"))
	}

	// Display warnings
	for _, warning := range result.Warnings {
		utils.DisplayWarning(warning)
	}

	// Display errors
	for _, err := range result.Errors {
		utils.DisplayError(fmt.Errorf("%s", err))
	}
}
