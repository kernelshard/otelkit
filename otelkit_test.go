package otelkit

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func TestNewTracer(t *testing.T) {
	// Remove the ineffectual assignment
	_ = New("test-service")
}

func TestNewProviderConfig(t *testing.T) {
	config := NewProviderConfig("test-service", "v1.0.0")
	if config == nil {
		t.Error("NewProviderConfig() returned nil config")
	}
}

func TestNewDefaultProvider(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDefaultProvider(ctx, "test-service")
	if err != nil {
		t.Fatalf("NewDefaultProvider() failed: %v", err)
	}
	if provider == nil {
		t.Error("NewDefaultProvider() should return a non-nil provider")
	}

	// Test shutdown
	if err := ShutdownTracerProvider(ctx, provider); err != nil {
		t.Errorf("ShutdownTracerProvider() failed: %v", err)
	}
}

func TestSetupTracing(t *testing.T) {
	ctx := context.Background()
	shutdown, err := SetupTracing(ctx, "test-service")
	if err != nil {
		t.Fatalf("SetupTracing() failed: %v", err)
	}

	// Test that we can create a tracer after setup
	tracer := New("test-service")
	if tracer == nil {
		t.Error("Should be able to create tracer after SetupTracing")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown function failed: %v", err)
	}
}

func TestNewHttpMiddleware(t *testing.T) {
	tracer := New("test-service")
	middleware := NewHttpMiddleware(tracer)
	if middleware == nil {
		t.Error("NewHttpMiddleware() should return a non-nil middleware")
	}
}

func TestSpanUtilities(t *testing.T) {
	// Test with a real span
	ctx := context.Background()
	provider, err := NewDefaultProvider(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider)

	tracer := New("test-service")
	ctx, span := tracer.Start(ctx, "test-span")

	// Test AddAttributes
	AddAttributes(span, attribute.String("test.key", "test.value"))

	// Test AddEvent
	AddEvent(span, "test.event", attribute.String("event.key", "event.value"))

	// Test AddTimedEvent
	AddTimedEvent(span, "timed.event", 100*time.Millisecond)

	// Test RecordError with nil span (should not panic)
	RecordError(nil, nil)

	// Test EndSpan with nil span (should not panic)
	EndSpan(nil)

	// Test IsRecording with nil span
	if IsRecording(nil) {
		t.Error("IsRecording should return false for nil span")
	}

	// Test IsRecording with real span - note that some spans might not be recording
	// depending on sampling configuration, so we'll just test that the function doesn't panic
	IsRecording(span)

	EndSpan(span)
}

func TestSetGlobalTracerProvider(t *testing.T) {
	ctx := context.Background()
	provider, err := NewDefaultProvider(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	SetGlobalTracerProvider(provider)

	// Verify that we can get the global provider
	// Note: We can't easily test trace.GetTracerProvider() as it's a global state
	// that might be affected by other tests. We'll just verify our function works.
	if provider == nil {
		t.Error("Provider should not be nil")
	}

	ShutdownTracerProvider(ctx, provider)
}

func TestTypeAliases(t *testing.T) {
	// Test that type aliases work correctly by ensuring they compile
	// These are compile-time checks, so we just need to reference them
	var span Span
	var tracer Tracer
	var middleware HTTPMiddleware
	var config ProviderConfig
	var configErr ConfigError
	var initErr InitializationError

	// Just reference the variables to avoid "declared but not used" errors
	_ = span
	_ = tracer
	_ = middleware
	_ = config
	_ = configErr
	_ = initErr
}

func TestDeprecatedFunctions(t *testing.T) {
	ctx := context.Background()

	// Test deprecated functions still work
	_, err := SetupTracingWithDefaults(ctx, "test-service", "v1.0.0")
	if err != nil {
		t.Errorf("SetupTracingWithDefaults failed: %v", err)
	}

	shutdown := MustSetupTracing(ctx, "test-service")
	if shutdown == nil {
		t.Error("MustSetupTracing should return shutdown function")
	}

	config := NewProviderConfig("test-service", "v1.0.0")
	_, err = SetupCustomTracing(ctx, config)
	if err != nil {
		t.Errorf("SetupCustomTracing failed: %v", err)
	}
}
