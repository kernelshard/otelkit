package tracer

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/samims/otelkit/internal/config"
	"github.com/samims/otelkit/provider"
)

// SetupTracing initializes OpenTelemetry tracing with sensible defaults.
// This is the simplest way to get started with tracing.
//
// Example:
//
//	shutdown, err := tracer.SetupTracing(ctx, "my-service")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer shutdown(ctx)
//
// The function reads configuration from environment variables and sets up:
// - OTLP exporter (HTTP by default, localhost:4318)
// - Probabilistic sampling (20% by default)
// - Batch span processor with sensible defaults
// - Resource with service information
func SetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) (func(context.Context) error, error) {
	// Use provided version or default
	version := "1.0.0"
	if len(serviceVersion) > 0 {
		version = serviceVersion[0]
	}

	// Create configuration from environment variables
	cfg := config.NewConfigFromEnv()
	cfg.ServiceName = serviceName
	cfg.ServiceVersion = version

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, &config.InitializationError{Component: "configuration", Cause: err}
	}

	// Create provider configuration
	providerCfg := &provider.ProviderConfig{
		Config:             cfg,
		BatchTimeout:       config.DefaultBatchTimeout,
		ExportTimeout:      config.DefaultExportTimeout,
		MaxExportBatchSize: config.DefaultMaxExportBatchSize,
		MaxQueueSize:       config.DefaultMaxQueueSize,
	}

	// Create and set the tracer provider
	tp, err := provider.NewProvider(ctx, providerCfg)
	if err != nil {
		return nil, err
	}

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		return provider.ShutdownTracerProvider(ctx, tp)
	}

	return shutdown, nil
}

// SetupTracingWithDefaults initializes tracing with hardcoded defaults.
// This is useful for quick setup without environment variables.
//
// It uses:
// - HTTP OTLP exporter to localhost:4318 (insecure)
// - Probabilistic sampling at 20%
// - Standard batch processing settings
func SetupTracingWithDefaults(ctx context.Context, serviceName, serviceVersion string) (func(context.Context) error, error) {
	tp, err := provider.NewDefaultProvider(ctx, serviceName, serviceVersion)
	if err != nil {
		return nil, err
	}

	// Return shutdown function
	shutdown := func(ctx context.Context) error {
		return provider.ShutdownTracerProvider(ctx, tp)
	}

	return shutdown, nil
}

// MustSetupTracing is like SetupTracing but panics on error.
// Use this for simple programs where you want to fail fast.
func MustSetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) func(context.Context) error {
	shutdown, err := SetupTracing(ctx, serviceName, serviceVersion...)
	if err != nil {
		panic(err)
	}
	return shutdown
}

// SetupCustomTracing provides full control over the tracing setup.
// Use this when you need custom configuration that goes beyond environment variables.
func SetupCustomTracing(ctx context.Context, cfg *provider.ProviderConfig) (*sdktrace.TracerProvider, error) {
	// Validate configuration
	if cfg == nil {
		return nil, &config.ConfigError{Field: "config", Message: "provider config cannot be nil"}
	}
	if cfg.Config == nil {
		return nil, &config.ConfigError{Field: "config.Config", Message: "tracer config cannot be nil"}
	}

	if err := cfg.Config.Validate(); err != nil {
		return nil, &config.InitializationError{Component: "configuration", Cause: err}
	}

	return provider.NewProvider(ctx, cfg)
}
