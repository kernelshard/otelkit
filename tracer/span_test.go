package tracer

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func TestAddAttributes(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	attrs := []attribute.KeyValue{
		attribute.String("test.attr1", "value1"),
		attribute.Int("test.attr2", 42),
		attribute.Bool("test.attr3", true),
	}

	// This should not panic
	AddAttributes(span, attrs...)

	// Test with nil span - should not panic
	AddAttributes(nil, attrs...)
}

func TestAddEvent(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	eventName := "test.event"
	attrs := []attribute.KeyValue{
		attribute.String("event.attr", "value"),
	}

	// This should not panic
	AddEvent(span, eventName, attrs...)

	// Test with nil span - should not panic
	AddEvent(nil, eventName, attrs...)
}

func TestAddTimedEvent(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	eventName := "timed.event"
	duration := 100 * time.Millisecond

	// This should not panic
	AddTimedEvent(span, eventName, duration)
}

func TestRecordError(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	testError := errors.New("test error")

	// This should not panic
	RecordError(span, testError)

	// Test with nil span - should not panic
	RecordError(nil, testError)

	// Test with nil error - should not panic
	RecordError(span, nil)

	// Test with both nil - should not panic
	RecordError(nil, nil)
}

func TestRecordErrorEnhanced(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	testError := errors.New("something went wrong")

	// Test basic enhanced error recording - should default to custom type
	RecordErrorEnhanced(span, testError)

	// Test with nil span - should not panic
	RecordErrorEnhanced(nil, testError)

	// Test with nil error - should not panic
	RecordErrorEnhanced(span, nil)

	// Test with explicit error type
	RecordErrorEnhanced(span, testError,
		WithErrorType(ErrorTypeValidation),
		WithStackTrace(false),
		WithErrorCode("TEST_ERROR"),
		WithErrorAttributes(attribute.String("test", "value")),
	)
}

func TestEndSpan(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")

	// This should not panic
	EndSpan(span)

	// Test with nil span - should not panic
	EndSpan(nil)
}

func TestIsRecording(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()

	// Test with nil span
	if IsRecording(nil) {
		t.Error("IsRecording(nil) should return false")
	}

	// Test with valid span
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	// For noop tracer, this will be false, but for real tracer it would be true
	// The test ensures the function works without panicking
	result := IsRecording(span)
	_ = result // We don't assert the value since it depends on the tracer type
}

func TestSpanUtilitiesWithRealSpan(t *testing.T) {
	// Test with real span to ensure the utilities work
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	// Test AddAttributes - should not panic
	attrs := []attribute.KeyValue{attribute.String("key", "value")}
	AddAttributes(span, attrs...)

	// Test AddEvent - should not panic
	AddEvent(span, "test-event", attrs...)

	// Test RecordError - should not panic
	testErr := errors.New("test error")
	RecordError(span, testErr)

	// All operations completed without panic
}

func TestSpanUtilitiesErrorHandling(t *testing.T) {
	// Test error handling with various edge cases
	tr := New("test-tracer")
	ctx := context.Background()
	_, span := tr.Start(ctx, "test-span")
	defer span.End()

	// Test empty attribute slice
	AddAttributes(span)

	// Test empty event name
	AddEvent(span, "")

	// Test zero duration
	AddTimedEvent(span, "zero-duration", 0)

	// Test negative duration
	AddTimedEvent(span, "negative-duration", -1*time.Second)
}
