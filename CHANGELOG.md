# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GitHub Actions CI/CD workflow
- Issue templates for bugs and feature requests

## [0.2.0] - 2025-01-24

### Changed
- **API Simplification**: Reduced the number of entry points for easier usage and clarity.
- **Deprecated Functions**: Marked several functions as deprecated to guide users towards preferred usage patterns.
- **Documentation Improvements**: Enhanced README and API reference to provide clearer guidance on usage.

### Added
- Comprehensive integration test suite
- Open source documentation and contribution guidelines
- Enhanced project structure with proper internal packages
- Improved error handling and validation
- Better documentation with practical examples

### Fixed
- Resource creation function to use proper OpenTelemetry SDK API
- ForceFlush method calls in integration tests
- Makefile handling of missing development tools
- Variadic parameter handling for service version

### Improved
- **Project Structure**: Reorganized into proper internal packages (config, provider, tracer, middleware)
- **Documentation clarity**: Enhanced comments with practical usage context and real-world considerations
- **Error handling**: Improved validation and error messages for better developer experience
- **Code readability**: Added human-readable explanations for design decisions
- **Testing**: Comprehensive test coverage and integration testing

### Features
- **Provider Configuration**: Fluent API for configuring tracer providers
- **Tracer Wrapper**: Simplified interface over OpenTelemetry tracer
- **HTTP Middleware**: Automatic tracing for HTTP handlers
- **Configuration Management**: Environment variable and programmatic configuration
- **Error Handling**: Comprehensive error types and validation

### Supported Exporters
- OTLP HTTP exporter
- OTLP gRPC exporter

### Supported Sampling Strategies
- Always On (sample all traces)
- Always Off (sample no traces)
- Probabilistic (sample percentage of traces)

### Documentation
- Getting Started guide
- Advanced Usage documentation
- Integration guides for various frameworks
- API reference
- Examples for basic and production usage
- Release permissions troubleshooting guide
- Comprehensive CHANGELOG and release checklist

## [0.1.0] - 2025-01-01

### Added
- Initial release of OtelKit OpenTelemetry tracing library
- Zero-configuration setup with sensible defaults
- Support for HTTP and gRPC OTLP exporters
- Multiple sampling strategies (probabilistic, always_on, always_off)
- Automatic resource management with service metadata
- HTTP middleware for automatic request tracing
- Context propagation for distributed tracing
- Error recording and span utilities
- Basic unit test coverage
- Production-ready examples and documentation
