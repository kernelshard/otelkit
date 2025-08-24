package config

import (
	"os"
	"testing"
	"time"
)

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name           string
		serviceName    string
		serviceVersion string
		wantName       string
		wantVersion    string
	}{
		{
			name:           "basic config creation",
			serviceName:    "test-service",
			serviceVersion: "1.0.0",
			wantName:       "test-service",
			wantVersion:    "1.0.0",
		},
		{
			name:           "empty service name",
			serviceName:    "",
			serviceVersion: "1.0.0",
			wantName:       "",
			wantVersion:    "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewConfig(tt.serviceName, tt.serviceVersion)

			if cfg.ServiceName != tt.wantName {
				t.Errorf("NewConfig() ServiceName = %v, want %v", cfg.ServiceName, tt.wantName)
			}
			if cfg.ServiceVersion != tt.wantVersion {
				t.Errorf("NewConfig() ServiceVersion = %v, want %v", cfg.ServiceVersion, tt.wantVersion)
			}

			// Check defaults
			if cfg.Environment != DefaultEnvironment {
				t.Errorf("NewConfig() Environment = %v, want %v", cfg.Environment, DefaultEnvironment)
			}
			if cfg.OTLPExporterEndpoint != DefaultOTLPExporterEndpoint {
				t.Errorf("NewConfig() OTLPExporterEndpoint = %v, want %v", cfg.OTLPExporterEndpoint, DefaultOTLPExporterEndpoint)
			}
			if cfg.SamplingRatio != DefaultSamplingRatio {
				t.Errorf("NewConfig() SamplingRatio = %v, want %v", cfg.SamplingRatio, DefaultSamplingRatio)
			}
			if cfg.InstanceID == "" {
				t.Error("NewConfig() InstanceID should not be empty")
			}
		})
	}
}

func TestNewConfigFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"OTEL_SERVICE_NAME",
		"OTEL_SERVICE_VERSION",
		"OTEL_ENVIRONMENT",
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_INSECURE",
		"OTEL_TRACES_SAMPLER",
		"OTEL_TRACES_SAMPLER_ARG",
	}

	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			originalEnv[env] = val
		}
		os.Unsetenv(env)
	}

	// Restore environment after test
	defer func() {
		for _, env := range envVars {
			os.Unsetenv(env)
			if val, exists := originalEnv[env]; exists {
				os.Setenv(env, val)
			}
		}
	}()

	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(*testing.T, *Config)
	}{
		{
			name:    "default config from empty env",
			envVars: map[string]string{},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ServiceName != DefaultServiceName {
					t.Errorf("Expected ServiceName %v, got %v", DefaultServiceName, cfg.ServiceName)
				}
				if cfg.Environment != DefaultEnvironment {
					t.Errorf("Expected Environment %v, got %v", DefaultEnvironment, cfg.Environment)
				}
			},
		},
		{
			name: "custom config from env",
			envVars: map[string]string{
				"OTEL_SERVICE_NAME":           "custom-service",
				"OTEL_SERVICE_VERSION":        "2.0.0",
				"OTEL_ENVIRONMENT":            "production",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "jaeger:14250",
				"OTEL_EXPORTER_OTLP_INSECURE": "true",
				"OTEL_TRACES_SAMPLER":         "always_on",
				"OTEL_TRACES_SAMPLER_ARG":     "1.0",
			},
			validate: func(t *testing.T, cfg *Config) {
				if cfg.ServiceName != "custom-service" {
					t.Errorf("Expected ServiceName custom-service, got %v", cfg.ServiceName)
				}
				if cfg.Environment != "production" {
					t.Errorf("Expected Environment production, got %v", cfg.Environment)
				}
				if cfg.OTLPExporterEndpoint != "jaeger:14250" {
					t.Errorf("Expected OTLPExporterEndpoint jaeger:14250, got %v", cfg.OTLPExporterEndpoint)
				}
				if !cfg.OTLPExporterInsecure {
					t.Error("Expected OTLPExporterInsecure to be true")
				}
				if cfg.SamplingType != "always_on" {
					t.Errorf("Expected SamplingType always_on, got %v", cfg.SamplingType)
				}
				if cfg.SamplingRatio != 1.0 {
					t.Errorf("Expected SamplingRatio 1.0, got %v", cfg.SamplingRatio)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			cfg := NewConfigFromEnv()
			tt.validate(t, cfg)

			// Clean up environment variables for next test
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errType string
	}{
		{
			name: "valid config",
			config: &Config{
				ServiceName:          "test-service",
				ServiceVersion:       "1.0.0",
				Environment:          "development",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        0.5,
				SamplingType:         "probabilistic",
				OTLPExporterProtocol: "grpc",
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: &Config{
				ServiceName:          "",
				ServiceVersion:       "1.0.0",
				Environment:          "development",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        0.5,
				SamplingType:         "probabilistic",
				OTLPExporterProtocol: "grpc",
			},
			wantErr: true,
			errType: "ServiceName",
		},
		{
			name: "invalid environment",
			config: &Config{
				ServiceName:          "test-service",
				ServiceVersion:       "1.0.0",
				Environment:          "invalid",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        0.5,
				SamplingType:         "probabilistic",
				OTLPExporterProtocol: "grpc",
			},
			wantErr: true,
			errType: "Environment",
		},
		{
			name: "invalid sampling ratio",
			config: &Config{
				ServiceName:          "test-service",
				ServiceVersion:       "1.0.0",
				Environment:          "development",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        1.5,
				SamplingType:         "probabilistic",
				OTLPExporterProtocol: "grpc",
			},
			wantErr: true,
			errType: "SamplingRatio",
		},
		{
			name: "invalid sampling type",
			config: &Config{
				ServiceName:          "test-service",
				ServiceVersion:       "1.0.0",
				Environment:          "development",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        0.5,
				SamplingType:         "invalid",
				OTLPExporterProtocol: "grpc",
			},
			wantErr: true,
			errType: "SamplingType",
		},
		{
			name: "invalid protocol",
			config: &Config{
				ServiceName:          "test-service",
				ServiceVersion:       "1.0.0",
				Environment:          "development",
				OTLPExporterEndpoint: "localhost:4317",
				SamplingRatio:        0.5,
				SamplingType:         "probabilistic",
				OTLPExporterProtocol: "invalid",
			},
			wantErr: true,
			errType: "OTLPExporterProtocol",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr && err == nil {
				t.Error("Validate() expected error but got none")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}

			if tt.wantErr && err != nil {
				if configErr, ok := err.(*ConfigError); ok {
					if configErr.Field != tt.errType {
						t.Errorf("Expected error field %v, got %v", tt.errType, configErr.Field)
					}
				} else {
					t.Errorf("Expected ConfigError, got %T", err)
				}
			}
		})
	}
}

func TestConfig_FluentAPI(t *testing.T) {
	cfg := NewConfig("test-service", "1.0.0")

	result := cfg.WithEnvironment("production").
		WithOTLPExporter("jaeger:14250", true, "http").
		WithSampling("always_on", 1.0)

	// Should return the same instance for chaining
	if result != cfg {
		t.Error("Fluent API should return the same instance")
	}

	if cfg.Environment != "production" {
		t.Errorf("Expected Environment production, got %v", cfg.Environment)
	}
	if cfg.OTLPExporterEndpoint != "jaeger:14250" {
		t.Errorf("Expected OTLPExporterEndpoint jaeger:14250, got %v", cfg.OTLPExporterEndpoint)
	}
	if !cfg.OTLPExporterInsecure {
		t.Error("Expected OTLPExporterInsecure to be true")
	}
	if cfg.OTLPExporterProtocol != "http" {
		t.Errorf("Expected OTLPExporterProtocol http, got %v", cfg.OTLPExporterProtocol)
	}
	if cfg.SamplingType != "always_on" {
		t.Errorf("Expected SamplingType always_on, got %v", cfg.SamplingType)
	}
	if cfg.SamplingRatio != 1.0 {
		t.Errorf("Expected SamplingRatio 1.0, got %v", cfg.SamplingRatio)
	}
}

func TestGetEnvHelpers(t *testing.T) {
	tests := []struct {
		name         string
		envKey       string
		envValue     string
		defaultValue interface{}
		testFunc     func(t *testing.T)
	}{
		{
			name:         "getEnvInt with valid value",
			envKey:       "TEST_INT",
			envValue:     "42",
			defaultValue: 10,
			testFunc: func(t *testing.T) {
				result := getEnvInt("TEST_INT", 10)
				if result != 42 {
					t.Errorf("Expected 42, got %d", result)
				}
			},
		},
		{
			name:         "getEnvBool with true",
			envKey:       "TEST_BOOL",
			envValue:     "true",
			defaultValue: false,
			testFunc: func(t *testing.T) {
				result := getEnvBool("TEST_BOOL", false)
				if !result {
					t.Error("Expected true, got false")
				}
			},
		},
		{
			name:         "getEnvFloat with valid value",
			envKey:       "TEST_FLOAT",
			envValue:     "3.14",
			defaultValue: 1.0,
			testFunc: func(t *testing.T) {
				result := getEnvFloat("TEST_FLOAT", 1.0)
				if result != 3.14 {
					t.Errorf("Expected 3.14, got %f", result)
				}
			},
		},
		{
			name:         "getEnvDuration with valid value",
			envKey:       "TEST_DURATION",
			envValue:     "5m",
			defaultValue: time.Minute,
			testFunc: func(t *testing.T) {
				result := getEnvDuration("TEST_DURATION", time.Minute)
				expected := 5 * time.Minute
				if result != expected {
					t.Errorf("Expected %v, got %v", expected, result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv(tt.envKey, tt.envValue)
			defer os.Unsetenv(tt.envKey)

			tt.testFunc(t)
		})
	}
}

func TestGenerateInstanceID(t *testing.T) {
	id1 := generateInstanceID()
	id2 := generateInstanceID()

	if id1 == "" {
		t.Error("generateInstanceID() should not return empty string")
	}
	if id1 == id2 {
		t.Error("generateInstanceID() should return unique IDs")
	}
}
