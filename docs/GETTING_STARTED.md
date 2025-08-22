# Getting Started with OtelKit

This guide will help you get up and running with OtelKit in just a few minutes.

## Prerequisites

- Go 1.22 or later
- An OpenTelemetry collector (Jaeger, Zipkin, or OTLP-compatible)
- Basic understanding of Go HTTP servers

## Quick Start (5 minutes)

### 1. Install OtelKit

```bash
go get github.com/samims/otelkit
```

### 2. Basic Setup

Create a simple HTTP server with tracing:

```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    // Initialize tracing with defaults
    provider, err := otelkit.NewDefaultProvider(ctx, "my-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        provider.Shutdown(shutdownCtx)
    }()

    // Create tracer
    tracer := otelkit.New("my-service")
    
    // Create HTTP middleware
    middleware := otelkit.NewHttpMiddleware(tracer)
    
    // Setup routes
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, OpenTelemetry!"))
    })
    
    // Start server
    handler := middleware.Middleware(mux)
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### 3. Run with Jaeger (Local Development)

```bash
# Start Jaeger in Docker
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest

# Run your service
go run main.go

# View traces at http://localhost:16686
```

## Configuration Options

### Environment Variables

Set these environment variables to configure OtelKit:

```bash
# Service identification
export OTEL_SERVICE_NAME=my-service
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_ENVIRONMENT=development

# Collector configuration
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_EXPORTER_OTLP_INSECURE=true

# Sampling
export OTEL_TRACES_SAMPLER=probabilistic
export OTEL_TRACES_SAMPLER_ARG=0.1  # 10% sampling
```

### Programmatic Configuration

```go
config := otelkit.NewProviderConfig("my-service", "v1.0.0").
    WithOTLPExporter("collector:4317", "grpc", false).
    WithSampling("probabilistic", 0.05)

provider, err := otelkit.NewProvider(ctx, config)
```

## Next Steps

- [Advanced Usage Guide](ADVANCED_USAGE.md) - Production configuration
- [Integration Guides](INTEGRATION_GUIDES.md) - Framework-specific setup
- [API Reference](API_REFERENCE.md) - Complete API documentation
