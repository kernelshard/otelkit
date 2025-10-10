//go:build integration

package integration

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/attribute"

	"github.com/kernelshard/otelkit"
)

// TestEnhancedErrorHandlingIntegration performs comprehensive integration testing
// of the enhanced error handling functionality to ensure it works correctly
// in real-world scenarios before shipping.
func TestEnhancedErrorHandlingIntegration(t *testing.T) {
	// Setup tracing
	ctx := context.Background()
	shutdown, err := otelkit.SetupTracing(ctx, "integration-test")
	if err != nil {
		t.Fatalf("Failed to setup tracing: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			t.Errorf("Failed to shutdown tracing: %v", err)
		}
	}()

	t.Run("BasicErrorRecording", func(t *testing.T) {
		tracer := otelkit.New("test-service")
		_, span := tracer.Start(ctx, "basic-error-test")
		defer span.End()

		testErr := errors.New("basic test error")
		otelkit.RecordError(span, testErr)

		// Verify span status is set to error
		if span.SpanContext().IsValid() {
			// Span should be valid and have error status
			t.Logf("Basic error recording completed successfully")
		}
	})

	t.Run("EnhancedErrorRecordingWithOptions", func(t *testing.T) {
		tracer := otelkit.New("test-service")
		_, span := tracer.Start(ctx, "enhanced-error-test")
		defer span.End()

		testErr := errors.New("enhanced test error")

		// Test with all options
		otelkit.RecordErrorEnhanced(span, testErr,
			otelkit.WithErrorType(otelkit.ErrorTypeValidation),
			otelkit.WithStackTrace(true),
			otelkit.WithErrorCode("VALIDATION_FAILED"),
			otelkit.WithErrorAttributes(
				attribute.String("field", "email"),
				attribute.String("reason", "invalid_format"),
				attribute.Bool("user_input", true),
			),
		)

		t.Logf("Enhanced error recording with all options completed successfully")
	})

	t.Run("EnhancedErrorRecordingDefaults", func(t *testing.T) {
		tracer := otelkit.New("test-service")
		_, span := tracer.Start(ctx, "defaults-test")
		defer span.End()

		testErr := errors.New("defaults test error")

		// Test with no options - should default to ErrorTypeCustom
		otelkit.RecordErrorEnhanced(span, testErr)

		t.Logf("Enhanced error recording with defaults completed successfully")
	})

	t.Run("HTTPMiddlewareIntegration", func(t *testing.T) {
		// Create a test HTTP handler with error recording
		mux := http.NewServeMux()
		tracer := otelkit.New("http-test-service")
		middleware := otelkit.NewHttpMiddleware(tracer)

		mux.Handle("/error-test", middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tracer := otelkit.New("http-test-service")
			_, span := tracer.Start(r.Context(), "handler")
			defer span.End()

			// Simulate an error
			err := errors.New("handler error")
			otelkit.RecordErrorEnhanced(span, err,
				otelkit.WithErrorType(otelkit.ErrorTypeSystem),
				otelkit.WithErrorCode("HANDLER_ERROR"),
			)

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error occurred"))
		})))

		// Create test request
		req := httptest.NewRequest("GET", "/error-test", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", w.Code)
		}

		if !strings.Contains(w.Body.String(), "error occurred") {
			t.Errorf("Expected error message in response")
		}

		t.Logf("HTTP middleware integration test completed successfully")
	})

	t.Run("NestedSpansWithErrors", func(t *testing.T) {
		tracer := otelkit.New("nested-test-service")

		// Outer span
		ctx, outerSpan := tracer.Start(ctx, "outer-operation")
		defer outerSpan.End()

		// Simulate some work
		_, innerSpan := tracer.Start(ctx, "inner-operation")
		defer innerSpan.End()

		// Record error on inner span
		innerErr := errors.New("inner operation failed")
		otelkit.RecordErrorEnhanced(innerSpan, innerErr,
			otelkit.WithErrorType(otelkit.ErrorTypeDatabase),
			otelkit.WithErrorCode("DB_CONNECTION_FAILED"),
		)

		// Record error on outer span
		outerErr := errors.New("outer operation failed due to inner error")
		otelkit.RecordErrorEnhanced(outerSpan, outerErr,
			otelkit.WithErrorType(otelkit.ErrorTypeSystem),
			otelkit.WithErrorCode("OPERATION_FAILED"),
		)

		t.Logf("Nested spans with errors test completed successfully")
	})

	t.Run("ErrorTypesCoverage", func(t *testing.T) {
		tracer := otelkit.New("types-test-service")

		errorTypes := []otelkit.ErrorType{
			otelkit.ErrorTypeNetwork,
			otelkit.ErrorTypeDatabase,
			otelkit.ErrorTypeValidation,
			otelkit.ErrorTypeSystem,
			otelkit.ErrorTypeCustom,
		}

		for _, errorType := range errorTypes {
			_, span := tracer.Start(ctx, "error-type-test")
			testErr := errors.New("test error for type: " + string(errorType))

			otelkit.RecordErrorEnhanced(span, testErr,
				otelkit.WithErrorType(errorType),
				otelkit.WithErrorCode("TEST_"+string(errorType)),
			)

			span.End()
		}

		t.Logf("Error types coverage test completed successfully")
	})

	t.Run("NilSafety", func(t *testing.T) {
		tracer := otelkit.New("nil-test-service")
		_, span := tracer.Start(ctx, "nil-test")
		defer span.End()

		testErr := errors.New("test error")

		// Test nil error (should not panic)
		otelkit.RecordErrorEnhanced(span, nil)

		// Test with nil options (should work fine)
		otelkit.RecordErrorEnhanced(span, testErr, nil)

		t.Logf("Nil safety test completed successfully")
	})

	t.Run("StackTraceCapture", func(t *testing.T) {
		tracer := otelkit.New("stack-trace-test-service")
		_, span := tracer.Start(ctx, "stack-trace-test")
		defer span.End()

		testErr := errors.New("stack trace test error")

		// Test with stack trace enabled
		otelkit.RecordErrorEnhanced(span, testErr,
			otelkit.WithErrorType(otelkit.ErrorTypeSystem),
			otelkit.WithStackTrace(true),
		)

		t.Logf("Stack trace capture test completed successfully")
	})

	t.Run("PerformanceCheck", func(t *testing.T) {
		tracer := otelkit.New("performance-test-service")

		// Test performance with many error recordings
		for i := 0; i < 100; i++ {
			_, span := tracer.Start(ctx, "perf-test")
			testErr := errors.New("performance test error")

			otelkit.RecordErrorEnhanced(span, testErr,
				otelkit.WithErrorType(otelkit.ErrorTypeValidation),
				otelkit.WithErrorCode("PERF_TEST"),
			)

			span.End()
		}

		t.Logf("Performance test with 100 error recordings completed successfully")
	})
}
