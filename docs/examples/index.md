# Examples

Practical, ready-to-use code examples for common use cases.

## Available Examples

### Basic Example

Simple HTTP server with automatic tracing.

[**View Basic Example →**](basic.md){ .md-button .md-button--primary }

Learn the fundamentals of setting up otelkit with a basic HTTP server.

---

### Gin Framework

Integration with the popular Gin web framework.

[**View Gin Example →**](gin.md){ .md-button }

See how to add tracing to your Gin applications.

---

### Production Setup

Production-ready configuration with all best practices.

[**View Production Example →**](production.md){ .md-button }

Complete example with error handling, custom sampling, and exporters.

---

### Traced HTTP Client

HTTP client with automatic trace propagation.

[**View HTTP Client Example →**](traced_http_client.md){ .md-button }

Learn how to trace outgoing HTTP requests and propagate context.

---

## Running Examples

All examples can be run directly:

```bash
cd examples/basic
go run main.go
```

Then view traces at [http://localhost:16686](http://localhost:16686) (Jaeger UI).
