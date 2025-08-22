# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Project parameters
BINARY_NAME=otelkit
COVERAGE_FILE=coverage.out

.PHONY: all build clean test test-race test-coverage deps fmt lint vet security-check help

# Default target
all: fmt lint vet test build

# Build the project
build:
	$(GOBUILD) -v ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(COVERAGE_FILE)

# Run tests
test:
	$(GOTEST) -v -race ./...

# Run tests with race detection
test-race:
	$(GOTEST) -v -race ./...

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_FILE) ./...
	$(GOCMD) tool cover -html=$(COVERAGE_FILE) -o coverage.html

# Run all tests including integration tests
test-all:
	$(GOTEST) -v -race -tags=integration ./...

# Run integration tests only
test-integration:
	$(GOTEST) -v -tags=integration -run="TestIntegration" ./...

# Run integration tests with collector running
test-integration-with-collector: integration-up
	@echo "Waiting for collector to be ready..."
	@sleep 5
	$(GOTEST) -v -tags=integration -run="TestIntegration" ./...
	$(MAKE) integration-down

# Start integration test environment
integration-up:
	docker-compose -f docker-compose.integration.yml up -d

# Stop integration test environment
integration-down:
	docker-compose -f docker-compose.integration.yml down

# View integration test logs
integration-logs:
	docker-compose -f docker-compose.integration.yml logs -f

# Check integration test setup
check-integration-setup:
	./scripts/check-integration-setup.sh

# Run benchmarks
bench:
	$(GOTEST) -bench=. -benchmem ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Format code
fmt:
	$(GOFMT) -s -w .
	$(GOCMD) fmt ./...

# Lint code
lint:
	@if command -v $(GOLINT) >/dev/null 2>&1; then \
		$(GOLINT) run; \
	else \
		echo "golangci-lint not installed, skipping linting"; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Vet code
vet:
	$(GOCMD) vet ./...

# Security check
security-check:
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping security check"; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Update dependencies
update-deps:
	$(GOMOD) get -u ./...
	$(GOMOD) tidy

# Generate mocks (if using mockgen)
generate:
	$(GOCMD) generate ./...

# Run examples
examples:
	@echo "Building examples..."
	@for dir in examples/*/; do \
		if [ -f "$$dir/main.go" ]; then \
			echo "Building $$dir"; \
			cd "$$dir" && $(GOBUILD) -o example . && cd ../..; \
		fi \
	done

# Docker build
docker-build:
	docker build -t $(BINARY_NAME):latest .

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Format, lint, vet, test, and build"
	@echo "  build         - Build the project"
	@echo "  clean         - Clean build artifacts"
	@echo "  test          - Run tests"
	@echo "  test-race     - Run tests with race detection"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-all      - Run all tests including integration"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-integration-with-collector - Run integration tests with collector"
	@echo "  integration-up - Start integration test environment"
	@echo "  integration-down - Stop integration test environment"
	@echo "  integration-logs - View integration test logs"
	@echo "  check-integration-setup - Check integration test setup"
	@echo "  bench         - Run benchmarks"
	@echo "  deps          - Download and tidy dependencies"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  vet           - Vet code"
	@echo "  security-check - Run security analysis"
	@echo "  install-tools - Install development tools"
	@echo "  update-deps   - Update dependencies"
	@echo "  generate      - Generate code"
	@echo "  examples      - Build examples"
	@echo "  docker-build  - Build Docker image"
	@echo "  help          - Show this help"
