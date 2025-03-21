package main

import (
	"log"
	"time"

	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/middleware/cors"
)

func main() {
	app := zinc.New()

	// Add CORS middleware with default configuration
	// This allows all origins with the common methods and headers
	app.Use(cors.New())

	// Basic route to demonstrate CORS in action
	app.Get("/", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"message": "Hello, CORS!",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API group with custom CORS configuration
	api := app.Group("/api")

	// Use custom CORS configuration for this group
	api.Use(cors.NewWithOptions(
		// Only allow requests from specified origins
		cors.WithAllowOrigins("http://localhost:3000", "https://example.com"),
		// Allow credentials
		cors.WithAllowCredentials(true),
		// Add custom allowed headers
		cors.WithAllowHeaders("X-API-Key", "Content-Type", "Accept", "Authorization"),
		// Cache preflight response for 1 hour
		cors.WithMaxAge(time.Hour),
		// Expose custom headers to the browser
		cors.WithExposeHeaders("X-Request-ID", "X-API-Version"),
	))

	// API endpoints
	api.Get("/users", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"users": []zinc.Map{
				{"id": 1, "name": "Alice"},
				{"id": 2, "name": "Bob"},
				{"id": 3, "name": "Charlie"},
			},
		})
	})

	api.Post("/users", func(c *zinc.Context) error {
		// Add a custom header that will be exposed to the browser
		c.Response.Header().Set("X-Request-ID", "12345")

		// Create user logic would go here
		return c.Status(201).JSON(zinc.Map{
			"message": "User created successfully",
		})
	})

	// Start the server
	log.Fatal(app.Serve())
}
