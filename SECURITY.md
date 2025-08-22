# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of OtelKit seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Private Disclosure Process

We ask that you do not use public channels to report security vulnerabilities.

**Please report security vulnerabilities by email:**
- Email: security@yourdomain.com (replace with actual security contact)
- You should receive a response within 48 hours

### Preferred Languages

We prefer all communications to be in English.

## Security Considerations

### OpenTelemetry Security

OtelKit uses the OpenTelemetry Go SDK. Please also review:
- [OpenTelemetry Security](https://github.com/open-telemetry/opentelemetry-go#security)
- [OpenTelemetry Security Advisories](https://github.com/open-telemetry/opentelemetry-go/security/advisories)

### Data Sensitivity

OtelKit handles tracing data which may contain:
- Service names and metadata
- Request/response information
- Potentially sensitive operation names
- Timing information

### Best Practices

1. **Use secure connections** for OTLP exporters (disable insecure mode in production)
2. **Validate configurations** to prevent misconfiguration
3. **Keep dependencies updated** with security patches
4. **Monitor for security advisories** in OpenTelemetry dependencies

## Dependency Security

We regularly update dependencies to address security vulnerabilities. You can check current dependencies with:

```bash
go list -m all
```

## Security Updates

Security updates will be released as patch versions (e.g., 1.0.1, 1.0.2). Critical security fixes may receive backports to previous major versions.

## Acknowledgments

We would like to thank security researchers and users who report vulnerabilities to us. Your efforts help make OtelKit more secure for everyone.
