// Package otelkit provides OpenTelemetry tracer provider configuration and initialization.
// This file contains the core provider setup that configures exporters, sampling,
// resource identification, and batch processing for the entire tracing system.
package otelkit

import (
	"context"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// setOnce ensures the global tracer provider is set only once across all
// provider creation calls, preventing conflicts in multi-initialization scenarios.
var setOnce sync.Once

// ProviderConfig holds comprehensive configuration for creating a TracerProvider.
// It combines basic tracing configuration with advanced options for batch processing,
// resource identification, and performance tuning. This allows fine-grained control
// over the tracing pipeline behavior.
//
// The configuration supports fluent method chaining for ease of use:
//
//	config := tracer.NewProviderConfig("my-service", "v1.0.0").
//	    WithOTLPExporter("localhost:4317", "grpc", true).
//	    WithSampling("probabilistic", 0.1).
//	    WithBatchOptions(5*time.Second, 30*time.Second, 512, 2048)
type ProviderConfig struct {
	// Config contains the core tracing configuration including service identification,
	// exporter settings, and sampling strategy.
	Config *Config

	// Resource provides custom resource attributes for service identification.
	// If nil, a default resource will be created using service name, version,
	// environment, hostname, and instance ID from Config.
	Resource *sdkresource.Resource

	// BatchTimeout is the maximum time the batch processor waits before
	// exporting spans. Lower values reduce latency but may increase overhead.
	// Default: 5 seconds.
	BatchTimeout time.Duration

	// ExportTimeout is the maximum time allowed for exporting a batch of spans.
	// Exports exceeding this timeout will be cancelled. Default: 30 seconds.
	ExportTimeout time.Duration

	// MaxExportBatchSize is the maximum number of spans to export in a single batch.
	// Larger batches improve throughput but use more memory. Default: 512.
	MaxExportBatchSize int

	// MaxQueueSize is the maximum number of spans that can be queued for export.
	// When the queue is full, new spans will be dropped. Default: 2048.
	MaxQueueSize int
}

// NewProviderConfig creates a new ProviderConfig with sensible defaults for advanced configuration.
// It initializes the configuration with default batch processing settings and creates a base Config
// using the provided service name and version. The returned config supports fluent method chaining
// for customization.
//
// Default settings:
//   - BatchTimeout: 5 seconds
//   - ExportTimeout: 30 seconds
//   - MaxExportBatchSize: 512 spans
//   - MaxQueueSize: 2048 spans
//   - OTLP HTTP exporter pointing to localhost:4318
//   - Probabilistic sampling at 20%
//
// Example:
//
//	config := tracer.NewProviderConfig("user-service", "v2.1.0")
//	provider, err := tracer.NewProvider(ctx, config)
func NewProviderConfig(serviceName, serviceVersion string) *ProviderConfig {
	return &ProviderConfig{
		Config:             NewConfig(serviceName, serviceVersion),
		BatchTimeout:       DefaultBatchTimeout,
		ExportTimeout:      DefaultExportTimeout,
		MaxExportBatchSize: DefaultMaxExportBatchSize,
		MaxQueueSize:       DefaultMaxQueueSize,
	}
}

// WithOTLPExporter configures the OTLP exporter settings for trace export.
// This method allows you to specify the endpoint, protocol, and security settings
// for sending traces to an OTLP-compatible backend.
//
// Parameters:
//   - endpoint: The URL or address of the OTLP collector (e.g., "localhost:4317", "https://api.honeycomb.io")
//   - protocol: Either "grpc" for gRPC transport or "http" for HTTP transport
//   - insecure: true to disable TLS (for development), false to use TLS (for production)
//
// Example:
//
//	config.WithOTLPExporter("https://api.honeycomb.io", "http", false)
//	config.WithOTLPExporter("localhost:4317", "grpc", true)  // Development
func (pc *ProviderConfig) WithOTLPExporter(endpoint, protocol string, insecure bool) *ProviderConfig {
	pc.Config.OTLPExporterEndpoint = endpoint
	pc.Config.OTLPExporterProtocol = protocol
	pc.Config.OTLPExporterInsecure = insecure
	return pc
}

// WithSampling configures the sampling strategy and ratio for trace collection.
// Sampling controls what percentage of traces are collected and exported, which is crucial
// for managing overhead in high-traffic applications.
//
// Parameters:
//   - samplingType: "probabilistic" (ratio-based), "always_on" (100%), or "always_off" (0%)
//   - ratio: For probabilistic sampling, the ratio of traces to sample (0.0 to 1.0)
//     Ignored for "always_on" and "always_off" strategies
//
// Example:
//
//	config.WithSampling("probabilistic", 0.01)  // 1% sampling for production
//	config.WithSampling("always_on", 0)        // 100% sampling for development
//	config.WithSampling("always_off", 0)       // Disable tracing
func (pc *ProviderConfig) WithSampling(samplingType string, ratio float64) *ProviderConfig {
	pc.Config.SamplingType = samplingType
	pc.Config.SamplingRatio = ratio
	return pc
}

// WithBatchOptions configures the batch processor settings for span export optimization.
// These settings control how spans are batched and exported, affecting both performance
// and resource usage. Tune these values based on your application's traffic patterns
// and latency requirements.
//
// Parameters:
//   - batchTimeout: Maximum time to wait before exporting (lower = less latency, higher = better throughput)
//   - exportTimeout: Maximum time allowed for export operations (prevents hanging exports)
//   - maxExportBatchSize: Maximum spans per batch (higher = better throughput, more memory usage)
//   - maxQueueSize: Maximum queued spans before dropping (higher = more memory, less data loss)
//
// Example:
//
//	// Low-latency configuration
//	config.WithBatchOptions(1*time.Second, 10*time.Second, 256, 1024)
//
//	// High-throughput configuration
//	config.WithBatchOptions(10*time.Second, 60*time.Second, 1024, 4096)
func (pc *ProviderConfig) WithBatchOptions(batchTimeout, exportTimeout time.Duration, maxExportBatchSize, maxQueueSize int) *ProviderConfig {
	pc.BatchTimeout = batchTimeout
	pc.ExportTimeout = exportTimeout
	pc.MaxExportBatchSize = maxExportBatchSize
	pc.MaxQueueSize = maxQueueSize
	return pc
}

// WithResource sets a custom OpenTelemetry resource for service identification.
// Resources contain attributes that identify the service, version, environment,
// and other metadata. If not provided, a default resource will be created automatically
// using the service name, version, and other attributes from the Config.
//
// Example:
//
//	resource, _ := resource.New(ctx,
//	    resource.WithAttributes(
//	        semconv.ServiceName("payment-service"),
//	        semconv.ServiceVersion("v1.2.3"),
//	        semconv.DeploymentEnvironment("production"),
//	        attribute.String("region", "us-west-2"),
//	    ),
//	)
//	config.WithResource(resource)
func (pc *ProviderConfig) WithResource(resource *sdkresource.Resource) *ProviderConfig {
	pc.Resource = resource
	return pc
}

// newDefaultProvider creates a tracer provider with opinionated defaults for quick setup.
// This is an internal function that provides sensible defaults for development and testing.
// It configures an HTTP OTLP exporter pointing to localhost:4318 with insecure connections,
// probabilistic sampling at the default rate, and standard batch processing settings.
//
// For production use or advanced configuration, use NewProvider with NewProviderConfig instead.
func newDefaultProvider(ctx context.Context, serviceName string, serviceVersion ...string) (*sdktrace.TracerProvider, error) {
	// Set default service version if not provided
	ver := DefaultServiceVersion
	if len(serviceVersion) > 0 {
		ver = serviceVersion[0]
	}

	// Create configuration with defaults
	cfg := NewConfig(serviceName, ver)
	cfg.OTLPExporterProtocol = "http"
	cfg.OTLPExporterEndpoint = "localhost:4318"
	cfg.OTLPExporterInsecure = true
	cfg.SamplingType = DefaultSamplingType
	cfg.SamplingRatio = DefaultSamplingRatio
	cfg.Environment = DefaultEnvironment

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, &InitializationError{Component: "default provider config", Cause: err}
	}

	// Create provider config with defaults
	providerCfg := &ProviderConfig{
		Config:             cfg,
		BatchTimeout:       DefaultBatchTimeout,
		ExportTimeout:      DefaultExportTimeout,
		MaxExportBatchSize: DefaultMaxExportBatchSize,
		MaxQueueSize:       DefaultMaxQueueSize,
	}

	return NewProvider(ctx, providerCfg)
}

// NewDefaultProvider creates a tracer provider with default settings and sets it as the global provider.
// This is a convenience function for quick setup in development or simple applications.
// It creates a provider with opinionated defaults:
//   - HTTP OTLP exporter to localhost:4318 (insecure)
//   - Probabilistic sampling at the default rate (typically 20%)
//   - Standard batch processing settings
//   - Automatic resource detection for service identification
//
// The provider is set as the global OpenTelemetry provider (only once per application).
// For production use or when you need custom configuration, use NewProvider with NewProviderConfig.
//
// Example:
//
//	provider, err := tracer.NewDefaultProvider(ctx, "my-service", "v1.0.0")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
func NewDefaultProvider(ctx context.Context, serviceName string, serviceVersion ...string) (*sdktrace.TracerProvider, error) {
	tp, err := newDefaultProvider(ctx, serviceName, serviceVersion...)
	if err != nil {
		return nil, err
	}
	setOnce.Do(func() {
		otel.SetTracerProvider(tp)
	})
	return tp, nil
}

// newProvider creates a new tracer provider based on the provided configuration.
// This is the core implementation that handles the complete provider setup including
// resource creation, exporter initialization, sampler configuration, and batch processor setup.
//
// The function performs the following operations:
//  1. Creates or uses the provided resource with service identification attributes
//  2. Initializes the appropriate exporter (HTTP or gRPC OTLP)
//  3. Configures the sampler based on the sampling strategy
//  4. Sets up the batch span processor with the specified options
//  5. Creates and returns the configured TracerProvider
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
		return nil, &ConfigError{Field: "OTLPExporterProtocol", Message: "must be 'grpc' or 'http'"}
	}
	if err != nil {
		return nil, &InitializationError{Component: "exporter", Cause: err}
	}

	// Create sampler
	sampler := createSampler(cfg.Config)

	// Set defaults for batch processor options if not provided
	if cfg.BatchTimeout == 0 {
		cfg.BatchTimeout = DefaultBatchTimeout
	}
	if cfg.ExportTimeout == 0 {
		cfg.ExportTimeout = DefaultExportTimeout
	}
	if cfg.MaxExportBatchSize == 0 {
		cfg.MaxExportBatchSize = DefaultMaxExportBatchSize
	}
	if cfg.MaxQueueSize == 0 {
		cfg.MaxQueueSize = DefaultMaxQueueSize
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

// NewProvider creates and configures a new TracerProvider using the provided configuration,
// then sets it as the global OpenTelemetry provider (only once per application lifecycle).
// This is the recommended way to initialize tracing when you need custom configuration.
//
// The function ensures that the global provider is set only once, even if called multiple times.
// This prevents conflicts and ensures consistent tracing behavior across the application.
//
// Example:
//
//	config := tracer.NewProviderConfig("payment-service", "v1.2.3").
//	    WithOTLPExporter("https://api.honeycomb.io", "http", false).
//	    WithSampling("probabilistic", 0.05)
//
//	provider, err := tracer.NewProvider(ctx, config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
func NewProvider(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error) {
	tp, err := newProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	setOnce.Do(func() {
		otel.SetTracerProvider(tp)
	})

	return tp, nil
}

// createGRPCExporter creates an OTLP gRPC exporter configured with the provided settings.
// gRPC exporters are typically more efficient than HTTP for high-throughput scenarios
// but require gRPC infrastructure. The exporter handles the binary OTLP protocol over gRPC.
func createGRPCExporter(ctx context.Context, cfg *Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPExporterEndpoint),
	}
	if cfg.OTLPExporterInsecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}
	return otlptracegrpc.New(ctx, opts...)
}

// createHTTPExporter creates an OTLP HTTP exporter configured with the provided settings.
// HTTP exporters are more widely compatible and easier to set up than gRPC, making them
// suitable for most use cases. The exporter uses HTTP/2 when available for better performance.
func createHTTPExporter(ctx context.Context, cfg *Config) (sdktrace.SpanExporter, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.OTLPExporterEndpoint),
	}
	if cfg.OTLPExporterInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}
	return otlptracehttp.New(ctx, opts...)
}

// createSampler creates a sampler instance based on the provided configuration.
// The sampler determines which traces should be collected and exported, which is crucial
// for controlling overhead in high-traffic applications.
//
// Supported sampling strategies:
//   - "probabilistic": Samples traces based on a probability ratio (0.0 to 1.0)
//   - "always_on": Samples all traces (100% sampling)
//   - "always_off": Samples no traces (0% sampling)
//
// The probabilistic sampler uses parent-based sampling, which means if a parent
// span was sampled, all child spans will also be sampled to maintain trace integrity.
func createSampler(cfg *Config) sdktrace.Sampler {
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
		sampler = sdktrace.ParentBased(sdktrace.TraceIDRatioBased(DefaultSamplingRatio))
	}
	return sampler
}

// ShutdownTracerProvider gracefully shuts down the tracer provider, ensuring all pending spans
// are exported before the application terminates. This function should be called during
// application shutdown, typically with a context that has a reasonable timeout.
//
// The shutdown process:
//  1. Stops accepting new spans
//  2. Exports all remaining spans in the queue
//  3. Closes the exporter connection
//  4. Releases any resources held by the provider
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	if err := tracer.ShutdownTracerProvider(ctx, provider); err != nil {
//	    log.Printf("Error during tracer shutdown: %v", err)
//	}
func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) error {
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}
