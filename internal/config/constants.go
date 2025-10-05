// Package otelkit provides a simple, production-ready OpenTelemetry tracing library for Go.
//
// This file contains all constants used throughout the library for consistent
// string management and maintainability.

package config

import (
	"net/http"
	"time"
)

// SamplingType represents the type of sampling strategy to use.
// This custom type provides type safety and prevents runtime errors from typos.
type SamplingType string

// Defined sampling types for type safety
const (
	SamplingProbabilistic SamplingType = "probabilistic"
	SamplingAlwaysOn      SamplingType = "always_on"
	SamplingAlwaysOff     SamplingType = "always_off"
)

// String returns the string representation of the sampling type
func (s SamplingType) String() string {
	return string(s)
}

// IsValid checks if the sampling type is one of the defined constants
func (s SamplingType) IsValid() bool {
	switch s {
	case SamplingProbabilistic, SamplingAlwaysOn, SamplingAlwaysOff:
		return true
	default:
		return false
	}
}

// Service configuration constants
const (
	DefaultServiceName          = "unknown-service"
	DefaultServiceVersion       = "1.0.0"
	DefaultEnvironment          = "development"
	DefaultOTLPExporterEndpoint = "localhost:4318"
	DefaultSamplingRatio        = 0.2
	DefaultSamplingType         = SamplingProbabilistic
	DefaultOTLPExporterProtocol = "http"
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
		http.MethodDelete, http.MethodPatch, http.MethodOptions,
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
