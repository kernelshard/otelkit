// Package otelkit provides a simplified, opinionated wrapper around OpenTelemetry tracing for Go applications.
//
// This package offers easy-to-use APIs for creating and managing distributed traces while hiding
// the complexity of the underlying OpenTelemetry SDK. It provides zero-configuration setup with
// sensible defaults, while still allowing full customization when needed.
//
// # Quick Start (Recommended)
//
// For most use cases, start with SetupTracing:
//
//	ctx := context.Background()
//	shutdown, err := otelkit.SetupTracing(ctx, "my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(ctx)
//
//	tracer := otelkit.New("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
//
// # Advanced Configuration
//
// For custom configuration, use NewProviderConfig with NewProvider:
//
//	config := otelkit.NewProviderConfig("my-service", "v1.0.0").
//	    WithOTLPExporter("https://api.honeycomb.io", "http", false).
//	    WithSampling("probabilistic", 0.05)
//	provider, err := otelkit.NewProvider(ctx, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
//
// # HTTP Middleware
//
// For HTTP request tracing:
//
//	tracer := otelkit.New("web-service")
//	middleware := otelkit.NewHttpMiddleware(tracer)
//	r.Use(middleware.Middleware)
//
// The package handles:
// - OTLP exporter configuration (HTTP/gRPC)
// - Sampling strategies (probabilistic, always_on, always_off)
// - Resource management with service metadata
// - Context propagation for distributed tracing
// - HTTP and gRPC instrumentation
// - Error recording and span utilities
//
// Configuration can be done via environment variables or programmatically.
// See the README for comprehensive examples and configuration options.
package otelkit

import (
	"context"
	"net/http"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/samims/otelkit/internal/config"
	"github.com/samims/otelkit/middleware"
	"github.com/samims/otelkit/provider"
	"github.com/samims/otelkit/tracer"
)

// New creates a new tracer instance with the given name.
// This is the main entry point for creating tracers that will be used
// throughout your application for manual span creation.
//
// Example:
//
//	tracer := otelkit.New("my-service")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
func New(name string) *tracer.Tracer {
	return tracer.New(name)
}

// NewProviderConfig creates a new ProviderConfig with sensible defaults for advanced configuration.
// This is the starting point for custom tracer provider configuration when you need more control
// than the default setup provides.
//
// Example:
//
//	config := otelkit.NewProviderConfig("my-service", "v1.0.0").
//	    WithOTLPExporter("https://api.honeycomb.io", "http", false).
//	    WithSampling("probabilistic", 0.05)
func NewProviderConfig(serviceName, serviceVersion string) *provider.ProviderConfig {
	return provider.NewProviderConfig(serviceName, serviceVersion)
}

// NewProvider creates and configures a new TracerProvider using the provided configuration,
// then sets it as the global OpenTelemetry provider (only once per application lifecycle).
// This is the recommended way to initialize tracing when you need custom configuration.
//
// Example:
//
//	config := otelkit.NewProviderConfig("payment-service", "v1.2.3").
//	    WithOTLPExporter("https://api.honeycomb.io", "http", false).
//	    WithSampling("probabilistic", 0.05)
//
//	provider, err := otelkit.NewProvider(ctx, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
func NewProvider(ctx context.Context, cfg *provider.ProviderConfig) (*sdktrace.TracerProvider, error) {
	return provider.NewProvider(ctx, cfg)
}

// NewDefaultProvider creates a tracer provider with default settings and sets it as the global provider.
// This is a convenience function for quick setup in development or simple applications.
//
// Example:
//
//	provider, err := otelkit.NewDefaultProvider(ctx, "my-service", "v1.0.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
func NewDefaultProvider(ctx context.Context, serviceName string, serviceVersion ...string) (*sdktrace.TracerProvider, error) {
	return provider.NewDefaultProvider(ctx, serviceName, serviceVersion...)
}

// SetupTracing initializes OpenTelemetry tracing with sensible defaults.
// This is the simplest way to get started with tracing.
//
// Example:
//
//	shutdown, err := otelkit.SetupTracing(ctx, "my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(ctx)
func SetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) (func(context.Context) error, error) {
	return tracer.SetupTracing(ctx, serviceName, serviceVersion...)
}

// SetupTracingWithDefaults initializes tracing with hardcoded defaults.
// This is useful for quick setup without environment variables.
//
// Deprecated: Use SetupTracing instead. This function will be removed in v1.0.0.
func SetupTracingWithDefaults(ctx context.Context, serviceName, serviceVersion string) (func(context.Context) error, error) {
	return tracer.SetupTracingWithDefaults(ctx, serviceName, serviceVersion)
}

// MustSetupTracing is like SetupTracing but panics on error.
// Use this for simple programs where you want to fail fast.
//
// Deprecated: Handle errors explicitly instead. This function will be removed in v1.0.0.
func MustSetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) func(context.Context) error {
	return tracer.MustSetupTracing(ctx, serviceName, serviceVersion...)
}

// SetupCustomTracing provides full control over the tracing setup.
// Use this when you need custom configuration that goes beyond environment variables.
//
// Deprecated: Use NewProviderConfig() with NewProvider() for advanced configuration.
// This function will be removed in v1.0.0.
func SetupCustomTracing(ctx context.Context, cfg *provider.ProviderConfig) (*sdktrace.TracerProvider, error) {
	return tracer.SetupCustomTracing(ctx, cfg)
}

// ShutdownTracerProvider gracefully shuts down the tracer provider, ensuring all pending spans
// are exported before the application terminates.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := otelkit.ShutdownTracerProvider(ctx, provider); err != nil {
//	    log.Printf("Error during tracer shutdown: %v", err)
//	}
func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) error {
	return provider.ShutdownTracerProvider(ctx, tp)
}

// NewHttpMiddleware creates HTTP middleware for automatic request tracing.
// This middleware automatically creates spans for HTTP requests and adds
// useful attributes like HTTP method, URL, status code, and user agent.
//
// Example:
//
//	tracer := otelkit.New("web-service")
//	middleware := otelkit.NewHttpMiddleware(tracer)
//	r.Use(middleware.Middleware)
func NewHttpMiddleware(tracer *tracer.Tracer) *middleware.HTTPMiddleware {
	return middleware.NewHttpMiddleware(tracer)
}

// AddAttributes safely adds one or more attributes to the given span.
// If the span is nil, this function is a no-op.
func AddAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	tracer.AddAttributes(span, attrs...)
}

// AddEvent safely adds a named event with optional attributes to the span.
// If the span is nil, this function is a no-op.
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	tracer.AddEvent(span, name, attrs...)
}

// AddTimedEvent adds an event with duration information to the span.
func AddTimedEvent(span trace.Span, name string, duration time.Duration) {
	tracer.AddTimedEvent(span, name, duration)
}

// RecordError safely records an error on the span and sets the span status to error.
// This function handles nil checks for both span and error.
func RecordError(span trace.Span, err error) {
	tracer.RecordError(span, err)
}

// EndSpan safely ends the given span.
// If the span is nil, this function is a no-op.
func EndSpan(span trace.Span) {
	tracer.EndSpan(span)
}

// IsRecording checks if the span is currently recording telemetry data.
// Returns false if the span is nil or if the span context is invalid.
func IsRecording(span trace.Span) bool {
	return tracer.IsRecording(span)
}

// SetGlobalTracerProvider sets the global tracer provider.
// This is typically called automatically by the setup functions,
// but can be called manually if needed.
func SetGlobalTracerProvider(tp *sdktrace.TracerProvider) {
	tracer.SetGlobalTracerProvider(tp)
}

// Span is an alias for the underlying OpenTelemetry span interface.
// This provides a cleaner API surface for users of this package.
type Span = trace.Span

// Tracer is an alias for the tracer wrapper type.
// This provides a cleaner API surface for users of this package.
type Tracer = tracer.Tracer

// HTTPMiddleware is an alias for the HTTP middleware type.
// This provides a cleaner API surface for users of this package.
type HTTPMiddleware = middleware.HTTPMiddleware

// ProviderConfig is an alias for the provider configuration type.
// This provides a cleaner API surface for users of this package.
type ProviderConfig = provider.ProviderConfig

// ConfigError represents a configuration validation error.
type ConfigError = config.ConfigError

// InitializationError represents an error during tracer provider initialization.
type InitializationError = config.InitializationError

// TracedHTTPClient is an alias for the traced HTTP client type.
type TracedHTTPClient = tracer.TracedHTTPClient

// NewTracedHTTPClient creates a new traced HTTP client.
func NewTracedHTTPClient(client *http.Client, tr *Tracer, service string) *TracedHTTPClient {
	return tracer.NewTracedHTTPClient(client, (*tracer.Tracer)(tr), service)
}

// RecordErrorWithCode safely records an error on the span with a custom error code and message.
func RecordErrorWithCode(span trace.Span, err error, code string, message string) {
	tracer.RecordErrorWithCode(span, err, code, message)
}

// ErrorCodeExternalService is a constant for external service errors.
const ErrorCodeExternalService = tracer.ErrorCodeExternalService
