package tracer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestNewTracedHTTPClient(t *testing.T) {
	tracer := New("test-tracer")
	client := NewTracedHTTPClient(nil, tracer, "test-service")

	if client == nil {
		t.Fatal("NewTracedHTTPClient returned nil")
	}
	if client.tracer != tracer {
		t.Error("tracer not set correctly")
	}
	if client.service != "test-service" {
		t.Error("service not set correctly")
	}
	if client.client == nil {
		t.Error("client should have default http.Client")
	}
}

func TestTracedHTTPClient_Do_Success(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	tracer := New("test-tracer")
	SetGlobalTracerProvider(tp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewTracedHTTPClient(nil, tracer, "test-service")

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	if span.Name != "http.external_call" {
		t.Errorf("expected span name 'http.external_call', got '%s'", span.Name)
	}

	// Check attributes
	foundService := false
	foundMethod := false
	foundStatus := false
	for _, attr := range span.Attributes {
		switch attr.Key {
		case attribute.Key("http.external.service"):
			if attr.Value.AsString() != "test-service" {
				t.Errorf("expected service 'test-service', got '%s'", attr.Value.AsString())
			}
			foundService = true
		case attribute.Key("http.external.method"):
			if attr.Value.AsString() != "GET" {
				t.Errorf("expected method 'GET', got '%s'", attr.Value.AsString())
			}
			foundMethod = true
		case attribute.Key("http.external.status_code"):
			if attr.Value.AsInt64() != int64(http.StatusOK) {
				t.Errorf("expected status %d, got %d", http.StatusOK, attr.Value.AsInt64())
			}
			foundStatus = true
		}
	}

	if !foundService {
		t.Error("http.external.service attribute not found")
	}
	if !foundMethod {
		t.Error("http.external.method attribute not found")
	}
	if !foundStatus {
		t.Error("http.external.status_code attribute not found")
	}
}

func TestTracedHTTPClient_Get(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	SetGlobalTracerProvider(tp)
	tracer := New("test-tracer")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewTracedHTTPClient(nil, tracer, "test-service")

	ctx := context.Background()
	resp, err := client.Get(ctx, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
}

func TestTracedHTTPClient_Post(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	SetGlobalTracerProvider(tp)
	tracer := New("test-tracer")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewTracedHTTPClient(nil, tracer, "test-service")

	ctx := context.Background()
	body := []byte(`{"test": "data"}`)
	resp, err := client.Post(ctx, server.URL, "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	foundSize := false
	for _, attr := range span.Attributes {
		if attr.Key == attribute.Key("http.external.request_size") {
			if attr.Value.AsInt64() != int64(len(body)) {
				t.Errorf("expected request size %d, got %d", len(body), attr.Value.AsInt64())
			}
			foundSize = true
		}
	}

	if !foundSize {
		t.Error("http.external.request_size attribute not found")
	}
}

func TestTracedHTTPClient_Do_Error(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() {
		_ = tp.Shutdown(context.Background())
	}()

	SetGlobalTracerProvider(tp)
	tracer := New("test-tracer")

	client := NewTracedHTTPClient(&http.Client{Timeout: 1 * time.Millisecond}, tracer, "test-service")

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://httpbin.org/delay/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := client.Do(ctx, req)
	if err == nil {
		t.Fatal("expected error due to timeout")
	}
	if resp != nil {
		t.Fatal("expected nil response on error")
	}

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}

	span := spans[0]
	foundFailed := false
	for _, attr := range span.Attributes {
		if attr.Key == attribute.Key("http.external.failed") {
			if attr.Value.AsBool() != true {
				t.Error("expected http.external.failed to be true")
			}
			foundFailed = true
		}
	}

	if !foundFailed {
		t.Error("http.external.failed attribute not found")
	}
}
