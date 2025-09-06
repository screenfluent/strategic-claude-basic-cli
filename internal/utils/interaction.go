package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InteractionService provides utilities for user interaction
type InteractionService struct {
	scanner *bufio.Scanner
}

// NewInteractionService creates a new interaction service
func NewInteractionService() *InteractionService {
	return &InteractionService{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// ConfirmPrompt displays a confirmation prompt and returns the user's choice
func (i *InteractionService) ConfirmPrompt(message string) (bool, error) {
	fmt.Printf("%s (y/N): ", message)

	if !i.scanner.Scan() {
		if err := i.scanner.Err(); err != nil {
			return false, fmt.Errorf("failed to read input: %w", err)
		}
		// EOF - treat as "no"
		return false, nil
	}

	response := strings.TrimSpace(strings.ToLower(i.scanner.Text()))
	return response == "y" || response == "yes", nil
}

// PromptWithDefault prompts for input with a default value
func (i *InteractionService) PromptWithDefault(message, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", message, defaultValue)
	} else {
		fmt.Printf("%s: ", message)
	}

	if !i.scanner.Scan() {
		if err := i.scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		// EOF - return default
		return defaultValue, nil
	}

	response := strings.TrimSpace(i.scanner.Text())
	if response == "" {
		return defaultValue, nil
	}
	return response, nil
}

// DisplayError displays an error message in a formatted way
func DisplayError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
}

// DisplaySuccess displays a success message
func DisplaySuccess(message string) {
	fmt.Printf("✅ %s\n", message)
}

// DisplayWarning displays a warning message
func DisplayWarning(message string) {
	fmt.Printf("⚠️  %s\n", message)
}

// DisplayInfo displays an informational message
func DisplayInfo(message string) {
	fmt.Printf("ℹ️  %s\n", message)
}

// VerbosePrintln prints a message only if verbose mode is enabled
func VerbosePrintln(verbose bool, message string) {
	if verbose {
		fmt.Printf("🔍 %s\n", message)
	}
}

// VerbosePrintf prints a formatted message only if verbose mode is enabled
func VerbosePrintf(verbose bool, format string, args ...interface{}) {
	if verbose {
		fmt.Printf("🔍 "+format, args...)
	}
}

// ConfirmCleanup displays a cleanup confirmation prompt with directory information
func (i *InteractionService) ConfirmCleanup(targetDir string) (bool, error) {
	fmt.Printf("\n⚠️  This will remove Strategic Claude Basic from: %s\n", targetDir)
	fmt.Println("This action will:")
	fmt.Println("  • Remove the .strategic-claude-basic directory")
	fmt.Println("  • Remove Strategic Claude symlinks from .claude directory")
	fmt.Println("  • Preserve any user-created content in .claude")
	fmt.Println()

	return i.ConfirmPrompt("Are you sure you want to proceed?")
}
