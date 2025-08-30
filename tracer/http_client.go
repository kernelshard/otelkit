package tracer

import (
	"bytes"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// TracedHTTPClient wraps an HTTP client with OpenTelemetry tracing
type TracedHTTPClient struct {
	client  *http.Client
	tracer  *Tracer
	service string
}

// NewTracedHTTPClient creates a new traced HTTP client
func NewTracedHTTPClient(client *http.Client, tracer *Tracer, service string) *TracedHTTPClient {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &TracedHTTPClient{
		client:  client,
		tracer:  tracer,
		service: service,
	}
}

// Do executes an HTTP request with tracing
func (t *TracedHTTPClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	ctx, span := t.tracer.StartClientSpan(ctx, "http.external_call")
	defer span.End()

	// Set span attributes for the request
	span.SetAttributes(
		attribute.String("http.external.method", req.Method),
		attribute.String("http.external.url", req.URL.String()),
		attribute.String("http.external.scheme", req.URL.Scheme),
		attribute.String("http.external.host", req.URL.Host),
		attribute.String("http.external.path", req.URL.Path),
		attribute.String("http.external.service", t.service),
	)
	if req.ContentLength > 0 {
		span.SetAttributes(attribute.Int("http.external.request_size", int(req.ContentLength)))
	}

	// Inject trace context into request headers
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, propagation.HeaderCarrier(req.Header))

	// Add event for request start
	span.AddEvent("external_request.start")

	// Execute the request
	start := time.Now()
	resp, err := t.client.Do(req)
	duration := time.Since(start)

	// Add event for response received
	span.AddEvent("external_request.complete")

	// Capture response details
	if resp != nil {
		span.SetAttributes(
			attribute.Int("http.external.status_code", resp.StatusCode),
			attribute.String("http.external.status_text", resp.Status),
			attribute.String("http.external.content_type", resp.Header.Get("Content-Type")),
			attribute.String("http.external.content_length", resp.Header.Get("Content-Length")),
			attribute.Float64("http.external.duration_ms", float64(duration.Milliseconds())),
		)

		// Capture non-sensitive response headers
		captureResponseHeaders(span, resp.Header)

		// Set span status based on HTTP status code
		if resp.StatusCode >= 400 {
			span.SetStatus(codes.Error, "External service returned error: "+strconv.Itoa(resp.StatusCode))
			span.SetAttributes(attribute.Bool("http.external.error", true))
		} else {
			span.SetStatus(codes.Ok, "External service call successful")
			span.SetAttributes(attribute.Bool("http.external.success", true))
		}
	}

	if err != nil {
		RecordErrorWithCode(span, err, ErrorCodeExternalService, "External HTTP call failed")
		span.SetAttributes(attribute.Bool("http.external.failed", true))
	}

	return resp, err
}

// Get performs a GET request with tracing
func (t *TracedHTTPClient) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return t.Do(ctx, req)
}

// Post performs a POST request with tracing
func (t *TracedHTTPClient) Post(ctx context.Context, url, contentType string, body []byte) (*http.Response, error) {
	var reqBody *bytes.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, reqBody)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.ContentLength = int64(len(body))
	}
	req.Header.Set("Content-Type", contentType)
	return t.Do(ctx, req)
}

// captureResponseHeaders captures non-sensitive response headers for tracing
func captureResponseHeaders(span trace.Span, headers http.Header) {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"set-cookie":    true,
		"x-api-key":     true,
		"token":         true,
	}

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		if !sensitiveHeaders[lowerKey] && len(values) > 0 {
			span.SetAttributes(attribute.String("http.external.response.header."+lowerKey, values[0]))
		}
	}
}
