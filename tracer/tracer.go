package tracer

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps OpenTelemetry tracer with convenience methods for easier tracing operations.
// It provides a simplified interface for creating spans, adding attributes, and managing
// trace context while maintaining full compatibility with OpenTelemetry standards.
//
// Example usage:
//
//	tracer := otelkit.New("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
//	// ... your code here
type Tracer struct {
	tracer trace.Tracer
}

// New creates a new Tracer instance with the specified name.
// The name is used to identify the tracer and appears in telemetry data.
// It's recommended to use your service or component name.
//
// Example:
//
//	tracer := otelkit.New("user-service")
//	tracer := otelkit.New("payment-processor")
func New(name string) *Tracer {
	return &Tracer{
		tracer: otel.Tracer(name),
	}
}

// OtelTracer returns the underlying OpenTelemetry tracer instance.
// This is useful when you need direct access to OpenTelemetry APIs
// or when integrating with other OpenTelemetry-compatible libraries.
//
// Example:
//
//	otelTracer := tracer.OtelTracer()
//	// Use with other OpenTelemetry libraries
func (t *Tracer) OtelTracer() trace.Tracer {
	return t.tracer
}

// SetGlobalTracerProvider sets the global OpenTelemetry tracer provider.
// This should typically be called once during application initialization.
// All subsequent tracer instances will use this provider.
//
// Example:
//
//	provider := setupTracerProvider()
//	tracer.SetGlobalTracerProvider(provider)
func SetGlobalTracerProvider(tp trace.TracerProvider) {
	otel.SetTracerProvider(tp)
}

// Start creates a new span with the given name and options.
// Returns a new context containing the span and the span itself.
// The span must be ended by calling span.End() when the operation completes.
//
// Example:
//
//	ctx, span := tracer.Start(ctx, "database-query")
//	defer span.End()
//	// ... perform database operation
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// StartServerSpan creates a new server span for incoming requests or operations.
// This is a convenience method that automatically sets the span kind to SpanKindServer
// and adds the provided attributes. Use this for HTTP handlers, gRPC server methods,
// message queue consumers, etc.
//
// Example:
//
//	ctx, span := tracer.StartServerSpan(ctx, "handle-user-request",
//	    attribute.String("user.id", userID),
//	    attribute.String("request.method", "POST"),
//	)
//	defer span.End()
func (t *Tracer) StartServerSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operation,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindServer),
	)
}

// StartClientSpan creates a new client span for outgoing requests or operations.
// This is a convenience method that automatically sets the span kind to SpanKindClient
// and adds the provided attributes. Use this for HTTP client requests, gRPC client calls,
// database queries, external API calls, etc.
//
// Example:
//
//	ctx, span := tracer.StartClientSpan(ctx, "call-payment-api",
//	    attribute.String("http.method", "POST"),
//	    attribute.String("http.url", "https://api.payment.com/charge"),
//	)
//	defer span.End()
func (t *Tracer) StartClientSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, operation,
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindClient),
	)
}

// GetTraceID extracts and returns the trace ID from the current span context.
// Returns an empty string if no valid span is found in the context.
// This is useful for correlation logging and debugging.
//
// Example:
//
//	traceID := tracer.GetTraceID(ctx)
//	log.WithField("trace_id", traceID).Info("Processing request")
func (t *Tracer) GetTraceID(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		return span.SpanContext().TraceID().String()
	}
	return ""
}
