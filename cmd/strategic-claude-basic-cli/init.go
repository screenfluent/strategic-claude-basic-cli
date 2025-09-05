package main

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	force     bool
	forceCore bool
	yes       bool
	noBackup  bool
	dryRun    bool
)

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Install Strategic Claude Basic framework",
	Long: `Install Strategic Claude Basic framework in the specified directory.

This command will:
- Clone the Strategic Claude Basic repository at a fixed commit
- Set up the framework directory structure
- Create or update CLAUDE.md configuration
- Create symlinks for framework integration
- Optionally backup existing files

Installation modes:
- New installation: Install in a clean directory
- Update core only (--force-core): Update only core framework files, preserve user content
- Full overwrite (--force): Replace all framework files

Examples:
  strategic-claude-basic-cli init                    # Install in current directory
  strategic-claude-basic-cli init ./my-project      # Install in specific directory
  strategic-claude-basic-cli init --force-core      # Update core files only
  strategic-claude-basic-cli init --dry-run         # Preview what would be done`,
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
			fmt.Printf("Target directory: %s\n", absTarget)
			fmt.Printf("Force: %v, Force Core: %v, Yes: %v, No Backup: %v, Dry Run: %v\n",
				force, forceCore, yes, noBackup, dryRun)
		}

		// TODO: Implement installation logic in Phase 4
		fmt.Printf("Init command stub - target: %s\n", absTarget)
		fmt.Println("Installation logic will be implemented in Phase 4")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "force installation, overwriting existing files")
	initCmd.Flags().BoolVar(&forceCore, "force-core", false, "update only core framework files, preserving user content")
	initCmd.Flags().BoolVarP(&yes, "yes", "y", false, "automatically answer yes to all prompts")
	initCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip creating backups of existing files")
	initCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
}
