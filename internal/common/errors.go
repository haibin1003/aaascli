// Package common provides shared utilities for command execution.
package common

import (
	"errors"
	"fmt"
)

// Common errors used across the application.
// These errors should be used for consistent error handling and testing.
var (
	// ErrResourceNotFound indicates the requested resource does not exist.
	ErrResourceNotFound = errors.New("resource not found")

	// ErrUnauthorized indicates authentication or authorization failure.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrInvalidInput indicates invalid user input.
	ErrInvalidInput = errors.New("invalid input")

	// ErrAPIFailure indicates the API request failed.
	ErrAPIFailure = errors.New("API request failed")

	// ErrConfigNotFound indicates configuration file not found.
	ErrConfigNotFound = errors.New("configuration file not found")

	// ErrWorkspaceRequired indicates workspace key is required but missing.
	ErrWorkspaceRequired = errors.New("workspace key is required")
)

// AutoDetectError represents an auto-detection failure with detailed information.
// This error type preserves structured information for JSON output.
type AutoDetectError struct {
	Message    string   `json:"message"`
	Details    string   `json:"details,omitempty"`
	Suggestion string   `json:"suggestion,omitempty"`
	Missing    []string `json:"missing,omitempty"` // List of missing parameters
}

// Error implements the error interface.
func (e *AutoDetectError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s", e.Message, e.Details)
	}
	return e.Message
}

// NewAutoDetectError creates a new auto-detect error with the given message.
func NewAutoDetectError(message string) *AutoDetectError {
	return &AutoDetectError{
		Message: message,
	}
}

// WithDetails adds details to the error.
func (e *AutoDetectError) WithDetails(details string) *AutoDetectError {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion to the error.
func (e *AutoDetectError) WithSuggestion(suggestion string) *AutoDetectError {
	e.Suggestion = suggestion
	return e
}

// WithMissing adds missing parameters to the error.
func (e *AutoDetectError) WithMissing(params ...string) *AutoDetectError {
	e.Missing = params
	return e
}

// ErrorCode represents error codes for API responses.
type ErrorCode string

const (
	ErrorCodeSuccess        ErrorCode = "00"
	ErrorCodeUnauthorized   ErrorCode = "01"
	ErrorCodeNotFound       ErrorCode = "404"
	ErrorCodeInvalidRequest ErrorCode = "400"
)

// IsNotFound checks if the error indicates resource not found.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrResourceNotFound)
}

// IsUnauthorized checks if the error indicates authentication failure.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}
