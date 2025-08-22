# Troubleshooting Guide

This guide helps you resolve common issues when using OtelKit.

## Common Issues and Solutions

### 1. No Traces Appearing

**Symptoms:**
- No traces visible in your collector UI
- Empty trace list in Jaeger/Zipkin

**Solutions:**

#### Check Collector Connectivity
```bash
# Test connectivity to collector
curl -v http://localhost:4318/v1/traces
# or for gRPC
grpcurl -plaintext localhost:4317 list
```

#### Verify Configuration
```go
// Add debug logging
config := otelkit.NewProviderConfig("service", "v1.0.0")
fmt.Printf("Config: %+v\n", config)
```

#### Check Sampling
```go
// Ensure sampling is enabled
config.WithSampling("always_on", 1.0) // For debugging
```

### 2. High Memory Usage

**Symptoms:**
- Application memory usage increases over time
- Out of memory errors

**Solutions:**

#### Reduce Batch Sizes
```go
config.WithBatchOptions(
    5*time.Second,
    30*time.Second,
    256,  // Reduced batch size
    1024, // Reduced queue size
)
```

#### Increase Sampling
```go
config.WithSampling("probabilistic", 0.1) // Increase sampling ratio
```

#### Monitor Span Creation
```go
// Add span count monitoring
spanCounter := 0
go func() {
    for {
        time.Sleep(10 * time.Second)
        fmt.Printf("Spans created: %d\n", spanCounter)
    }
}()
```

### 3. Performance Issues

**Symptoms:**
- High CPU usage
- Slow request processing

**Solutions:**

#### Tune Batch Processing
```go
config.WithBatchOptions(
    10*time.Second,  // Longer batch timeout
    60*time.Second,  // Longer export timeout
    1024,           // Larger batch size
    4096,           // Larger queue size
)
```

#### Use Probabilistic Sampling
```go
config.WithSampling("probabilistic", 0.01) // 1% sampling
```

#### Async Span Creation
```go
// Use goroutines for expensive operations
go func() {
    ctx, span := tracer.Start(ctx, "background-work")
    defer span.End()
    // Your background work
}()
```

### 4. Configuration Errors

**Symptoms:**
- Application fails to start
- Configuration validation errors

**Solutions:**

#### Validate Configuration
```go
if err := config.Validate(); err != nil {
    fmt.Printf("Configuration error: %v\n", err)
    // Handle error appropriately
}
```

#### Check Environment Variables
```bash
# Print all OTEL environment variables
env | grep OTEL
```

#### Debug Configuration Loading
```go
cfg := otelkit.NewConfigFromEnv()
fmt.Printf("Loaded config: %+v\n", cfg)
```

### 5. Collector Connection Issues

**Symptoms:**
- Connection refused errors
- TLS certificate issues

**Solutions:**

#### Test Connectivity
```bash
# Test HTTP endpoint
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{"resourceSpans":[]}'
```

#### Check TLS Configuration
```go
// For insecure connections (development)
config.WithOTLPExporter("localhost:4317", "grpc", true)

// For secure connections (production)
config.WithOTLPExporter("collector.example.com:4317", "grpc", false)
```

### 6. Context Propagation Issues

**Symptoms:**
- Broken trace chains
- Missing parent spans

**Solutions:**

#### Verify Context Usage
```go
// Always pass context
func processRequest(ctx context.Context, data string) error {
    ctx, span := tracer.Start(ctx, "process-request")
    defer span.End()
    
    // Pass context to downstream calls
    return downstreamService.Process(ctx, data)
}
```

#### Check Middleware Order
```go
// Ensure middleware is applied correctly
r := chi.NewRouter()
r.Use(otelkitMiddleware) // Applied before routes
r.Get("/api", handler)
```

## Debugging Tools

### 1. Enable Debug Logging
```go
import "go.opentelemetry.io/otel/sdk/trace"

// Enable debug logging
logger := log.New(os.Stdout, "OTEL: ", log.LstdFlags)
tp := sdktrace.NewTracerProvider(
    sdktrace.WithSyncer(exporter),
    sdktrace.WithResource(res),
)
```

### 2. In-Memory Testing
```go
import "go.opentelemetry.io/otel/sdk/trace/tracetest"

// Use in-memory exporter for testing
exporter := tracetest.NewInMemoryExporter()
provider := sdktrace.NewTracerProvider(
    sdktrace.WithSyncer(exporter),
)

// After operations
spans := exporter.GetSpans()
fmt.Printf("Captured %d spans\n", len(spans))
```

### 3. Health Check Endpoint
```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "health-check")
    defer span.End()
    
    // Perform health check
    if err := checkCollectorHealth(ctx); err != nil {
        span.SetStatus(codes.Error, err.Error())
        http.Error(w, "Unhealthy", http.StatusServiceUnavailable)
        return
    }
    
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

## Environment-Specific Issues

### Docker Issues
```dockerfile
# Ensure proper networking
FROM golang:1.22-alpine
WORKDIR /app
COPY . .
RUN go build -o app .

# Use host network for local development
docker run --network host my-app
```

### Kubernetes Issues
```yaml
# Ensure proper service discovery
apiVersion: v1
kind: Service
metadata:
  name: otel-collector
spec:
  selector:
    app: otel-collector
  ports:
  - port: 4317
    targetPort: 4317
```

### Cloud Provider Issues
```go
// AWS ECS
config := otelkit.NewProviderConfig("service", "v1.0.0").
    WithOTLPExporter("otel-collector.local:4317", "grpc", false)

// GCP Cloud Run
config := otelkit.NewProviderConfig("service", "v1.0.0").
    WithOTLPExporter("otel-collector:4317", "grpc", false)
```

## Performance Monitoring

### Memory Profiling
```go
import "runtime"

// Monitor memory usage
var m runtime.MemStats
runtime.ReadMemStats(&m)
fmt.Printf("Memory usage: %d MB\n", m.Alloc/1024/1024)
```

### Span Count Monitoring
```go
// Monitor span creation rate
spanCounter := 0
go func() {
    for {
        time.Sleep(10 * time.Second)
        fmt.Printf("Spans per second: %d\n", spanCounter/10)
        spanCounter = 0
    }
}()
```

## Common Error Messages

### "connection refused"
- Check if collector is running
- Verify endpoint configuration
- Check firewall settings

### "certificate verify failed"
- Use insecure mode for development
- Install proper certificates for production
- Check certificate validity

### "context deadline exceeded"
- Increase timeout values
- Check network connectivity
- Verify collector availability

## Getting Help

### Debug Information to Collect
1. Configuration values (sanitized)
2. Environment variables
3. Collector logs
4. Application logs
5. Network connectivity tests

### Support Channels
1. Check GitHub issues
2. Review documentation
3. Test with minimal reproduction
4. Provide debug information

### Example Debug Script
```bash
#!/bin/bash
echo "=== OtelKit Debug Information ==="
echo "Go version: $(go version)"
echo "OtelKit version: $(go list -m github.com/samims/otelkit)"
echo "Environment variables:"
env | grep OTEL
echo "=== Collector Health Check ==="
curl -v http://localhost:4318/v1/traces || echo "Collector not accessible"
