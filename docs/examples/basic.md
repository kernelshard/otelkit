# Basic Example


```go
// Package main demonstrates basic OpenTelemetry tracing with HTTP middleware.
//
// This example shows how to:
// 1. Configure tracing with proper provider setup
// 2. Set up HTTP middleware for automatic tracing
// 3. Create custom spans manually
// 4. Add attributes and events to spans
// 5. Handle errors with tracing
//
// Usage:
//
//	go run main.go
//	curl http://localhost:8080/hello
//	curl http://localhost:8080/error
//
// View traces at: http://localhost:16686
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kernelshard/otelkit"
	"go.opentelemetry.io/otel/attribute"
)

func main() {
	ctx := context.Background()

	// Create provider configuration
	config := otelkit.NewProviderConfig("basic-http-example", "1.0.0").
		WithOTLPExporter("localhost:4317", "grpc", true).
		WithSampling("probabilistic", 1.0)

	// Initialize tracer provider
	provider, err := otelkit.NewProvider(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create tracer provider: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := otelkit.ShutdownTracerProvider(shutdownCtx, provider); err != nil {
			log.Printf("Error shutting down provider: %v", err)
		}
	}()

	// Create tracer
	tracer := otelkit.New("basic-http-example")

	// Create middleware
	middleware := otelkit.NewHttpMiddleware(tracer)

	// Setup HTTP handlers
	mux := http.NewServeMux()
	// ...rest of code omitted for brevity...
```

A minimal example of using otelkit for HTTP tracing.