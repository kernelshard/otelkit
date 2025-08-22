# Advanced Usage Guide

This guide covers advanced configuration and production-ready usage patterns.

## Production Configuration

### High-Performance Setup

```go
config := otelkit.NewProviderConfig("payment-service", "v2.1.0").
    WithEnvironment("production").
    WithOTLPExporter("otel-collector:4317", "grpc", false).
    WithSampling("probabilistic", 0.01). // 1% sampling
    WithBatchOptions(
        2*time.Second,   // batch timeout
        30*time.Second,  // export timeout
        1024,           // max batch size
        4096,           // max queue size
    )

provider, err := otelkit.NewProvider(ctx, config)
```

### Custom Resource Attributes

```go
import (
    "go.opentelemetry.io/otel/sdk/resource"
    semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

res, err := resource.New(ctx,
    resource.WithAttributes(
        semconv.ServiceName("payment-service"),
        semconv.ServiceVersion("v2.1.0"),
        semconv.DeploymentEnvironment("production"),
        semconv.HostName("payment-01"),
        semconv.ServiceInstanceID("instance-123"),
        attribute.String("cloud.provider", "aws"),
        attribute.String("cloud.region", "us-west-2"),
        attribute.String("k8s.cluster.name", "production-cluster"),
    ),
)

config := otelkit.NewProviderConfig("payment-service", "v2.1.0").
    WithResource(res)
```

## Sampling Strategies

### Probabilistic Sampling
```go
// 1% sampling for production
config.WithSampling("probabilistic", 0.01)
```

### Always On (Development)
```go
// 100% sampling for development
config.WithSampling("always_on", 0)
```

### Always Off (Testing)
```go
// Disable tracing for tests
config.WithSampling("always_off", 0)
```

## Batch Processing Configuration

### Low Latency
```go
config.WithBatchOptions(
    500*time.Millisecond, // Fast batching
    10*time.Second,       // Quick export timeout
    256,                  // Small batches
    1024,                 // Limited queue
)
```

### High Throughput
```go
config.WithBatchOptions(
    10*time.Second,       // Longer batching
    60*time.Second,       // Longer export timeout
    2048,                 // Large batches
    8192,                 // Large queue
)
```

## Error Handling and Recovery

### Graceful Degradation
```go
func initTracing(ctx context.Context) *sdktrace.TracerProvider {
    provider, err := otelkit.NewProvider(ctx, config)
    if err != nil {
        log.Printf("Failed to initialize tracing: %v", err)
        // Return noop provider for graceful degradation
        return sdktrace.NewTracerProvider()
    }
    return provider
}
```

### Health Checks
```go
func tracingHealthCheck(ctx context.Context) error {
    // Create a test span to verify tracing is working
    tracer := otelkit.New("health-check")
    _, span := tracer.Start(ctx, "health-check")
    defer span.End()
    
    // Add health check attributes
    span.SetAttributes(
        attribute.String("check.type", "tracing"),
        attribute.Bool("check.passed", true),
    )
    
    return nil
}
```

## Performance Monitoring

### Memory Usage
```go
// Monitor span queue size
config.WithBatchOptions(
    5*time.Second,
    30*time.Second,
    512,
    2048, // Monitor this value
)
```

### Export Metrics
```go
// Add custom metrics for tracing
import "go.opentelemetry.io/otel/metric"

meter := global.Meter("otelkit")
spanCounter, _ := meter.Int64Counter("spans_exported")
errorCounter, _ := meter.Int64Counter("export_errors")
```

## Security Considerations

### TLS Configuration
```go
config := otelkit.NewProviderConfig("service", "v1.0.0").
    WithOTLPExporter("collector.example.com:4317", "grpc", false) // TLS enabled
```

### Header Authentication
```go
// For custom authentication headers
import "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

exporter, err := otlptracehttp.New(ctx,
    otlptracehttp.WithEndpoint("collector.example.com"),
    otlptracehttp.WithHeaders(map[string]string{
        "Authorization": "Bearer " + token,
    }),
)
```

## Multi-Service Setup

### Service Mesh Integration
```go
// Configure for service mesh environments
config := otelkit.NewProviderConfig("service", "v1.0.0").
    WithOTLPExporter("otel-collector.mesh.local:4317", "grpc", false).
    WithSampling("probabilistic", 0.1)
```

### Kubernetes Configuration
```yaml
# Deployment configuration
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-service
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: OTEL_SERVICE_NAME
          value: "my-service"
        - name: OTEL_SERVICE_VERSION
          valueFrom:
            fieldRef:
              fieldPath: metadata.labels['version']
        - name: OTEL_RESOURCE_ATTRIBUTES
          value: "k8s.namespace.name=$(NAMESPACE),k8s.pod.name=$(POD_NAME)"
```

## Troubleshooting

### Common Issues

#### 1. No Traces Appearing
- Check collector endpoint connectivity
- Verify sampling configuration
- Check for dropped spans (queue full)

#### 2. High Memory Usage
- Reduce batch size and queue size
- Increase sampling ratio
- Monitor span creation rate

#### 3. Performance Impact
- Use probabilistic sampling
- Tune batch processing parameters
- Consider async span creation

### Debug Configuration
```go
// Enable debug logging
import "go.opentelemetry.io/otel/sdk/trace"

tp := sdktrace.NewTracerProvider(
    sdktrace.WithSyncer(exporter), // Synchronous for debugging
    sdktrace.WithResource(res),
)
