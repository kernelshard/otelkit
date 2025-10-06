# Production Example


```go
// Package main demonstrates a production-ready OpenTelemetry tracing setup with comprehensive configuration,
// error handling, and best practices for real-world applications.
//
// This example shows:
// 1. Advanced configuration with environment-based setup
// 2. Proper error handling and recovery
// 3. Health checks and metrics endpoints
// 4. Graceful shutdown handling
// 5. Structured logging with trace correlation
// 6. Production-ready middleware stack
// 7. Database tracing with connection pooling
// 8. Circuit breaker pattern for resilience
//
// Usage:
//
//	go run main.go
//	curl http://localhost:8080/api/users
//	curl http://localhost:8080/health
//	curl http://localhost:8080/metrics
//
// Environment Variables:
//
//	OTEL_SERVICE_NAME=production-service
//	OTEL_SERVICE_VERSION=1.0.0
//	OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
//	OTEL_TRACES_SAMPLER=probabilistic
//	OTEL_TRACES_SAMPLER_ARG=0.1
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kernelshard/otelkit"
	"github.com/kernelshard/otelkit/internal/config"
	"go.opentelemetry.io/otel/attribute"
)

// User represents a user entity
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// App represents the application with all dependencies
type App struct {
	db     *sql.DB
	tracer *otelkit.Tracer
}
// ...rest of code omitted for brevity...
```

A production-ready setup for otelkit.