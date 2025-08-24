package otelkit // import "github.com/samims/otelkit"

Package otelkit provides HTTP middleware for automatic request tracing.
The middleware integrates seamlessly with any HTTP framework that supports
the standard http.Handler interface, including gorilla/mux, chi, gin, echo,
and others.

Package otelkit provides OpenTelemetry tracer provider configuration and
initialization. This file contains the core provider setup that configures
exporters, sampling, resource identification, and batch processing for the
entire tracing system.

Package tracer provides span utility functions for OpenTelemetry tracing.
These utilities offer safe, convenient methods for common span operations with
built-in nil checks and error handling.

CONSTANTS

const (
	DefaultServiceName          = "unknown-service"
	DefaultServiceVersion       = "1.0.0"
	DefaultEnvironment          = "development"
	DefaultOTLPExporterEndpoint = "localhost:4317"
	DefaultSamplingRatio        = 0.2
	DefaultSamplingType         = "probabilistic"
	DefaultOTLPExporterProtocol = "grpc"
	DefaultBatchTimeout         = 5 * time.Second
	DefaultExportTimeout        = 30 * time.Second
	DefaultMaxExportBatchSize   = 512
	DefaultMaxQueueSize         = 2048
)
    Service configuration constants

const (
	AttrHTTPMethod     = "http.method"
	AttrHTTPURL        = "http.url"
	AttrHTTPUserAgent  = "http.user_agent"
	AttrHTTPStatusCode = "http.status_code"
)
    OpenTelemetry semantic convention constants

const (
	ErrServiceNameRequired     = "service name is required"
	ErrServiceVersionRequired  = "service version is required"
	ErrInvalidEnvironment      = "invalid environment"
	ErrInvalidSamplingType     = "invalid sampling type"
	ErrInvalidSamplingRatio    = "sampling ratio must be between 0 and 1"
	ErrInvalidExporterProtocol = "invalid exporter protocol"
	ErrInvalidExporterEndpoint = "exporter endpoint is required"
)
    Error message constants

const (
	EnvServiceName          = "OTEL_SERVICE_NAME"
	EnvServiceVersion       = "OTEL_SERVICE_VERSION"
	EnvEnvironment          = "OTEL_ENVIRONMENT"
	EnvOTLPExporterEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	EnvOTLPExporterInsecure = "OTEL_EXPORTER_OTLP_INSECURE"
	EnvOTLPExporterProtocol = "OTEL_EXPORTER_OTLP_PROTOCOL"
	EnvBatchTimeout         = "OTEL_BSP_TIMEOUT"
	EnvExportTimeout        = "OTEL_EXPORTER_TIMEOUT"
	EnvMaxExportBatchSize   = "OTEL_BSP_MAX_EXPORT_BATCH_SIZE"
	EnvMaxQueueSize         = "OTEL_BSP_MAX_QUEUE_SIZE"
	EnvSamplingType         = "OTEL_TRACES_SAMPLER"
	EnvSamplingRatio        = "OTEL_TRACES_SAMPLER_ARG"
	EnvInstanceID           = "OTEL_RESOURCE_ATTRIBUTES_SERVICE_INSTANCE_ID"
)
    Environment variable constants


VARIABLES

var (
	ValidEnvironments  = []string{"development", "staging", "production"}
	ValidSamplingTypes = []string{"probabilistic", "always_on", "always_off"}
	ValidHTTPMethods   = []string{
		http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions,
	}
	ValidOTLPProtocols = []string{"grpc", "http"}
)
    Valid configuration options


FUNCTIONS

func AddAttributes(span trace.Span, attrs ...attribute.KeyValue)
    AddAttributes safely adds one or more attributes to the given span. If the
    span is nil, this function is a no-op. This is useful for adding contextual
    information to spans such as user IDs, request parameters, or business logic
    details.

    Example:

        AddAttributes(span,
            attribute.String("user.id", "12345"),
            attribute.Int("request.size", 1024),
            attribute.Bool("cache.hit", true),
        )

func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue)
    AddEvent safely adds a named event with optional attributes to the span.
    Events are timestamped markers that can help understand the flow of
    execution. If the span is nil, this function is a no-op.

    Example:

        AddEvent(span, "cache.miss",
            attribute.String("key", cacheKey),
            attribute.String("reason", "expired"),
        )

func AddTimedEvent(span trace.Span, name string, duration time.Duration)
    AddTimedEvent adds an event with duration information to the span. This is
    useful for recording the time taken for specific operations within a larger
    span. The duration is added as a string attribute.

    Example:

        start := time.Now()
        // ... perform operation
        AddTimedEvent(span, "database.query", time.Since(start))

func EndSpan(span trace.Span)
    EndSpan safely ends the given span. If the span is nil, this function is
    a no-op. This provides a safe way to end spans without worrying about nil
    checks.

    Example:

        defer EndSpan(span)

func ExtractTraceContext(req *http.Request) context.Context
    ExtractTraceContext extracts trace context from HTTP request headers into
    the context.

func InjectTraceContext(ctx context.Context, req *http.Request)
    InjectTraceContext injects the current trace context into the HTTP request
    headers.

func InjectTraceIDIntoContext(ctx context.Context, span trace.Span) context.Context
    InjectTraceIDIntoContext adds trace ID into the context (as a new value).

func IsRecording(span trace.Span) bool
    IsRecording checks if the span is currently recording telemetry data.
    Returns false if the span is nil or if the span context is invalid. This
    can be used to avoid expensive operations when tracing is disabled or when
    working with noop spans.

    Example:

        if IsRecording(span) {
            // Perform expensive attribute computation
            span.SetAttributes(expensiveAttributes()...)
        }

func MustSetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) func(context.Context) error
    MustSetupTracing is like SetupTracing but panics on error. Use this for
    simple programs where you want to fail fast.

func NewConfigError(field, message string) error
    NewConfigError creates a new ConfigError.

func NewDefaultProvider(ctx context.Context, serviceName string, serviceVersion ...string) (*sdktrace.TracerProvider, error)
    NewDefaultProvider creates a tracer provider with default settings and sets
    it as the global provider. This is a convenience function for quick setup in
    development or simple applications. It creates a provider with opinionated
    defaults:
      - HTTP OTLP exporter to localhost:4318 (insecure)
      - Probabilistic sampling at the default rate (typically 20%)
      - Standard batch processing settings
      - Automatic resource detection for service identification

    The provider is set as the global OpenTelemetry provider (only once per
    application). For production use or when you need custom configuration,
    use NewProvider with NewProviderConfig.

    Note: This is the function most users will start with. It's designed to
    "just work" for local development and testing scenarios.

    Example:

        provider, err := tracer.NewDefaultProvider(ctx, "my-service", "v1.0.0")
        if err != nil {
            log.Fatal(err)
        }
        defer provider.Shutdown(ctx)

func NewInitializationError(component string, cause error) error
    NewInitializationError creates a new InitializationError.

func NewInstrumentedGRPCClientDialOptions() []grpc.DialOption
    NewInstrumentedGRPCClientDialOptions returns grpc.DialOption slice with
    OpenTelemetry instrumentation for client connections. Use this in grpc.Dial
    for instrumented client connections.

func NewInstrumentedGRPCServer(opts ...grpc.ServerOption) *grpc.Server
    NewInstrumentedGRPCServer creates a new gRPC server with OpenTelemetry unary
    and stream interceptors attached automatically.

func NewInstrumentedHTTPClient(baseTransport http.RoundTripper) *http.Client
    NewInstrumentedHTTPClient returns an *http.Client with OpenTelemetry
    transport for automatic HTTP tracing. You can customize it by passing a base
    transport; if nil, http.DefaultTransport is used.

func NewInstrumentedHTTPHandler(handler http.Handler, operationName string) http.Handler
    NewInstrumentedHTTPHandler wraps an http.Handler with OpenTelemetry
    instrumentation and returns the wrapped handler. Usage: http.Handle("/path",
    NewInstrumentedHTTPHandler(yourHandler, "operationName"))

func NewPropagationError(operation string, cause error) error
    NewPropagationError creates a new PropagationError.

func NewProvider(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error)
    NewProvider creates and configures a new TracerProvider using the provided
    configuration, then sets it as the global OpenTelemetry provider (only
    once per application lifecycle). This is the recommended way to initialize
    tracing when you need custom configuration.

    The function ensures that the global provider is set only once, even if
    called multiple times. This prevents conflicts and ensures consistent
    tracing behavior across the application.

    Example:

        config := tracer.NewProviderConfig("payment-service", "v1.2.3").
            WithOTLPExporter("https://api.honeycomb.io", "http", false).
            WithSampling("probabilistic", 0.05)

        provider, err := tracer.NewProvider(ctx, config)
        if err != nil {
            log.Fatal(err)
        }
        defer provider.Shutdown(ctx)

func RecordError(span trace.Span, err error)
    RecordError safely records an error on the span and sets the span status
    to error. This function handles nil checks for both span and error. When
    an error is recorded, the span status is automatically set to codes.Error
    with the error message. This is essential for proper error tracking in
    distributed tracing.

    Example:

        if err := doSomething(); err != nil {
            RecordError(span, err)
            return err
        }

func SetGlobalTracerProvider(tp trace.TracerProvider)
    SetGlobalTracerProvider sets the global OpenTelemetry tracer provider.
    This should typically be called once during application initialization.
    All subsequent tracer instances will use this provider.

    Example:

        provider := setupTracerProvider()
        tracer.SetGlobalTracerProvider(provider)

func SetupCustomTracing(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error)
    SetupCustomTracing provides full control over the tracing setup. Use this
    when you need custom configuration that goes beyond environment variables.

func SetupTracing(ctx context.Context, serviceName string, serviceVersion ...string) (func(context.Context) error, error)
    SetupTracing initializes OpenTelemetry tracing with sensible defaults.
    This is the simplest way to get started with tracing.

    Example:

        shutdown, err := tracer.SetupTracing(ctx, "my-service")
        if err != nil {
            log.Fatal(err)
        }
        defer shutdown(ctx)

    The function reads configuration from environment variables and sets up:
    - OTLP exporter (HTTP by default, localhost:4318) - Probabilistic sampling
    (20% by default) - Batch span processor with sensible defaults - Resource
    with service information

func SetupTracingWithDefaults(ctx context.Context, serviceName, serviceVersion string) (func(context.Context) error, error)
    SetupTracingWithDefaults initializes tracing with hardcoded defaults.
    This is useful for quick setup without environment variables.

    It uses: - HTTP OTLP exporter to localhost:4318 (insecure) - Probabilistic
    sampling at 20% - Standard batch processing settings

func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) error
    ShutdownTracerProvider gracefully shuts down the tracer provider,
    ensuring all pending spans are exported before the application terminates.
    This function should be called during application shutdown, typically with a
    context that has a reasonable timeout.

    The shutdown process:
     1. Stops accepting new spans
     2. Exports all remaining spans in the queue
     3. Closes the exporter connection
     4. Releases any resources held by the provider

    Example:

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        if err := tracer.ShutdownTracerProvider(ctx, provider); err != nil {
            log.Printf("Error during tracer shutdown: %v", err)
        }

func TraceIDFromContext(ctx context.Context) string
    TraceIDFromContext tries to retrieve trace ID from context.


TYPES

type Config struct {
	// Service identification metadata
	ServiceName    string // Name of the service (required)
	ServiceVersion string // Version of the service (required)
	Environment    string // Deployment environment (development/staging/production)

	// OTLP exporter settings
	OTLPExporterEndpoint string // Collector endpoint (host:port)
	OTLPExporterInsecure bool   // Disable TLS verification
	OTLPExporterProtocol string // Protocol for OTLP exporter (default: grpc)

	// Batch processing configuration
	BatchTimeout       time.Duration // Timeout for batch processing (default: 5s)
	ExportTimeout      time.Duration // Timeout for export requests (default: 30s)
	MaxExportBatchSize int           // Maximum batch size for exports (default: 512)
	MaxQueueSize       int           // Maximum queue size for spans (default: 2048)

	// Sampling configuration
	SamplingRatio float64 // Ratio of traces to sample (0.0 - 1.0)
	SamplingType  string  // Sampling strategy

	// Resource attributes
	InstanceID string // Unique instance identifier
	Hostname   string // Host machine name
}
    Config defines tracing configuration parameters

func NewConfig(serviceName, serviceVersion string) *Config
    NewConfig creates a configuration with sensible defaults

func NewConfigFromEnv() *Config
    NewConfigFromEnv creates configuration from environment variables

func (c *Config) Validate() error
    Validate ensures configuration parameters are correct

func (c *Config) WithEnvironment(env string) *Config
    WithEnvironment sets the deployment environment

func (c *Config) WithOTLPExporter(endpoint string, insecure bool, protocol string) *Config
    WithOTLPExporter configures the OTLP exporter (endpoint, insecure mode,
    and protocol)

func (c *Config) WithSampling(samplingType string, ratio float64) *Config
    WithSampling configures the sampling strategy

type ConfigError struct {
	Field   string
	Message string
}
    ConfigError represents a validation error in configuration.

func (e *ConfigError) Error() string
    Error returns a string representation of the error.

type HTTPMiddleware struct {
	// Has unexported fields.
}
    HTTPMiddleware provides HTTP middleware for automatic request tracing.
    It extracts trace context from incoming requests, creates server spans,
    and automatically records HTTP-specific attributes like method, URL, status
    code, and user agent. The middleware handles trace context propagation
    according to W3C Trace Context and B3 propagation standards.

    The middleware is compatible with any HTTP framework that uses the standard
    http.Handler interface.

func NewHttpMiddleware(tracer *Tracer) *HTTPMiddleware
    NewHttpMiddleware creates a new HTTPMiddleware instance using the provided
    Tracer. The tracer will be used to create spans for all incoming HTTP
    requests.

    Example:

        tracer := otelkit.New("http-service")
        middleware := tracer.NewHttpMiddleware(tracer)

        // With gorilla/mux
        r := mux.NewRouter()
        r.Use(middleware.Middleware)

        // With chi
        r := chi.NewRouter()
        r.Use(middleware.Middleware)

        // With standard http.ServeMux
        mux := http.NewServeMux()
        handler := middleware.Middleware(mux)

func (m *HTTPMiddleware) Middleware(next http.Handler) http.Handler
    Middleware returns an HTTP handler middleware function that automatically
    traces incoming requests.

    The middleware performs the following operations:
     1. Extracts trace context from incoming request headers (supports W3C Trace
        Context and B3)
     2. Creates a new server span with operation name "METHOD /path"
     3. Adds standard HTTP attributes: method, URL, user agent
     4. Wraps the response writer to capture the HTTP status code
     5. Propagates the trace context to downstream handlers
     6. Records the final HTTP status code when the request completes

    Example usage:

        middleware := tracer.NewHttpMiddleware(tracer)

        http.Handle("/api/", middleware.Middleware(apiHandler))

        // Or with a router:
        r := mux.NewRouter()
        r.Use(middleware.Middleware)
        r.HandleFunc("/users/{id}", getUserHandler)

type InitializationError struct {
	Component string
	Cause     error
}
    InitializationError wraps failures during component setup.

func (e *InitializationError) Error() string
    Error returns a string representation of the error.

func (e *InitializationError) Unwrap() error
    Unwrap returns the underlying error.

type PropagationError struct {
	Operation string
	Cause     error
}
    PropagationError wraps errors related to context propagation.

func (e *PropagationError) Error() string
    Error returns a string representation of the error.

func (e *PropagationError) Unwrap() error
    Unwrap returns the underlying error.

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
    ProviderConfig holds comprehensive configuration for creating a
    TracerProvider. It combines basic tracing configuration with advanced
    options for batch processing, resource identification, and performance
    tuning. This allows fine-grained control over the tracing pipeline behavior.

    The configuration supports fluent method chaining for ease of use:

        config := tracer.NewProviderConfig("my-service", "v1.0.0").
            WithOTLPExporter("localhost:4317", "grpc", true).
            WithSampling("probabilistic", 0.1).
            WithBatchOptions(5*time.Second, 30*time.Second, 512, 2048)

func NewProviderConfig(serviceName, serviceVersion string) *ProviderConfig
    NewProviderConfig creates a new ProviderConfig with sensible defaults for
    advanced configuration. It initializes the configuration with default batch
    processing settings and creates a base Config using the provided service
    name and version. The returned config supports fluent method chaining for
    customization.

    Default settings:
      - BatchTimeout: 5 seconds
      - ExportTimeout: 30 seconds
      - MaxExportBatchSize: 512 spans
      - MaxQueueSize: 2048 spans
      - OTLP HTTP exporter pointing to localhost:4318
      - Probabilistic sampling at 20%

    Example:

        config := tracer.NewProviderConfig("user-service", "v2.1.0")
        provider, err := tracer.NewProvider(ctx, config)

func (pc *ProviderConfig) WithBatchOptions(batchTimeout, exportTimeout time.Duration, maxExportBatchSize, maxQueueSize int) *ProviderConfig
    WithBatchOptions configures the batch processor settings for span export
    optimization. These settings control how spans are batched and exported,
    affecting both performance and resource usage. Tune these values based on
    your application's traffic patterns and latency requirements.

    Parameters:
      - batchTimeout: Maximum time to wait before exporting (lower = less
        latency, higher = better throughput)
      - exportTimeout: Maximum time allowed for export operations (prevents
        hanging exports)
      - maxExportBatchSize: Maximum spans per batch (higher = better throughput,
        more memory usage)
      - maxQueueSize: Maximum queued spans before dropping (higher = more
        memory, less data loss)

    Example:

        // Low-latency configuration
        config.WithBatchOptions(1*time.Second, 10*time.Second, 256, 1024)

        // High-throughput configuration
        config.WithBatchOptions(10*time.Second, 60*time.Second, 1024, 4096)

func (pc *ProviderConfig) WithOTLPExporter(endpoint, protocol string, insecure bool) *ProviderConfig
    WithOTLPExporter configures the OTLP exporter settings for trace export.
    This method allows you to specify the endpoint, protocol, and security
    settings for sending traces to an OTLP-compatible backend.

    Parameters:
      - endpoint: The URL or address of the OTLP collector (e.g.,
        "localhost:4317", "https://api.honeycomb.io")
      - protocol: Either "grpc" for gRPC transport or "http" for HTTP transport
      - insecure: true to disable TLS (for development), false to use TLS (for
        production)

    Example:

        config.WithOTLPExporter("https://api.honeycomb.io", "http", false)
        config.WithOTLPExporter("localhost:4317", "grpc", true)  // Development

func (pc *ProviderConfig) WithResource(resource *sdkresource.Resource) *ProviderConfig
    WithResource sets a custom OpenTelemetry resource for service
    identification. Resources contain attributes that identify the service,
    version, environment, and other metadata. If not provided, a default
    resource will be created automatically using the service name, version,
    and other attributes from the Config.

    Example:

        resource, _ := resource.New(ctx,
            resource.WithAttributes(
                semconv.ServiceName("payment-service"),
                semconv.ServiceVersion("v1.2.3"),
                semconv.DeploymentEnvironment("production"),
                attribute.String("region", "us-west-2"),
            ),
        )
        config.WithResource(resource)

func (pc *ProviderConfig) WithSampling(samplingType string, ratio float64) *ProviderConfig
    WithSampling configures the sampling strategy and ratio for trace
    collection. Sampling controls what percentage of traces are collected
    and exported, which is crucial for managing overhead in high-traffic
    applications.

    Parameters:
      - samplingType: "probabilistic" (ratio-based), "always_on" (100%),
        or "always_off" (0%)
      - ratio: For probabilistic sampling, the ratio of traces to sample (0.0 to
        1.0) Ignored for "always_on" and "always_off" strategies

    Example:

        config.WithSampling("probabilistic", 0.01)  // 1% sampling for production
        config.WithSampling("always_on", 0)        // 100% sampling for development
        config.WithSampling("always_off", 0)       // Disable tracing

type Tracer struct {
	// Has unexported fields.
}
    Tracer wraps OpenTelemetry tracer with convenience methods for easier
    tracing operations. It provides a simplified interface for creating spans,
    adding attributes, and managing trace context while maintaining full
    compatibility with OpenTelemetry standards.

    Example usage:

        tracer := otelkit.New("my-service")
        ctx, span := tracer.Start(ctx, "operation-name")
        defer span.End()
        // ... your code here

func New(name string) *Tracer
    New creates a new Tracer instance with the specified name. The name is used
    to identify the tracer and appears in telemetry data. It's recommended to
    use your service or component name.

    Example:

        tracer := otelkit.New("user-service")
        tracer := otelkit.New("payment-processor")

func (t *Tracer) GetTraceID(ctx context.Context) string
    GetTraceID extracts and returns the trace ID from the current span context.
    Returns an empty string if no valid span is found in the context. This is
    useful for correlation logging and debugging.

    Example:

        traceID := tracer.GetTraceID(ctx)
        log.WithField("trace_id", traceID).Info("Processing request")

func (t *Tracer) OtelTracer() trace.Tracer
    OtelTracer returns the underlying OpenTelemetry tracer instance. This is
    useful when you need direct access to OpenTelemetry APIs or when integrating
    with other OpenTelemetry-compatible libraries.

    Example:

        otelTracer := tracer.OtelTracer()
        // Use with other OpenTelemetry libraries

func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
    Start creates a new span with the given name and options. Returns a new
    context containing the span and the span itself. The span must be ended by
    calling span.End() when the operation completes.

    Example:

        ctx, span := tracer.Start(ctx, "database-query")
        defer span.End()
        // ... perform database operation

func (t *Tracer) StartClientSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
    StartClientSpan creates a new client span for outgoing requests or
    operations. This is a convenience method that automatically sets the span
    kind to SpanKindClient and adds the provided attributes. Use this for HTTP
    client requests, gRPC client calls, database queries, external API calls,
    etc.

    Example:

        ctx, span := tracer.StartClientSpan(ctx, "call-payment-api",
            attribute.String("http.method", "POST"),
            attribute.String("http.url", "https://api.payment.com/charge"),
        )
        defer span.End()

func (t *Tracer) StartServerSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
    StartServerSpan creates a new server span for incoming requests or
    operations. This is a convenience method that automatically sets the span
    kind to SpanKindServer and adds the provided attributes. Use this for HTTP
    handlers, gRPC server methods, message queue consumers, etc.

    Example:

        ctx, span := tracer.StartServerSpan(ctx, "handle-user-request",
            attribute.String("user.id", userID),
            attribute.String("request.method", "POST"),
        )
        defer span.End()

Generated on Sun Aug 24 06:06:11 UTC 2025
