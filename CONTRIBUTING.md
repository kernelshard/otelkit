# Contributing to OtelKit

Thank you for your interest in contributing to OtelKit! This document provides guidelines and instructions for contributing to this OpenTelemetry tracing library.

## Code of Conduct

By participating in this project, you are expected to uphold our Code of Conduct:
- Be respectful and inclusive
- Exercise consideration and respect in your speech and actions
- Attempt collaboration before conflict
- Refrain from demeaning, discriminatory, or harassing behavior

## Getting Started

### Prerequisites
- Go 1.22 or later
- Basic understanding of OpenTelemetry concepts
- Familiarity with Go modules

### Development Setup

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/your-username/otelkit.git
   cd otelkit
   ```
3. Install development tools:
   ```bash
   make install-tools
   ```

## Development Workflow

### Branch Strategy
- `main` - Stable production branch
- `develop` - Development branch (if using GitFlow)
- Feature branches: `feature/your-feature-name`
- Bug fix branches: `fix/issue-description`

### Making Changes

1. Create a feature branch from `main` or `develop`
2. Make your changes with clear, focused commits
3. Add tests for new functionality
4. Update documentation as needed
5. Run the test suite:
   ```bash
   make all
   ```

### Testing

#### Unit Tests
```bash
make test
```

#### Integration Tests
```bash
# Requires Docker for OTLP collectors
make integration-up
make test-integration
make integration-down
```

#### Code Quality
```bash
make fmt      # Format code
make lint     # Lint code (requires golangci-lint)
make vet      # Go vet
```

### Pull Request Process

1. Ensure all tests pass
2. Update documentation if needed
3. Add a clear description of changes
4. Reference any related issues
5. Request review from maintainers

## Code Style Guidelines

### Go Code
- Follow standard Go formatting (`gofmt`)
- Use descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and concise

### Documentation
- Update README.md for user-facing changes
- Add godoc comments for all exported elements
- Include examples for new features

### Testing
- Write tests for new functionality
- Maintain high test coverage
- Include both unit and integration tests
- Use table-driven tests where appropriate

## Adding New Features

### Feature Proposal
Before implementing major features, please:
1. Open an issue to discuss the feature
2. Get feedback from maintainers
3. Outline the implementation approach

### Backward Compatibility
- Maintain backward compatibility when possible
- Use semantic versioning for breaking changes
- Provide migration guides for major changes

## Bug Reports

When reporting bugs, please include:
- OtelKit version
- Go version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Relevant logs or error messages

## Release Process

Releases follow semantic versioning (MAJOR.MINOR.PATCH):

1. **Patch releases** (x.x.1): Bug fixes
2. **Minor releases** (x.1.x): New features, backward compatible
3. **Major releases** (1.x.x): Breaking changes

### Release Steps
1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release tag
4. Build and test release artifacts
5. Publish release notes

## Community

- Join discussions in GitHub Issues
- Ask questions in GitHub Discussions
- Report security issues privately to maintainers

## License

By contributing, you agree that your contributions will be licensed under the same MIT License that covers the project.

Thank you for contributing to OtelKit! 🚀
