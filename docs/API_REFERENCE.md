# API Reference

Complete API documentation for OtelKit.

## Core Types

### Config
Configuration structure for OpenTelemetry setup.

```go
type Config struct {
    ServiceName          string
    ServiceVersion       string
    Environment          string
    OTLPExporterEndpoint string
    OTLPExporterInsecure bool
    OTLPExporterProtocol string
    BatchTimeout         time.Duration
    ExportTimeout        time.Duration
    MaxExportBatchSize   int
    MaxQueueSize         int
    SamplingRatio        float64
    SamplingType         string
    InstanceID           string
    Hostname             string
}
```

### ProviderConfig
Advanced configuration for tracer provider.

```go
type ProviderConfig struct {
    Config   *Config
    Resource *sdkresource.Resource
    BatchTimeout time.Duration
    ExportTimeout time.Duration
    MaxExportBatchSize int
    MaxQueueSize int
}
```

### Tracer
Main tracer wrapper with convenience methods.

```go
type Tracer struct {
    // Contains filtered or unexported fields
}
```

## Functions

### NewConfig
Creates a new configuration with sensible defaults.

```go
func NewConfig(serviceName, serviceVersion string) *Config
```

**Parameters:**
- `serviceName`: Name of the service
- `serviceVersion`: Version of the service

**Returns:**
- `*Config`: New configuration instance

**Example:**
```go
cfg := otelkit.NewConfig("my-service", "v1.0.0")
```

### NewConfigFromEnv
Creates configuration from environment variables.

```go
func NewConfigFromEnv() *Config
```

**Returns:**
- `*Config`: Configuration loaded from environment

**Example:**
```go
cfg := otelkit.NewConfigFromEnv()
```

### NewProviderConfig
Creates a new provider configuration.

```go
func NewProviderConfig(serviceName, serviceVersion string) *ProviderConfig
```

**Parameters:**
- `serviceName`: Name of the service
- `serviceVersion`: Version of the service

**Returns:**
- `*ProviderConfig`: New provider configuration

**Example:**
```go
config := otelkit.NewProviderConfig("service", "v1.0.0")
```

### NewDefaultProvider
Creates a tracer provider with default settings.

```go
func NewDefaultProvider(ctx context.Context, serviceName string, serviceVersion ...string) (*sdktrace.TracerProvider, error)
```

**Parameters:**
- `ctx`: Context for initialization
- `serviceName`: Name of the service
- `serviceVersion`: Optional version (defaults to "1.0.0")

**Returns:**
- `*sdktrace.TracerProvider`: Configured tracer provider
- `error`: Any initialization error

**Example:**
```go
provider, err := otelkit.NewDefaultProvider(ctx, "my-service")
```

### NewProvider
Creates a tracer provider with custom configuration.

```go
func NewProvider(ctx context.Context, cfg *ProviderConfig) (*sdktrace.TracerProvider, error)
```

**Parameters:**
- `ctx`: Context for initialization
- `cfg`: Provider configuration

**Returns:**
- `*sdktrace.TracerProvider`: Configured tracer provider
- `error`: Any initialization error

**Example:**
```go
config := otelkit.NewProviderConfig("service", "v1.0.0")
provider, err := otelkit.NewProvider(ctx, config)
```

### New
Creates a new tracer instance.

```go
func New(name string) *Tracer
```

**Parameters:**
- `name`: Name for the tracer

**Returns:**
- `*Tracer`: New tracer instance

**Example:**
```go
tracer := otelkit.New("my-service")
```

### ShutdownTracerProvider
Gracefully shuts down the tracer provider.

```go
func ShutdownTracerProvider(ctx context.Context, tp *sdktrace.TracerProvider) error
```

**Parameters:**
- `ctx`: Context with timeout
- `tp`: Tracer provider to shutdown

**Returns:**
- `error`: Any shutdown error

**Example:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := otelkit.ShutdownTracerProvider(ctx, provider)
```

## Tracer Methods

### Start
Creates a new span.

```go
func (t *Tracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
```

**Parameters:**
- `ctx`: Parent context
- `name`: Span name
- `opts`: Optional span options

**Returns:**
- `context.Context`: New context with span
- `trace.Span`: The created span

**Example:**
```go
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()
```

### StartServerSpan
Creates a new server span with attributes.

```go
func (t *Tracer) StartServerSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
```

**Parameters:**
- `ctx`: Parent context
- `operation`: Operation name
- `attrs`: Optional attributes

**Returns:**
- `context.Context`: New context with span
- `trace.Span`: The created span

**Example:**
```go
ctx, span := tracer.StartServerSpan(ctx, "handle-request",
    attribute.String("http.method", "GET"),
    attribute.String("http.route", "/users"),
)
```

### StartClientSpan
Creates a new client span with attributes.

```go
func (t *Tracer) StartClientSpan(ctx context.Context, operation string, attrs ...attribute.KeyValue) (context.Context, trace.Span)
```

**Parameters:**
- `ctx`: Parent context
- `operation`: Operation name
- `attrs`: Optional attributes

**Returns:**
- `context.Context`: New context with span
- `trace.Span`: The created span

**Example:**
```go
ctx, span := tracer.StartClientSpan(ctx, "http-request",
    attribute.String("http.url", "https://api.example.com"),
    attribute.String("http.method", "POST"),
)
```

### GetTraceID
Extracts the trace ID from context.

```go
func (t *Tracer) GetTraceID(ctx context.Context) string
```

**Parameters:**
- `ctx`: Context to extract from

**Returns:**
- `string`: Trace ID or empty string if not found

**Example:**
```go
traceID := tracer.GetTraceID(ctx)
log.Printf("Trace ID: %s", traceID)
```

### OtelTracer
Returns the underlying OpenTelemetry tracer.

```go
func (t *Tracer) OtelTracer() trace.Tracer
```

**Returns:**
- `trace.Tracer`: Underlying OpenTelemetry tracer

**Example:**
```go
otelTracer := tracer.OtelTracer()
```

## Utility Functions

### AddAttributes
Safely adds attributes to a span.

```go
func AddAttributes(span trace.Span, attrs ...attribute.KeyValue)
```

**Parameters:**
- `span`: Span to add attributes to
- `attrs`: Attributes to add

**Example:**
```go
otelkit.AddAttributes(span,
    attribute.String("user.id", "123"),
    attribute.Int("request.size", 1024),
)
```

### AddEvent
Safely adds an event to a span.

```go
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue)
```

**Parameters:**
- `span`: Span to add event to
- `name`: Event name
- `attrs`: Optional event attributes

**Example:**
```go
otelkit.AddEvent(span, "cache.miss",
    attribute.String("key", "user:123"),
)
```

### RecordError
Records an error on a span.

```go
func RecordError(span trace.Span, err error)
```

**Parameters:**
- `span`: Span to record error on
- `err`: Error to record

**Example:**
```go
if err := doSomething(); err != nil {
    otelkit.RecordError(span, err)
    return err
}
```

### EndSpan
Safely ends a span.

```go
func EndSpan(span trace.Span)
```

**Parameters:**
- `span`: Span to end

**Example:**
```go
defer otelkit.EndSpan(span)
```

### IsRecording
Checks if a span is recording.

```go
func IsRecording(span trace.Span) bool
```

**Parameters:**
- `span`: Span to check

**Returns:**
- `bool`: True if span is recording

**Example:**
```go
if otelkit.IsRecording(span) {
    // Perform expensive operations
}
```

## HTTP Middleware

### NewHttpMiddleware
Creates new HTTP middleware.

```go
func NewHttpMiddleware(tracer *Tracer) *HTTPMiddleware
```

**Parameters:**
- `tracer`: Tracer instance

**Returns:**
- `*HTTPMiddleware`: New middleware instance

**Example:**
```go
middleware := otelkit.NewHttpMiddleware(tracer)
```

### Middleware
Returns HTTP middleware handler.

```go
func (m *HTTPMiddleware) Middleware(next http.Handler) http.Handler
```

**Parameters:**
- `next`: Next handler in chain

**Returns:**
- `http.Handler`: Middleware handler

**Example:**
```go
handler := middleware.Middleware(router)
```

## Configuration Methods

### WithEnvironment
Sets the deployment environment.

```go
func (c *Config) WithEnvironment(env string) *Config
```

### WithOTLPExporter
Configures the OTLP exporter.

```go
func (c *Config) WithOTLPExporter(endpoint string, protocol string, insecure bool) *Config
```

### WithSampling
Configures sampling strategy.

```go
func (c *Config) WithSampling(samplingType string, ratio float64) *Config
```

### WithBatchOptions
Configures batch processing options.

```go
func (pc *ProviderConfig) WithBatchOptions(batchTimeout, exportTimeout time.Duration, maxExportBatchSize, maxQueueSize int) *ProviderConfig
```

### WithResource
Sets custom resource.

```go
func (pc *ProviderConfig) WithResource(resource *sdkresource.Resource) *ProviderConfig
```

## Error Types

### ConfigError
Configuration validation error.

```go
type ConfigError struct {
    Field   string
    Message string
}
```

### InitializationError
Initialization error.

```go
type InitializationError struct {
    Component string
    Cause     error
}
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OTEL_SERVICE_NAME` | Service name | "unknown-service" |
| `OTEL_SERVICE_VERSION` | Service version | "1.0.0" |
| `OTEL_ENVIRONMENT` | Environment | "development" |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Collector endpoint | "localhost:4317" |
| `OTEL_EXPORTER_OTLP_INSECURE` | Disable TLS | false |
| `OTEL_EXPORTER_OTLP_PROTOCOL` | Protocol (grpc/http) | "grpc" |
| `OTEL_TRACES_SAMPLER` | Sampling type | "probabilistic" |
| `OTEL_TRACES_SAMPLER_ARG` | Sampling ratio | 0.2 |
| `OTEL_BSP_TIMEOUT` | Batch timeout | "5s" |
| `OTEL_EXPORTER_TIMEOUT` | Export timeout | "30s" |
| `OTEL_BSP_MAX_EXPORT_BATCH_SIZE` | Max batch size | 512 |
| `OTEL_BSP_MAX_QUEUE_SIZE` | Max queue size | 2048 |
