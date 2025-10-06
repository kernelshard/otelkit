package tracer

import (
	"context"
	"os"
	"testing"

	"go.opentelemetry.io/otel"

	"github.com/kernelshard/otelkit/internal/config"
	"github.com/kernelshard/otelkit/provider"
)

func TestSetupTracing(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-setup-service"
	serviceVersion := "1.2.3"

	// Save original global provider and environment
	originalProvider := otel.GetTracerProvider()
	originalEnv := make(map[string]string)
	envVars := []string{
		"OTEL_SERVICE_NAME",
		"OTEL_SERVICE_VERSION",
		"OTEL_ENVIRONMENT",
	}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}

	defer func() {
		otel.SetTracerProvider(originalProvider)
		for _, env := range envVars {
			os.Unsetenv(env)
			if val, exists := originalEnv[env]; exists {
				os.Setenv(env, val)
			}
		}
	}()

	// Test successful setup
	shutdown, err := SetupTracing(ctx, serviceName, serviceVersion)
	if err != nil {
		t.Fatalf("SetupTracing failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("Shutdown function should not be nil")
	}

	// Just check that we can get a global provider (may be the same reference)
	globalProvider := otel.GetTracerProvider()
	if globalProvider == nil {
		t.Error("Global tracer provider should not be nil")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestSetupTracing_WithDefaults(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-setup-service"

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// Test with default version
	shutdown, err := SetupTracing(ctx, serviceName)
	if err != nil {
		t.Fatalf("SetupTracing failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("Shutdown function should not be nil")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestSetupTracing_InvalidConfig(t *testing.T) {
	ctx := context.Background()
	serviceName := "" // Invalid service name

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// Test with invalid config
	shutdown, err := SetupTracing(ctx, serviceName)
	if err == nil {
		t.Error("Expected error for invalid service name")
	}
	if shutdown != nil {
		t.Error("Shutdown function should be nil on error")
	}
}

func TestSetupTracingWithDefaults(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-defaults-service"
	serviceVersion := "2.0.0"

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	shutdown, err := SetupTracingWithDefaults(ctx, serviceName, serviceVersion)
	if err != nil {
		t.Fatalf("SetupTracingWithDefaults failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("Shutdown function should not be nil")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestMustSetupTracing(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-must-service"

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// This should not panic with valid input
	shutdown := MustSetupTracing(ctx, serviceName)
	if shutdown == nil {
		t.Fatal("Shutdown function should not be nil")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestMustSetupTracing_Panic(t *testing.T) {
	ctx := context.Background()
	serviceName := "" // Invalid service name

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// This should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustSetupTracing should panic with invalid config")
		}
	}()

	MustSetupTracing(ctx, serviceName)
}

func TestSetupCustomTracing(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		config  *provider.ProviderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &provider.ProviderConfig{
				Config: &config.Config{
					ServiceName:          "test-custom",
					ServiceVersion:       "1.0.0",
					Environment:          "development",
					OTLPExporterEndpoint: "localhost:4318",
					OTLPExporterProtocol: "http",
					OTLPExporterInsecure: true,
					SamplingType:         "probabilistic",
					SamplingRatio:        0.5,
					InstanceID:           "test-instance",
					Hostname:             "test-host",
				},
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "provider config cannot be nil",
		},
		{
			name: "nil inner config",
			config: &provider.ProviderConfig{
				Config: nil,
			},
			wantErr: true,
			errMsg:  "tracer config cannot be nil",
		},
		{
			name: "invalid inner config",
			config: &provider.ProviderConfig{
				Config: &config.Config{
					ServiceName:    "", // Invalid
					ServiceVersion: "1.0.0",
					Environment:    "development",
				},
			},
			wantErr: true,
		},
	}

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp, err := SetupCustomTracing(ctx, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tp != nil {
					t.Error("Provider should be nil on error")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tp == nil {
				t.Error("Provider should not be nil")
				return
			}

			// Test shutdown
			if err := provider.ShutdownTracerProvider(ctx, tp); err != nil {
				t.Errorf("Shutdown failed: %v", err)
			}
		})
	}
}

func TestSetupTracingEnvironmentVariables(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-env-service"

	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"OTEL_SERVICE_NAME",
		"OTEL_ENVIRONMENT",
		"OTEL_TRACES_SAMPLER",
		"OTEL_TRACES_SAMPLER_ARG",
	}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}

	defer func() {
		for _, env := range envVars {
			os.Unsetenv(env)
			if val, exists := originalEnv[env]; exists {
				os.Setenv(env, val)
			}
		}
	}()

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// Set environment variables
	os.Setenv("OTEL_ENVIRONMENT", "production")
	os.Setenv("OTEL_TRACES_SAMPLER", "always_on")
	os.Setenv("OTEL_TRACES_SAMPLER_ARG", "1.0")

	shutdown, err := SetupTracing(ctx, serviceName)
	if err != nil {
		t.Fatalf("SetupTracing failed: %v", err)
	}
	if shutdown == nil {
		t.Fatal("Shutdown function should not be nil")
	}

	// Test shutdown
	if err := shutdown(ctx); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestSetupFunctions_Integration(t *testing.T) {
	ctx := context.Background()

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	// Test that all setup functions work together
	tests := []struct {
		name string
		fn   func() (func(context.Context) error, error)
	}{
		{
			name: "SetupTracing",
			fn: func() (func(context.Context) error, error) {
				return SetupTracing(ctx, "integration-test-1")
			},
		},
		{
			name: "SetupTracingWithDefaults",
			fn: func() (func(context.Context) error, error) {
				return SetupTracingWithDefaults(ctx, "integration-test-2", "1.0.0")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shutdown, err := tt.fn()
			if err != nil {
				t.Errorf("%s failed: %v", tt.name, err)
				return
			}

			if shutdown == nil {
				t.Errorf("%s returned nil shutdown function", tt.name)
				return
			}

			// Test that we can create a tracer
			tracer := New("test-tracer")
			if tracer == nil {
				t.Error("Failed to create tracer after setup")
			}

			// Test shutdown
			if err := shutdown(ctx); err != nil {
				t.Errorf("Shutdown failed for %s: %v", tt.name, err)
			}
		})
	}
}
