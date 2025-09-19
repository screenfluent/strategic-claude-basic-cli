package main

import (
	"fmt"
	"runtime"

	"github.com/Fomo-Driven-Development/strategic-claude-basic-cli/internal/templates"

	"github.com/spf13/cobra"
)

var (
	version = "0.1.0"
	commit  = "dev"
	date    = "unknown"
)

func getVersion() string {
	return fmt.Sprintf("%s (%s)", version, commit)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information including version number, commit hash, build date, and Go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("strategic-claude-basic-cli version %s\n", version)
		fmt.Printf("Git commit: %s\n", commit)
		fmt.Printf("Build date: %s\n", date)
		fmt.Printf("Go version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		fmt.Printf("\nTemplate Registry:\n")
		templateList := templates.ListTemplates()
		for _, template := range templateList {
			fmt.Printf("  %-4s: %s (%s @ %s)\n",
				template.ID,
				template.Name,
				template.Commit[:7],
				template.Branch)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
