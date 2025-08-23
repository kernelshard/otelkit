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

	"github.com/samims/otelkit"
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

	// Hello handler with manual span creation
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "handle-hello")
		defer span.End()

		// Add request attributes
		span.SetAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
		)

		// Simulate some processing
		time.Sleep(100 * time.Millisecond)

		// Add a custom event
		span.AddEvent("processing_request")

		// Create a child span for database operation
		_, dbSpan := tracer.Start(ctx, "database-query")
		dbSpan.SetAttributes(
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
			attribute.String("db.statement", "SELECT * FROM users WHERE id = ?"),
		)

		// Simulate database query
		time.Sleep(50 * time.Millisecond)
		dbSpan.End()

		// Add response attributes
		span.SetAttributes(attribute.Int("http.status_code", 200))

		response := map[string]string{
			"message":  "Hello, OpenTelemetry!",
			"trace_id": span.SpanContext().TraceID().String(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	})

	// Error handler demonstrating error recording
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "handle-error")
		defer span.End()

		// Simulate an error
		err := fmt.Errorf("something went wrong")
		otelkit.RecordError(span, err)
		span.SetAttributes(attribute.Bool("error", true))

		w.WriteHeader(http.StatusInternalServerError)
		if err := json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		}); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	})

	// Health check handler
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	})

	// Wrap the mux with tracing middleware
	handler := middleware.Middleware(mux)

	// Start server with proper timeout configuration
	port := getEnv("PORT", "8080")
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("Starting server on :%s", port)
	log.Printf("View traces at: http://localhost:16686")
	log.Fatal(server.ListenAndServe())
}

func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
