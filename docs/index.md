---
title: Home
description: Welcome to otelkit - Seamless OpenTelemetry tracing for Go applications
---

<div class="hero">
  <h1>🚀 otelkit</h1>
  <p>HTTP middleware for automatic request tracing with OpenTelemetry integration</p>
  <a href="GETTING_STARTED.html" class="md-button md-button--primary">Get Started</a>
  <a href="https://github.com/kernelshard/otelkit" class="md-button">View on GitHub</a>
</div>

## ✨ Features

<div class="feature-grid">
  <div class="feature-card">
    <span class="twemoji">🔗</span>
    <h3>HTTP Middleware</h3>
    <p>Seamlessly integrate tracing into your HTTP handlers with zero configuration.</p>
  </div>
  <div class="feature-card">
    <span class="twemoji">⚡</span>
    <h3>Framework Agnostic</h3>
    <p>Works with gorilla/mux, chi, gin, echo, and any http.Handler compatible framework.</p>
  </div>
  <div class="feature-card">
    <span class="twemoji">📊</span>
    <h3>Rich Configuration</h3>
    <p>Flexible sampling, exporters, and batch processing options for production use.</p>
  </div>
  <div class="feature-card">
    <span class="twemoji">🛡️</span>
    <h3>Safe Operations</h3>
    <p>Built-in nil checks and error handling for robust tracing operations.</p>
  </div>
</div>

## 📦 Installation

```bash
go get github.com/kernelshard/otelkit
```

## 🚀 Quick Example

```go
package main

import (
    "net/http"
    "github.com/kernelshard/otelkit"
)

func main() {
    // Initialize tracing
    shutdown, err := otelkit.SetupTracing(ctx, "my-service")
    if err != nil {
        log.Fatal(err)
    }
    defer shutdown(ctx)

    // Create instrumented handler
    handler := otelkit.NewInstrumentedHTTPHandler(
        http.HandlerFunc(myHandler),
        "my-operation",
    )

    http.ListenAndServe(":8080", handler)
}
```

## 📚 Documentation

- **[Getting Started](GETTING_STARTED.md)** - Quick setup guide
- **[API Reference](API_REFERENCE.md)** - Complete function reference
- **[Advanced Usage](ADVANCED_USAGE.md)** - Advanced configuration
- **[Integration Guides](INTEGRATION_GUIDES.md)** - Framework-specific guides

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](https://github.com/kernelshard/otelkit/blob/main/CONTRIBUTING.md) for details.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/kernelshard/otelkit/blob/main/LICENSE) file for details.