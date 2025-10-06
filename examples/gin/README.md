# Gin Framework with OpenTelemetry Example

This example demonstrates how to integrate OpenTelemetry tracing with the [Gin](https://gin-gonic.com/) web framework using OtelKit and the official `otelgin` middleware.

## Features

- ✅ **Automatic HTTP Request Tracing** using `otelgin.Middleware`
- ✅ **Custom Business Logic Spans** with context propagation
- ✅ **Error Tracing** with proper status codes and error recording
- ✅ **Span Attributes** for detailed trace context
- ✅ **Event Recording** for important operations
- ✅ **Best Practices** following OpenTelemetry standards

## Prerequisites

- Go 1.22 or later
- Running OTLP collector (Jaeger, or any OTLP-compatible backend)

## Installation

```bash
# Install dependencies
go get github.com/gin-gonic/gin
go get go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
go get github.com/kernelshard/otelkit
```

## Quick Start

### 1. Start an OTLP Collector (Jaeger)

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### 2. Run the Example

```bash
go run main.go
```

### 3. Test the Endpoints

```bash
# Health check
curl http://localhost:8080/health

# Get all users
curl http://localhost:8080/api/users

# Get specific user
curl http://localhost:8080/api/users/1

# Test error handling (returns 404)
curl http://localhost:8080/api/users/error

# Create a new user
curl -X POST http://localhost:8080/api/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Charlie","email":"charlie@example.com"}'
```

### 4. View Traces

Open Jaeger UI: http://localhost:16686

## How It Works

### Automatic HTTP Tracing

The example uses the official `otelgin.Middleware` to automatically create spans for all HTTP requests:

```go
r := gin.Default()
r.Use(otelgin.Middleware("gin-example"))
```

This middleware:
- Creates a span for each HTTP request
- Captures HTTP method, URL, status code
- Propagates trace context
- Handles errors automatically

### Custom Business Logic Spans

For detailed tracing of business logic, create custom child spans:

```go
api.GET("/users", func(c *gin.Context) {
    // Create a child span for business logic
    ctx, span := createCustomSpan(c, "fetch-all-users")
    defer span.End()
    
    // Add custom attributes
    span.AddEvent("querying database")
    span.SetAttributes(attribute.String("db.operation", "SELECT"))
    
    // Your business logic here...
    users := fetchUsers()
    
    span.SetAttributes(attribute.Int("user.count", len(users)))
    c.JSON(http.StatusOK, users)
})
```

### Error Handling

Properly record errors with trace context:

```go
if err != nil {
    span.SetStatus(codes.Error, "User not found")
    span.RecordError(err)
    c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
    return
}
span.SetStatus(codes.Ok, "User fetched successfully")
```

## Configuration

### Environment Variables

You can configure the OTLP exporter using environment variables:

```bash
export OTEL_SERVICE_NAME=gin-example
export OTEL_SERVICE_VERSION=1.0.0
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_PROTOCOL=grpc
export OTEL_TRACES_SAMPLER=probabilistic
export OTEL_TRACES_SAMPLER_ARG=1.0
```

Then use:

```go
shutdown, err := otelkit.SetupTracing(ctx, "gin-example", "1.0.0")
defer shutdown(ctx)
```

### Programmatic Configuration

For more control, use the provider configuration:

```go
config := otelkit.NewProviderConfig("gin-example", "1.0.0").
    WithOTLPExporter("localhost:4317", "grpc", true).
    WithSampling("probabilistic", 1.0)

provider, err := otelkit.NewProvider(ctx, config)
defer provider.Shutdown(ctx)
```

## Integration with Other Services

### Database Tracing

For database operations, use `otelsql`:

```go
import "go.nhat.io/otelsql"

func initDB() (*sql.DB, error) {
    driverName, err := otelsql.Register("postgres",
        otelsql.AllowRoot(),
        otelsql.TraceQueryWithoutArgs(),
        otelsql.WithDatabaseName("mydb"),
    )
    
    db, err := sql.Open(driverName, dsn)
    if err := otelsql.RecordStats(db); err != nil {
        return nil, err
    }
    
    return db, nil
}
```

### Redis Tracing

For Redis operations, use `redisotel`:

```go
import "github.com/redis/go-redis/extra/redisotel/v9"

func initRedis() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    
    // Setup traces for redis
    if err := redisotel.InstrumentTracing(rdb); err != nil {
        log.Fatalf("failed to instrument Redis: %v", err)
    }
    
    return rdb
}
```

## Production Recommendations

### 1. Use Appropriate Sampling

For production, use lower sampling rates to reduce overhead:

```go
config.WithSampling("probabilistic", 0.01) // 1% sampling
```

### 2. Add Custom Attributes

Add business-relevant attributes to spans:

```go
span.SetAttributes(
    attribute.String("user.id", userID),
    attribute.String("tenant.id", tenantID),
    attribute.String("api.version", "v1"),
)
```

### 3. Handle Shutdown Gracefully

Always ensure proper shutdown to flush pending spans:

```go
defer func() {
    shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := provider.Shutdown(shutdownCtx); err != nil {
        log.Printf("Error shutting down tracer provider: %v", err)
    }
}()
```

### 4. Use Structured Logging with Trace IDs

Correlate logs with traces by including trace IDs:

```go
import "go.opentelemetry.io/otel/trace"

func logWithTrace(c *gin.Context, message string) {
    span := trace.SpanFromContext(c.Request.Context())
    traceID := span.SpanContext().TraceID().String()
    log.Printf("[trace_id=%s] %s", traceID, message)
}
```

## Advanced Features

### Custom Middleware

Create custom middleware for specific needs:

```go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ctx, span := otelkit.New("gin-app").Start(c.Request.Context(), "auth-check")
        defer span.End()
        
        token := c.GetHeader("Authorization")
        span.SetAttributes(attribute.Bool("auth.present", token != ""))
        
        // Your auth logic...
        
        c.Next()
    }
}
```

### Distributed Tracing

The `otelgin` middleware automatically handles context propagation for distributed tracing:

```go
// Client side - make request to another service
client := otelkit.NewTracedHTTPClient(nil, tracer, "external-service")
resp, err := client.Get(c.Request.Context(), "http://other-service/api")
```

## Troubleshooting

### Traces Not Appearing

1. Check OTLP collector is running:
   ```bash
   curl http://localhost:4318/v1/traces  # Should return 405 Method Not Allowed
   ```

2. Verify endpoint configuration:
   ```bash
   export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
   ```

3. Check sampling rate (set to 1.0 for testing):
   ```go
   config.WithSampling("probabilistic", 1.0)
   ```

### Missing Spans

Ensure you're calling `defer span.End()` for all custom spans:

```go
ctx, span := createCustomSpan(c, "operation")
defer span.End()  // Don't forget this!
```

### Context Not Propagated

Always pass the context from the Gin context:

```go
ctx := c.Request.Context()  // Get context from request
ctx, span := tracer.Start(ctx, "operation")  // Use it for span creation
```

## References

- [OtelKit Documentation](https://github.com/kernelshard/otelkit)
- [Gin Framework](https://gin-gonic.com/)
- [otelgin Middleware](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin)
- [OpenTelemetry Go](https://opentelemetry.io/docs/languages/go/)
- [Last9 Gin Integration Guide](https://last9.io/docs/integrations-opentelemetry-gin/)

## License

This example is part of the OtelKit project and is licensed under the MIT License.
