---
title: otelkit - OpenTelemetry Tracing for Go | Zero Boilerplate Distributed Tracing
description: Production-ready OpenTelemetry distributed tracing for Go applications. Zero boilerplate, works with Gin, Echo, Chi, Gorilla Mux, Fiber, and any http.Handler. Flexible sampling, multiple exporters (Jaeger, OTLP, Zipkin).
keywords: 
  - OpenTelemetry
  - Go
  - Golang
  - distributed tracing
  - observability
  - tracing middleware
  - Gin
  - Echo
  - Chi
  - Gorilla Mux
  - Fiber
  - OTLP
  - Jaeger
  - Zipkin
  - http.Handler
  - microservices
  - monitoring
---

# Home

**Zero boilerplate OpenTelemetry distributed tracing for Go applications.**

Production-ready distributed tracing that works with any Go web framework including [Gin](INTEGRATION_GUIDES.md#gin-framework), [Echo](INTEGRATION_GUIDES.md#echo-framework), [Chi](INTEGRATION_GUIDES.md#chi-router), [Gorilla Mux](INTEGRATION_GUIDES.md#gorilla-mux), and standard [net/http](INTEGRATION_GUIDES.md#standard-library-nethttp).

[![Go Reference](https://pkg.go.dev/badge/github.com/kernelshard/otelkit.svg)](https://pkg.go.dev/github.com/kernelshard/otelkit)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernelshard/otelkit)](https://goreportcard.com/report/github.com/kernelshard/otelkit)

Documentation: [https://kernelshard.github.io/otelkit/](https://kernelshard.github.io/otelkit/)

Source Code: [https://github.com/kernelshard/otelkit](https://github.com/kernelshard/otelkit)

## Why use otelkit?

- **Framework Agnostic** - Works seamlessly with [Gin](INTEGRATION_GUIDES.md#gin-framework), [Echo](INTEGRATION_GUIDES.md#echo-framework), [Chi](INTEGRATION_GUIDES.md#chi-router), [Gorilla Mux](INTEGRATION_GUIDES.md#gorilla-mux), and any Go `http.Handler`. Perfect for microservices and REST APIs.
- **Production Ready** - Flexible sampling strategies, multiple exporters (Jaeger, OTLP, Zipkin), intelligent batch processing for high-performance applications.
- **Robust & Safe** - Built-in nil checks, graceful error handling, context-aware operations ensure reliability in production.
- **Easy to Learn** - Minimal configuration with sensible defaults. Get distributed tracing running in your Go application in minutes.

## Installation

Install using `go get`:

```bash
go get github.com/kernelshard/otelkit
```

## Quick Start

Create a `main.go` with automatic HTTP request tracing:

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/kernelshard/otelkit"
)

func main() {
    ctx := context.Background()
    
    // Initialize tracing
    shutdown, err := otelkit.SetupTracing(ctx, "my-service")
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)
    
    // Wrap your HTTP handler
    handler := otelkit.NewInstrumentedHTTPHandler(
        http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.Write([]byte("Hello, world!"))
        }),
        "hello-handler",
    )
    
    http.ListenAndServe(":8080", handler)
}
```

!!! note "View Your Traces"
    This example automatically traces all HTTP requests. View traces in Jaeger UI at [http://localhost:16686](http://localhost:16686) or your configured OpenTelemetry backend.

## Production Configuration

Configure the tracing provider with custom sampling and exporters:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/kernelshard/otelkit"
)

func main() {
    ctx := context.Background()
    
    // Create custom configuration
    config := otelkit.NewProviderConfig("my-service", "v1.0.0").
        WithEnvironment("production").
        WithOTLPExporter("otel-collector:4317", "grpc", false).
        WithSampling("probabilistic", 0.1). // 10% sampling
        WithBatchOptions(
            5*time.Second,   // batch timeout
            30*time.Second,  // export timeout
            512,             // max batch size
            2048,            // max queue size
        )
    
    provider, err := otelkit.NewProvider(ctx, config)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)
    
    // Your application code here
}
```

## Framework Integration

otelkit works seamlessly with popular Go web frameworks.

### Gin Framework

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
    "github.com/kernelshard/otelkit"
)

func main() {
    ctx := context.Background()
    shutdown, err := otelkit.SetupTracing(ctx, "gin-service")
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)
    
    r := gin.Default()
    
    // Add OpenTelemetry middleware
    r.Use(otelgin.Middleware("gin-service"))
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Hello, world!"})
    })
    
    r.Run(":8080")
}
```

### Chi Router

```go
package main

import (
    "context"
    "log"
    "net/http"
    
    "github.com/go-chi/chi/v5"
    "github.com/kernelshard/otelkit"
)

func main() {
    ctx := context.Background()
    shutdown, err := otelkit.SetupTracing(ctx, "chi-service")
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)
    
    r := chi.NewRouter()
    
    // Add otelkit middleware
    tracer := otelkit.New("chi-service")
    middleware := otelkit.NewHttpMiddleware(tracer)
    r.Use(func(next http.Handler) http.Handler {
        return middleware.Middleware(next)
    })
    
    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, world!"))
    })
    
    http.ListenAndServe(":8080", r)
}
```

!!! tip "More Framework Examples"
    See the [Integration Guides](INTEGRATION_GUIDES.md) for complete examples with Echo, Fiber, Gorilla Mux, and other popular Go frameworks.

## Learn More

**[Getting Started](GETTING_STARTED.md)** - Complete setup tutorial

**[API Reference](API_REFERENCE.md)** - Full API documentation

**[Advanced Usage](ADVANCED_USAGE.md)** - Production configurations and optimization

**[Troubleshooting](TROUBLESHOOTING.md)** - Common issues and solutions

## Contributing


For guidance on setting up a development environment and how to make a contribution to otelkit, see [Contributing to otelkit](https://github.com/kernelshard/otelkit/blob/main/CONTRIBUTING.md).
