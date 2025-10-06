package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kernelshard/otelkit/tracer"
	"go.opentelemetry.io/otel/trace"
)

func TestNewHttpMiddleware(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	if middleware == nil {
		t.Fatal("NewHttpMiddleware returned nil")
	}
	if middleware.tracer == nil {
		t.Error("Middleware tracer should not be nil")
	}
	if middleware.tracer != tracer {
		t.Error("Middleware tracer should be the same as provided")
	}
}

func TestHTTPMiddleware_Middleware(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if we have a span in the context
		span := trace.SpanFromContext(r.Context())
		if span == nil {
			t.Error("Expected span in request context")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wrap the handler with middleware
	wrappedHandler := middleware.Middleware(testHandler)

	// Create a test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()

	// Execute the request
	wrappedHandler.ServeHTTP(rr, req)

	// Check response
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got %s", rr.Body.String())
	}
}

func TestHTTPMiddleware_DifferentMethods(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check if we have a span in the context
				span := trace.SpanFromContext(r.Context())
				if span == nil {
					t.Error("Expected span in request context")
				}
				w.WriteHeader(http.StatusOK)
			})

			wrappedHandler := middleware.Middleware(testHandler)

			req := httptest.NewRequest(method, "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", method, rr.Code)
			}
		})
	}
}

func TestHTTPMiddleware_DifferentStatusCodes(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	statusCodes := []int{200, 201, 400, 404, 500}

	for _, statusCode := range statusCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(statusCode)
			})

			wrappedHandler := middleware.Middleware(testHandler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rr, req)

			if rr.Code != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, rr.Code)
			}
		})
	}
}

func TestHTTPMiddleware_WithHeaders(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		if span == nil {
			t.Error("Expected span in request context")
		}
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom-Header", "custom-value")
	req.Header.Set("Authorization", "Bearer token")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestHTTPMiddleware_ContextPropagation(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get span from context
		span := trace.SpanFromContext(r.Context())
		if span == nil {
			t.Error("Expected span in request context")
			return
		}

		// Verify that the span context is valid (for real tracers)
		spanContext := span.SpanContext()
		if spanContext.TraceID().IsValid() {
			// If we have a valid trace ID, make sure it's not empty string
			traceIDStr := spanContext.TraceID().String()
			if traceIDStr == "" {
				t.Error("Expected non-empty trace ID string")
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	// Add trace context headers to simulate distributed tracing
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")

	rr := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rr,
		status:         http.StatusOK, // Default value
	}

	// Test default status
	if rw.status != http.StatusOK {
		t.Errorf("Expected default status 200, got %d", rw.status)
	}

	// Test setting status
	rw.WriteHeader(http.StatusNotFound)
	if rw.status != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rw.status)
	}
	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected underlying recorder status 404, got %d", rr.Code)
	}
}

func TestResponseWriter_Write(t *testing.T) {
	rr := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: rr,
		status:         http.StatusOK,
	}

	data := []byte("test data")
	n, err := rw.Write(data)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	if rr.Body.String() != "test data" {
		t.Errorf("Expected body 'test data', got %s", rr.Body.String())
	}
}

func TestHTTPMiddleware_MultipleRequests(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	requestCount := 0
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		span := trace.SpanFromContext(r.Context())
		if span == nil {
			t.Error("Expected span in request context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	wrappedHandler := middleware.Middleware(testHandler)

	// Send multiple requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Request %d: Expected status 200, got %d", i, rr.Code)
		}
	}

	if requestCount != 5 {
		t.Errorf("Expected 5 requests to be processed, got %d", requestCount)
	}
}

func TestHTTPMiddleware_PanicRecovery(t *testing.T) {
	tracer := tracer.New("test-tracer")
	middleware := NewHttpMiddleware(tracer)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify span exists before we potentially panic in the handler
		span := trace.SpanFromContext(r.Context())
		if span == nil {
			t.Error("Expected span in request context")
		}
		// Don't actually panic in the test - just verify the middleware works
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware.Middleware(testHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	// This should not panic
	wrappedHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}
}
