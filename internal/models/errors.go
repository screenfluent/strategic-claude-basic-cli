package models

import (
	"errors"
	"fmt"
)

// ErrorCode represents different types of errors that can occur
type ErrorCode string

const (
	// Git operation errors
	ErrorCodeGitCloneFailed    ErrorCode = "GIT_CLONE_FAILED"
	ErrorCodeGitCheckoutFailed ErrorCode = "GIT_CHECKOUT_FAILED"
	ErrorCodeGitNotInstalled   ErrorCode = "GIT_NOT_INSTALLED"

	// File system errors
	ErrorCodeDirectoryNotFound     ErrorCode = "DIRECTORY_NOT_FOUND"
	ErrorCodeDirectoryNotEmpty     ErrorCode = "DIRECTORY_NOT_EMPTY"
	ErrorCodePermissionDenied      ErrorCode = "PERMISSION_DENIED"
	ErrorCodeFileAlreadyExists     ErrorCode = "FILE_ALREADY_EXISTS"
	ErrorCodeSymlinkCreationFailed ErrorCode = "SYMLINK_CREATION_FAILED"
	ErrorCodeSymlinkInvalid        ErrorCode = "SYMLINK_INVALID"

	// Installation errors
	ErrorCodeInstallationFailed ErrorCode = "INSTALLATION_FAILED"
	ErrorCodeAlreadyInstalled   ErrorCode = "ALREADY_INSTALLED"
	ErrorCodeNotInstalled       ErrorCode = "NOT_INSTALLED"
	ErrorCodeBackupFailed       ErrorCode = "BACKUP_FAILED"
	ErrorCodeRestoreFailed      ErrorCode = "RESTORE_FAILED"

	// Validation errors
	ErrorCodeInvalidPath          ErrorCode = "INVALID_PATH"
	ErrorCodeInvalidConfiguration ErrorCode = "INVALID_CONFIGURATION"
	ErrorCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"

	// Network errors
	ErrorCodeNetworkTimeout ErrorCode = "NETWORK_TIMEOUT"
	ErrorCodeNetworkError   ErrorCode = "NETWORK_ERROR"

	// User interaction errors
	ErrorCodeUserCancelled ErrorCode = "USER_CANCELLED"
	ErrorCodeInputError    ErrorCode = "INPUT_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Cause   error                  `json:"-"` // Original error, not serialized
	Context map[string]interface{} `json:"context,omitempty"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error code
func (e *AppError) Is(target error) bool {
	if appErr, ok := target.(*AppError); ok {
		return e.Code == appErr.Code
	}
	return false
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// Predefined error constructors for common scenarios

// NewGitError creates a git-related error
func NewGitError(code ErrorCode, operation string, cause error) *AppError {
	return NewAppError(code, fmt.Sprintf("Git operation failed: %s", operation), cause).
		WithContext("operation", operation)
}

// NewFileSystemError creates a file system related error
func NewFileSystemError(code ErrorCode, path string, cause error) *AppError {
	return NewAppError(code, fmt.Sprintf("File system operation failed for path: %s", path), cause).
		WithContext("path", path)
}

// NewInstallationError creates an installation-related error
func NewInstallationError(code ErrorCode, targetDir string, cause error) *AppError {
	return NewAppError(code, fmt.Sprintf("Installation failed in directory: %s", targetDir), cause).
		WithContext("target_dir", targetDir)
}

// NewValidationError creates a validation error
func NewValidationError(field string, value interface{}, message string) *AppError {
	return NewAppError(ErrorCodeValidationFailed, fmt.Sprintf("Validation failed for %s: %s", field, message), nil).
		WithContext("field", field).
		WithContext("value", value)
}

// IsErrorCode checks if the error has the specified error code
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// IsGitError checks if the error is a git-related error
func IsGitError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrorCodeGitCloneFailed, ErrorCodeGitCheckoutFailed, ErrorCodeGitNotInstalled:
			return true
		}
	}
	return false
}

// IsFileSystemError checks if the error is a file system related error
func IsFileSystemError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrorCodeDirectoryNotFound, ErrorCodeDirectoryNotEmpty, ErrorCodePermissionDenied,
			ErrorCodeFileAlreadyExists, ErrorCodeSymlinkCreationFailed, ErrorCodeSymlinkInvalid:
			return true
		}
	}
	return false
}

// GetUserFriendlyMessage returns a user-friendly error message
func GetUserFriendlyMessage(err error) string {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		return err.Error()
	}

	switch appErr.Code {
	case ErrorCodeGitNotInstalled:
		return "Git is not installed or not available in PATH. Please install Git and try again."
	case ErrorCodeGitCloneFailed:
		return "Failed to download the Strategic Claude Basic repository. Please check your internet connection."
	case ErrorCodePermissionDenied:
		return "Permission denied. Please check that you have write permissions to the target directory."
	case ErrorCodeAlreadyInstalled:
		return "Strategic Claude Basic is already installed in this directory. Use --force to reinstall or --force-core to update core files only."
	case ErrorCodeNotInstalled:
		return "Strategic Claude Basic is not installed in this directory."
	case ErrorCodeUserCancelled:
		return "Operation cancelled by user."
	case ErrorCodeDirectoryNotFound:
		return "The specified directory does not exist."
	case ErrorCodeInvalidPath:
		return "The specified path is invalid or inaccessible."
	default:
		return appErr.Message
	}
}
