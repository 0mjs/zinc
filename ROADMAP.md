# Zinc Framework Roadmap

Zinc aims to provide a high-performance, minimal API framework for Go with a focus on simplicity and developer productivity. This roadmap outlines features to implement, inspired by popular frameworks like Fiber, Echo, and Chi, but adapted to fit Zinc's philosophy.

## Core Components

### Current Features

- ✅ Fast HTTP routing with static and parameter-based routes
- ✅ Route grouping with middleware support
- ✅ Context-based request/response handling
- ✅ JSON, XML, and form data binding and validation
- ✅ File upload handling
- ✅ Template rendering
- ✅ WebSocket support
- ✅ Cron scheduler
- ✅ Service dependency injection
- ✅ Content negotiation

### Planned Enhancements

## 1. Middleware Ecosystem

### Request/Response Handling

- [ ] **CORS Middleware**
  - Configure allowed origins, methods, headers
  - Support for credentials and preflight requests
  - Options for custom handling of complex scenarios

- [ ] **Compression Middleware**
  - Support for gzip, deflate, and brotli
  - Configurable compression levels
  - Content-type filtering

- [ ] **Cache Control Middleware**
  - ETag generation and validation
  - Configurable cache headers
  - Conditional requests (If-None-Match, If-Modified-Since)

- [ ] **Rate Limiting**
  - Request-based limiting with configurable windows
  - IP-based throttling
  - Token bucket algorithm implementation
  - Custom storage backends (memory, Redis)

- [ ] **Body Size Limiting**
  - Configurable maximum request body size
  - Early termination for oversized requests
  - Custom error responses

### Security

- [ ] **JWT Authentication**
  - Token validation and verification
  - Role-based access control
  - Custom claim validators
  - Multiple signing methods (HS256, RS256, etc.)

- [ ] **CSRF Protection**
  - Token generation and validation
  - Multiple storage options
  - Customizable error handling

- [ ] **Security Headers**
  - Content-Security-Policy
  - X-XSS-Protection
  - X-Content-Type-Options
  - Referrer-Policy
  - Permissions-Policy

- [ ] **IP Filtering**
  - Allow/deny lists
  - CIDR notation support
  - Customizable responses

### Observability

- [ ] **Structured Logging**
  - Request/response logging
  - Performance metrics
  - Log levels and formatting
  - Integration with popular logging packages

- [ ] **Request Tracing**
  - OpenTelemetry integration
  - Correlation IDs
  - Span propagation

- [ ] **Error Reporting**
  - Structured error handling
  - Stack trace capture
  - Integration with error reporting services

- [ ] **Metrics Collection**
  - Request counts, latencies, status codes
  - Custom metric registration
  - Prometheus integration

### Development

- [ ] **Hot Reload**
  - Development mode with automatic rebuilding
  - Template reloading
  - Graceful restart

- [ ] **Request Validator**
  - Enhanced struct validation
  - Custom validation rules
  - Error translation

## 2. Extended Core Features

- [ ] **Improved Router**
  - Optional path normalization
  - Regex path matching
  - Optional trailing slash handling
  - Custom 404 and 405 handlers

- [ ] **Enhanced Context**
  - Flash messages
  - Context cancellation
  - Enhanced error handling with status codes
  - Context cloning

- [ ] **File Serving**
  - Static file server with directory browsing option
  - File compression on-the-fly
  - Range requests support
  - Cache control headers

- [ ] **Enhanced Template Engine**
  - Multiple template engine support
  - Layout and partial support
  - Hot reloading of templates
  - Custom function registry

- [ ] **WebSocket Enhancements**
  - Event-based API
  - Room-based broadcasting
  - Connection pools
  - Heartbeat and auto-reconnect

- [ ] **Enhanced Content Negotiation**
  - Content type-based response formatting
  - Accept header-based negotiation
  - Custom serializers/deserializers

## 3. Additional Features

- [ ] **Configuration Management**
  - Environment-based configuration
  - Multiple sources (files, env vars, flags)
  - Hot reloading of configuration
  - Validation

- [ ] **Database Integration**
  - Connection pools
  - Transaction management
  - Migration helpers
  - Query builders

- [ ] **Testing Utilities**
  - Request/response testing helpers
  - Mocking utilities
  - Benchmark tools
  - Test middleware

- [ ] **Graceful Shutdown**
  - Signal handling
  - Connection draining
  - Timeout configuration
  - Shutdown hooks

- [ ] **Health Checks**
  - Customizable health check endpoints
  - Dependency health monitoring
  - Readiness and liveness probes
  - Status reporting

## Implementation Strategy

### Phase 1: Middleware Ecosystem

Focus on implementing the most commonly used middleware components:

1. CORS middleware
2. Basic security headers
3. JWT authentication
4. Request logging
5. Rate limiting
6. Compression

### Phase 2: Core Enhancements

Improve existing core components:

1. Router enhancements
2. Context extensions
3. File serving improvements
4. WebSocket enhancements

### Phase 3: Additional Features

Add new capabilities:

1. Configuration management
2. Testing utilities
3. Enhanced health checks
4. Database integration helpers

## Implementation Guidelines

1. **Performance First**: All implementations should maintain or improve Zinc's performance characteristics.
2. **Minimal Dependencies**: Limit external dependencies to maintain a small footprint.
3. **Go Idioms**: Follow Go best practices and idioms.
4. **Comprehensive Tests**: Each feature should include benchmark and unit tests.
5. **Clear Documentation**: All features should include clear documentation and examples.
6. **Backwards Compatibility**: Maintain compatibility with existing Zinc API where possible.

## Feature Implementation Details

This section will be expanded with detailed implementation plans for each feature as they are scheduled for development.

### CORS Middleware Implementation

```go
// Example implementation plan
func CORSMiddleware(options ...CORSOption) Middleware {
    // Initialize default options
    // Apply custom options
    // Return middleware that handles CORS
}

type CORSOptions struct {
    AllowOrigins     []string
    AllowMethods     []string
    AllowHeaders     []string
    ExposeHeaders    []string
    AllowCredentials bool
    MaxAge           int
}

type CORSOption func(*CORSOptions)

func AllowOrigins(origins ...string) CORSOption {
    return func(o *CORSOptions) {
        o.AllowOrigins = origins
    }
}

// Additional option functions...
```

### JWT Authentication Implementation

```go
// Example implementation plan
type JWTConfig struct {
    SigningKey      interface{}
    SigningMethod   jwt.SigningMethod
    ContextKey      string
    TokenLookup     string
    TokenHeadName   string
    AuthScheme      string
    Claims          jwt.Claims
    KeyFunc         jwt.Keyfunc
    ErrorHandler    JWTErrorHandler
    Skipper         Skipper
}

func JWTMiddleware(config JWTConfig) Middleware {
    // Initialize with defaults if needed
    // Return middleware that validates JWT tokens
}
```

## Contribution Guidelines

When implementing features for Zinc:

1. Start with a proposal in the form of an issue describing the feature
2. Follow the implementation guidelines above
3. Include examples and documentation
4. Add both unit and benchmark tests
5. Submit a PR with a clear description of the changes 