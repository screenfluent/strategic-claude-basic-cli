package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"strategic-claude-basic-cli/internal/models"
	"strategic-claude-basic-cli/internal/services/git"
	"strategic-claude-basic-cli/internal/services/installer"
	"strategic-claude-basic-cli/internal/templates"
	"strategic-claude-basic-cli/internal/utils"

	"github.com/spf13/cobra"
)

var (
	force      bool
	forceCore  bool
	yes        bool
	noBackup   bool
	dryRun     bool
	templateID string
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

Template selection:
- Use --template to specify a template ID directly
- Without --template, you'll be prompted to choose interactively

Examples:
  strategic-claude-basic-cli init                      # Install with template selection
  strategic-claude-basic-cli init --template=main     # Install main template
  strategic-claude-basic-cli init --template=ccr      # Install CCR template
  strategic-claude-basic-cli init ./my-project        # Install in specific directory
  strategic-claude-basic-cli init --force-core        # Update core files only
  strategic-claude-basic-cli init --dry-run           # Preview what would be done`,
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
	initCmd.Flags().StringVar(&templateID, "template", "", "template ID to install (main, ccr, etc.)")

	// Custom completion for directory argument
	initCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return []string{}, cobra.ShellCompDirectiveFilterDirs
		}
		return []string{}, cobra.ShellCompDirectiveNoFileComp
	}

	// Add completion for template flag
	if err := initCmd.RegisterFlagCompletionFunc("template", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templateIDs := templates.GetTemplateIDs()
		return templateIDs, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		// This should not happen in normal operation, but we handle it for completeness
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for --template flag: %v\n", err)
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
	utils.VerbosePrintf(verbose, "Flags - Force: %v, Force Core: %v, Yes: %v, No Backup: %v, Dry Run: %v, Template: %s\n",
		force, forceCore, yes, noBackup, dryRun, templateID)

	// Handle template selection
	selectedTemplateID, err := selectTemplate(templateID, yes)
	if err != nil {
		utils.DisplayError(err)
		return err
	}

	utils.VerbosePrintf(verbose, "Selected template: %s\n", selectedTemplateID)

	// Validate prerequisites
	if err := validatePrerequisites(); err != nil {
		utils.DisplayError(err)
		return err
	}

	// Create install configuration
	installConfig := models.InstallConfig{
		TargetDir:   absTarget,
		TemplateID:  selectedTemplateID,
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

// selectTemplate handles template selection based on flags and user input
func selectTemplate(templateFlag string, skipPrompt bool) (string, error) {
	// If template is specified via flag, validate and use it
	if templateFlag != "" {
		if err := templates.ValidateTemplateID(templateFlag); err != nil {
			return "", fmt.Errorf("invalid template ID '%s': %w", templateFlag, err)
		}
		return templateFlag, nil
	}

	// If skipping prompts, use default template
	if skipPrompt {
		return templates.DefaultTemplateID, nil
	}

	// Interactive template selection
	return selectTemplateInteractively()
}

// selectTemplateInteractively presents template options to the user for selection
func selectTemplateInteractively() (string, error) {
	availableTemplates := templates.ListActiveTemplates()
	if len(availableTemplates) == 0 {
		return "", fmt.Errorf("no templates available")
	}

	// If only one template, use it automatically
	if len(availableTemplates) == 1 {
		template := availableTemplates[0]
		fmt.Printf("Using template: %s (%s)\n", template.DisplayName(), template.ID)
		return template.ID, nil
	}

	// Display template options
	fmt.Println()
	fmt.Println("Available templates:")
	for i, template := range availableTemplates {
		fmt.Printf("  %d. %s (%s)\n", i+1, template.DisplayName(), template.ID)
		if template.Description != "" {
			fmt.Printf("     %s\n", template.Description)
		}
	}
	fmt.Println()

	// Get user selection
	interactionService := utils.NewInteractionService()
	for {
		input, err := interactionService.PromptWithDefault(fmt.Sprintf("Select template (1-%d)", len(availableTemplates)), "")
		if err != nil {
			return "", fmt.Errorf("failed to get user input: %w", err)
		}

		choice, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || choice < 1 || choice > len(availableTemplates) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(availableTemplates))
			continue
		}

		selectedTemplate := availableTemplates[choice-1]
		fmt.Printf("Selected: %s (%s)\n", selectedTemplate.DisplayName(), selectedTemplate.ID)
		return selectedTemplate.ID, nil
	}
}

// getInstallationConfirmation displays the installation plan and asks for user confirmation
func getInstallationConfirmation(plan *models.InstallationPlan) (bool, error) {
	fmt.Println() // Empty line for readability
	fmt.Printf("Target directory: %s\n", plan.TargetDir)
	fmt.Printf("Installation type: %s\n", plan.InstallationType)

	// Display template information
	template := plan.Template
	fmt.Printf("Template: %s (%s)\n", template.DisplayName(), template.ID)
	if template.Description != "" {
		fmt.Printf("Description: %s\n", template.Description)
	}
	fmt.Printf("Branch: %s\n", template.Branch)
	fmt.Printf("Commit: %s\n", template.Commit)
	fmt.Println()

	// Display what will happen
	if len(plan.WillCreate) > 0 {
		fmt.Println("Files/directories to be created:")
		for _, item := range plan.WillCreate {
			fmt.Printf("  + %s\n", item)
		}
		fmt.Println()
	}

	if len(plan.SymlinksToCreate) > 0 {
		fmt.Println("Symlinks to be created:")
		for _, symlink := range plan.SymlinksToCreate {
			fmt.Printf("  → %s\n", symlink)
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
	fmt.Printf("Use 'strategic-claude-basic-cli status -t %s' to check installation status.\n", plan.TargetDir)
}
