# zinc

![Zinc](https://img.shields.io/badge/Zinc-%20A%20web%20framework%20for%20Go-silver)
![Version](https://img.shields.io/badge/version-0.0.58-red)
![Go Version](https://img.shields.io/badge/Go-1.22+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Zinc is a high-performance, minimal API framework for Go that focuses on speed, simplicity, and developer velocity. Designed to compete with the most popular frameworks around today in performance and usability.

## Features

- **Fast**: Optimized routing and minimal middleware overhead
- **Simple API**: Intuitive and expressive API that follows Go idioms
- **Powerful Router**: Support for static routes, path parameters, route groups, and middleware
- **Template Engine**: Built-in HTML templating with custom functions and template caching
- **WebSocket Support**: Real-time communication with room-based broadcasting
- **File Uploads**: Easy file upload handling with size limits and type validation
- **Cron Scheduler**: Built-in cron jobs for scheduled tasks
- **Memory Efficient**: Utilizes sync.Pool and fixed-size data structures to minimize allocations
- **Well-Tested**: Comprehensive test suite ensures reliability

## Installation

```bash
go get github.com/0mjs/zinc
```

## Quick Start

```go
package main

import (
    "github.com/0mjs/zinc"
    "log"
)

func main() {
    app := zinc.New()
    
    // Simple route
    app.Get("/", func(c *zinc.Context) {
        c.Send("Hello, World!")
    })
    
    // Path parameters
    app.Get("/users/:id", func(c *zinc.Context) {
        c.JSON(zinc.Map{
            "message": "User ID: " + c.Param("id"),
        })
    })
    
    // Route grouping
    api := app.Group("/api")
    api.Get("/users", func(c *zinc.Context) {
        c.JSON(zinc.Map{
            "users": []string{"matthew", "mark", "luke", "john"},
        })
    })
    
    // Middleware
    app.Use(LoggerMiddleware())
    
    log.Fatal(app.Serve())
}

func LoggerMiddleware() zinc.Middleware {
    return func(c *zinc.Context) {
        // Log before request handling
        c.Next() // Pass control to the next middleware or handler
        // Log after request handling
    }
}
```

## Documentation

For complete documentation, visit:

- [Pkg.go.dev Documentation](https://pkg.go.dev/github.com/0mjs/zinc)
- [Zinc Docs](https://github.com/0mjs/zinc/docs](https://zinc.0mjs.dev/)

## Benchmarks

Zinc is designed for high performance, with benchmarks showing it to be competitive with or faster than other popular Go frameworks:

- Static routes: ~800ns/op
- Dynamic routes: ~1.2μs/op
- Middleware chain: ~2.0μs/op

## Typed Service Dependency Injection

Zinc provides a powerful type-safe service dependency injection system that allows you to register and retrieve services by their concrete types.

### Registering Services

You can register services using either the string-based approach or the new type-based approach:

```go
// String-based service registration (legacy)
app.Service("userService", userService)

// Type-based service registration (recommended)
app.Register(userService)
```

### Retrieving Services

There are several ways to retrieve services:

1. String-based retrieval (legacy):

```go
userService := c.Service("userService").(*UserService)
```

2. Type-based retrieval using generics:

```go
// From App instance
userService, ok := zinc.ServiceOf[*UserService](app)
if !ok {
    // Handle service not found
}

// From Context
userService, ok := zinc.ContextServiceOf[*UserService](c)
if !ok {
    // Handle service not found
}
```

The type-based approach provides several advantages:
- Compile-time type safety
- No need for type assertions
- No string literals that could contain typos
- Better IDE support with code completion

### Examples

#### Basic Usage

```go
// Register a service
app.Register(userService)

// Use the service in a handler
app.Get("/users", func(c *zinc.Context) error {
    service, ok := zinc.ContextServiceOf[*UserService](c)
    if !ok {
        return c.Status(zinc.StatusInternalServerError).String("Service not available")
    }
    return service.GetUsers(c)
})
```

#### Using Services in Middleware

Services can be accessed directly from middleware functions:

```go
// Middleware that uses typed services
authMiddleware := func(c *zinc.Context) error {
    // Access the auth service directly from context
    authService, ok := zinc.ContextServiceOf[*AuthService](c)
    if !ok {
        return c.Status(zinc.StatusInternalServerError).String("Auth service not available")
    }
    
    // Use the service
    token := c.Request.Header.Get("Authorization")
    if err := authService.ValidateToken(token); err != nil {
        return c.Status(zinc.StatusUnauthorized).String("Invalid token")
    }
    
    // Continue with the next handler
    return c.Next()
}

// Apply the middleware to routes or groups
app.Get("/protected", authMiddleware, func(c *zinc.Context) error {
    return c.String("Protected resource")
})
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
