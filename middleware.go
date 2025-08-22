// Package otelkit provides HTTP middleware for automatic request tracing.
// The middleware integrates seamlessly with any HTTP framework that supports
// the standard http.Handler interface, including gorilla/mux, chi, gin, echo, and others.
package otelkit

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// HTTPMiddleware provides HTTP middleware for automatic request tracing.
// It extracts trace context from incoming requests, creates server spans,
// and automatically records HTTP-specific attributes like method, URL, status code,
// and user agent. The middleware handles trace context propagation according to
// W3C Trace Context and B3 propagation standards.
//
// The middleware is compatible with any HTTP framework that uses the standard
// http.Handler interface.
type HTTPMiddleware struct {
	tracer *Tracer
}

// NewHttpMiddleware creates a new HTTPMiddleware instance using the provided Tracer.
// The tracer will be used to create spans for all incoming HTTP requests.
//
// Example:
//
//	tracer := otelkit.New("http-service")
//	middleware := tracer.NewHttpMiddleware(tracer)
//
//	// With gorilla/mux
//	r := mux.NewRouter()
//	r.Use(middleware.Middleware)
//
//	// With chi
//	r := chi.NewRouter()
//	r.Use(middleware.Middleware)
//
//	// With standard http.ServeMux
//	mux := http.NewServeMux()
//	handler := middleware.Middleware(mux)
func NewHttpMiddleware(tracer *Tracer) *HTTPMiddleware {
	return &HTTPMiddleware{
		tracer: tracer,
	}
}

// Middleware returns an HTTP handler middleware function that automatically traces incoming requests.
//
// The middleware performs the following operations:
//  1. Extracts trace context from incoming request headers (supports W3C Trace Context and B3)
//  2. Creates a new server span with operation name "METHOD /path"
//  3. Adds standard HTTP attributes: method, URL, user agent
//  4. Wraps the response writer to capture the HTTP status code
//  5. Propagates the trace context to downstream handlers
//  6. Records the final HTTP status code when the request completes
//
// Example usage:
//
//	middleware := tracer.NewHttpMiddleware(tracer)
//
//	http.Handle("/api/", middleware.Middleware(apiHandler))
//
//	// Or with a router:
//	r := mux.NewRouter()
//	r.Use(middleware.Middleware)
//	r.HandleFunc("/users/{id}", getUserHandler)
func (m *HTTPMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// extract trace information from the incoming request header
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

		// Start a new server-kind span for the operation
		operationName := r.Method + " " + r.URL.Path

		ctx, span := m.tracer.Start(ctx, operationName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String(AttrHTTPMethod, r.Method),
				attribute.String(AttrHTTPURL, r.URL.String()),
				attribute.String(AttrHTTPUserAgent, r.UserAgent()),
			),
		)
		defer span.End()

		// Wrap the ResponseWriter to capture the status code.
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Call the next handler in the chain.
		next.ServeHTTP(rw, r.WithContext(ctx))

		// Record the status code as an attribute.
		span.SetAttributes(attribute.Int(AttrHTTPStatusCode, rw.status))
	})
}

// responseWriter is a wrapper around http.ResponseWriter that captures the HTTP status code.
// It implements the http.ResponseWriter interface and transparently passes through all
// method calls while recording the status code for tracing purposes.
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader captures the HTTP status code before calling the underlying WriteHeader method.
// This allows the middleware to record the final response status in the trace span.
func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
