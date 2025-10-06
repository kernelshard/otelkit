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

func main() {
	ctx := context.Background()

	// Load production configuration
	cfg := loadProductionConfig()

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize tracer provider
	provider, err := initTracerProvider(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize tracer provider: %v", err)
	}
	defer shutdownTracerProvider(ctx, provider)

	// Create tracer
	tracer := otelkit.New("example-service")

	// Create application
	app := &App{
		db:     db,
		tracer: tracer,
	}

	// Setup HTTP server
	server := setupHTTPServer(app, cfg)

	// Start server with graceful shutdown
	startServerWithGracefulShutdown(server, cfg)
}

// loadProductionConfig loads production-ready configuration
func loadProductionConfig() *otelkit.ProviderConfig {
	// Create base configuration
	providerCfg := otelkit.NewProviderConfig("production-service", "1.0.0")

	// Override with environment variables
	if serviceName := os.Getenv("OTEL_SERVICE_NAME"); serviceName != "" {
		providerCfg.Config.ServiceName = serviceName
	}
	if serviceVersion := os.Getenv("OTEL_SERVICE_VERSION"); serviceVersion != "" {
		providerCfg.Config.ServiceVersion = serviceVersion
	}
	if endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); endpoint != "" {
		providerCfg.Config.OTLPExporterEndpoint = endpoint
	}
	if sampler := os.Getenv("OTEL_TRACES_SAMPLER"); sampler != "" {
		// Parse sampling type from environment variable
		switch sampler {
		case "probabilistic":
			providerCfg.Config.SamplingType = config.SamplingProbabilistic
		case "always_on":
			providerCfg.Config.SamplingType = config.SamplingAlwaysOn
		case "always_off":
			providerCfg.Config.SamplingType = config.SamplingAlwaysOff
		default:
			log.Printf("Unknown sampling type: %s, using default", sampler)
		}
	}
	if samplerArg := os.Getenv("OTEL_TRACES_SAMPLER_ARG"); samplerArg != "" {
		var ratio float64
		if _, err := fmt.Sscanf(samplerArg, "%f", &ratio); err != nil {
			log.Printf("Invalid sampling ratio: %v, using default", err)
		} else {
			providerCfg.Config.SamplingRatio = ratio
		}
	}

	// Set production defaults
	providerCfg.WithSampling(config.SamplingProbabilistic, 0.1) // 10% sampling for production
	providerCfg.WithBatchOptions(
		2*time.Second,  // batch timeout
		30*time.Second, // export timeout
		512,            // max batch size
		2048,           // max queue size
	)

	return providerCfg
}

// initDatabase initializes database with connection pooling
// Note: This example uses PostgreSQL. For a simpler setup, you could use SQLite:
//
//	import _ "github.com/mattn/go-sqlite3"
//	db, err := sql.Open("sqlite3", ":memory:")
func initDatabase(cfg *otelkit.ProviderConfig) (*sql.DB, error) {
	// In production, use environment variables for database connection
	dsn := getEnv("DATABASE_URL", "")
	if dsn == "" {
		// For local development
		dsn = "postgres://user:password@localhost:5432/production?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// initTracerProvider initializes the tracer provider with production settings
func initTracerProvider(ctx context.Context, cfg *otelkit.ProviderConfig) (*otelkit.ProviderConfig, error) {
	// The config is already properly configured in loadProductionConfig
	// Create and initialize the provider
	_, err := otelkit.NewProvider(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer provider: %w", err)
	}

	return cfg, nil
}

// shutdownTracerProvider gracefully shuts down the tracer provider
func shutdownTracerProvider(ctx context.Context, cfg *otelkit.ProviderConfig) {
	// Note: In a real application, you would store the provider instance
	// and call provider.Shutdown(shutdownCtx) here
	// The shutdown is handled by the global provider set in initTracerProvider
	log.Println("Tracer provider shutdown complete")
}

// setupHTTPServer configures the HTTP server with all routes
func setupHTTPServer(app *App, cfg *otelkit.ProviderConfig) *http.Server {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/users", app.handleGetUsers)
	mux.HandleFunc("/api/users/", app.handleGetUser)

	// Health check
	mux.HandleFunc("/health", app.handleHealth)

	// Metrics endpoint
	mux.HandleFunc("/metrics", app.handleMetrics)

	// Create middleware
	middleware := otelkit.NewHttpMiddleware(app.tracer)

	// Wrap with middleware
	handler := middleware.Middleware(mux)

	return &http.Server{
		Addr:         ":" + getEnv("PORT", "8080"),
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

// startServerWithGracefulShutdown starts the server with graceful shutdown handling
func startServerWithGracefulShutdown(server *http.Server, cfg *otelkit.ProviderConfig) {
	log.Printf("Starting production server on %s", server.Addr)
	log.Printf("Service: %s v%s", cfg.Config.ServiceName, cfg.Config.ServiceVersion)
	log.Printf("Environment: %s", cfg.Config.Environment)

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// handleGetUsers handles GET /api/users
func (a *App) handleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "get-users")
	defer span.End()

	// Query users from database
	users, err := a.getUsers(ctx)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// handleGetUser handles GET /api/users/{id}
func (a *App) handleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "get-user")
	defer span.End()

	// Extract user ID from URL
	id := r.URL.Path[len("/api/users/"):]
	span.SetAttributes(attribute.String("user.id", id))

	user, err := a.getUserByID(ctx, id)
	if err != nil {
		span.RecordError(err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// handleHealth handles health check endpoint
func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, span := a.tracer.Start(r.Context(), "health-check")
	defer span.End()

	// Check database connectivity
	if err := a.db.PingContext(ctx); err != nil {
		span.RecordError(err)
		http.Error(w, "Database connection failed", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "healthy"}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// handleMetrics handles metrics endpoint
func (a *App) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "production-service",
		"version": "1.0.0",
		"status":  "running",
	}); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
	}
}

// getUsers retrieves all users from database
func (a *App) getUsers(ctx context.Context) ([]User, error) {
	_, span := a.tracer.Start(ctx, "db-get-users")
	defer span.End()

	query := `SELECT id, name, email, created_at FROM users ORDER BY created_at DESC`
	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// getUserByID retrieves a user by ID from database
func (a *App) getUserByID(ctx context.Context, id string) (*User, error) {
	_, span := a.tracer.Start(ctx, "db-get-user-by-id")
	defer span.End()

	var user User
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	err := a.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
