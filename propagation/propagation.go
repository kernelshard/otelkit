package propagation

import (
	"context"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// InjectTraceContext injects the current trace context into the HTTP request headers.
func InjectTraceContext(ctx context.Context, req *http.Request) {
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractTraceContext extracts trace context from HTTP request headers into the context.
func ExtractTraceContext(req *http.Request) context.Context {
	propagator := otel.GetTextMapPropagator()
	ctx := propagator.Extract(context.Background(), propagation.HeaderCarrier(req.Header))
	return ctx
}
