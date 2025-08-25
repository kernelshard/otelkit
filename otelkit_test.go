package otelkit

import (
	"testing"
)

func TestNewTracer(t *testing.T) {
	tracer := New("test-service")
	if tracer == nil {
		t.Error("New() returned nil tracer")
	}
}

func TestNewProviderConfig(t *testing.T) {
	config := NewProviderConfig("test-service", "v1.0.0")
	if config == nil {
		t.Error("NewProviderConfig() returned nil config")
	}
}
