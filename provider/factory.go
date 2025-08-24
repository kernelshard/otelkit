// Package provider provides OpenTelemetry tracer provider configuration and initialization.
package provider

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/samims/otelkit/internal/config"
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

// newProvider creates a new tracer provider based on the provided configuration.
func newProvider(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error) {
	// Use provided resource or create a new one
	var res *sdkresource.Resource
	var err error
	if cfg.Resource != nil {
		res = cfg.Resource
	} else {
		res, err = sdkresource.New(ctx,
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
	}

	// Create exporter
	var exporter sdktrace.SpanExporter
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

	// Create sampler
	sampler := createSampler(cfg.Config)

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

	// Batch span processor with configurable options
	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(cfg.BatchTimeout),
		sdktrace.WithExportTimeout(cfg.ExportTimeout),
		sdktrace.WithMaxExportBatchSize(cfg.MaxExportBatchSize),
		sdktrace.WithMaxQueueSize(cfg.MaxQueueSize),
	)

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

// createSampler creates a sampler instance based on the provided configuration.
func createSampler(cfg *config.Config) sdktrace.Sampler {
	var sampler sdktrace.Sampler

	switch cfg.SamplingType {
	case "probabilistic":
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SamplingRatio))
	case "always_on":
		sampler = sdktrace.AlwaysSample()
	case "always_off":
		sampler = sdktrace.NeverSample()
	default:
		// Fallback to probabilistic sampling
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(config.DefaultSamplingRatio))
	}
	return sampler
}
