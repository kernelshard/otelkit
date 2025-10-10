// Package tracer provides enhanced error handling utilities for OpenTelemetry tracing.
// This package offers error recording with explicit classification, optional stack traces,
// and flexible configuration options.
package tracer

import (
	"runtime"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ErrorType represents the classification of an error for better observability.
type ErrorType string

const (
	// ErrorTypeNetwork represents network-related errors (connection failures, timeouts, DNS issues)
	ErrorTypeNetwork ErrorType = "network"
	// ErrorTypeDatabase represents database-related errors (connection issues, query failures, constraint violations)
	ErrorTypeDatabase ErrorType = "database"
	// ErrorTypeValidation represents validation errors (invalid input, business rule violations)
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeSystem represents system-level errors (file I/O, memory issues, OS errors)
	ErrorTypeSystem ErrorType = "system"
	// ErrorTypeCustom represents application-specific custom errors
	ErrorTypeCustom ErrorType = "custom"
)

// ErrorOption represents a configuration option for error recording.
type ErrorOption func(*errorConfig)

type errorConfig struct {
	errorType   ErrorType
	stackTrace  bool
	errorCode   string
	customAttrs []attribute.KeyValue
}

// WithErrorType sets the error type for classification.
func WithErrorType(errorType ErrorType) ErrorOption {
	return func(c *errorConfig) {
		c.errorType = errorType
	}
}

// WithStackTrace enables or disables stack trace capture.
func WithStackTrace(enabled bool) ErrorOption {
	return func(c *errorConfig) {
		c.stackTrace = enabled
	}
}

// WithErrorCode sets a custom error code for the error.
func WithErrorCode(code string) ErrorOption {
	return func(c *errorConfig) {
		c.errorCode = code
	}
}

// WithErrorAttributes adds custom attributes to the error span.
func WithErrorAttributes(attrs ...attribute.KeyValue) ErrorOption {
	return func(c *errorConfig) {
		c.customAttrs = append(c.customAttrs, attrs...)
	}
}

// captureStackTrace captures the current stack trace with dynamic sizing.
func captureStackTrace() string {
	// Start with a reasonable buffer, but allow it to grow if needed
	buf := make([]byte, 2048)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			return string(buf[:n])
		}
		// Buffer was too small, double it and try again
		buf = make([]byte, len(buf)*2)
	}
}

// RecordErrorEnhanced records an error on the span with explicit classification and options.
// This function provides error recording with user-specified classification, optional stack traces,
// and flexible configuration through options.
//
// Error classification must be explicitly specified using WithErrorType() - there is no
// automatic classification to avoid confusion and ensure accuracy.
//
// Example:
//
//	// Basic usage - defaults to custom type
//	otelkit.RecordErrorEnhanced(span, err)
//
//	// With explicit classification and options
//	otelkit.RecordErrorEnhanced(span, err,
//	    otelkit.WithErrorType(otelkit.ErrorTypeValidation),
//	    otelkit.WithStackTrace(true),
//	    otelkit.WithErrorCode("VALIDATION_FAILED"),
//	    otelkit.WithErrorAttributes(
//	        attribute.String("field", "email"),
//	        attribute.String("reason", "invalid_format"),
//	    ),
//	)
func RecordErrorEnhanced(span trace.Span, err error, opts ...ErrorOption) {
	if span == nil || err == nil {
		return
	}

	// Default configuration - explicit classification required
	config := &errorConfig{
		errorType:  ErrorTypeCustom, // Default to custom, user must specify
		stackTrace: false,
		errorCode:  "",
	}

	// Apply options
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}

	// Record the error
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())

	// Add error classification attributes
	attrs := []attribute.KeyValue{
		attribute.String("error.type", string(config.errorType)),
	}

	// Add error code if provided
	if config.errorCode != "" {
		attrs = append(attrs, attribute.String("error.code", config.errorCode))
	}

	// Add stack trace if requested
	if config.stackTrace {
		stackTrace := captureStackTrace()
		attrs = append(attrs, attribute.String("error.stack_trace", stackTrace))
	}

	// Add custom attributes
	attrs = append(attrs, config.customAttrs...)

	// Set all attributes on the span
	span.SetAttributes(attrs...)
}
