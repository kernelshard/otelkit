package config

import (
	"fmt"
)

// ConfigError represents a validation error in configuration.
type ConfigError struct {
	Field   string
	Message string
}

// Error returns a string representation of the error.
func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s: %s", e.Field, e.Message)
}

// InitializationError wraps failures during component setup.
type InitializationError struct {
	Component string
	Cause     error
}

// Error returns a string representation of the error.
func (e *InitializationError) Error() string {
	return fmt.Sprintf("initialization failed for %s: %v", e.Component, e.Cause)
}

// Unwrap returns the underlying error.
func (e *InitializationError) Unwrap() error {
	return e.Cause
}

// PropagationError wraps errors related to context propagation.
type PropagationError struct {
	Operation string
	Cause     error
}

// Error returns a string representation of the error.
func (e *PropagationError) Error() string {
	return fmt.Sprintf("context propagation failed during %s: %v", e.Operation, e.Cause)
}

// Unwrap returns the underlying error.
func (e *PropagationError) Unwrap() error {
	return e.Cause
}

// NewConfigError creates a new ConfigError.
func NewConfigError(field, message string) error {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}

// NewInitializationError creates a new InitializationError.
func NewInitializationError(component string, cause error) error {
	return &InitializationError{
		Component: component,
		Cause:     cause,
	}
}

// NewPropagationError creates a new PropagationError.
func NewPropagationError(operation string, cause error) error {
	return &PropagationError{
		Operation: operation,
		Cause:     cause,
	}
}
