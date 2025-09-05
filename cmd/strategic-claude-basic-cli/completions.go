package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionsCmd = &cobra.Command{
	Use:   "completions [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for strategic-claude-basic-cli.

The completion script must be sourced to take effect.

Bash:
  # Add to ~/.bashrc or ~/.bash_profile
  source <(strategic-claude-basic-cli completions bash)

  # Or save to a file and source it
  strategic-claude-basic-cli completions bash > ~/.strategic-claude-basic-cli-completion
  echo "source ~/.strategic-claude-basic-cli-completion" >> ~/.bashrc

Zsh:
  # If shell completion is not already enabled, run:
  echo "autoload -U compinit; compinit" >> ~/.zshrc

  # Add completion
  strategic-claude-basic-cli completions zsh > "${fpath[1]}/_strategic-claude-basic-cli"

Fish:
  strategic-claude-basic-cli completions fish | source

  # Or save to fish completion directory
  strategic-claude-basic-cli completions fish > ~/.config/fish/completions/strategic-claude-basic-cli.fish

PowerShell:
  strategic-claude-basic-cli completions powershell | Out-String | Invoke-Expression

  # Or add to PowerShell profile
  strategic-claude-basic-cli completions powershell >> $PROFILE`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionsCmd)
}
