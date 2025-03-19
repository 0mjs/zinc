package main

import (
	"fmt"
	"log"

	"github.com/0mjs/zinc"
)

func main() {
	// Create a new Zinc application
	app := zinc.New(zinc.Config{
		CaseSensitive: true,
	})

	// Create a POST route for testing validation
	app.Post("/users", func(c *zinc.Context) error {
		// Define a user struct with validation tags
		var user struct {
			Name  string `json:"name" validate:"required"`
			Email string `json:"email" validate:"required,email"`
		}

		// Try to parse and validate
		err := c.BodyParser(&user)
		if err != nil {
			log.Printf("Validation error: %v", err)
			return c.Status(zinc.StatusBadRequest).String(fmt.Sprintf("Validation failed: %v", err))
		}

		// Return the user data
		return c.JSON(user)
	})

	// Start the server
	log.Println("Server started on http://localhost:8080")
	app.Serve()
}
