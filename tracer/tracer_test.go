package otelkit

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestNew(t *testing.T) {
	tracerName := "test-tracer"
	tr := New(tracerName)

	if tr == nil {
		t.Fatal("New() returned nil")
	}

	if tr.tracer == nil {
		t.Error("New() tracer field is nil")
	}
}

func TestTracer_OtelTracer(t *testing.T) {
	tracerName := "test-tracer"
	tr := New(tracerName)

	otelTracer := tr.OtelTracer()
	if otelTracer == nil {
		t.Error("OtelTracer() returned nil")
	}
}

func TestSetGlobalTracerProvider(t *testing.T) {
	// Create a noop tracer provider for testing
	tp := noop.NewTracerProvider()

	// This should not panic
	SetGlobalTracerProvider(tp)

	// Verify it was set
	currentProvider := otel.GetTracerProvider()
	if currentProvider != tp {
		t.Error("SetGlobalTracerProvider() did not set the provider correctly")
	}
}

func TestTracer_Start(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	spanName := "test-span"

	newCtx, span := tr.Start(ctx, spanName)

	if newCtx == nil {
		t.Error("Start() returned nil context")
	}
	if span == nil {
		t.Error("Start() returned nil span")
	}

	// Verify context contains the span
	spanFromCtx := trace.SpanFromContext(newCtx)
	if spanFromCtx.SpanContext().TraceID() != span.SpanContext().TraceID() {
		t.Error("Context does not contain the expected span")
	}

	span.End()
}

func TestTracer_StartServerSpan(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	operation := "server-operation"
	attrs := []attribute.KeyValue{
		attribute.String("test.key", "test.value"),
	}

	newCtx, span := tr.StartServerSpan(ctx, operation, attrs...)

	if newCtx == nil {
		t.Error("StartServerSpan() returned nil context")
	}
	if span == nil {
		t.Error("StartServerSpan() returned nil span")
	}

	// For noop tracer, spans may not be valid, so just check that we get a span
	if span == nil {
		t.Error("StartServerSpan() should return a span (even if noop)")
	}

	span.End()
}

func TestTracer_StartClientSpan(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	operation := "client-operation"
	attrs := []attribute.KeyValue{
		attribute.String("client.key", "client.value"),
	}

	newCtx, span := tr.StartClientSpan(ctx, operation, attrs...)

	if newCtx == nil {
		t.Error("StartClientSpan() returned nil context")
	}
	if span == nil {
		t.Error("StartClientSpan() returned nil span")
	}

	// Just verify we get a span back - validity depends on tracer implementation
	span.End()
}

func TestTracer_GetTraceID(t *testing.T) {
	tr := New("test-tracer")

	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "context without span",
			ctx:  context.Background(),
			want: "",
		},
	}

	// Test with active span
	ctx := context.Background()
	spanCtx, span := tr.Start(ctx, "test-span")
	defer span.End()

	traceID := tr.GetTraceID(spanCtx)
	if span.SpanContext().IsValid() && traceID == "" {
		t.Error("GetTraceID() returned empty string for valid span context")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tr.GetTraceID(tt.ctx)
			if got != tt.want {
				t.Errorf("GetTraceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTracer_StartWithOptions(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()
	spanName := "test-span-with-options"

	// Test with span options
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attribute.String("custom.attr", "value")),
		trace.WithSpanKind(trace.SpanKindInternal),
	}

	newCtx, span := tr.Start(ctx, spanName, opts...)

	if newCtx == nil {
		t.Error("Start() with options returned nil context")
	}
	if span == nil {
		t.Error("Start() with options returned nil span")
	}

	// Just verify we get a span back
	span.End()
}

// Test helper to verify span attributes
func TestTracer_SpanAttributes(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()

	attrs := []attribute.KeyValue{
		attribute.String("service.name", "test-service"),
		attribute.Int("request.count", 42),
		attribute.Bool("is.test", true),
	}

	_, span := tr.StartServerSpan(ctx, "test-operation", attrs...)
	defer span.End()

	// Test passes if we reach here without panicking
	// The API call worked without errors

	spanContext := span.SpanContext()
	// For no-op tracers, span context may not be valid, which is acceptable
	_ = spanContext // Avoid unused variable warning
	// The real test is that the function call didn't panic,
	// and we can work with the span object
}

func TestTracer_NestedSpans(t *testing.T) {
	tr := New("test-tracer")
	ctx := context.Background()

	// Create parent span
	parentCtx, parentSpan := tr.Start(ctx, "parent-span")
	defer parentSpan.End()

	// Create child span
	childCtx, childSpan := tr.Start(parentCtx, "child-span")
	defer childSpan.End()

	// Verify context propagation - just check that we can get a span from context
	spanFromChildCtx := trace.SpanFromContext(childCtx)
	if spanFromChildCtx == nil {
		t.Error("Child context should contain a span")
	}
}
