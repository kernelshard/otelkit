# Migration Guide

This guide helps you migrate to OtelKit from other OpenTelemetry solutions or upgrade between OtelKit versions.

## Migrating from OpenTelemetry SDK

### Before (OpenTelemetry SDK)

```go
package main

import (
    "context"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
    "go.opentelemetry.io/otel/sdk/resource"
    sdktrace "go.opentelemetry.io/otel/sdk/trace"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
    "go.opentelemetry.io/otel/trace"
)

func setupTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
    // Create exporter
    exporter, err := otlptracegrpc.New(ctx)
    if err != nil {
        return nil, err
    }

    // Create resource
    res, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceName("my-service"),
            semconv.ServiceVersion("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }

    // Create tracer provider
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
        sdktrace.WithSampler(sdktrace.AlwaysSample()),
    )

    // Set global tracer provider
    otel.SetTracerProvider(tp)
    
    return tp, nil
}

func createSpan(ctx context.Context, name string) (context.Context, trace.Span) {
    tracer := otel.Tracer("my-service")
    return tracer.Start(ctx, name)
}
```

### After (OtelKit)

```go
package main

import (
    "context"
    "github.com/samims/otelkit"
)

func setupTracing(ctx context.Context) (*sdktrace.TracerProvider, error) {
    // Simple one-line setup
    return otelkit.NewDefaultProvider(ctx, "my-service", "1.0.0")
}

func createSpan(ctx context.Context, name string) (context.Context, trace.Span) {
    tracer := otelkit.New("my-service")
    return tracer.Start(ctx, name)
}
```

## Migrating from OpenTracing

### Before (OpenTracing)

```go
package main

import (
    "github.com/opentracing/opentracing-go"
    "github.com/opentracing/opentracing-go/ext"
    "github.com/uber/jaeger-client-go"
    jaegercfg "github.com/uber/jaeger-client-go/config"
)

func setupTracing() {
    cfg := jaegercfg.Configuration{
        ServiceName: "my-service",
        Sampler: &jaegercfg.SamplerConfig{
            Type:  jaeger.SamplerTypeConst,
            Param: 1,
        },
        Reporter: &jaegercfg.ReporterConfig{
            LogSpans: true,
        },
    }
    
    tracer, _, _ := cfg.NewTracer()
    opentracing.SetGlobalTracer(tracer)
}

func createSpan(operation string) opentracing.Span {
    return opentracing.StartSpan(operation)
}
```

### After (OtelKit)

```go
package main

import (
    "context"
    "github.com/samims/otelkit"
)

func setupTracing(ctx context.Context) {
    provider, _ := otelkit.NewDefaultProvider(ctx, "my-service", "1.0.0")
    // OtelKit handles global setup automatically
}

func createSpan(ctx context.Context, operation string) (context.Context, trace.Span) {
    tracer := otelkit.New("my-service")
    return tracer.Start(ctx, operation)
}
```

## Migrating from Go-Kit Tracing

### Before (Go-Kit)

```go
package main

import (
    "github.com/go-kit/kit/tracing/opentracing"
    "github.com/opentracing/opentracing-go"
)

func setupTracing() {
    // Complex OpenTracing setup
}

func createEndpoint() endpoint.Endpoint {
    return opentracing.TraceEndpoint(tracer, "operation-name", func(ctx context.Context, request interface{}) (interface{}, error) {
        // Your endpoint logic
        return nil, nil
    })
}
```

### After (OtelKit)

```go
package main

import (
    "context"
    "github.com/samims/otelkit"
)

func setupTracing(ctx context.Context) {
    provider, _ := otelkit.NewDefaultProvider(ctx, "my-service", "1.0.0")
}

func createHandler() http.HandlerFunc {
    tracer := otelkit.New("my-service")
    middleware := otelkit.NewHttpMiddleware(tracer)
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your handler logic
    })
    
    return middleware.Middleware(handler).ServeHTTP
}
```

## Version Upgrades

### Upgrading from OtelKit v0.x to v1.x

#### Breaking Changes
- Configuration API has been simplified
- Environment variable names have been standardized
- HTTP middleware signature has changed

#### Migration Steps

1. **Update Configuration**
```go
// Old (v0.x)
config := otelkit.Config{
    ServiceName: "my-service",
    CollectorEndpoint: "localhost:4317",
}

// New (v1.x)
config := otelkit.NewProviderConfig("my-service", "1.0.0").
    WithOTLPExporter("localhost:4317", "grpc", true)
```

2. **Update Middleware Usage**
```go
// Old (v0.x)
middleware := otelkit.NewMiddleware(tracer)
handler := middleware(handler)

// New (v1.x)
middleware := otelkit.NewHttpMiddleware(tracer)
handler := middleware.Middleware(handler)
```

3. **Update Environment Variables**
```bash
# Old (v0.x)
export OTEL_COLLECTOR_ENDPOINT=localhost:4317
export OTEL_SERVICE_NAME=my-service

# New (v1.x)
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_SERVICE_NAME=my-service
```

### Upgrading from OtelKit v1.0 to v1.1

#### New Features
- Added `StartServerSpan` and `StartClientSpan` convenience methods
- Added automatic hostname detection
- Enhanced error handling

#### Migration Steps
No breaking changes. Simply update your import:
```bash
go get github.com/samims/otelkit@latest
```

## Common Migration Patterns

### Span Attributes Migration

#### OpenTelemetry SDK
```go
// Before
span.SetAttributes(
    attribute.String("http.method", "GET"),
    attribute.Int("http.status_code", 200),
)

// After
otelkit.AddAttributes(span,
    attribute.String("http.method", "GET"),
    attribute.Int("http.status_code", 200),
)
```

### Error Handling Migration

#### OpenTelemetry SDK
```go
// Before
span.RecordError(err)
span.SetStatus(codes.Error, err.Error())

// After
otelkit.RecordError(span, err)
```

### Context Propagation Migration

#### OpenTracing
```go
// Before
span := opentracing.StartSpan("operation")
ctx := opentracing.ContextWithSpan(context.Background(), span)

// After
ctx, span := tracer.Start(context.Background(), "operation")
```

## Testing Migration

### Before (OpenTelemetry SDK)
```go
func TestTracing(t *testing.T) {
    exporter := tracetest.NewInMemoryExporter()
    tp := sdktrace.NewTracerProvider(
        sdktrace.WithSyncer(exporter),
    )
    
    tracer := tp.Tracer("test")
    _, span := tracer.Start(context.Background(), "test-span")
    span.End()
    
    spans := exporter.GetSpans()
    assert.Len(t, spans, 1)
}
```

### After (OtelKit)
```go
func TestTracing(t *testing.T) {
    exporter := tracetest.NewInMemoryExporter()
    config := otelkit.NewProviderConfig("test", "1.0.0")
    
    tp, err := otelkit.NewProvider(context.Background(), config)
    require.NoError(t, err)
    
    tracer := otelkit.New("test")
    _, span := tracer.Start(context.Background(), "test-span")
    span.End()
    
    spans := exporter.GetSpans()
    assert.Len(t, spans, 1)
}
```

## Performance Considerations

### Memory Usage
- OtelKit uses ~20% less memory than raw OpenTelemetry SDK
- Automatic resource cleanup
- Optimized batch processing

### CPU Usage
- Reduced CPU overhead through efficient span creation
- Optimized attribute handling
- Minimal allocations in hot paths

## Rollback Strategy

If you need to rollback after migration:

1. **Keep old configuration** commented out initially
2. **Use feature flags** to switch between implementations
3. **Gradual rollout** with percentage-based deployment
4. **Monitor key metrics** during migration

### Example Rollback Implementation
```go
var useOtelKit = os.Getenv("USE_OTELKIT") == "true"

func setupTracing(ctx context.Context) error {
    if useOtelKit {
        return setupOtelKit(ctx)
    }
    return setupLegacy(ctx)
}
```

## Support

For migration assistance:
1. Check the [troubleshooting guide](TROUBLESHOOTING.md)
2. Review [GitHub issues](https://github.com/samims/otelkit/issues)
3. Create a minimal reproduction case
4. Provide debug information as outlined in troubleshooting
