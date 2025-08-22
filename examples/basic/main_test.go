package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/samims/otelkit"
)

const (
	testSleepDuration    = 10 * time.Millisecond
	dbQuerySleepDuration = 5 * time.Millisecond
	testTimeout          = 5 * time.Second
	testServiceName      = "test-service"
	testServiceVersion   = "test-version"
	testTracerName       = "test-tracer"
	testSpanNameHello    = "handle-hello"
	testSpanNameError    = "handle-error"
	testSpanNameDBQuery  = "database-query"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Set test environment variables
	os.Setenv("OTEL_SERVICE_NAME", testServiceName)
	os.Setenv("OTEL_SERVICE_VERSION", testServiceVersion)
	os.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "http/protobuf")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")

	code := m.Run()
	os.Exit(code)
}

// setupTestTracer creates a test tracer with in-memory exporter
func setupTestTracer(t *testing.T) (*tracetest.InMemoryExporter, *trace.TracerProvider) {
	t.Helper()

	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)

	// Use fresh context for shutdown to avoid cancellation issues
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown tracer provider: %v", err)
		}
	})

	// Always use otel.SetTracerProvider for consistency
	otel.SetTracerProvider(tp)

	return exporter, tp
}

// TestHelloHandler tests the /hello endpoint
func TestHelloHandler(t *testing.T) {
	exporter, tp := setupTestTracer(t)
	tr := otelkit.New(testTracerName)

	req := httptest.NewRequest("GET", "/hello", nil)
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tr.Start(r.Context(), testSpanNameHello)
		defer otelkit.EndSpan(span)

		otelkit.AddAttributes(span,
			attribute.String("http.method", r.Method),
			attribute.String("http.url", r.URL.String()),
			attribute.String("http.user_agent", r.UserAgent()),
		)

		time.Sleep(testSleepDuration)

		otelkit.AddEvent(span, "processing_request")

		// Simulate database query
		_, dbSpan := tr.Start(ctx, testSpanNameDBQuery)
		otelkit.AddAttributes(dbSpan,
			attribute.String("db.system", "postgresql"),
			attribute.String("db.operation", "SELECT"),
		)
		time.Sleep(dbQuerySleepDuration)
		otelkit.EndSpan(dbSpan)

		otelkit.AddAttributes(span, attribute.Int("http.status_code", 200))

		response := map[string]string{
			"message":  "Hello, OpenTelemetry!",
			"trace_id": tr.GetTraceID(ctx),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["message"] != "Hello, OpenTelemetry!" {
		t.Errorf("Unexpected message: got %v", response["message"])
	}

	if response["trace_id"] == "" {
		t.Error("Trace ID should be present in response")
	}

	tp.ForceFlush(context.Background())
	spans := exporter.GetSpans()

	if len(spans) < 2 {
		t.Errorf("Expected at least 2 spans, got %d", len(spans))
	}

	foundHelloSpan := false
	for _, span := range spans {
		if span.Name == testSpanNameHello {
			foundHelloSpan = true
			break
		}
	}

	if !foundHelloSpan {
		t.Error("Expected to find 'handle-hello' span")
	}
}

// TestErrorHandler tests the /error endpoint
func TestErrorHandler(t *testing.T) {
	exporter, tp := setupTestTracer(t)
	tr := otelkit.New(testTracerName)

	req := httptest.NewRequest("GET", "/error", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, span := tr.Start(r.Context(), testSpanNameError)
		defer otelkit.EndSpan(span)

		err := fmt.Errorf("something went wrong")
		otelkit.RecordError(span, err)
		otelkit.AddAttributes(span, attribute.Bool("error", true))

		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	tp.ForceFlush(context.Background())
	spans := exporter.GetSpans()

	if len(spans) == 0 {
		t.Fatal("No spans recorded")
	}

	errorSpan := spans[0]
	if errorSpan.Name != testSpanNameError {
		t.Errorf("Expected span name '%s', got '%s'", testSpanNameError, errorSpan.Name)
	}

	hasErrorAttr := false
	for _, attr := range errorSpan.Attributes {
		if attr.Key == "error" && attr.Value.AsBool() == true {
			hasErrorAttr = true
			break
		}
	}

	if !hasErrorAttr {
		t.Error("Error attribute should be set to true")
	}
}

// TestHealthHandler tests the /health endpoint
func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if rr.Body.String() != "OK" {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), "OK")
	}
}

// TestMiddlewareIntegration tests the HTTP middleware integration
func TestMiddlewareIntegration(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(
		trace.WithSyncer(exporter),
	)
	defer tp.Shutdown(context.Background())

	otel.SetTracerProvider(tp)
	tr := otelkit.New("test-tracer")

	// Create middleware using the actual otelkit implementation
	middleware := otelkit.NewHttpMiddleware(tr)

	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	})

	handler := middleware.Middleware(mux)

	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	// Check that middleware creates spans
	tp.ForceFlush(context.Background())
	spans := exporter.GetSpans()

	if len(spans) == 0 {
		t.Error("Middleware should create spans for HTTP requests")
	}
}
