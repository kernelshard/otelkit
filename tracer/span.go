// Package tracer provides span utility functions for OpenTelemetry tracing.
// These utilities offer safe, convenient methods for common span operations
// with built-in nil checks and error handling.
package tracer

import (
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// AddAttributes safely adds one or more attributes to the given span.
// If the span is nil, this function is a no-op. This is useful for adding
// contextual information to spans such as user IDs, request parameters,
// or business logic details.
//
// Example:
//
//	AddAttributes(span,
//	    attribute.String("user.id", "12345"),
//	    attribute.Int("request.size", 1024),
//	    attribute.Bool("cache.hit", true),
//	)
func AddAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span != nil {
		span.SetAttributes(attrs...)
	}
}

// AddEvent safely adds a named event with optional attributes to the span.
// Events are timestamped markers that can help understand the flow of execution.
// If the span is nil, this function is a no-op.
//
// Example:
//
//	AddEvent(span, "cache.miss",
//	    attribute.String("key", cacheKey),
//	    attribute.String("reason", "expired"),
//	)
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	if span != nil {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// AddTimedEvent adds an event with duration information to the span.
// This is useful for recording the time taken for specific operations
// within a larger span. The duration is added as a string attribute.
//
// Example:
//
//	start := time.Now()
//	// ... perform operation
//	AddTimedEvent(span, "database.query", time.Since(start))
func AddTimedEvent(span trace.Span, name string, duration time.Duration) {
	span.AddEvent(name, trace.WithAttributes(
		attribute.String("duration", duration.String()),
	))
}

// RecordError safely records an error on the span and sets the span status to error.
// This function handles nil checks for both span and error. When an error is recorded,
// the span status is automatically set to codes.Error with the error message.
// This is essential for proper error tracking in distributed tracing.
//
// Example:
//
//	if err := doSomething(); err != nil {
//	    RecordError(span, err)
//	    return err
//	}
func RecordError(span trace.Span, err error) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// EndSpan safely ends the given span.
// If the span is nil, this function is a no-op.
// This provides a safe way to end spans without worrying about nil checks.
//
// Example:
//
//	defer EndSpan(span)
func EndSpan(span trace.Span) {
	if span != nil {
		span.End()
	}
}

// IsRecording checks if the span is currently recording telemetry data.
// Returns false if the span is nil or if the span context is invalid.
// This can be used to avoid expensive operations when tracing is disabled
// or when working with noop spans.
//
// Example:
//
//	if IsRecording(span) {
//	    // Perform expensive attribute computation
//	    span.SetAttributes(expensiveAttributes()...)
//	}
func IsRecording(span trace.Span) bool {
	if span == nil {
		return false
	}
	return span.SpanContext().IsValid()
}
