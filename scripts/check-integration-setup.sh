#!/bin/bash

# Integration Test Setup Check Script
# This script verifies that the environment is ready for integration tests

set -e

echo "🔍 Checking OtelKit Integration Test Setup"
echo "=========================================="

# Check Docker
echo "Checking Docker..."
if command -v docker &> /dev/null; then
    echo "✅ Docker is installed"
    
    # Check Docker daemon is running
    if docker info &> /dev/null; then
        echo "✅ Docker daemon is running"
    else
        echo "❌ Docker daemon is not running"
        exit 1
    fi
else
    echo "❌ Docker is not installed"
    exit 1
fi

# Check Docker Compose
echo "Checking Docker Compose..."
if command -v docker-compose &> /dev/null; then
    echo "✅ Docker Compose is installed"
else
    echo "❌ Docker Compose is not installed"
    exit 1
fi

# Check Go version
echo "Checking Go version..."
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_GO="1.24"

if [ "$(printf '%s\n' "$REQUIRED_GO" "$GO_VERSION" | sort -V | head -n1)" = "$REQUIRED_GO" ]; then
    echo "✅ Go version $GO_VERSION is sufficient (requires $REQUIRED_GO+)"
else
    echo "❌ Go version $GO_VERSION is too old (requires $REQUIRED_GO+)"
    exit 1
fi

# Check if ports are available
echo "Checking port availability..."
PORTS=(4317 4318 16686 9411)

for port in "${PORTS[@]}"; do
    if lsof -i :$port > /dev/null 2>&1; then
        echo "⚠️  Port $port is already in use - may conflict with integration tests"
    else
        echo "✅ Port $port is available"
    fi
done

# Check if integration test files exist
echo "Checking integration test files..."
FILES=(
    "integration_test.go"
    "docker-compose.integration.yml" 
    "testdata/otel-collector-config.yaml"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo "✅ $file exists"
    else
        echo "❌ $file is missing"
        exit 1
    fi
done

# Check build tags
echo "Checking build tags..."
if grep -q "//go:build integration" integration_test.go; then
    echo "✅ Integration build tag found"
else
    echo "❌ Integration build tag missing from integration_test.go"
    exit 1
fi

echo ""
echo "🎉 Setup check completed successfully!"
echo ""
echo "Next steps:"
echo "1. Run 'make integration-up' to start the collector"
echo "2. Run 'make test-integration' to run integration tests"
echo "3. Run 'make integration-down' to stop the collector"
echo ""
echo "View traces at: http://localhost:16686 (Jaeger)"
echo "View traces at: http://localhost:9411 (Zipkin)"
echo "View metrics at: http://localhost:9090 (Prometheus)"
echo "View dashboards at: http://localhost:3000 (Grafana - admin/admin)"
