# CORS Middleware for Zinc

This middleware provides Cross-Origin Resource Sharing (CORS) support for Zinc applications.

## Features

- Full support for standard CORS headers
- Allow specific origins or all origins
- Configurable allowed methods and headers
- Support for credentials
- Configurable preflight caching
- Custom header exposure
- Flexible configuration with functional options

## Usage

### Basic Usage with Default Configuration

The default configuration allows all origins, common HTTP methods, and standard headers:

```go
package main

import (
  "github.com/0mjs/zinc"
  "github.com/0mjs/zinc/middleware/cors"
)

func main() {
  app := zinc.New()
  
  // Add CORS middleware with default configuration
  app.Use(cors.New())
  
  app.Get("/", func(c *zinc.Context) error {
    return c.JSON(zinc.Map{
      "message": "Hello, World!",
    })
  })
  
  app.Serve()
}
```

### Custom Configuration

For more control, use the options functions to customize the CORS behavior:

```go
package main

import (
  "time"
  
  "github.com/0mjs/zinc"
  "github.com/0mjs/zinc/middleware/cors"
)

func main() {
  app := zinc.New()
  
  // Add CORS middleware with custom configuration
  app.Use(cors.NewWithOptions(
    // Only allow requests from these origins
    cors.WithAllowOrigins("https://example.com", "https://app.example.com"),
    
    // Allow credentials (cookies, authorization headers, etc.)
    cors.WithAllowCredentials(true),
    
    // Allow specific methods
    cors.WithAllowMethods("GET", "POST", "PUT", "DELETE"),
    
    // Allow specific headers
    cors.WithAllowHeaders("Content-Type", "Authorization", "X-API-Key"),
    
    // Expose custom headers to the client
    cors.WithExposeHeaders("X-Request-ID", "X-API-Version"),
    
    // Cache preflight requests for 1 hour
    cors.WithMaxAge(time.Hour),
  ))
  
  // Routes
  app.Get("/api/users", getUsersHandler)
  
  app.Serve()
}
```

### Per-Group Configuration

You can apply different CORS settings to different route groups:

```go
package main

import (
  "github.com/0mjs/zinc"
  "github.com/0mjs/zinc/middleware/cors"
)

func main() {
  app := zinc.New()
  
  // Public API with permissive CORS
  publicAPI := app.Group("/public")
  publicAPI.Use(cors.New()) // Default configuration allows all origins
  
  // Private API with strict CORS
  privateAPI := app.Group("/api")
  privateAPI.Use(cors.NewWithOptions(
    cors.WithAllowOrigins("https://admin.example.com"),
    cors.WithAllowCredentials(true),
    cors.WithAllowHeaders("Authorization", "Content-Type"),
  ))
  
  // Configure routes
  publicAPI.Get("/status", statusHandler)
  privateAPI.Get("/users", getUsersHandler)
  
  app.Serve()
}
```

### Skipping CORS for Specific Requests

If you need to skip CORS handling for certain requests, use the skipper option:

```go
app.Use(cors.NewWithOptions(
  // Other options...
  
  // Skip CORS for internal requests
  cors.WithSkipper(func(c *zinc.Context) bool {
    // Skip if request comes from trusted internal network
    return strings.HasPrefix(c.Request.RemoteAddr, "10.0.") || 
           strings.HasPrefix(c.Request.RemoteAddr, "192.168.")
  }),
))
```

## Configuration Options

| Option                 | Description                                        | Default                                                 |
| ---------------------- | -------------------------------------------------- | ------------------------------------------------------- |
| `WithAllowOrigins`     | Origins allowed to access the resource             | `["*"]`                                                 |
| `WithAllowMethods`     | HTTP methods allowed for CORS requests             | `["GET", "POST", "HEAD", "PUT", "DELETE", "PATCH"]`     |
| `WithAllowHeaders`     | Headers allowed in actual requests                 | `["Origin", "Content-Type", "Accept", "Authorization"]` |
| `WithExposeHeaders`    | Headers exposed to the browser                     | `[]`                                                    |
| `WithAllowCredentials` | Allows sending credentials with requests           | `false`                                                 |
| `WithMaxAge`           | Duration for caching preflight requests            | `0` (no caching)                                        |
| `WithMaxAgeSeconds`    | Duration in seconds for caching preflight requests | `0` (no caching)                                        |
| `WithSkipper`          | Function to determine if CORS should be skipped    | `nil` (never skip)                                      |

## Performance Considerations

The CORS middleware is designed to be lightweight and add minimal overhead to requests. It uses optimizations like:

- Pre-computed header strings to avoid repeated joins
- Early returns for non-CORS requests
- Map-based origin lookup for performance
- Minimal allocations for common paths

## Security Considerations

When configuring CORS, consider these security best practices:

1. **Avoid using wildcards** for `AllowOrigins` in production, especially when `AllowCredentials` is enabled
2. **Only expose necessary headers** to minimize attack surface
3. **Limit allowed methods** to only those needed by your API
4. **Be cautious with credentials** - only enable if needed 