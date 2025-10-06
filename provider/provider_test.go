package provider

import (
	"context"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/kernelshard/otelkit/internal/config"
)

func TestNewProviderConfig(t *testing.T) {
	serviceName := "test-service"
	serviceVersion := "1.0.0"

	pc := NewProviderConfig(serviceName, serviceVersion)

	if pc == nil {
		t.Fatal("NewProviderConfig returned nil")
	}
	if pc.Config == nil {
		t.Error("Config should not be nil")
	}
	if pc.Config.ServiceName != serviceName {
		t.Errorf("Expected ServiceName %s, got %s", serviceName, pc.Config.ServiceName)
	}
	if pc.Config.ServiceVersion != serviceVersion {
		t.Errorf("Expected ServiceVersion %s, got %s", serviceVersion, pc.Config.ServiceVersion)
	}

	// Check defaults
	if pc.BatchTimeout != config.DefaultBatchTimeout {
		t.Errorf("Expected BatchTimeout %v, got %v", config.DefaultBatchTimeout, pc.BatchTimeout)
	}
	if pc.ExportTimeout != config.DefaultExportTimeout {
		t.Errorf("Expected ExportTimeout %v, got %v", config.DefaultExportTimeout, pc.ExportTimeout)
	}
	if pc.MaxExportBatchSize != config.DefaultMaxExportBatchSize {
		t.Errorf("Expected MaxExportBatchSize %d, got %d", config.DefaultMaxExportBatchSize, pc.MaxExportBatchSize)
	}
	if pc.MaxQueueSize != config.DefaultMaxQueueSize {
		t.Errorf("Expected MaxQueueSize %d, got %d", config.DefaultMaxQueueSize, pc.MaxQueueSize)
	}
}

func TestProviderConfig_WithOTLPExporter(t *testing.T) {
	pc := NewProviderConfig("test", "1.0.0")
	endpoint := "jaeger:14250"
	protocol := "grpc"
	insecure := true

	result := pc.WithOTLPExporter(endpoint, protocol, insecure)

	// Should return same instance for chaining
	if result != pc {
		t.Error("WithOTLPExporter should return same instance")
	}

	if pc.Config.OTLPExporterEndpoint != endpoint {
		t.Errorf("Expected endpoint %s, got %s", endpoint, pc.Config.OTLPExporterEndpoint)
	}
	if pc.Config.OTLPExporterProtocol != protocol {
		t.Errorf("Expected protocol %s, got %s", protocol, pc.Config.OTLPExporterProtocol)
	}
	if pc.Config.OTLPExporterInsecure != insecure {
		t.Errorf("Expected insecure %v, got %v", insecure, pc.Config.OTLPExporterInsecure)
	}
}

func TestProviderConfig_WithSampling(t *testing.T) {
	pc := NewProviderConfig("test", "1.0.0")
	samplingType := config.SamplingAlwaysOn
	ratio := 1.0

	result := pc.WithSampling(samplingType, ratio)

	// Should return same instance for chaining
	if result != pc {
		t.Error("WithSampling should return same instance")
	}

	if pc.Config.SamplingType != samplingType {
		t.Errorf("Expected sampling type %s, got %s", samplingType, pc.Config.SamplingType)
	}
	if pc.Config.SamplingRatio != ratio {
		t.Errorf("Expected sampling ratio %f, got %f", ratio, pc.Config.SamplingRatio)
	}
}

func TestProviderConfig_WithBatchOptions(t *testing.T) {
	pc := NewProviderConfig("test", "1.0.0")
	batchTimeout := 10 * time.Second
	exportTimeout := 60 * time.Second
	maxBatchSize := 1000
	maxQueueSize := 4000

	result := pc.WithBatchOptions(batchTimeout, exportTimeout, maxBatchSize, maxQueueSize)

	// Should return same instance for chaining
	if result != pc {
		t.Error("WithBatchOptions should return same instance")
	}

	if pc.BatchTimeout != batchTimeout {
		t.Errorf("Expected BatchTimeout %v, got %v", batchTimeout, pc.BatchTimeout)
	}
	if pc.ExportTimeout != exportTimeout {
		t.Errorf("Expected ExportTimeout %v, got %v", exportTimeout, pc.ExportTimeout)
	}
	if pc.MaxExportBatchSize != maxBatchSize {
		t.Errorf("Expected MaxExportBatchSize %d, got %d", maxBatchSize, pc.MaxExportBatchSize)
	}
	if pc.MaxQueueSize != maxQueueSize {
		t.Errorf("Expected MaxQueueSize %d, got %d", maxQueueSize, pc.MaxQueueSize)
	}
}

func TestProviderConfig_WithResource(t *testing.T) {
	pc := NewProviderConfig("test", "1.0.0")
	ctx := context.Background()

	// Create a custom resource
	resource, err := sdkresource.New(ctx,
		sdkresource.WithFromEnv(),
	)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	result := pc.WithResource(resource)

	// Should return same instance for chaining
	if result != pc {
		t.Error("WithResource should return same instance")
	}

	if pc.Resource != resource {
		t.Error("Resource was not set correctly")
	}
}

func TestNewDefaultProvider(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service"
	serviceVersion := "2.0.0"

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	provider, err := NewDefaultProvider(ctx, serviceName, serviceVersion)
	if err != nil {
		t.Fatalf("NewDefaultProvider failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Just check that we can get a global provider
	globalProvider := otel.GetTracerProvider()
	if globalProvider == nil {
		t.Error("Global tracer provider should not be nil")
	}

	// Test shutdown
	if err := ShutdownTracerProvider(ctx, provider); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestNewDefaultProvider_WithDefaults(t *testing.T) {
	ctx := context.Background()
	serviceName := "test-service"

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	provider, err := NewDefaultProvider(ctx, serviceName)
	if err != nil {
		t.Fatalf("NewDefaultProvider failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	// Test shutdown
	if err := ShutdownTracerProvider(ctx, provider); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}
}

func TestNewProvider(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		config  *ProviderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid HTTP config",
			config: &ProviderConfig{
				Config: &config.Config{
					ServiceName:          "test-service",
					ServiceVersion:       "1.0.0",
					Environment:          "development",
					OTLPExporterEndpoint: "localhost:4318",
					OTLPExporterProtocol: "http",
					OTLPExporterInsecure: true,
					SamplingType:         config.SamplingProbabilistic,
					SamplingRatio:        0.5,
					InstanceID:           "test-instance",
					Hostname:             "test-host",
				},
				BatchTimeout:       5 * time.Second,
				ExportTimeout:      30 * time.Second,
				MaxExportBatchSize: 512,
				MaxQueueSize:       2048,
			},
			wantErr: false,
		},
		{
			name: "valid gRPC config",
			config: &ProviderConfig{
				Config: &config.Config{
					ServiceName:          "test-service",
					ServiceVersion:       "1.0.0",
					Environment:          "development",
					OTLPExporterEndpoint: "localhost:4317",
					OTLPExporterProtocol: "grpc",
					OTLPExporterInsecure: true,
					SamplingType:         config.SamplingAlwaysOn,
					SamplingRatio:        1.0,
					InstanceID:           "test-instance",
					Hostname:             "test-host",
				},
				BatchTimeout:       5 * time.Second,
				ExportTimeout:      30 * time.Second,
				MaxExportBatchSize: 512,
				MaxQueueSize:       2048,
			},
			wantErr: false,
		},
		{
			name: "invalid protocol",
			config: &ProviderConfig{
				Config: &config.Config{
					ServiceName:          "test-service",
					ServiceVersion:       "1.0.0",
					Environment:          "development",
					OTLPExporterEndpoint: "localhost:4317",
					OTLPExporterProtocol: "invalid",
					SamplingType:         config.SamplingProbabilistic,
					SamplingRatio:        0.5,
				},
			},
			wantErr: true,
			errMsg:  "must be 'grpc' or 'http'",
		},
	}

	// Save original global provider
	originalProvider := otel.GetTracerProvider()
	defer func() {
		otel.SetTracerProvider(originalProvider)
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewProvider(ctx, tt.config)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if provider == nil {
				t.Error("Provider should not be nil")
				return
			}

			// Test shutdown
			if err := ShutdownTracerProvider(ctx, provider); err != nil {
				t.Errorf("Shutdown failed: %v", err)
			}
		})
	}
}

func TestCreateSampler(t *testing.T) {
	tests := []struct {
		name         string
		config       *config.Config
		expectedType string // We can't easily check exact type, so we'll just ensure it doesn't panic
	}{
		{
			name: "probabilistic sampler",
			config: &config.Config{
				SamplingType:  config.SamplingProbabilistic,
				SamplingRatio: 0.5,
			},
		},
		{
			name: "always_on sampler",
			config: &config.Config{
				SamplingType:  config.SamplingAlwaysOn,
				SamplingRatio: 1.0,
			},
		},
		{
			name: "always_off sampler",
			config: &config.Config{
				SamplingType:  config.SamplingAlwaysOff,
				SamplingRatio: 0.0,
			},
		},
		{
			name: "invalid sampler falls back to probabilistic",
			config: &config.Config{
				SamplingType:  config.SamplingType("invalid"),
				SamplingRatio: 0.3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This should not panic
			sampler := createSampler(tt.config)
			if sampler == nil {
				t.Error("Sampler should not be nil")
			}
		})
	}
}

func TestShutdownTracerProvider(t *testing.T) {
	ctx := context.Background()

	// Test with nil provider
	err := ShutdownTracerProvider(ctx, nil)
	if err != nil {
		t.Errorf("Shutdown with nil provider should not return error, got: %v", err)
	}

	// Test with real provider
	provider, err := NewDefaultProvider(ctx, "test-shutdown")
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}

	err = ShutdownTracerProvider(ctx, provider)
	if err != nil {
		t.Errorf("Shutdown should not return error, got: %v", err)
	}
}

func TestProviderConfigFluentAPI(t *testing.T) {
	// Test chaining of fluent API
	pc := NewProviderConfig("test", "1.0.0").
		WithOTLPExporter("endpoint", "grpc", true).
		WithSampling(config.SamplingAlwaysOn, 1.0).
		WithBatchOptions(1*time.Second, 5*time.Second, 100, 1000)

	if pc.Config.OTLPExporterEndpoint != "endpoint" {
		t.Error("Fluent API chain failed for OTLP exporter")
	}
	if pc.Config.SamplingType != config.SamplingAlwaysOn {
		t.Error("Fluent API chain failed for sampling")
	}
	if pc.BatchTimeout != 1*time.Second {
		t.Error("Fluent API chain failed for batch options")
	}
}

func TestCreateResource_CustomResource(t *testing.T) {
	ctx := context.Background()
	pc := NewProviderConfig("test-service", "1.0.0")

	customRes, err := sdkresource.New(ctx,
		sdkresource.WithAttributes(),
	)
	if err != nil {
		t.Fatalf("Failed to create custom resource: %v", err)
	}

	pc.WithResource(customRes)

	res, err := createResource(ctx, pc)
	if err != nil {
		t.Fatalf("createResource returned error: %v", err)
	}

	if res != customRes {
		t.Error("Expected createResource to return the custom resource")
	}
}

func TestCreateResource_DefaultResource(t *testing.T) {
	ctx := context.Background()
	pc := NewProviderConfig("test-service", "1.0.0")

	res, err := createResource(ctx, pc)
	if err != nil {
		t.Fatalf("createResource returned error: %v", err)
	}

	if res == nil {
		t.Error("Expected createResource to return a resource")
	}
}

func TestCreateExporter_InvalidProtocol(t *testing.T) {
	ctx := context.Background()
	pc := NewProviderConfig("test-service", "1.0.0")
	pc.Config.OTLPExporterProtocol = "invalid"

	_, err := createExporter(ctx, pc)
	if err == nil {
		t.Error("Expected error for invalid OTLPExporterProtocol")
	}
}

func TestCreateBatchProcessor_Defaults(t *testing.T) {
	pc := NewProviderConfig("test-service", "1.0.0")
	pc.BatchTimeout = 0
	pc.ExportTimeout = 0
	pc.MaxExportBatchSize = 0
	pc.MaxQueueSize = 0

	exporter := &mockExporter{}

	sp := createBatchProcessor(exporter, pc)
	if sp == nil {
		t.Error("Expected createBatchProcessor to return a SpanProcessor")
	}
}

// mockExporter implements sdktrace.SpanExporter for testing
type mockExporter struct{}

func (m *mockExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	return nil
}

func (m *mockExporter) Shutdown(ctx context.Context) error {
	return nil
}

func TestNewProvider_Integration(t *testing.T) {
	ctx := context.Background()
	pc := NewProviderConfig("test-service", "1.0.0")

	tp, err := newProvider(ctx, pc)
	if err != nil {
		t.Fatalf("newProvider returned error: %v", err)
	}
	if tp == nil {
		t.Error("Expected newProvider to return a TracerProvider")
	}
}
