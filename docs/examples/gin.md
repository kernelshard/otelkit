# Gin Example


```go
// Package main demonstrates OpenTelemetry tracing with Gin framework integration.
//
// This example shows how to integrate OpenTelemetry tracing with Gin web framework
// using the official otelgin middleware from OpenTelemetry contrib.
//
// Usage:
//
//	go run main.go
//	curl http://localhost:8080/api/users
//	curl http://localhost:8080/api/users/1
//	curl http://localhost:8080/health
//
// Installation:
//
//	go get github.com/gin-gonic/gin
//	go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
//
// The example uses:
// - otelgin.Middleware for automatic HTTP request tracing
// - Custom spans for business logic
// - Context propagation for distributed tracing
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/kernelshard/otelkit"
)

// User represents a user entity
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// getTracerFromContext retrieves the tracer from Gin context or creates a new one
func getTracerFromContext(c *gin.Context) *otelkit.Tracer {
	// Get the global tracer
	return otelkit.New("gin-example")
}

// createCustomSpan creates a custom span within the current request context
func createCustomSpan(c *gin.Context, operationName string) (context.Context, trace.Span) {
	// Get the current span context from the request
	ctx := c.Request.Context()
	tracer := getTracerFromContext(c)

	// Create a child span
	return tracer.Start(ctx, operationName)
}
// ...rest of code omitted for brevity...
```

How to use otelkit with the Gin framework.