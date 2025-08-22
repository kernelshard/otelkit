# OtelKit

[![Go Report Card](https://goreportcard.com/badge/github.com/samims/otelkit)](https://goreportcard.com/report/github.com/samims/otelkit)
[![GoDoc](https://godoc.org/github.com/samims/otelkit?status.svg)](https://godoc.org/github.com/samims/otelkit)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simplified, opinionated wrapper around OpenTelemetry tracing for Go applications. OtelKit provides an easy-to-use API for creating and managing distributed traces while hiding the complexity of the underlying OpenTelemetry SDK.

## Features

🚀 **Simple API** - Easy-to-use wrapper around OpenTelemetry  
🔧 **Flexible Configuration** - Supports both simple and advanced configurations  
🌐 **HTTP Middleware** - Built-in middleware for popular HTTP frameworks  
📊 **Multiple Exporters** - Support for Jaeger, OTLP HTTP, and OTLP gRPC  
⚡ **Performance Optimized** - Configurable sampling and batch processing  
🛡️ **Production Ready** - Comprehensive error handling and graceful shutdown  

## Quick Start

### Installation

```bash
go get github.com/samims/otelkit
```

### Basic Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/samims/otelkit"
    "go.opentelemetry.io/otel/attribute"
)

func main() {
    // Initialize with default configuration
    provider, err := otelkit.NewDefaultProvider(context.Background(), "my-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())

    // Create a tracer
    tracer := otelkit.New("my-service")
    
    // Create a span
    ctx, span := tracer.Start(context.Background(), "do-work")
    defer span.End()
    
    // Add attributes
    otelkit.AddAttributes(span,
        attribute.String("user.id", "12345"),
        attribute.String("operation", "process-payment"),
    )
    
    // Your business logic here
    doWork(ctx)
}

func doWork(ctx context.Context) {
    // Work happens here...
}
```

### HTTP Middleware

```go
package main

import (
    "context"
    "net/http"
    
    "github.com/gorilla/mux"
    "github.com/samims/otelkit"
)

func main() {
    // Setup tracing
    provider, _ := otelkit.NewDefaultProvider(context.Background(), "web-service")
    defer provider.Shutdown(context.Background())
    
    tracer := otelkit.New("web-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    // Setup router with middleware
    r := mux.NewRouter()
    r.Use(middleware.Middleware)
    r.HandleFunc("/users/{id}", getUserHandler)
    
    http.ListenAndServe(":8080", r)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
    // Handler automatically traced by middleware
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("User data"))
}
```

## Advanced Configuration

For production environments, you'll want more control over the configuration:

```go
package main

import (
    "context"
    "time"
    
    "github.com/samims/otelkit"
)

func main() {
    // Advanced configuration
    config := otelkit.NewProviderConfig("payment-service", "v2.1.0").
        WithOTLPExporter("https://api.honeycomb.io", "http", false).
        WithSampling("probabilistic", 0.05). // 5% sampling
        WithBatchOptions(
            2*time.Second,  // batch timeout
            10*time.Second, // export timeout  
            1024,          // max batch size
            4096,          // max queue size
        )
    
    provider, err := otelkit.NewProvider(context.Background(), config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(context.Background())
    
    // Use tracer as normal
    tracer := otelkit.New("payment-service")
    // ...
}
```

## Configuration Options

### Sampling Strategies

- **`probabilistic`** - Sample based on probability ratio (0.0 to 1.0)
- **`always_on`** - Sample all traces (100%)
- **`always_off`** - Sample no traces (0%)

### Exporters

- **OTLP HTTP** - HTTP-based OTLP exporter (default)
- **OTLP gRPC** - gRPC-based OTLP exporter (more efficient for high throughput)

### Batch Processing

Fine-tune performance with batch processor settings:

```go
config.WithBatchOptions(
    batchTimeout,       // Max time to wait before exporting
    exportTimeout,      // Max time for export operation
    maxExportBatchSize, // Max spans per batch
    maxQueueSize,       // Max queued spans before dropping
)
```

## API Reference

### Core Types

- **`Tracer`** - Main tracer wrapper with convenience methods
- **`ProviderConfig`** - Configuration for tracer provider
- **`HTTPMiddleware`** - HTTP middleware for automatic request tracing

### Key Functions

- **`New(name)`** - Create new tracer instance
- **`NewDefaultProvider()`** - Quick setup with sensible defaults
- **`NewProvider(config)`** - Advanced setup with custom configuration
- **`NewHttpMiddleware(tracer)`** - Create HTTP middleware

### Utility Functions

- **`AddAttributes(span, ...attrs)`** - Safely add attributes to span
- **`AddEvent(span, name, ...attrs)`** - Add timestamped event to span
- **`RecordError(span, err)`** - Record error and set span status
- **`EndSpan(span)`** - Safely end span

## Examples

Check the `/examples` directory for complete working examples:

- **[Basic Usage](examples/basic/main.go)** - Simple tracing setup
- **[HTTP Server](examples/http/main.go)** - HTTP server with middleware
- **[Advanced Config](examples/advanced/main.go)** - Production configuration
- **[Database Tracing](examples/database/main.go)** - Database operation tracing

## Best Practices

### 1. Use Appropriate Sampling in Production
```go
// Development: 100% sampling
config.WithSampling("always_on", 0)

// Production: Low sampling rate
config.WithSampling("probabilistic", 0.01) // 1%
```

### 2. Always Defer Span Ending
```go
ctx, span := tracer.Start(ctx, "operation")
defer span.End() // Always defer this
```

### 3. Add Meaningful Attributes
```go
otelkit.AddAttributes(span,
    attribute.String("user.id", userID),
    attribute.String("operation.type", "payment"),
    attribute.Int64("amount", amount),
)
```

### 4. Handle Errors Properly
```go
if err := doSomething(); err != nil {
    otelkit.RecordError(span, err)
    return err
}
```

### 5. Use Context Propagation
```go
// Always pass context to maintain trace continuity
func processPayment(ctx context.Context, amount int64) error {
    ctx, span := tracer.Start(ctx, "process-payment")
    defer span.End()
    
    return callPaymentAPI(ctx, amount) // Pass ctx along
}
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Built on top of [OpenTelemetry Go](https://github.com/open-telemetry/opentelemetry-go)
- Inspired by the need for simpler tracing setup in Go applications
