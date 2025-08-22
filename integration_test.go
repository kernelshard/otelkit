//go:build integration

package otelkit

import (
	"context"
	"net"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
)

// TestIntegration_HTTPExporter tests integration with an HTTP OTLP exporter
// This test requires a running OTLP collector on localhost:4318
func TestIntegration_HTTPExporter(t *testing.T) {
	if !isPortOpen("localhost:4318") {
		t.Skip("OTLP HTTP collector not available on localhost:4318")
	}

	ctx := context.Background()
	serviceName := "integration-http-service"
	serviceVersion := "1.0.0"

	// Create provider with HTTP exporter
	config := NewProviderConfig(serviceName, serviceVersion).
		WithOTLPExporter("localhost:4318", "http", true).
		WithSampling("always_on", 1.0)

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider)

	// Create tracer and test spans
	tracer := New("integration-test")
	_, span := tracer.Start(ctx, "integration-test-span")
	span.SetAttributes(attribute.String("test.attribute", "integration-value"))
	span.End()

	// Force flush to ensure spans are exported
	if err := provider.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed (may be expected): %v", err)
	}

	// Give some time for export
	time.Sleep(100 * time.Millisecond)

	// Test that we can create multiple spans
	for i := 0; i < 3; i++ {
		_, span := tracer.Start(ctx, "test-span", trace.WithAttributes(
			attribute.Int("iteration", i),
			attribute.String("test.type", "integration"),
		))
		span.End()
	}

	t.Logf("Integration test completed with HTTP exporter")
}

// TestIntegration_GRPCExporter tests integration with a gRPC OTLP exporter
// This test requires a running OTLP collector on localhost:4317
func TestIntegration_GRPCExporter(t *testing.T) {
	if !isPortOpen("localhost:4317") {
		t.Skip("OTLP gRPC collector not available on localhost:4317")
	}

	ctx := context.Background()
	serviceName := "integration-grpc-service"
	serviceVersion := "1.0.0"

	// Create provider with gRPC exporter
	config := NewProviderConfig(serviceName, serviceVersion).
		WithOTLPExporter("localhost:4317", "grpc", true).
		WithSampling("always_on", 1.0)

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider)

	// Test different sampling strategies
	tracer := New("integration-grpc-test")

	// Test probabilistic sampling
	_, span := tracer.Start(ctx, "probabilistic-test")
	span.SetAttributes(attribute.String("sampling.type", "probabilistic"))
	span.End()

	// Test batch processing with multiple spans
	for i := 0; i < 5; i++ {
		_, span := tracer.Start(ctx, "batch-test-span", trace.WithAttributes(
			attribute.Int("batch.index", i),
			attribute.String("exporter", "grpc"),
		))
		time.Sleep(10 * time.Millisecond) // Simulate some work
		span.End()
	}

	// Force flush
	if err := provider.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed: %v", err)
	}

	t.Logf("Integration test completed with gRPC exporter")
}

// TestIntegration_BatchProcessing tests batch processing behavior
func TestIntegration_BatchProcessing(t *testing.T) {
	if !isPortOpen("localhost:4318") {
		t.Skip("OTLP collector not available for batch processing test")
	}

	ctx := context.Background()
	serviceName := "batch-test-service"

	// Configure with short batch timeout for testing
	config := NewProviderConfig(serviceName, "1.0.0").
		WithOTLPExporter("localhost:4318", "http", true).
		WithSampling("always_on", 1.0).
		WithBatchOptions(100*time.Millisecond, 5*time.Second, 10, 100)

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider)

	tracer := New("batch-test")

	// Create spans that should be batched
	start := time.Now()
	for i := 0; i < 15; i++ {
		_, span := tracer.Start(ctx, "batch-span", trace.WithAttributes(
			attribute.Int("span.number", i),
			attribute.String("test.scenario", "batch-processing"),
		))
		span.End()
	}

	// Wait for batch timeout
	time.Sleep(150 * time.Millisecond)

	// Force flush any remaining spans
	if err := provider.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed: %v", err)
	}

	elapsed := time.Since(start)
	t.Logf("Batch processing test completed in %v", elapsed)
}

// TestIntegration_SamplingStrategies tests different sampling strategies
func TestIntegration_SamplingStrategies(t *testing.T) {
	if !isPortOpen("localhost:4318") {
		t.Skip("OTLP collector not available for sampling test")
	}

	ctx := context.Background()
	serviceName := "sampling-test-service"

	testCases := []struct {
		name        string
		sampling    string
		ratio       float64
		description string
	}{
		{"always_on", "always_on", 1.0, "Sample all traces"},
		{"always_off", "always_off", 0.0, "Sample no traces"},
		{"probabilistic_low", "probabilistic", 0.1, "Sample 10% of traces"},
		{"probabilistic_high", "probabilistic", 0.9, "Sample 90% of traces"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := NewProviderConfig(serviceName, "1.0.0").
				WithOTLPExporter("localhost:4318", "http", true).
				WithSampling(tc.sampling, tc.ratio)

			provider, err := NewProvider(ctx, config)
			if err != nil {
				t.Fatalf("Failed to create provider for %s: %v", tc.name, err)
			}
			defer ShutdownTracerProvider(ctx, provider)

			tracer := New("sampling-test")

			// Create test spans
			for i := 0; i < 10; i++ {
				_, span := tracer.Start(ctx, tc.name+"-span", trace.WithAttributes(
					attribute.String("sampling.strategy", tc.name),
					attribute.Float64("sampling.ratio", tc.ratio),
				))
				span.End()
			}

			// Force flush
			if err := provider.ForceFlush(ctx); err != nil {
				t.Logf("ForceFlush failed for %s: %v", tc.name, err)
			}

			t.Logf("Sampling test %s completed", tc.name)
		})
	}
}

// TestIntegration_ResourceAttributes tests resource attribute propagation
func TestIntegration_ResourceAttributes(t *testing.T) {
	if !isPortOpen("localhost:4318") {
		t.Skip("OTLP collector not available for resource test")
	}

	ctx := context.Background()
	serviceName := "resource-test-service"
	serviceVersion := "2.3.4"

	config := NewProviderConfig(serviceName, serviceVersion).
		WithOTLPExporter("localhost:4318", "http", true).
		WithSampling("always_on", 1.0)

	// Set custom environment for resource detection
	config.Config.Environment = "integration-test"
	config.Config.InstanceID = "test-instance-123"

	provider, err := NewProvider(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider)

	tracer := New("resource-test")

	// Create spans that should include resource attributes
	_, span := tracer.Start(ctx, "resource-test-span")
	span.SetAttributes(attribute.String("test.resource", "custom-attributes"))
	span.End()

	// Force flush
	if err := provider.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed: %v", err)
	}

	t.Logf("Resource attributes test completed")
}

// TestIntegration_MultipleProviders tests multiple provider instances
func TestIntegration_MultipleProviders(t *testing.T) {
	if !isPortOpen("localhost:4318") {
		t.Skip("OTLP collector not available for multiple providers test")
	}

	ctx := context.Background()

	// Create first provider
	provider1, err := NewProvider(ctx, NewProviderConfig("service-1", "1.0.0").
		WithOTLPExporter("localhost:4318", "http", true))
	if err != nil {
		t.Fatalf("Failed to create first provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider1)

	// Create second provider with different configuration
	provider2, err := NewProvider(ctx, NewProviderConfig("service-2", "2.0.0").
		WithOTLPExporter("localhost:4318", "http", true).
		WithSampling("probabilistic", 0.5))
	if err != nil {
		t.Fatalf("Failed to create second provider: %v", err)
	}
	defer ShutdownTracerProvider(ctx, provider2)

	tracer1 := New("multi-provider-1")
	tracer2 := New("multi-provider-2")

	// Create spans from both providers
	_, span1 := tracer1.Start(ctx, "provider-1-span")
	span1.End()

	_, span2 := tracer2.Start(ctx, "provider-2-span")
	span2.End()

	// Force flush both providers
	if err := provider1.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed for provider 1: %v", err)
	}
	if err := provider2.ForceFlush(ctx); err != nil {
		t.Logf("ForceFlush failed for provider 2: %v", err)
	}

	t.Logf("Multiple providers test completed")
}

// TestIntegration_InMemoryExporter is a fallback test that doesn't require external collector
// but still tests the full integration pipeline with in-memory exporter
func TestIntegration_InMemoryExporter(t *testing.T) {
	ctx := context.Background()
	serviceName := "in-memory-test-service"

	// Use in-memory exporter for testing without external dependencies
	exporter := tracetest.NewInMemoryExporter()

	// Create a custom provider config that uses our in-memory exporter
	config := &ProviderConfig{
		Config: &Config{
			ServiceName:    serviceName,
			ServiceVersion: "1.0.0",
			Environment:    "test",
			SamplingType:   "always_on",
			SamplingRatio:  1.0,
		},
	}

	// Create resource
	resource, err := createTestResource(config.Config)
	if err != nil {
		t.Fatalf("Failed to create resource: %v", err)
	}

	// Create sampler
	sampler := createSampler(config.Config)

	// Create batch processor with in-memory exporter
	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithBatchTimeout(10*time.Millisecond),
		sdktrace.WithExportTimeout(5*time.Second),
		sdktrace.WithMaxExportBatchSize(10),
		sdktrace.WithMaxQueueSize(100),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sampler),
		sdktrace.WithSpanProcessor(bsp),
	)

	defer tp.Shutdown(ctx)

	// Set as global provider
	otel.SetTracerProvider(tp)
	tracer := New("in-memory-test")

	// Test span creation and export
	_, span := tracer.Start(ctx, "in-memory-test-span")
	span.SetAttributes(attribute.String("test.type", "in-memory"))
	span.End()

	// Force flush
	tp.ForceFlush(ctx)

	// Verify spans were exported
	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Error("No spans were exported to in-memory exporter")
	}

	if len(spans) > 0 {
		span := spans[0]
		if span.Name != "in-memory-test-span" {
			t.Errorf("Expected span name 'in-memory-test-span', got '%s'", span.Name)
		}
	}

	t.Logf("In-memory exporter test completed with %d spans", len(spans))
}

// Helper function to check if a port is open (for collector detection)
func isPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Helper function to create test resource
func createTestResource(cfg *Config) (*sdkresource.Resource, error) {
	// Use the same logic as in provider.go but simplified for testing
	ctx := context.Background()
	attrs := []attribute.KeyValue{
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", cfg.ServiceVersion),
		attribute.String("deployment.environment", cfg.Environment),
		attribute.String("service.instance.id", cfg.InstanceID),
		attribute.String("host.name", cfg.Hostname),
	}

	return sdkresource.New(ctx,
		sdkresource.WithAttributes(attrs...),
	)
}
