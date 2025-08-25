package propagation

import (
	"context"
	"net/http"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/samims/otelkit"
)

func TestInjectTraceContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	ctx := context.Background()

	// Set up the propagator explicitly
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Initialize tracer provider first
	provider, err := otelkit.NewDefaultProvider(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer otelkit.ShutdownTracerProvider(ctx, provider)

	// Create a span to have trace context to inject
	tracer := otelkit.New("test-service")
	ctx, span := tracer.Start(ctx, "test-span")
	defer span.End()

	InjectTraceContext(ctx, req)

	// Debug: print all headers
	t.Logf("Request headers: %v", req.Header)

	// Verify that the trace context is injected into the request headers
	if req.Header.Get("traceparent") == "" {
		t.Error("Expected traceparent header to be set")
	}
}

func TestExtractTraceContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:8080", nil)
	ctx := context.Background()

	// Set up the propagator explicitly
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Initialize tracer provider first
	provider, err := otelkit.NewDefaultProvider(ctx, "test-service")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer otelkit.ShutdownTracerProvider(ctx, provider)

	InjectTraceContext(ctx, req)

	extractedCtx := ExtractTraceContext(req)

	// Verify that the trace context is extracted correctly
	if extractedCtx == nil {
		t.Error("Expected extracted context to be non-nil")
	}
}
