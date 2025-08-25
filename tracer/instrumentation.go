package tracer

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewInstrumentedGRPCServer creates a new gRPC server with OpenTelemetry unary and stream interceptors attached automatically.
func NewInstrumentedGRPCServer(opts ...grpc.ServerOption) *grpc.Server {
	opts = append(opts,
		grpc.StatsHandler(otelgrpc.NewServerHandler()))

	return grpc.NewServer(opts...)
}

// NewInstrumentedGRPCClientDialOptions returns grpc.DialOption slice with OpenTelemetry instrumentation for client connections.
// Use this in grpc.Dial for instrumented client connections.
func NewInstrumentedGRPCClientDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Replace with TLS credentials as needed
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}

// NewInstrumentedHTTPHandler wraps an http.Handler with OpenTelemetry instrumentation and returns the wrapped handler.
// Usage: http.Handle("/path", NewInstrumentedHTTPHandler(yourHandler, "operationName"))
func NewInstrumentedHTTPHandler(handler http.Handler, operationName string) http.Handler {
	return otelhttp.NewHandler(handler, operationName)
}

// NewInstrumentedHTTPClient returns an *http.Client with OpenTelemetry transport for automatic HTTP tracing.
// You can customize it by passing a base transport; if nil, http.DefaultTransport is used.
func NewInstrumentedHTTPClient(baseTransport http.RoundTripper) *http.Client {
	if baseTransport == nil {
		baseTransport = http.DefaultTransport
	}
	return &http.Client{
		Transport: otelhttp.NewTransport(baseTransport),
	}
}
