package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose   bool
	targetDir string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "strategic-claude-basic-cli",
	Short: "CLI tool for managing Strategic Claude Basic framework installations",
	Long: `Strategic Claude Basic CLI is a command-line tool that simplifies the integration
of the Strategic Claude Basic framework into your development projects.

It provides commands to install, update, check status, and clean up the framework
installation while preserving your custom configurations and user content.`,
	Version: getVersion(),
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVarP(&targetDir, "target", "t", ".", "target directory for operations")

	// Custom completions for flags
	if err := rootCmd.RegisterFlagCompletionFunc("target", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{}, cobra.ShellCompDirectiveFilterDirs
	}); err != nil {
		// This should not happen in normal operation, but we handle it for completeness
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for --target flag: %v\n", err)
	}
}
