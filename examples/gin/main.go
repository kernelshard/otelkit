// Package main demonstrates OpenTelemetry tracing with Gin framework integration.
//
// This example shows how to integrate OpenTelemetry tracing with Gin web framework
// including route-specific middleware, error handling, and database operations.
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
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"

	"github.com/samims/otelkit"
)

// User represents a user entity
type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// GinTracerMiddleware creates Gin middleware for OpenTelemetry tracing
func GinTracerMiddleware(tracer *otelkit.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, span := tracer.Start(c.Request.Context(), c.FullPath())
		defer span.End()

		span.SetAttributes(
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.url", c.Request.URL.String()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)

		c.Set("trace_context", ctx)
		c.Next()

		span.SetAttributes(
			attribute.Int("http.status_code", c.Writer.Status()),
		)
	}
}

func main() {
	ctx := context.Background()

	// Create provider configuration
	config := otelkit.NewProviderConfig("gin-example", "1.0.0").
		WithOTLPExporter("localhost:4317", "grpc", true).
		WithSampling("probabilistic", 1.0)

	// Initialize tracer provider
	provider, err := otelkit.NewProvider(ctx, config)
	if err != nil {
		panic(err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = otelkit.ShutdownTracerProvider(shutdownCtx, provider)
	}()

	// Create tracer
	tracer := otelkit.New("gin-example")

	// Create Gin router
	r := gin.Default()

	// Add tracing middleware
	r.Use(GinTracerMiddleware(tracer))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := r.Group("/api")
	{
		// Get all users
		api.GET("/users", func(c *gin.Context) {
			users := []User{
				{ID: "1", Name: "Alice"},
				{ID: "2", Name: "Bob"},
			}
			c.JSON(http.StatusOK, users)
		})

		// Get user by ID
		api.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			user := User{ID: id, Name: "User " + id}
			c.JSON(http.StatusOK, user)
		})

		// Create user
		api.POST("/users", func(c *gin.Context) {
			var user User
			if err := c.ShouldBindJSON(&user); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			user.ID = "123"
			c.JSON(http.StatusCreated, user)
		})
	}

	// Start server
	r.Run(":8080")
}
