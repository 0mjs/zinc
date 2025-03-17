# zinc.

![Zinc](https://img.shields.io/badge/Zinc-%20A%20web%20framework%20for%20Go-silver)
![Version](https://img.shields.io/badge/version-0.0.54-red)
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
- [Official Guide]([https://github.com/0mjs/zinc/docs](https://zinc.0mjs.dev/)

## Benchmarks

Zinc is designed for high performance, with benchmarks showing it to be competitive with or faster than other popular Go frameworks:

- Static routes: ~800ns/op
- Dynamic routes: ~1.2μs/op
- Middleware chain: ~2.0μs/op

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
