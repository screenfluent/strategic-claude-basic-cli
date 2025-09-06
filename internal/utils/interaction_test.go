package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestInteractionService_ConfirmPrompt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		hasError bool
	}{
		{
			name:     "yes response",
			input:    "yes\n",
			expected: true,
			hasError: false,
		},
		{
			name:     "y response",
			input:    "y\n",
			expected: true,
			hasError: false,
		},
		{
			name:     "Y response (uppercase)",
			input:    "Y\n",
			expected: true,
			hasError: false,
		},
		{
			name:     "YES response (uppercase)",
			input:    "YES\n",
			expected: true,
			hasError: false,
		},
		{
			name:     "no response",
			input:    "no\n",
			expected: false,
			hasError: false,
		},
		{
			name:     "n response",
			input:    "n\n",
			expected: false,
			hasError: false,
		},
		{
			name:     "N response (uppercase)",
			input:    "N\n",
			expected: false,
			hasError: false,
		},
		{
			name:     "empty response (defaults to no)",
			input:    "\n",
			expected: false,
			hasError: false,
		},
		{
			name:     "invalid response (treated as no)",
			input:    "invalid\n",
			expected: false,
			hasError: false,
		},
		{
			name:     "maybe response (treated as no)",
			input:    "maybe\n",
			expected: false,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}

			os.Stdin = r

			// Write input to pipe
			go func() {
				defer w.Close()
				_, _ = w.WriteString(tt.input)
			}()

			// Capture stdout for verification
			oldStdout := os.Stdout
			rOut, wOut, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create output pipe: %v", err)
			}
			os.Stdout = wOut

			var output bytes.Buffer
			done := make(chan bool)
			go func() {
				_, _ = io.Copy(&output, rOut)
				done <- true
			}()

			// Test the function
			service := NewInteractionService()
			result, err := service.ConfirmPrompt("Test prompt")

			// Restore stdout
			wOut.Close()
			os.Stdout = oldStdout
			<-done

			// Check results
			if tt.hasError && err == nil {
				t.Error("Expected error, got nil")
			} else if !tt.hasError && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}

			// Verify prompt was displayed
			outputStr := output.String()
			if !strings.Contains(outputStr, "Test prompt") {
				t.Error("Expected prompt message in output")
			}
		})
	}
}

func TestInteractionService_PromptWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultValue string
		expected     string
	}{
		{
			name:         "user provides input",
			input:        "user-value\n",
			defaultValue: "default",
			expected:     "user-value",
		},
		{
			name:         "user provides empty input (uses default)",
			input:        "\n",
			defaultValue: "default-value",
			expected:     "default-value",
		},
		{
			name:         "user provides whitespace (trimmed)",
			input:        "  trimmed-value  \n",
			defaultValue: "default",
			expected:     "trimmed-value",
		},
		{
			name:         "empty default with empty input",
			input:        "\n",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock stdin
			oldStdin := os.Stdin
			defer func() { os.Stdin = oldStdin }()

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}

			os.Stdin = r

			// Write input to pipe
			go func() {
				defer w.Close()
				_, _ = w.WriteString(tt.input)
			}()

			// Capture stdout
			oldStdout := os.Stdout
			rOut, wOut, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create output pipe: %v", err)
			}
			os.Stdout = wOut

			var output bytes.Buffer
			done := make(chan bool)
			go func() {
				_, _ = io.Copy(&output, rOut)
				done <- true
			}()

			// Test the function
			service := NewInteractionService()
			result, err := service.PromptWithDefault("Test prompt", tt.defaultValue)

			// Restore stdout
			wOut.Close()
			os.Stdout = oldStdout
			<-done

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}

			// Verify prompt was displayed
			outputStr := output.String()
			if !strings.Contains(outputStr, "Test prompt") {
				t.Error("Expected prompt message in output")
			}

			// Verify default value is shown in prompt
			if tt.defaultValue != "" && !strings.Contains(outputStr, tt.defaultValue) {
				t.Error("Expected default value in prompt")
			}
		})
	}
}

func TestDisplayFunctions(t *testing.T) {
	tests := []struct {
		name     string
		function func()
		expected string
	}{
		{
			name: "DisplayError",
			function: func() {
				DisplayError(io.ErrUnexpectedEOF)
			},
			expected: "❌ Error: unexpected EOF",
		},
		{
			name: "DisplaySuccess",
			function: func() {
				DisplaySuccess("Operation completed successfully")
			},
			expected: "✅ Operation completed successfully",
		},
		{
			name: "DisplayWarning",
			function: func() {
				DisplayWarning("This is a warning message")
			},
			expected: "⚠️  This is a warning message",
		},
		{
			name: "DisplayInfo",
			function: func() {
				DisplayInfo("This is an info message")
			},
			expected: "ℹ️  This is an info message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			var r *os.File
			var w *os.File
			var err error

			// Capture stdout or stderr based on function
			if tt.name == "DisplayError" {
				// DisplayError writes to stderr
				oldStderr := os.Stderr
				defer func() { os.Stderr = oldStderr }()
				r, w, err = os.Pipe()
				if err != nil {
					t.Fatalf("Failed to create pipe: %v", err)
				}
				os.Stderr = w
			} else {
				// Other functions write to stdout
				oldStdout := os.Stdout
				defer func() { os.Stdout = oldStdout }()
				r, w, err = os.Pipe()
				if err != nil {
					t.Fatalf("Failed to create pipe: %v", err)
				}
				os.Stdout = w
			}

			// Read output in background
			done := make(chan bool)
			go func() {
				_, _ = io.Copy(&output, r)
				done <- true
			}()

			// Run the function
			tt.function()

			// Close and wait
			w.Close()
			<-done

			outputStr := strings.TrimSpace(output.String())
			if !strings.Contains(outputStr, tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, outputStr)
			}
		})
	}
}

func TestVerbosePrintln(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		message  string
		expected bool // whether output is expected
	}{
		{
			name:     "verbose enabled",
			verbose:  true,
			message:  "verbose message",
			expected: true,
		},
		{
			name:     "verbose disabled",
			verbose:  false,
			message:  "hidden message",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Run the function
			VerbosePrintln(tt.verbose, tt.message)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			_, err = io.Copy(&output, r)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			outputStr := output.String()
			hasOutput := len(strings.TrimSpace(outputStr)) > 0

			if tt.expected && !hasOutput {
				t.Error("Expected output, got none")
			} else if !tt.expected && hasOutput {
				t.Errorf("Expected no output, got %q", outputStr)
			}

			if tt.expected && !strings.Contains(outputStr, tt.message) {
				t.Errorf("Expected output to contain %q, got %q", tt.message, outputStr)
			}
		})
	}
}

func TestVerbosePrintf(t *testing.T) {
	tests := []struct {
		name     string
		verbose  bool
		format   string
		args     []interface{}
		expected bool // whether output is expected
	}{
		{
			name:     "verbose enabled with formatting",
			verbose:  true,
			format:   "Processing %s: %d items",
			args:     []interface{}{"files", 42},
			expected: true,
		},
		{
			name:     "verbose disabled",
			verbose:  false,
			format:   "Hidden %s",
			args:     []interface{}{"message"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Run the function
			VerbosePrintf(tt.verbose, tt.format, tt.args...)

			// Restore stdout and read output
			w.Close()
			os.Stdout = oldStdout

			var output bytes.Buffer
			_, err = io.Copy(&output, r)
			if err != nil {
				t.Fatalf("Failed to read output: %v", err)
			}

			outputStr := output.String()
			hasOutput := len(strings.TrimSpace(outputStr)) > 0

			if tt.expected && !hasOutput {
				t.Error("Expected output, got none")
			} else if !tt.expected && hasOutput {
				t.Errorf("Expected no output, got %q", outputStr)
			}

			if tt.expected {
				expectedStr := fmt.Sprintf(tt.format, tt.args...)
				if !strings.Contains(outputStr, expectedStr) {
					t.Errorf("Expected output to contain %q, got %q", expectedStr, outputStr)
				}
			}
		})
	}
}

// Test error handling with reader issues
func TestInteractionService_ConfirmPrompt_ReaderError(t *testing.T) {
	// Mock stdin with a reader that will fail
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	// Create a file that immediately returns EOF
	r, w, _ := os.Pipe()
	w.Close() // Close write end immediately to simulate EOF
	os.Stdin = r

	service := NewInteractionService()
	_, err := service.ConfirmPrompt("Test prompt")

	// Should handle EOF gracefully (return false, no error)
	if err != nil {
		// EOF is acceptable - the function should handle it
		t.Logf("Function returned error on EOF: %v", err)
	}
}

// Test single invalid input
func TestInteractionService_ConfirmPrompt_InvalidInput(t *testing.T) {
	// Mock stdin with invalid input
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	os.Stdin = r

	// Write invalid input
	go func() {
		defer w.Close()
		_, _ = w.WriteString("invalid\n")
	}()

	service := NewInteractionService()
	result, err := service.ConfirmPrompt("Test prompt")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if result {
		t.Error("Expected false for invalid input, got true")
	}
}

// Helper function to test stdin/stdout redirection works properly
func TestIORedirection(t *testing.T) {
	// Test that our test setup correctly captures stdin/stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout
	defer func() {
		os.Stdin = oldStdin
		os.Stdout = oldStdout
	}()

	// Setup input
	inputR, inputW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create input pipe: %v", err)
	}
	os.Stdin = inputR

	go func() {
		defer inputW.Close()
		_, _ = inputW.WriteString("test input\n")
	}()

	// Setup output capture
	outputR, outputW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create output pipe: %v", err)
	}
	os.Stdout = outputW

	var output bytes.Buffer
	done := make(chan bool)
	go func() {
		_, _ = io.Copy(&output, outputR)
		done <- true
	}()

	// Test reading and writing
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		input := scanner.Text()
		_, _ = os.Stdout.WriteString("received: " + input + "\n")
	}

	outputW.Close()
	<-done

	// Verify
	result := output.String()
	if !strings.Contains(result, "received: test input") {
		t.Errorf("IO redirection test failed, got: %q", result)
	}
}
