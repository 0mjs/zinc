package main

import (
	"log"
	"strconv"
	"time"

	"github.com/0mjs/zinc"
)

func main() {
	// Create a new Zinc app with custom configuration
	app := zinc.New(zinc.Config{
		// Server configuration
		DefaultAddr:  "127.0.0.1:3000",
		ServerHeader: "MyAwesomeAPI",
		AppName:      "My Zinc App",

		// Timeouts
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     60 * time.Second,
		ShutdownTimeout: 30 * time.Second,

		// Router settings
		CaseSensitive:  true,             // /Users and /users are different routes
		StrictRouting:  true,             // /api and /api/ are different routes
		BodyLimit:      10 * 1024 * 1024, // 10MB max request size
		Concurrency:    1000,             // Maximum of 1000 concurrent connections
		RouteCacheSize: 10000,            // Cache up to 10,000 routes

		// Proxy settings
		EnableTrustedProxyCheck: true,
		TrustedProxies:          []string{"127.0.0.1", "10.0.0.0/8"},
		ProxyHeader:             "X-Forwarded-For",

		// Miscellaneous
		DisableKeepalive:          false,
		DisableDefaultContentType: false,
		DisableStartupMessage:     false,
		EnablePrintRoutes:         true,
	})

	// Add a route
	app.Get("/", func(c *zinc.Context) error {
		return c.Send("Hello from configured Zinc server!")
	})

	// Route that shows IP address - demonstrates proxy settings
	app.Get("/ip", func(c *zinc.Context) error {
		return c.Send("Your IP is: " + c.IP())
	})

	// Route that reads request body - demonstrates body limit
	app.Post("/upload", func(c *zinc.Context) error {
		body, err := c.Body()
		if err != nil {
			return c.Status(400).Send("Error reading body: " + err.Error())
		}
		return c.Send("Received body of length: " + strconv.Itoa(len(body)))
	})

	// Routes that demonstrate case sensitivity
	app.Get("/Users", func(c *zinc.Context) error {
		return c.Send("This is the /Users route (uppercase)")
	})

	app.Get("/users", func(c *zinc.Context) error {
		return c.Send("This is the /users route (lowercase)")
	})

	// Routes that demonstrate strict routing
	app.Get("/api", func(c *zinc.Context) error {
		return c.Send("This is the /api route (no trailing slash)")
	})

	app.Get("/api/", func(c *zinc.Context) error {
		return c.Send("This is the /api/ route (with trailing slash)")
	})

	// Start the server
	log.Fatal(app.Serve())
}
