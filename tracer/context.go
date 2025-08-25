package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type traceContextKey string

const traceIDKey traceContextKey = "trace_id"

// InjectTraceIDIntoContext adds trace ID into the context (as a new value).
func InjectTraceIDIntoContext(ctx context.Context, span trace.Span) context.Context {
	if span == nil {
		return ctx
	}
	return context.WithValue(ctx, traceIDKey, span.SpanContext().TraceID().String())
}

// TraceIDFromContext tries to retrieve trace ID from context.
func TraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}
