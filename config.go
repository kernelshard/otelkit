// Package tracer provides a simple, production-ready OpenTelemetry tracing library for Go.
//
// This package offers zero-configuration setup with sensible defaults,
// while still providing full customization when needed.
//
// Basic usage:
//   ctx := context.Background()
//   shutdown, err := tracer.SetupTracing(ctx, "my-service")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer shutdown(ctx)
//
// The package handles:
// - OTLP exporter configuration (HTTP/gRPC)
// - Sampling strategies (probabilistic, always_on, always_off)
// - Resource management with service metadata
// - Context propagation for distributed tracing
// - HTTP and gRPC instrumentation
// - Error recording and span utilities
//
// Configuration can be done via environment variables or programmatically.
// See the README for comprehensive examples and configuration options.
//
// This file defines the Config struct, default values, environment-based
// configuration loading, validation logic, and fluent helper methods.
//
// Usage:
//   cfg := tracer.NewConfigFromEnv()
//   if err := cfg.Validate(); err != nil {
//       log.Fatalf("invalid config: %v", err)
//   }
//
// Environment Variables Supported:
//   - OTEL_SERVICE_NAME                         (e.g., "user-service")
//   - OTEL_SERVICE_VERSION                      (e.g., "1.2.3")
//   - OTEL_ENVIRONMENT                          ("development", "staging", "production")
//   - OTEL_EXPORTER_OTLP_ENDPOINT               (e.g., "localhost:4317")
//   - OTEL_EXPORTER_OTLP_INSECURE               (true/false)
//   - OTEL_EXPORTER_OTLP_PROTOCOL               ("grpc" or "http")
//   - OTEL_BSP_TIMEOUT                          (e.g., "5s")
//   - OTEL_EXPORTER_TIMEOUT                     (e.g., "30s")
//   - OTEL_BSP_MAX_EXPORT_BATCH_SIZE            (e.g., "512")
//   - OTEL_BSP_MAX_QUEUE_SIZE                   (e.g., "2048")
//   - OTEL_TRACES_SAMPLER                       ("probabilistic", "always_on", "always_off")
//   - OTEL_TRACES_SAMPLER_ARG                   (e.g., "0.25")
//   - OTEL_RESOURCE_ATTRIBUTES_SERVICE_INSTANCE_ID (optional unique instance ID)
//
// Customization:
//   Use fluent-style helper methods like `WithEnvironment`, `WithSampling`, etc.,
//   or rely on environment variables.
//
// Note:
//   Validation must be called after construction to ensure correctness.

package otelkit

import (
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Import constants from constants.go
// Default values are now defined in constants.go

// Config defines tracing configuration parameters
type Config struct {
	// Service identification metadata
	ServiceName    string // Name of the service (required)
	ServiceVersion string // Version of the service (required)
	Environment    string // Deployment environment (development/staging/production)

	// OTLP exporter settings
	OTLPExporterEndpoint string // Collector endpoint (host:port)
	OTLPExporterInsecure bool   // Disable TLS verification
	OTLPExporterProtocol string // Protocol for OTLP exporter (default: grpc)

	// Batch processing configuration
	BatchTimeout       time.Duration // Timeout for batch processing (default: 5s)
	ExportTimeout      time.Duration // Timeout for export requests (default: 30s)
	MaxExportBatchSize int           // Maximum batch size for exports (default: 512)
	MaxQueueSize       int           // Maximum queue size for spans (default: 2048)

	// Sampling configuration
	SamplingRatio float64 // Ratio of traces to sample (0.0 - 1.0)
	SamplingType  string  // Sampling strategy

	// Resource attributes
	InstanceID string // Unique instance identifier
	Hostname   string // Host machine name
}

// NewConfig creates a configuration with sensible defaults
func NewConfig(serviceName, serviceVersion string) *Config {
	hostname, _ := os.Hostname()
	return &Config{
		ServiceName:          serviceName,
		ServiceVersion:       serviceVersion,
		Environment:          DefaultEnvironment,
		OTLPExporterEndpoint: DefaultOTLPExporterEndpoint,
		OTLPExporterInsecure: false,
		SamplingRatio:        DefaultSamplingRatio,
		SamplingType:         DefaultSamplingType,
		InstanceID:           generateInstanceID(),
		Hostname:             hostname,
		OTLPExporterProtocol: DefaultOTLPExporterProtocol,
	}
}

// NewConfigFromEnv creates configuration from environment variables
func NewConfigFromEnv() *Config {
	cfg := NewConfig(
		getEnv(EnvServiceName, DefaultServiceName),
		getEnv(EnvServiceVersion, DefaultServiceVersion),
	)

	// Apply environment overrides
	cfg.Environment = getEnv(EnvEnvironment, DefaultEnvironment)

	cfg.OTLPExporterProtocol = getEnv(EnvOTLPExporterProtocol, DefaultOTLPExporterProtocol)
	cfg.BatchTimeout = getEnvDuration(EnvBatchTimeout, DefaultBatchTimeout)
	cfg.ExportTimeout = getEnvDuration(EnvExportTimeout, DefaultExportTimeout)
	cfg.MaxExportBatchSize = getEnvInt(EnvMaxExportBatchSize, DefaultMaxExportBatchSize)
	cfg.MaxQueueSize = getEnvInt(EnvMaxQueueSize, DefaultMaxQueueSize)

	cfg.OTLPExporterEndpoint = getEnv(EnvOTLPExporterEndpoint, DefaultOTLPExporterEndpoint)
	cfg.OTLPExporterInsecure = getEnvBool(EnvOTLPExporterInsecure, false)
	cfg.SamplingRatio = getEnvFloat(EnvSamplingRatio, DefaultSamplingRatio)
	cfg.SamplingType = getEnv(EnvSamplingType, DefaultSamplingType)
	cfg.InstanceID = getEnv(EnvInstanceID, cfg.InstanceID)

	return cfg
}

// Validate ensures configuration parameters are correct
func (c *Config) Validate() error {
	if c.ServiceName == "" {
		return &ConfigError{Field: "ServiceName", Message: ErrServiceNameRequired}
	}
	if c.ServiceVersion == "" {
		return &ConfigError{Field: "ServiceVersion", Message: ErrServiceVersionRequired}
	}
	if !contains(ValidEnvironments, c.Environment) {
		return &ConfigError{Field: "Environment", Message: ErrInvalidEnvironment}
	}
	if c.OTLPExporterEndpoint == "" {
		return &ConfigError{Field: "OTLPExporterEndpoint", Message: ErrInvalidExporterEndpoint}
	}
	if c.SamplingRatio < 0 || c.SamplingRatio > 1 {
		return &ConfigError{Field: "SamplingRatio", Message: ErrInvalidSamplingRatio}
	}
	if !contains(ValidSamplingTypes, c.SamplingType) {
		return &ConfigError{Field: "SamplingType", Message: ErrInvalidSamplingType}
	}
	if !contains(ValidOTLPProtocols, c.OTLPExporterProtocol) {
		return &ConfigError{Field: "OTLPExporterProtocol", Message: ErrInvalidExporterProtocol}
	}

	return nil
}

// WithEnvironment sets the deployment environment
func (c *Config) WithEnvironment(env string) *Config {
	c.Environment = env
	return c
}

// WithOTLPExporter configures the OTLP exporter (endpoint, insecure mode, and protocol)
func (c *Config) WithOTLPExporter(endpoint string, insecure bool, protocol string) *Config {
	c.OTLPExporterEndpoint = endpoint
	c.OTLPExporterInsecure = insecure
	c.OTLPExporterProtocol = protocol
	return c
}

// WithSampling configures the sampling strategy
func (c *Config) WithSampling(samplingType string, ratio float64) *Config {
	c.SamplingType = samplingType
	c.SamplingRatio = ratio
	return c
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func generateInstanceID() string {
	return uuid.NewString()
}
