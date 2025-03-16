package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/0mjs/zinc"
)

type User struct {
	ID        int    `json:"id" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Age       int    `json:"age" validate:"min=18"`
	CreatedAt string `json:"created_at,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required,email"`
	Password string `json:"password" form:"password" validate:"required,min=8"`
}

func main() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Create new Zinc app
	app := zinc.New()

	// Configure with custom timeouts by creating a custom config
	config := zinc.Config{
		DefaultAddr:     "0.0.0.0:8080",
		ShutdownTimeout: 30 * time.Second,
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    15 * time.Second,
		IdleTimeout:     120 * time.Second,
	}
	app.SetConfig(&config)

	// Setup a signal handler to demonstrate graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("\nReceived an interrupt, starting graceful shutdown...")
		// The actual graceful shutdown is handled by the Serve method
	}()

	// JSON binding example
	app.Post("/users", func(c *zinc.Context) {
		var user User
		if err := c.BindJSON(&user); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		// Set creation timestamp
		user.CreatedAt = time.Now().Format(time.RFC3339)

		c.Status(201).JSON(user)
	})

	// Form binding example
	app.Post("/login", func(c *zinc.Context) {
		var login LoginRequest
		if err := c.BindForm(&login); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		// Simulate authentication
		if login.Email == "user@example.com" && login.Password == "password123" {
			c.JSON(zinc.Map{"status": "success", "message": "Login successful"})
		} else {
			c.Status(401).JSON(zinc.Map{"status": "error", "message": "Invalid credentials"})
		}
	})

	// Query binding example
	app.Get("/search", func(c *zinc.Context) {
		type SearchParams struct {
			Query  string `query:"q" validate:"required"`
			Limit  int    `query:"limit" validate:"min=1,max=100"`
			Offset int    `query:"offset" validate:"min=0"`
		}

		var params SearchParams
		// Set default values
		params.Limit = 10

		if err := c.BindQuery(&params); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		c.JSON(zinc.Map{
			"query":   params.Query,
			"limit":   params.Limit,
			"offset":  params.Offset,
			"results": []string{"result1", "result2", "result3"},
		})
	})

	// Automatic binding based on Content-Type
	app.Post("/auto-bind", func(c *zinc.Context) {
		var login LoginRequest
		if err := c.Bind(&login); err != nil {
			c.Status(400).JSON(zinc.Map{"error": err.Error()})
			return
		}

		c.JSON(zinc.Map{"email": login.Email, "status": "success"})
	})

	// File upload example
	app.Post("/upload", func(c *zinc.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.Status(400).JSON(zinc.Map{"error": "Failed to get file"})
			return
		}

		// Save the file
		dst := fmt.Sprintf("./uploads/%s", file.Filename)
		if err := c.SaveFile(file, dst); err != nil {
			c.Status(500).JSON(zinc.Map{"error": "Failed to save file"})
			return
		}

		c.JSON(zinc.Map{"filename": file.Filename, "size": file.Size})
	})

	// Health check endpoint
	app.Get("/health", func(c *zinc.Context) {
		c.JSON(zinc.Map{"status": "ok", "timestamp": time.Now().Unix()})
	})

	fmt.Println("Server starting... Press Ctrl+C to initiate graceful shutdown")
	// Start the server with graceful shutdown support
	if err := app.Serve("8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
