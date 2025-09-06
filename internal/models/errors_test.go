package models

import (
	"errors"
	"fmt"
	"testing"
)

func TestNewAppError(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name    string
		code    ErrorCode
		message string
		cause   error
	}{
		{
			name:    "error with cause",
			code:    ErrorCodeFileSystemError,
			message: "test error message",
			cause:   originalErr,
		},
		{
			name:    "error without cause",
			code:    ErrorCodeValidationFailed,
			message: "validation failed",
			cause:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAppError(tt.code, tt.message, tt.cause)

			if err.Code != tt.code {
				t.Errorf("Expected code %s, got %s", tt.code, err.Code)
			}

			if err.Message != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, err.Message)
			}

			if err.Cause != tt.cause {
				t.Errorf("Expected cause %v, got %v", tt.cause, err.Cause)
			}

			if err.Context == nil {
				t.Error("Expected context to be initialized")
			}
		})
	}
}

func TestAppError_Error(t *testing.T) {
	originalErr := errors.New("original error")

	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error with cause",
			appError: &AppError{
				Code:    ErrorCodeFileSystemError,
				Message: "test error message",
				Cause:   originalErr,
			},
			expected: "FILE_SYSTEM_ERROR: test error message (caused by: original error)",
		},
		{
			name: "error without cause",
			appError: &AppError{
				Code:    ErrorCodeValidationFailed,
				Message: "validation failed",
				Cause:   nil,
			},
			expected: "VALIDATION_FAILED: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")

	appErr := &AppError{
		Code:    ErrorCodeFileSystemError,
		Message: "test message",
		Cause:   originalErr,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Expected unwrapped error %v, got %v", originalErr, unwrapped)
	}

	// Test with no cause
	appErrNoCause := &AppError{
		Code:    ErrorCodeValidationFailed,
		Message: "test message",
		Cause:   nil,
	}

	unwrappedNil := appErrNoCause.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("Expected nil unwrapped error, got %v", unwrappedNil)
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := &AppError{Code: ErrorCodeFileSystemError, Message: "error 1"}
	err2 := &AppError{Code: ErrorCodeFileSystemError, Message: "error 2"}
	err3 := &AppError{Code: ErrorCodeValidationFailed, Message: "error 3"}
	regularErr := errors.New("regular error")

	tests := []struct {
		name     string
		err      *AppError
		target   error
		expected bool
	}{
		{
			name:     "same error code",
			err:      err1,
			target:   err2,
			expected: true,
		},
		{
			name:     "different error code",
			err:      err1,
			target:   err3,
			expected: false,
		},
		{
			name:     "regular error",
			err:      err1,
			target:   regularErr,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Is(tt.target)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	appErr := &AppError{
		Code:    ErrorCodeFileSystemError,
		Message: "test message",
		Context: make(map[string]interface{}),
	}

	// Add context
	result := appErr.WithContext("key1", "value1")

	// Should return the same instance
	if result != appErr {
		t.Error("WithContext should return the same instance")
	}

	// Check context was added
	if appErr.Context["key1"] != "value1" {
		t.Errorf("Expected context value %q, got %v", "value1", appErr.Context["key1"])
	}

	// Add more context
	_ = appErr.WithContext("key2", 42)
	if appErr.Context["key2"] != 42 {
		t.Errorf("Expected context value %d, got %v", 42, appErr.Context["key2"])
	}

	// Test with nil context (should initialize)
	errWithNilContext := &AppError{
		Code:    ErrorCodeValidationFailed,
		Message: "test",
		Context: nil,
	}

	_ = errWithNilContext.WithContext("test", "value")
	if errWithNilContext.Context == nil {
		t.Error("Expected context to be initialized")
	}
	if errWithNilContext.Context["test"] != "value" {
		t.Error("Expected context value to be set")
	}
}

func TestNewGitError(t *testing.T) {
	originalErr := errors.New("git command failed")

	gitErr := NewGitError(ErrorCodeGitCloneError, "clone", originalErr)

	if gitErr.Code != ErrorCodeGitCloneError {
		t.Errorf("Expected code %s, got %s", ErrorCodeGitCloneError, gitErr.Code)
	}

	if gitErr.Message != "Git operation failed: clone" {
		t.Errorf("Expected specific message format, got %q", gitErr.Message)
	}

	if gitErr.Cause != originalErr {
		t.Errorf("Expected cause %v, got %v", originalErr, gitErr.Cause)
	}

	if gitErr.Context["operation"] != "clone" {
		t.Errorf("Expected operation context %q, got %v", "clone", gitErr.Context["operation"])
	}
}

func TestNewFileSystemError(t *testing.T) {
	originalErr := errors.New("permission denied")
	path := "/test/path"

	fsErr := NewFileSystemError(ErrorCodePermissionDenied, path, originalErr)

	if fsErr.Code != ErrorCodePermissionDenied {
		t.Errorf("Expected code %s, got %s", ErrorCodePermissionDenied, fsErr.Code)
	}

	expectedMessage := fmt.Sprintf("File system operation failed for path: %s", path)
	if fsErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, fsErr.Message)
	}

	if fsErr.Cause != originalErr {
		t.Errorf("Expected cause %v, got %v", originalErr, fsErr.Cause)
	}

	if fsErr.Context["path"] != path {
		t.Errorf("Expected path context %q, got %v", path, fsErr.Context["path"])
	}
}

func TestNewInstallationError(t *testing.T) {
	originalErr := errors.New("installation failed")
	targetDir := "/target/dir"

	installErr := NewInstallationError(ErrorCodeInstallationFailed, targetDir, originalErr)

	if installErr.Code != ErrorCodeInstallationFailed {
		t.Errorf("Expected code %s, got %s", ErrorCodeInstallationFailed, installErr.Code)
	}

	expectedMessage := fmt.Sprintf("Installation failed in directory: %s", targetDir)
	if installErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, installErr.Message)
	}

	if installErr.Cause != originalErr {
		t.Errorf("Expected cause %v, got %v", originalErr, installErr.Cause)
	}

	if installErr.Context["target_dir"] != targetDir {
		t.Errorf("Expected target_dir context %q, got %v", targetDir, installErr.Context["target_dir"])
	}
}

func TestNewValidationError(t *testing.T) {
	field := "test_field"
	value := "test_value"
	message := "validation failed"

	validationErr := NewValidationError(field, value, message)

	if validationErr.Code != ErrorCodeValidationFailed {
		t.Errorf("Expected code %s, got %s", ErrorCodeValidationFailed, validationErr.Code)
	}

	expectedMessage := fmt.Sprintf("Validation failed for %s: %s", field, message)
	if validationErr.Message != expectedMessage {
		t.Errorf("Expected message %q, got %q", expectedMessage, validationErr.Message)
	}

	if validationErr.Cause != nil {
		t.Errorf("Expected no cause, got %v", validationErr.Cause)
	}

	if validationErr.Context["field"] != field {
		t.Errorf("Expected field context %q, got %v", field, validationErr.Context["field"])
	}

	if validationErr.Context["value"] != value {
		t.Errorf("Expected value context %q, got %v", value, validationErr.Context["value"])
	}
}

func TestIsErrorCode(t *testing.T) {
	appErr := NewAppError(ErrorCodeFileSystemError, "test", nil)
	regularErr := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		code     ErrorCode
		expected bool
	}{
		{
			name:     "matching app error",
			err:      appErr,
			code:     ErrorCodeFileSystemError,
			expected: true,
		},
		{
			name:     "non-matching app error",
			err:      appErr,
			code:     ErrorCodeValidationFailed,
			expected: false,
		},
		{
			name:     "regular error",
			err:      regularErr,
			code:     ErrorCodeFileSystemError,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			code:     ErrorCodeFileSystemError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsErrorCode(tt.err, tt.code)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsGitError(t *testing.T) {
	regularErr := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "git clone error",
			err:      NewAppError(ErrorCodeGitCloneError, "test", nil),
			expected: true,
		},
		{
			name:     "git checkout error",
			err:      NewAppError(ErrorCodeGitCheckoutError, "test", nil),
			expected: true,
		},
		{
			name:     "git not found error",
			err:      NewAppError(ErrorCodeGitNotFound, "test", nil),
			expected: true,
		},
		{
			name:     "file system error",
			err:      NewAppError(ErrorCodeFileSystemError, "fs error", nil),
			expected: false,
		},
		{
			name:     "regular error",
			err:      regularErr,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsFileSystemError(t *testing.T) {
	regularErr := errors.New("regular error")

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "directory not found error",
			err:      NewAppError(ErrorCodeDirectoryNotFound, "test", nil),
			expected: true,
		},
		{
			name:     "permission denied error",
			err:      NewAppError(ErrorCodePermissionDenied, "test", nil),
			expected: true,
		},
		{
			name:     "symlink creation failed error",
			err:      NewAppError(ErrorCodeSymlinkCreationFailed, "test", nil),
			expected: true,
		},
		{
			name:     "git error",
			err:      NewAppError(ErrorCodeGitCloneError, "git error", nil),
			expected: false,
		},
		{
			name:     "regular error",
			err:      regularErr,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsFileSystemError(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetUserFriendlyMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "git not installed",
			err:      NewAppError(ErrorCodeGitNotInstalled, "git not found", nil),
			expected: "Git is not installed or not available in PATH. Please install Git and try again.",
		},
		{
			name:     "git clone failed",
			err:      NewAppError(ErrorCodeGitCloneFailed, "clone failed", nil),
			expected: "Failed to download the Strategic Claude Basic repository. Please check your internet connection.",
		},
		{
			name:     "permission denied",
			err:      NewAppError(ErrorCodePermissionDenied, "no access", nil),
			expected: "Permission denied. Please check that you have write permissions to the target directory.",
		},
		{
			name:     "already installed",
			err:      NewAppError(ErrorCodeAlreadyInstalled, "exists", nil),
			expected: "Strategic Claude Basic is already installed in this directory. Use --force to reinstall or --force-core to update core files only.",
		},
		{
			name:     "user cancelled",
			err:      NewAppError(ErrorCodeUserCancelled, "cancelled", nil),
			expected: "Operation cancelled by user.",
		},
		{
			name:     "unknown app error",
			err:      NewAppError("UNKNOWN_ERROR", "unknown error", nil),
			expected: "unknown error",
		},
		{
			name:     "regular error",
			err:      errors.New("regular error message"),
			expected: "regular error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetUserFriendlyMessage(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestErrorCode_String(t *testing.T) {
	// Test that error codes can be converted to strings
	code := ErrorCodeFileSystemError
	str := string(code)
	if str != "FILE_SYSTEM_ERROR" {
		t.Errorf("Expected FILE_SYSTEM_ERROR, got %s", str)
	}
}

func TestError_Chain(t *testing.T) {
	// Test error chaining works with errors.Is and errors.As
	originalErr := errors.New("original")
	appErr := NewAppError(ErrorCodeFileSystemError, "wrapped", originalErr)

	// Test errors.Is
	if !errors.Is(appErr, originalErr) {
		t.Error("errors.Is should find the original error")
	}

	// Test errors.As
	var targetAppErr *AppError
	if !errors.As(appErr, &targetAppErr) {
		t.Error("errors.As should find the AppError")
	}

	if targetAppErr.Code != ErrorCodeFileSystemError {
		t.Errorf("Expected error code %s, got %s", ErrorCodeFileSystemError, targetAppErr.Code)
	}
}

func TestError_Context_Modification(t *testing.T) {
	// Test that context can be modified after creation
	err := NewAppError(ErrorCodeValidationFailed, "test", nil)

	// Add multiple context values
	_ = err.WithContext("key1", "value1").
		WithContext("key2", 123).
		WithContext("key3", true)

	// Verify all values are present
	if err.Context["key1"] != "value1" {
		t.Error("Expected key1 context value")
	}
	if err.Context["key2"] != 123 {
		t.Error("Expected key2 context value")
	}
	if err.Context["key3"] != true {
		t.Error("Expected key3 context value")
	}

	// Test overwriting context values
	_ = err.WithContext("key1", "new_value1")
	if err.Context["key1"] != "new_value1" {
		t.Error("Expected context value to be overwritten")
	}
}
