/*
Package provider implements the factory and builder patterns for creating and configuring
OpenTelemetry tracer providers. It encapsulates the complexity of resource creation,
exporter setup, batch processing, and sampler configuration.

The package provides a clean, modular API for constructing tracer providers with
sensible defaults and extensible configuration options.

Key components:
- InitializationError: Custom error type for initialization failures
- createResource: Creates or returns an OpenTelemetry resource for service identification
- createExporter: Factory method for OTLP exporters (HTTP or gRPC)
- createBatchProcessor: Configures batch span processor with performance tuning options
- newProvider: Orchestrates creation of the tracer provider from components
- createSampler: Strategy pattern for sampler selection based on config

Usage example:

	ctx := context.Background()
	cfg := NewProviderConfig("my-service", "v1.0.0")
	provider, err := newProvider(ctx, cfg)
	if err != nil {
	    log.Fatal(err)
	}

Design patterns:
- Factory Method: createExporter, createResource, createBatchProcessor
- Builder: ProviderConfig fluent API
- Strategy: createSampler selects sampling strategy

This package follows SOLID principles and Go best practices for maintainable,
testable, and extensible code.
*/
package provider

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/kernelshard/otelkit/internal/config"
)

// InitializationError represents an error during provider initialization
type InitializationError struct {
	Component string
	Cause     error
}

// Error implements the error interface
func (e *InitializationError) Error() string {
	return e.Component + ": " + e.Cause.Error()
}

// Unwrap returns the underlying error
func (e *InitializationError) Unwrap() error {
	return e.Cause
}

// createResource creates an OpenTelemetry resource based on the provided configuration or returns the existing one.
func createResource(ctx context.Context, cfg *ProviderConfig) (*sdkresource.Resource, error) {
	if cfg.Resource != nil {
		return cfg.Resource, nil
	}
	res, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(
			semconv.ServiceName(cfg.Config.ServiceName),
			semconv.ServiceVersion(cfg.Config.ServiceVersion),
			semconv.DeploymentEnvironment(cfg.Config.Environment),
			semconv.HostName(cfg.Config.Hostname),
			semconv.ServiceInstanceID(cfg.Config.InstanceID),
		),
		sdkresource.WithContainer(),
		sdkresource.WithHost(),
		sdkresource.WithOSType(),
	)
	if err != nil {
		return nil, &InitializationError{Component: "resource", Cause: err}
	}
	return res, nil
}

// createExporter creates an OTLP exporter based on the configuration.
func createExporter(ctx context.Context, cfg *ProviderConfig) (sdktrace.SpanExporter, error) {
	var exporter sdktrace.SpanExporter
	var err error
	switch cfg.Config.OTLPExporterProtocol {
	case "http":
		exporter, err = createHTTPExporter(ctx, cfg.Config)
	case "grpc":
		exporter, err = createGRPCExporter(ctx, cfg.Config)
	default:
		return nil, &config.ConfigError{Field: "OTLPExporterProtocol", Message: config.ErrInvalidExporterProtocol}
	}
	if err != nil {
		return nil, &InitializationError{Component: "exporter", Cause: err}
	}
	return exporter, nil
}

// createBatchProcessor creates a batch span processor with the given exporter and configuration.
func createBatchProcessor(exporter sdktrace.SpanExporter, cfg *ProviderConfig) sdktrace.SpanProcessor {
	// Set defaults for batch processor options if not provided
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = config.DefaultBatchTimeout
	}
	if cfg.ExportTimeout == 0 {
		cfg.ExportTimeout = config.DefaultExportTimeout
	}
	if cfg.MaxExportBatchSize == 0 {
		cfg.MaxExportBatchSize = config.DefaultMaxExportBatchSize
	}
	if cfg.MaxQueueSize == 0 {
		cfg.MaxQueueSize = config.DefaultMaxQueueSize
	}

	return sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		sdktrace.WithExportTimeout(cfg.ExportTimeout),
		sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
	)
}

// newProvider creates a new tracer provider based on the provided configuration.
func newProvider(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error) {
	res, err := createResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	exporter, err := createExporter(ctx, cfg)
	if err != nil {
		return nil, err
	}

	sampler := createSampler(cfg.Config)

	bsp := createBatchProcessor(exporter, cfg)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(bsp),
	)

	return tp, nil
}

// createGRPCExporter creates an OTLP gRPC exporter configured with the provided settings.
func createGRPCExporter(ctx context.Context, cfg *config.Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPExporterEndpoint),
	}
	if cfg.OTLPExporterInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPExporter creates an OTLP HTTP exporter configured with the provided settings.
func createHTTPExporter(ctx context.Context, cfg *config.Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLPExporterEndpoint),
	}
	if cfg.OTLPExporterInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	return otlptracehttp.New(ctx, opts...)
}

// SamplerFactory defines the interface for creating samplers.
// This allows for extensible sampler creation without modifying existing code.
type SamplerFactory interface {
	CreateSampler(cfg *config.Config) sdktrace.Sampler
}

// ProbabilisticSamplerFactory creates probabilistic samplers
type ProbabilisticSamplerFactory struct{}

func (f *ProbabilisticSamplerFactory) CreateSampler(cfg *config.Config) sdktrace.Sampler {
	return sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRatio))
}

// AlwaysOnSamplerFactory creates always-on samplers
type AlwaysOnSamplerFactory struct{}

func (f *AlwaysOnSamplerFactory) CreateSampler(cfg *config.Config) sdktrace.Sampler {
	return sdktrace.AlwaysSample()
}

// AlwaysOffSamplerFactory creates always-off samplers
type AlwaysOffSamplerFactory struct{}

func (f *AlwaysOffSamplerFactory) CreateSampler(cfg *config.Config) sdktrace.Sampler {
	return sdktrace.NeverSample()
}

// samplerFactories holds the mapping of sampler types to their factories
var samplerFactories = map[config.SamplingType]SamplerFactory{
	config.SamplingProbabilistic: &ProbabilisticSamplerFactory{},
	config.SamplingAlwaysOn:      &AlwaysOnSamplerFactory{},
	config.SamplingAlwaysOff:     &AlwaysOffSamplerFactory{},
}

// createSampler creates a sampler instance based on the provided configuration.
// This implementation follows the Open-Closed Principle by using the Strategy pattern
// with factory interfaces, allowing new sampler types to be added without modifying
// this function.
func createSampler(cfg *config.Config) sdktrace.Sampler {
	if factory, exists := samplerFactories[cfg.SamplingType]; exists {
		return factory.CreateSampler(cfg)
	}

	// Fallback to probabilistic sampling for unknown types
	return samplerFactories["probabilistic"].CreateSampler(cfg)
}
