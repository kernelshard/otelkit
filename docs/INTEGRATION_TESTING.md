# Integration Testing Guide

This guide explains how to set up and run integration tests for OtelKit with real OpenTelemetry collectors.

## Overview

Integration tests verify that OtelKit works correctly with actual OpenTelemetry collectors (OTLP exporters). These tests require Docker and Docker Compose to run the collector infrastructure.

## Prerequisites

- Docker and Docker Compose installed
- Go 1.24+ (for integration test build tags)

## Running Integration Tests

### Option 1: Run All Tests (Including Integration Tests)

```bash
make test-all
```

This runs all tests, including integration tests that require the `integration` build tag.

### Option 2: Run Only Integration Tests

```bash
make test-integration
```

Runs only the integration tests (skips if collector is not available).

### Option 3: Full Integration Test with Collector

```bash
make test-integration-with-collector
```

This command:
1. Starts the OpenTelemetry collector using Docker Compose
2. Waits for the collector to be ready
3. Runs all integration tests
4. Stops the collector when tests complete

### Option 4: Manual Collector Management

Start the collector:
```bash
make integration-up
```

Run integration tests:
```bash
make test-integration
```

View collector logs:
```bash
make integration-logs
```

Stop the collector:
```bash
make integration-down
```

## Test Environment

The integration test environment includes:

- **OpenTelemetry Collector**: Listens on ports 4317 (gRPC) and 4318 (HTTP)
- **Jaeger UI**: Available at http://localhost:16686 for trace visualization
- **Zipkin**: Available at http://localhost:9411 for alternative trace visualization
- **Prometheus**: Available at http://localhost:9090 for metrics
- **Grafana**: Available at http://localhost:3000 for metrics visualization (admin/admin)

## Integration Test Types

### 1. HTTP Exporter Tests
Tests integration with OTLP HTTP exporter (port 4318)

### 2. gRPC Exporter Tests  
Tests integration with OTLP gRPC exporter (port 4317)

### 3. Batch Processing Tests
Tests batch processing behavior with different configurations

### 4. Sampling Strategy Tests
Tests different sampling strategies (always_on, always_off, probabilistic)

### 5. Resource Attribute Tests
Tests resource attribute propagation to collectors

### 6. Multiple Provider Tests
Tests multiple tracer provider instances

### 7. In-Memory Exporter Tests
Fallback tests that don't require external collectors

## Test Configuration

Integration tests use the `//go:build integration` build tag. Tests will automatically skip if:

- The required collector ports (4317, 4318) are not available
- Docker/Docker Compose is not installed
- The integration build tag is not specified

## Writing Integration Tests

### Basic Structure

```go
//go:build integration

package otelkit

import (
    "testing"
)

func TestIntegration_Example(t *testing.T) {
    if !isPortOpen("localhost:4318") {
        t.Skip("Collector not available")
    }
    
    // Test logic here
}
```

### Best Practices

1. **Always check for collector availability** using `isPortOpen()`
2. **Use descriptive test names** starting with `TestIntegration_`
3. **Clean up resources** with proper shutdown/defer statements
4. **Include logging** to help with debugging
5. **Test error scenarios** and edge cases

## Troubleshooting

### Collector Not Starting
- Check Docker is running: `docker info`
- Check ports 4317/4318 are not already in use
- View logs: `make integration-logs`

### Tests Skipping
- Ensure collector is running: `make integration-up`
- Wait for collector to be ready (5-10 seconds after startup)

### Connection Issues
- Verify firewall settings allow Docker networking
- Check Docker network configuration

### Performance Issues
Integration tests may be slower due to:
- Docker container startup time
- Network latency to collector
- Batch processing timeouts

## CI/CD Integration

For CI/CD pipelines, include:

```yaml
steps:
  - name: Run integration tests
    run: |
      make integration-up
      sleep 10  # Wait for collector
      make test-integration
      make integration-down
```

## Manual Testing

You can also test manually using curl:

```bash
# Test HTTP endpoint
curl -X GET http://localhost:4318/health

# Test gRPC endpoint (requires grpcurl)
grpcurl -plaintext localhost:4317 list
```

## Monitoring Test Results

After running integration tests, you can view the results in:

- **Jaeger UI**: http://localhost:16686
- **Zipkin UI**: http://localhost:9411  
- **Collector Logs**: `make integration-logs`

The integration tests create spans with specific attributes that can be filtered in the UI:

- `test.type`: "integration"
- `sampling.strategy`: indicates the sampling type tested
- `batch.index`: for batch processing tests

## Next Steps

- Add more comprehensive error scenario testing
- Include performance benchmarking
- Add load testing scenarios
- Test with different collector configurations
