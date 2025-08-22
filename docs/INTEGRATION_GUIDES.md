# Integration Guides

This guide provides framework-specific integration instructions for popular Go web frameworks.

## Standard Library (net/http)

### Basic Integration
```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    provider, err := otelkit.NewDefaultProvider(ctx, "http-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    tracer := otelkit.New("http-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })

    handler := middleware.Middleware(mux)
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### With Custom Handlers
```go
mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
    ctx, span := tracer.Start(r.Context(), "get-users")
    defer span.End()
    
    // Your business logic here
    span.SetAttributes(attribute.Int("user.count", 42))
    
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(users)
})
```

## Gin Framework

### Installation
```bash
go get github.com/gin-gonic/gin
```

### Basic Setup
```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    provider, err := otelkit.NewDefaultProvider(ctx, "gin-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    tracer := otelkit.New("gin-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    r := gin.Default()
    r.Use(func(c *gin.Context) {
        // Convert Gin context to http.Handler
        middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Next()
        })).ServeHTTP(c.Writer, c.Request)
    })

    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })

    r.Run(":8080")
}
```

### Gin Middleware Wrapper
```go
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/samims/otelkit"
    "go.opentelemetry.io/otel/attribute"
)

func OtelKitMiddleware(tracer *otelkit.Tracer) gin.HandlerFunc {
    middleware := otelkit.NewHttpMiddleware(tracer)
    
    return func(c *gin.Context) {
        handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            c.Request = r
            c.Next()
        }))
        handler.ServeHTTP(c.Writer, c.Request)
    }
}

// Usage
func main() {
    r := gin.Default()
    tracer := otelkit.New("gin-service")
    r.Use(OtelKitMiddleware(tracer))
    
    r.GET("/users/:id", getUser)
    r.Run(":8080")
}

func getUser(c *gin.Context) {
    userID := c.Param("id")
    
    // Create custom span
    ctx, span := tracer.Start(c.Request.Context(), "get-user-by-id")
    defer span.End()
    
    span.SetAttributes(attribute.String("user.id", userID))
    
    // Your business logic
    c.JSON(200, gin.H{"user": userID})
}
```

## Chi Router

### Installation
```bash
go get github.com/go-chi/chi/v5
```

### Basic Setup
```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    provider, err := otelkit.NewDefaultProvider(ctx, "chi-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    tracer := otelkit.New("chi-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    r := chi.NewRouter()
    r.Use(func(next http.Handler) http.Handler {
        return middleware.Middleware(next)
    })

    r.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, Chi!"))
    })

    http.ListenAndServe(":8080", r)
}
```

### Chi Middleware
```go
package main

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/samims/otelkit"
)

func OtelKitMiddleware(tracer *otelkit.Tracer) func(http.Handler) http.Handler {
    middleware := otelkit.NewHttpMiddleware(tracer)
    return middleware.Middleware
}

// Usage
func main() {
    r := chi.NewRouter()
    tracer := otelkit.New("chi-service")
    
    r.Use(OtelKitMiddleware(tracer))
    
    r.Route("/api/v1", func(r chi.Router) {
        r.Get("/users", listUsers)
        r.Post("/users", createUser)
        r.Get("/users/{id}", getUser)
    })
    
    http.ListenAndServe(":8080", r)
}
```

## Echo Framework

### Installation
```bash
go get github.com/labstack/echo/v4
```

### Basic Setup
```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    provider, err := otelkit.NewDefaultProvider(ctx, "echo-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    tracer := otelkit.New("echo-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    e := echo.New()
    e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                c.SetRequest(r)
                next(c)
            }))
            handler.ServeHTTP(c.Response(), c.Request())
            return nil
        }
    })

    e.GET("/", func(c echo.Context) error {
        return c.String(http.StatusOK, "Hello, Echo!")
    })

    e.Start(":8080")
}
```

### Echo Middleware
```go
package main

import (
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/samims/otelkit"
)

func OtelKitMiddleware(tracer *otelkit.Tracer) echo.MiddlewareFunc {
    middleware := otelkit.NewHttpMiddleware(tracer)
    
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                c.SetRequest(r)
                if err := next(c); err != nil {
                    c.Error(err)
                }
            }))
            handler.ServeHTTP(c.Response(), c.Request())
            return nil
        }
    }
}

// Usage
func main() {
    e := echo.New()
    tracer := otelkit.New("echo-service")
    
    e.Use(OtelKitMiddleware(tracer))
    
    e.GET("/users/:id", getUser)
    e.Logger.Fatal(e.Start(":8080"))
}

func getUser(c echo.Context) error {
    userID := c.Param("id")
    return c.JSON(http.StatusOK, map[string]string{"user": userID})
}
```

## Gorilla Mux

### Installation
```bash
go get github.com/gorilla/mux
```

### Basic Setup
```go
package main

import (
    "context"
    "log"
    "net/http"
    "time"

    "github.com/gorilla/mux"
    "github.com/samims/otelkit"
)

func main() {
    ctx := context.Background()
    
    provider, err := otelkit.NewDefaultProvider(ctx, "mux-service", "v1.0.0")
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    tracer := otelkit.New("mux-service")
    middleware := otelkit.NewHttpMiddleware(tracer)

    r := mux.NewRouter()
    r.Use(middleware.Middleware)

    r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        title := vars["title"]
        page := vars["page"]
        
        w.Write([]byte("You've requested the book: " + title + " on page " + page))
    })

    http.ListenAndServe(":8080", r)
}
```

## Database Integration

### SQL Database Tracing
```go
package main

import (
    "context"
    "database/sql"
    "time"

    "github.com/samims/otelkit"
    _ "github.com/lib/pq"
    "go.opentelemetry.io/otel/attribute"
)

func queryUsers(ctx context.Context, db *sql.DB) ([]User, error) {
    tracer := otelkit.New("user-service")
    ctx, span := tracer.Start(ctx, "query-users")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("db.system", "postgresql"),
        attribute.String("db.operation", "SELECT"),
    )
    
    rows, err := db.QueryContext(ctx, "SELECT id, name, email FROM users")
    if err != nil {
        otelkit.RecordError(span, err)
        return nil, err
    }
    defer rows.Close()
    
    var users []User
    for rows.Next() {
        var user User
        if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
            otelkit.RecordError(span, err)
            return nil, err
        }
        users = append(users, user)
    }
    
    span.SetAttributes(attribute.Int("db.rows_affected", len(users)))
    return users, nil
}
```

## Testing Integration

### Unit Tests with Mock Tracer
```go
package main

import (
    "context"
    "testing"

    "github.com/samims/otelkit"
    "github.com/stretchr/testify/assert"
    "go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestHandler(t *testing.T) {
    // Create in-memory exporter for testing
    exporter := tracetest.NewInMemoryExporter()
    provider := sdktrace.NewTracerProvider(
        sdktrace.WithSyncer(exporter),
    )
    
    tracer := otelkit.New("test-service")
    handler := createHandler(tracer)
    
    // Test your handler
    // Assert on exported spans
    spans := exporter.GetSpans()
    assert.Len(t, spans, 1)
}
```

## Configuration Examples

### Development
```go
config := otelkit.NewProviderConfig("dev-service", "1.0.0").
    WithOTLPExporter("localhost:4317", "grpc", true).
    WithSampling("always_on", 1.0)
```

### Staging
```go
config := otelkit.NewProviderConfig("staging-service", "1.0.0").
    WithOTLPExporter("staging-collector:4317", "grpc", false).
    WithSampling("probabilistic", 0.5)
```

### Production
```go
config := otelkit.NewProviderConfig("prod-service", "1.0.0").
    WithOTLPExporter("prod-collector:4317", "grpc", false).
    WithSampling("probabilistic", 0.01).
    WithBatchOptions(5*time.Second, 30*time.Second, 512, 2048)
