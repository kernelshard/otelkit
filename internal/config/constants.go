// Package otelkit provides a simple, production-ready OpenTelemetry tracing library for Go.
//
// This file contains all constants used throughout the library for consistent
// string management and maintainability.

package config

import (
	"net/http"
	"time"
)

// Service configuration constants
const (
	DefaultServiceName          = "unknown-service"
	DefaultServiceVersion       = "1.0.0"
	DefaultEnvironment          = "development"
	DefaultOTLPExporterEndpoint = "localhost:4317"
	DefaultSamplingRatio        = 0.2
	DefaultSamplingType         = "probabilistic"
	DefaultOTLPExporterProtocol = "grpc"
	DefaultBatchTimeout         = 5 * time.Second
	DefaultExportTimeout        = 30 * time.Second
	DefaultMaxExportBatchSize   = 512
	DefaultMaxQueueSize         = 2048
)

// Valid configuration options
var (
	ValidEnvironments  = []string{"development", "staging", "production"}
	ValidSamplingTypes = []string{"probabilistic", "always_on", "always_off"}
	ValidHTTPMethods   = []string{
		http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions,
	}
	ValidOTLPProtocols = []string{"grpc", "http"}
)

// OpenTelemetry semantic convention constants
const (
	AttrHTTPMethod     = "http.method"
	AttrHTTPURL        = "http.url"
	AttrHTTPUserAgent  = "http.user_agent"
	AttrHTTPStatusCode = "http.status_code"
)

// Error message constants
const (
	ErrServiceNameRequired     = "service name is required"
	ErrServiceVersionRequired  = "service version is required"
	ErrInvalidEnvironment      = "invalid environment"
	ErrInvalidSamplingType     = "invalid sampling type"
	ErrInvalidSamplingRatio    = "sampling ratio must be between 0 and 1"
	ErrInvalidExporterProtocol = "invalid exporter protocol"
	ErrInvalidExporterEndpoint = "exporter endpoint is required"
)

// Environment variable constants
const (
	EnvServiceName          = "OTEL_SERVICE_NAME"
	EnvServiceVersion       = "OTEL_SERVICE_VERSION"
	EnvEnvironment          = "OTEL_ENVIRONMENT"
	EnvOTLPExporterEndpoint = "OTEL_EXPORTER_OTLP_ENDPOINT"
	EnvOTLPExporterInsecure = "OTEL_EXPORTER_OTLP_INSECURE"
	EnvOTLPExporterProtocol = "OTEL_EXPORTER_OTLP_PROTOCOL"
	EnvBatchTimeout         = "OTEL_BSP_TIMEOUT"
	EnvExportTimeout        = "OTEL_EXPORTER_TIMEOUT"
	EnvMaxExportBatchSize   = "OTEL_BSP_MAX_EXPORT_BATCH_SIZE"
	EnvMaxQueueSize         = "OTEL_BSP_MAX_QUEUE_SIZE"
	EnvSamplingType         = "OTEL_TRACES_SAMPLER"
	EnvSamplingRatio        = "OTEL_TRACES_SAMPLER_ARG"
	EnvInstanceID           = "OTEL_RESOURCE_ATTRIBUTES_SERVICE_INSTANCE_ID"
)
