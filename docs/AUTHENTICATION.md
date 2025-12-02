# API Authentication

This document describes the authentication system for the Brain-Salad API.

## Overview

The API supports optional authentication that is **disabled by default** for local CLI usage. When enabled, it provides API key-based authentication for securing the API in production deployments.

## Local Development (Default)

Authentication is **disabled by default** to allow easy local development and CLI usage without additional configuration.

```bash
# No authentication required - just start the server
go run ./cmd/web
```

All API endpoints are accessible without credentials in local development mode.

## Production Deployment

Enable authentication for public API deployments to secure your endpoints.

### Environment Variables

```bash
# Enable authentication
export AUTH_ENABLED=true

# Set authentication mode (currently only "api-key" is supported)
export AUTH_MODE=api-key

# Configure API keys (format: key1:description1,key2:description2)
export AUTH_API_KEYS="sk_prod_abc123:Production Client,sk_dev_xyz789:Development"
```

### Configuration Example

```bash
# Example production configuration
export AUTH_ENABLED=true
export AUTH_MODE=api-key
export AUTH_API_KEYS="sk_prod_abc123:Production Client,sk_dev_xyz789:Development Client"

# Start the server
go run ./cmd/web
```

The server will log the authentication status on startup:

```
{"level":"info","time":"2025-11-21T...","message":"Authentication enabled","mode":"api-key"}
{"level":"info","time":"2025-11-21T...","message":"API keys configured","api_keys":2}
```

## Using the API with Authentication

When authentication is enabled, include the API key in the `Authorization` header using the Bearer scheme:

```bash
# Make authenticated request
curl -H "Authorization: Bearer sk_prod_abc123" \
     http://localhost:8080/api/v1/ideas
```

### Request Format

```http
GET /api/v1/ideas HTTP/1.1
Host: localhost:8080
Authorization: Bearer sk_prod_abc123
```

### Response Codes

- **200 OK**: Request successful with valid API key
- **401 Unauthorized**: Missing, invalid, or malformed API key
  - Missing `Authorization` header
  - Invalid `Authorization` header format (must be `Bearer <key>`)
  - Invalid API key

## Bypassed Endpoints

The following endpoints bypass authentication even when it's enabled (for monitoring purposes):

- `/health` - Health check endpoint
- `/metrics` - Metrics endpoint

These endpoints should always be accessible for monitoring systems.

## Generating API Keys

Use secure random strings for API keys. Example using OpenSSL:

```bash
# Generate a secure API key
openssl rand -base64 32
```

Example output: `7h8JKL9mN0pQRsTuVwXyZ1a2B3c4D5e6F7g8H9i0J1k=`

### Recommended Key Format

For better organization, use a prefix to identify key types:

- `sk_prod_` - Production keys
- `sk_dev_` - Development keys
- `sk_test_` - Testing keys

Example: `sk_prod_7h8JKL9mN0pQRsTuVwXyZ1a2B3c4D5e6F7g8H9i0J1k=`

## Security Best Practices

### 1. Use HTTPS in Production

**Always** use HTTPS in production to prevent API keys from being transmitted in plain text:

```bash
# Bad - API key transmitted in plain text
curl -H "Authorization: Bearer sk_prod_abc123" \
     http://api.example.com/api/v1/ideas

# Good - API key encrypted in transit
curl -H "Authorization: Bearer sk_prod_abc123" \
     https://api.example.com/api/v1/ideas
```

### 2. Rotate Keys Regularly

Rotate API keys on a regular schedule (e.g., every 90 days) to minimize the impact of compromised keys.

### 3. One Key Per Client

Issue separate API keys for each client or service. This allows:
- Granular access control
- Better tracking of API usage
- Easy revocation of specific clients

### 4. Monitor API Key Usage

Monitor logs for:
- Failed authentication attempts
- Unusual usage patterns
- API calls from unexpected IPs

The server logs authentication events:

```json
{"level":"warn","time":"...","message":"Authentication failed: invalid API key","path":"/api/v1/ideas","method":"GET"}
{"level":"debug","time":"...","message":"Authentication successful","path":"/api/v1/ideas","method":"GET"}
```

### 5. Revoke Compromised Keys Immediately

If an API key is compromised:

1. Remove it from the `AUTH_API_KEYS` environment variable
2. Restart the server
3. Generate and distribute a new key to the affected client

### 6. Store Keys Securely

- **Never** commit API keys to version control
- Use environment variables or secure secret management systems
- Restrict access to production environment variables

## Architecture Notes

### Design Decisions

1. **Optional by Default**: Authentication is disabled by default to maintain backward compatibility and support local CLI usage without friction.

2. **Middleware-Based**: Authentication is implemented as HTTP middleware, making it easy to enable/disable and maintain.

3. **Monitoring Bypass**: Health and metrics endpoints bypass authentication to ensure monitoring systems can always check service health.

4. **Case-Sensitive Keys**: API keys are case-sensitive to maximize entropy and security.

### Future Enhancements

Planned authentication features:

- **JWT Authentication**: Token-based authentication with expiration
- **OAuth2 Support**: Third-party authentication providers
- **Rate Limiting per Key**: Different rate limits for different API keys
- **Key Permissions**: Granular permissions per API key (read-only, write, admin)

## Testing

### Unit Tests

Run authentication middleware tests:

```bash
go test ./internal/api -run TestAuth -v
```

### Integration Testing

Test authentication in integration tests:

```bash
# Test with auth disabled (default)
go run ./cmd/web &
WEB_PID=$!
sleep 2
curl -f http://localhost:8080/health || echo "Failed"
curl -f http://localhost:8080/api/v1/ideas || echo "Failed"
kill $WEB_PID

# Test with auth enabled
AUTH_ENABLED=true AUTH_API_KEYS="test:client" go run ./cmd/web &
WEB_PID=$!
sleep 2
curl -f http://localhost:8080/health || echo "Failed"  # Should work
curl -f http://localhost:8080/api/v1/ideas && echo "Should have failed!" || echo "Correctly blocked"
curl -f -H "Authorization: Bearer test" http://localhost:8080/api/v1/ideas || echo "Failed with valid key"
kill $WEB_PID
```

## Troubleshooting

### Authentication not working

1. Verify `AUTH_ENABLED=true` is set
2. Check that `AUTH_API_KEYS` is properly formatted
3. Verify the `Authorization` header format: `Bearer <key>`
4. Check server logs for authentication errors

### All requests return 401

1. Verify API key is in the `AUTH_API_KEYS` list
2. Check for typos in the API key
3. Ensure no extra whitespace in the header
4. Verify case matches exactly (keys are case-sensitive)

### Health endpoint returns 401

This should not happen - health and metrics endpoints bypass authentication. If this occurs:
1. Check that the request path is exactly `/health` or `/metrics`
2. Verify the middleware configuration in server.go
3. Check server logs for errors

## Support

For issues or questions about authentication:

1. Check this documentation
2. Review server logs for authentication errors
3. Run the test suite to verify functionality
4. Open an issue with authentication logs (redact actual API keys)
