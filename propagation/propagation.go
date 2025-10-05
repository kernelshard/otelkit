package propagation

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTraceContext injects the current trace context into the HTTP request headers.
// If req is nil, this function is a no-op to prevent panics.
func InjectTraceContext(ctx context.Context, req *http.Request) {
	if req == nil {
		return
	}
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractTraceContext extracts trace context from HTTP request headers into the context.
// If req is nil, returns the background context.
func ExtractTraceContext(req *http.Request) context.Context {
	if req == nil {
		return context.Background()
	}
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(req.Header))
	return ctx
}
