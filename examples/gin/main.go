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

func main() {
	ctx := context.Background()

	// Create provider configuration
	// You can also use environment variables: OTEL_EXPORTER_OTLP_ENDPOINT, OTEL_SERVICE_NAME, etc.
	config := otelkit.NewProviderConfig("gin-example", "1.0.0").
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
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// Create Gin router
	r := gin.Default()

	// Add OpenTelemetry middleware using the official otelgin package
	// This automatically traces all HTTP requests
	r.Use(otelgin.Middleware("gin-example"))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Get all users with custom span
		api.GET("/users", func(c *gin.Context) {
			// Create a custom child span for business logic
			ctx, span := createCustomSpan(c, "fetch-all-users")
			defer span.End()

			// Simulate database query
			span.AddEvent("querying database")
			span.SetAttributes(attribute.String("db.operation", "SELECT"))

			users := []User{
				{ID: "1", Name: "Alice", Email: "alice@example.com"},
				{ID: "2", Name: "Bob", Email: "bob@example.com"},
			}

			span.SetAttributes(attribute.Int("user.count", len(users)))
			_ = ctx // Use context if needed for further operations

			c.JSON(http.StatusOK, users)
		})

		// Get user by ID with error handling
		api.GET("/users/:id", func(c *gin.Context) {
			ctx, span := createCustomSpan(c, "fetch-user-by-id")
			defer span.End()

			id := c.Param("id")
			span.SetAttributes(
				attribute.String("user.id", id),
				attribute.String("operation", "get_user"),
			)

			// Simulate error case
			if id == "error" {
				span.SetStatus(codes.Error, "User not found")
				span.RecordError(http.ErrMissingFile)
				c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
				return
			}

			user := User{ID: id, Name: "User " + id, Email: id + "@example.com"}
			span.SetStatus(codes.Ok, "User fetched successfully")
			_ = ctx

			c.JSON(http.StatusOK, user)
		})

		// Create user with validation
		api.POST("/users", func(c *gin.Context) {
			ctx, span := createCustomSpan(c, "create-user")
			defer span.End()

			var user User
			if err := c.ShouldBindJSON(&user); err != nil {
				span.SetStatus(codes.Error, "Invalid request body")
				span.RecordError(err)
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			// Simulate user creation
			span.AddEvent("creating user in database")
			user.ID = "123"
			span.SetAttributes(
				attribute.String("user.id", user.ID),
				attribute.String("user.name", user.Name),
			)
			span.SetStatus(codes.Ok, "User created successfully")
			_ = ctx

			c.JSON(http.StatusCreated, user)
		})
	}

	// Start server with proper error handling
	log.Println("Starting Gin server on :8080")
	log.Println("Try:")
	log.Println("  curl http://localhost:8080/health")
	log.Println("  curl http://localhost:8080/api/users")
	log.Println("  curl http://localhost:8080/api/users/1")
	log.Println("  curl http://localhost:8080/api/users/error  # To see error tracing")
	log.Println("  curl -X POST http://localhost:8080/api/users -H 'Content-Type: application/json' -d '{\"name\":\"Charlie\",\"email\":\"charlie@example.com\"}'")

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
