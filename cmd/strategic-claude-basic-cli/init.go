package main

import (
	"fmt"
	"path/filepath"

	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/git"
	"strategic-claude-basic-cli/internal/services/installer"
	"strategic-claude-basic-cli/internal/utils"

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
		return runInit(args)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&force, "force", "f", false, "force installation, overwriting existing files")
	initCmd.Flags().BoolVar(&forceCore, "force-core", false, "update only core framework files, preserving user content")
	initCmd.Flags().BoolVarP(&yes, "yes", "y", false, "automatically answer yes to all prompts")
	initCmd.Flags().BoolVar(&noBackup, "no-backup", false, "skip creating backups of existing files")
	initCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")

	// Custom completion for directory argument
	initCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{}, cobra.ShellCompDirectiveFilterDirs
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}
}

// runInit executes the init command logic
func runInit(args []string) error {
	// Determine target directory
	target := targetDir
	if len(args) > 0 {
		target = args[0]
	}

	// Convert to absolute path
	absTarget, err := filepath.Abs(target)
	if err != nil {
		utils.DisplayError(fmt.Errorf("failed to resolve target directory: %w", err))
		return err
	}

	utils.VerbosePrintf(verbose, "Target directory: %s\n", absTarget)
	utils.VerbosePrintf(verbose, "Flags - Force: %v, Force Core: %v, Yes: %v, No Backup: %v, Dry Run: %v\n",
		force, forceCore, yes, noBackup, dryRun)

	// Validate prerequisites
	if err := validatePrerequisites(); err != nil {
		utils.DisplayError(err)
		return err
	}

	// Create install configuration
	installConfig := models.InstallConfig{
		TargetDir:   absTarget,
		Force:       force,
		ForceCore:   forceCore,
		SkipConfirm: yes,
		NoBackup:    noBackup,
		Verbose:     verbose,
	}

	// Validate install configuration
	if err := installConfig.Validate(); err != nil {
		utils.DisplayError(err)
		return err
	}

	// Create installer service
	installerService := installer.New()

	// Step 1: Analyze installation requirements
	utils.VerbosePrintln(verbose, "Analyzing installation requirements...")
	plan, err := installerService.AnalyzeInstallation(installConfig)
	if err != nil {
		utils.DisplayError(fmt.Errorf("installation analysis failed: %w", err))
		return err
	}

	// Step 2: Display installation plan and get confirmation
	if dryRun {
		return displayDryRun(plan)
	}

	if !installConfig.SkipConfirm {
		confirmed, err := getInstallationConfirmation(plan)
		if err != nil {
			utils.DisplayError(fmt.Errorf("confirmation failed: %w", err))
			return err
		}
		if !confirmed {
			utils.DisplayInfo("Installation cancelled by user")
			return nil
		}
	}

	// Step 3: Perform installation
	utils.DisplayInfo(fmt.Sprintf("Installing Strategic Claude Basic in %s...", plan.TargetDir))

	if err := installerService.Install(installConfig); err != nil {
		utils.DisplayError(fmt.Errorf("installation failed: %w", err))
		return err
	}

	// Step 4: Display success message
	utils.DisplaySuccess("Strategic Claude Basic installation completed successfully!")
	displayPostInstallInfo(plan)

	return nil
}

// validatePrerequisites checks that all required tools are available
func validatePrerequisites() error {
	utils.VerbosePrintln(verbose, "Validating prerequisites...")

	// Check if git is installed
	gitService := git.New()
	if err := gitService.ValidateGitInstalled(); err != nil {
		return fmt.Errorf("git validation failed: %w", err)
	}

	return nil
}

// getInstallationConfirmation displays the installation plan and asks for user confirmation
func getInstallationConfirmation(plan *models.InstallationPlan) (bool, error) {
	fmt.Println() // Empty line for readability
	fmt.Printf("Target directory: %s\n", plan.TargetDir)
	fmt.Printf("Installation type: %s\n", plan.InstallationType)
	fmt.Println()

	// Display what will happen
	if len(plan.WillCreate) > 0 {
		fmt.Println("Files/directories to be created:")
		for _, item := range plan.WillCreate {
			fmt.Printf("  + %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.WillReplace) > 0 {
		fmt.Println("Files/directories to be replaced:")
		for _, item := range plan.WillReplace {
			fmt.Printf("  ~ %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.WillPreserve) > 0 {
		fmt.Println("User content to be preserved:")
		for _, item := range plan.WillPreserve {
			fmt.Printf("  ✓ %s\n", item)
		}
		fmt.Println()
	}

	if plan.BackupRequired {
		fmt.Printf("Backup will be created at: %s\n", plan.BackupDir)
		fmt.Println()
	}

	if len(plan.Warnings) > 0 {
		fmt.Println("⚠️  Warnings:")
		for _, warning := range plan.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()
	}

	// Ask for confirmation
	interactionService := utils.NewInteractionService()
	return interactionService.ConfirmPrompt("This will install Strategic Claude Basic in the above directory.\nAre you sure you want to proceed?")
}

// displayDryRun shows what would happen without making changes
func displayDryRun(plan *models.InstallationPlan) error {
	fmt.Println("=== DRY RUN MODE ===")
	fmt.Println("This shows what would happen without making any changes.")
	fmt.Println()

	fmt.Printf("Target directory: %s\n", plan.TargetDir)
	fmt.Printf("Installation type: %s\n", plan.InstallationType)
	fmt.Println()

	if len(plan.WillCreate) > 0 {
		fmt.Println("Would create:")
		for _, item := range plan.WillCreate {
			fmt.Printf("  + %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.WillReplace) > 0 {
		fmt.Println("Would replace:")
		for _, item := range plan.WillReplace {
			fmt.Printf("  ~ %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.WillPreserve) > 0 {
		fmt.Println("Would preserve:")
		for _, item := range plan.WillPreserve {
			fmt.Printf("  ✓ %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.DirectoriesToCreate) > 0 {
		fmt.Println("Would create directories:")
		for _, dir := range plan.DirectoriesToCreate {
			fmt.Printf("  + %s/\n", dir)
		}
		fmt.Println()
	}

	if len(plan.SymlinksToCreate) > 0 {
		fmt.Println("Would create symlinks:")
		for _, symlink := range plan.SymlinksToCreate {
			fmt.Printf("  → %s\n", symlink)
		}
		fmt.Println()
	}

	if len(plan.SymlinksToUpdate) > 0 {
		fmt.Println("Would update symlinks:")
		for _, symlink := range plan.SymlinksToUpdate {
			fmt.Printf("  ↻ %s\n", symlink)
		}
		fmt.Println()
	}

	if plan.BackupRequired {
		fmt.Printf("Would create backup at: %s\n", plan.BackupDir)
		fmt.Println()
	}

	if len(plan.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range plan.Warnings {
			utils.DisplayWarning(warning)
		}
		fmt.Println()
	}

	if len(plan.Errors) > 0 {
		fmt.Println("Errors that would prevent installation:")
		for _, err := range plan.Errors {
			utils.DisplayError(fmt.Errorf("%s", err))
		}
		return fmt.Errorf("installation plan has errors")
	}

	fmt.Println("=== END DRY RUN ===")
	return nil
}

// displayPostInstallInfo shows helpful information after successful installation
func displayPostInstallInfo(plan *models.InstallationPlan) {
	fmt.Println()
	fmt.Println("🎉 Strategic Claude Basic has been installed!")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("1. Review the Strategic Claude Basic guides in .strategic-claude-basic/guides/")
	fmt.Println("2. Customize your project-specific configuration")
	fmt.Println("3. Start using Claude Code with the strategic agents, commands, and hooks")
	fmt.Println()
	fmt.Printf("Use 'strategic-claude-basic-cli status -t %s' to check installation status.\n", plan.TargetDir)
}
