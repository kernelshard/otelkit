# Tracing Examples

This directory contains comprehensive examples demonstrating how to use the tracing package in various scenarios.

## Examples Overview

### 1. Basic HTTP Server (`basic/`)
A simple HTTP server with OpenTelemetry integration showing:
- Basic tracer setup
- HTTP middleware usage
- Manual span creation
- Attribute setting
- Context propagation

### 2. Database Integration (`database/`)
Shows how to trace database operations:
- SQL query tracing
- Connection pool instrumentation
- Transaction spans
- Error handling and recording

### 3. gRPC Service (`grpc/`)
Demonstrates gRPC service tracing:
- Server-side interceptors
- Client-side interceptors
- Metadata propagation
- Streaming call tracing

### 4. Kafka Integration (`kafka/`)
Message queue tracing example:
- Producer span creation
- Consumer span linking
- Message headers for trace context
- Batch processing spans

### 5. Microservices (`microservices/`)
Complete microservices setup:
- Service mesh tracing
- Distributed context propagation
- Service dependency visualization
- Error tracking across services

### 6. Advanced Configuration (`advanced/`)
Advanced usage patterns:
- Custom samplers
- Resource attributes
- Multiple exporters
- Batch processor configuration

## Quick Start

1. Start Jaeger for trace collection:
```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

2. Run any example:
```bash
cd examples/basic
go run main.go
```

3. View traces at http://localhost:16686

## Environment Variables

All examples support standard OpenTelemetry environment variables:
- `OTEL_SERVICE_NAME`: Service identifier
- `OTEL_EXPORTER_OTLP_ENDPOINT`: Collector endpoint
- `OTEL_TRACES_SAMPLER`: Sampling strategy
- `OTEL_RESOURCE_ATTRIBUTES`: Additional resource attributes
