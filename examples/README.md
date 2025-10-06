# OtelKit Examples

This directory contains comprehensive examples demonstrating how to use OtelKit with various frameworks and scenarios.

## Examples Overview

### 1. Basic HTTP Server (`basic/`)
A simple HTTP server with OpenTelemetry integration showing:
- Basic tracer setup with OtelKit
- HTTP middleware usage
- Manual span creation
- Attribute setting
- Context propagation
- Error handling

**Quick Start:**
```bash
cd basic
go run main.go
curl http://localhost:8080/hello
```

### 2. Gin Framework (`gin/`)
**Complete Gin framework integration using official `otelgin` middleware:**
- Automatic HTTP request tracing with `otelgin.Middleware`
- Custom business logic spans
- Error tracing with proper status codes
- Span attributes and events
- Production-ready patterns

**Features:**
- ✅ Official OpenTelemetry Gin middleware
- ✅ Custom span creation for business logic
- ✅ Error handling best practices
- ✅ Database and Redis integration examples
- ✅ Distributed tracing support

**Quick Start:**
```bash
cd gin
go run main.go
curl http://localhost:8080/api/users
```

See [gin/README.md](gin/README.md) for detailed documentation.

### 3. Production Example (`production/`)
Production-ready setup demonstrating:
- Advanced configuration
- Database connection pooling
- Health checks and metrics
- Graceful shutdown
- Environment-based configuration
- Error handling patterns

**Quick Start:**
```bash
cd production
go run main.go
```

### 4. Traced HTTP Client (`traced_http_client/`)
Shows how to trace outbound HTTP requests:
- Automatic context propagation
- External service call tracking
- Request/response attributes
- Error handling for network calls

**Quick Start:**
```bash
cd traced_http_client
go run main.go
```

### 5. Dummy Service (`dummy_service/`)
Simple service for testing trace collection:
- Minimal setup
- Quick testing
- Basic span creation

## Quick Start

### 1. Start Jaeger for Trace Collection

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### 2. Run Any Example

```bash
# Basic example
cd examples/basic && go run main.go

# Gin framework example
cd examples/gin && go run main.go

# Production example
cd examples/production && go run main.go
```

### 3. View Traces

Open Jaeger UI: **http://localhost:16686**

## Configuration Options

All examples support both environment variables and programmatic configuration.

### Environment Variables

```bash
export OTEL_SERVICE_NAME=my-service
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_TRACES_SAMPLER=probabilistic
export OTEL_TRACES_SAMPLER_ARG=0.1
```

### Programmatic Configuration

```go
config := otelkit.NewProviderConfig("my-service", "1.0.0").
    WithOTLPExporter("localhost:4317", "grpc", true).
    WithSampling("probabilistic", 0.1)

provider, err := otelkit.NewProvider(ctx, config)
```

## Framework Integrations

### Gin Framework
Use the official `otelgin` middleware:
```bash
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
```

See [gin/README.md](gin/README.md) for complete guide.

### Database (SQL)
Use `otelsql` for automatic database tracing:
```bash
go get go.nhat.io/otelsql
```

### Redis
Use `redisotel` for Redis instrumentation:
```bash
# For go-redis v9
go get github.com/redis/go-redis/extra/redisotel/v9

# For go-redis v8
go get github.com/go-redis/redis/extra/redisotel/v8
```

## Common Patterns

### Simple Setup (Recommended)

```go
ctx := context.Background()
shutdown, err := otelkit.SetupTracing(ctx, "my-service", "1.0.0")
if err != nil {
    log.Fatal(err)
}
defer shutdown(ctx)

tracer := otelkit.New("my-service")
```

### Custom Configuration

```go
config := otelkit.NewProviderConfig("my-service", "1.0.0").
    WithOTLPExporter("localhost:4317", "grpc", true).
    WithSampling("probabilistic", 0.05)

provider, err := otelkit.NewProvider(ctx, config)
if err != nil {
    log.Fatal(err)
}
defer provider.Shutdown(ctx)
```

### Creating Spans

```go
ctx, span := tracer.Start(ctx, "operation-name")
defer span.End()

// Add attributes
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.Int("items.count", count),
)

// Record errors
if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, "operation failed")
}
```

## Testing Examples

Each example can be tested with curl or any HTTP client:

```bash
# Basic example
curl http://localhost:8080/hello
curl http://localhost:8080/error

# Gin example  
curl http://localhost:8080/api/users
curl http://localhost:8080/api/users/1
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Alice","email":"alice@example.com"}'

# Production example
curl http://localhost:8080/health
curl http://localhost:8080/api/users
curl http://localhost:8080/metrics
```

## Troubleshooting

### Traces Not Appearing

1. **Check OTLP collector is running:**
   ```bash
   docker ps | grep jaeger
   curl http://localhost:4318/v1/traces  # Should return 405
   ```

2. **Verify endpoint configuration:**
   ```bash
   echo $OTEL_EXPORTER_OTLP_ENDPOINT
   ```

3. **Set sampling to 100% for testing:**
   ```go
   config.WithSampling("probabilistic", 1.0)
   ```

### Common Issues

- **Port conflicts**: Check if ports 4317, 4318, or 8080 are already in use
- **Missing spans**: Ensure `defer span.End()` is called
- **Context not propagated**: Always pass context through function calls

## Additional Resources

- [OtelKit Documentation](https://github.com/kernelshard/otelkit)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [Integration Tests Guide](../INTEGRATION_TESTS.md)

## Contributing

Want to add a new example? Great! Please:
1. Follow the existing structure
2. Include a README.md with usage instructions
3. Add tests where applicable
4. Update this main README

## License

All examples are part of the OtelKit project and are licensed under the MIT License.
