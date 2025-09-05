package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
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

		// TODO: Implement cleanup logic in Phase 5
		fmt.Printf("Clean command stub - target: %s\n", absTarget)
		fmt.Println("Cleanup logic will be implemented in Phase 5")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVarP(&cleanForce, "force", "f", false, "force cleanup without confirmation")
}
