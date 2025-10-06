package main

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/attribute"

	"github.com/kernelshard/otelkit"
)

// Storage simulates a data storage layer
type Storage struct{}

// NewStorage creates a new storage instance
func NewStorage() *Storage {
	return &Storage{}
}

// GetMessage retrieves a message from storage
func (s *Storage) GetMessage(ctx context.Context) (string, error) {
	return "Hello from dummy service!", nil
}

// Service holds the business logic for the dummy service
type Service struct {
	storage *Storage
	tracer  *otelkit.Tracer
}

// NewService creates a new Service instance
func NewService(storage *Storage, tracer *otelkit.Tracer) *Service {
	return &Service{
		storage: storage,
		tracer:  tracer,
	}
}

// GetHelloMessage returns a hello message with tracing
func (s *Service) GetHelloMessage(ctx context.Context) (string, error) {
	ctx, span := s.tracer.Start(ctx, "service-get-hello-message")
	defer span.End()

	// Simulate some business logic
	message, err := s.storage.GetMessage(ctx)
	if err != nil {
		otelkit.RecordError(span, err)
		return "", err
	}

	otelkit.AddAttributes(span, attribute.String("message", message))
	return message, nil
}

func main() {
	// Initialize tracing
	ctx := context.Background()
	shutdown, err := otelkit.SetupTracing(ctx, "dummy-service")
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(ctx)

	// Create service components
	tracer := otelkit.New("dummy-service")
	storage := NewStorage()
	service := NewService(storage, tracer)

	// Use the service
	ctx, span := tracer.Start(ctx, "main-operation")
	defer span.End()

	message, err := service.GetHelloMessage(ctx)
	if err != nil {
		otelkit.RecordError(span, err)
		log.Fatal(err)
	}

	fmt.Printf("Service response: %s\n", message)
	fmt.Println("Dummy service example completed. Check your tracing backend for spans.")
}
