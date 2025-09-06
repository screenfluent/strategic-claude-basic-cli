package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCompletionsCommand(t *testing.T) {
	tests := []struct {
		name         string
		shell        string
		expectError  bool
		expectOutput string
	}{
		{
			name:         "bash completions",
			shell:        "bash",
			expectError:  false,
			expectOutput: "# bash completion for strategic-claude-basic-cli",
		},
		{
			name:         "zsh completions",
			shell:        "zsh",
			expectError:  false,
			expectOutput: "#compdef strategic-claude-basic-cli",
		},
		{
			name:         "fish completions",
			shell:        "fish",
			expectError:  false,
			expectOutput: "# fish completion for strategic-claude-basic-cli",
		},
		{
			name:         "powershell completions",
			shell:        "powershell",
			expectError:  false,
			expectOutput: "Register-ArgumentCompleter",
		},
		{
			name:        "unsupported shell",
			shell:       "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create a new root command for testing
			testRootCmd := &cobra.Command{
				Use:   "strategic-claude-basic-cli",
				Short: "Test CLI",
			}

			// Create completions command
			testCompletionsCmd := &cobra.Command{
				Use:                   "completions [bash|zsh|fish|powershell]",
				Short:                 "Generate shell completion scripts",
				DisableFlagsInUseLine: true,
				ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
				Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
				RunE: func(cmd *cobra.Command, args []string) error {
					switch args[0] {
					case "bash":
						return testRootCmd.GenBashCompletion(&buf)
					case "zsh":
						return testRootCmd.GenZshCompletion(&buf)
					case "fish":
						return testRootCmd.GenFishCompletion(&buf, true)
					case "powershell":
						return testRootCmd.GenPowerShellCompletionWithDesc(&buf)
					default:
						return fmt.Errorf("unsupported shell: %s", args[0])
					}
				},
			}

			testRootCmd.AddCommand(testCompletionsCmd)
			testRootCmd.SetOut(&buf)
			testRootCmd.SetErr(&buf)

			// Execute the command
			testRootCmd.SetArgs([]string{"completions", tt.shell})
			err := testRootCmd.Execute()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for shell %s, but got none", tt.shell)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for shell %s: %v", tt.shell, err)
				return
			}

			// Check output content
			output := buf.String()
			if !strings.Contains(output, tt.expectOutput) {
				t.Errorf("Expected output to contain '%s' for shell %s, but it didn't.\nOutput: %s",
					tt.expectOutput, tt.shell, output)
			}

			// Verify output is not empty
			if len(output) == 0 {
				t.Errorf("Expected non-empty output for shell %s", tt.shell)
			}
		})
	}
}

func TestCompletionsCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "no arguments",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "too many arguments",
			args:        []string{"bash", "zsh"},
			expectError: true,
		},
		{
			name:        "valid single argument",
			args:        []string{"bash"},
			expectError: false,
		},
		{
			name:        "invalid shell argument",
			args:        []string{"invalid-shell"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create a new root command for testing
			testRootCmd := &cobra.Command{
				Use:   "strategic-claude-basic-cli",
				Short: "Test CLI",
			}

			// Add the actual completions command
			testRootCmd.AddCommand(completionsCmd)
			testRootCmd.SetOut(&buf)
			testRootCmd.SetErr(&buf)

			// Execute the command with test args
			args := append([]string{"completions"}, tt.args...)
			testRootCmd.SetArgs(args)
			err := testRootCmd.Execute()

			// Check error expectation
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for args %v, but got none", tt.args)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for args %v: %v", tt.args, err)
				}
			}
		})
	}
}

func TestTargetFlagCompletion(t *testing.T) {
	// Test that the --target flag completion function exists and returns the correct directive
	targetCompFunc, exists := rootCmd.GetFlagCompletionFunc("target")
	if !exists {
		t.Error("Expected --target flag to have a completion function")
		return
	}

	if targetCompFunc == nil {
		t.Error("Expected --target flag completion function to be non-nil")
		return
	}

	// Test the completion function
	completions, directive := targetCompFunc(rootCmd, []string{}, "")

	// Should return empty completions but FilterDirs directive
	if len(completions) != 0 {
		t.Errorf("Expected empty completions, got %v", completions)
	}

	if directive != cobra.ShellCompDirectiveFilterDirs {
		t.Errorf("Expected ShellCompDirectiveFilterDirs (%d), got %d", cobra.ShellCompDirectiveFilterDirs, directive)
	}
}

func TestDirectoryArgumentCompletion(t *testing.T) {
	tests := []struct {
		name         string
		cmd          *cobra.Command
		args         []string
		expectDirs   bool
		expectNoComp bool
	}{
		{
			name:         "init command - no args",
			cmd:          initCmd,
			args:         []string{},
			expectDirs:   true,
			expectNoComp: false,
		},
		{
			name:         "init command - one arg",
			cmd:          initCmd,
			args:         []string{"/some/path"},
			expectDirs:   false,
			expectNoComp: true,
		},
		{
			name:         "clean command - no args",
			cmd:          cleanCmd,
			args:         []string{},
			expectDirs:   true,
			expectNoComp: false,
		},
		{
			name:         "clean command - one arg",
			cmd:          cleanCmd,
			args:         []string{"/some/path"},
			expectDirs:   false,
			expectNoComp: true,
		},
		{
			name:         "status command - no args",
			cmd:          statusCmd,
			args:         []string{},
			expectDirs:   true,
			expectNoComp: false,
		},
		{
			name:         "status command - one arg",
			cmd:          statusCmd,
			args:         []string{"/some/path"},
			expectDirs:   false,
			expectNoComp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd.ValidArgsFunction == nil {
				t.Errorf("Expected %s to have ValidArgsFunction", tt.cmd.Use)
				return
			}

			completions, directive := tt.cmd.ValidArgsFunction(tt.cmd, tt.args, "")

			// Should always return empty completions for directory completion
			if len(completions) != 0 {
				t.Errorf("Expected empty completions, got %v", completions)
			}

			if tt.expectDirs && directive != cobra.ShellCompDirectiveFilterDirs {
				t.Errorf("Expected ShellCompDirectiveFilterDirs (%d), got %d", cobra.ShellCompDirectiveFilterDirs, directive)
			}

			if tt.expectNoComp && directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("Expected ShellCompDirectiveNoFileComp (%d), got %d", cobra.ShellCompDirectiveNoFileComp, directive)
			}
		})
	}
}

func TestCompletionsCommandHelp(t *testing.T) {
	// Test that the help text contains installation instructions for all supported shells
	expectedShells := []string{"bash", "zsh", "fish", "powershell"}
	helpText := completionsCmd.Long

	for _, shell := range expectedShells {
		if !strings.Contains(strings.ToLower(helpText), shell) {
			t.Errorf("Expected help text to contain installation instructions for %s", shell)
		}
	}

	// Test that help contains basic usage patterns
	expectedPatterns := []string{
		"source <(",
		"completion",
		"profile",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(strings.ToLower(helpText), strings.ToLower(pattern)) {
			t.Errorf("Expected help text to contain pattern '%s'", pattern)
		}
	}
}
