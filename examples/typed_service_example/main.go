package main

import (
	"log"
	"time"

	"github.com/0mjs/zinc"
)

// UserService handles user-related operations
type UserService struct {
	users []User
}

// User represents a user entity
type User struct {
	Name  string `json:"name" validate:"required,min=3,max=10"`
	Age   int    `json:"age" validate:"required"`
	Email string `json:"email" validate:"required,email,min=3,max=50"`
	Phone string `json:"phone" validate:"required,min=11,max=15"`
}

// GetUsers returns all users
func (s *UserService) GetUsers(c *zinc.Context) error {
	return c.JSON(s.users)
}

// CreateUser creates a new user
func (s *UserService) CreateUser(c *zinc.Context) error {
	var user User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(zinc.StatusBadRequest).JSON(zinc.Map{
			"error":   "Bad request",
			"message": err.Error(),
		})
	}

	s.users = append(s.users, user)
	log.Printf("User created: %s", user.Email)
	return c.JSON(user)
}

// GetUserByEmail finds a user by email
func (s *UserService) GetUserByEmail(email string) *User {
	for _, user := range s.users {
		if user.Email == email {
			return &user
		}
	}
	return nil
}

// ProductService handles product-related operations
type ProductService struct {
	products []Product
}

// Product represents a product entity
type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// GetProducts returns all products
func (s *ProductService) GetProducts(c *zinc.Context) error {
	return c.JSON(s.products)
}

func main() {
	app := zinc.New(zinc.Config{
		ServerHeader:      "Zinc-Example",
		AppName:           "Typed Service Example",
		AppVersion:        "1.0.0",
		IdleTimeout:       3 * time.Second,
		EnablePrintRoutes: true,
	})

	// Initialize services
	userService := &UserService{
		users: []User{
			{Name: "Alice", Age: 30, Email: "alice@example.com", Phone: "12345678901"},
		},
	}

	productService := &ProductService{
		products: []Product{
			{ID: "p1", Name: "Laptop", Price: 999.99},
			{ID: "p2", Name: "Phone", Price: 699.99},
		},
	}

	// Register services with type-based registration
	app.Register(userService)
	app.Register(productService)

	// Routes using type-safe service access with ServiceOf
	app.Get("/api/users", func(c *zinc.Context) error {
		service, ok := zinc.ServiceOf[*UserService](app)
		if !ok {
			return c.Status(zinc.StatusInternalServerError).String("UserService not available")
		}
		return service.GetUsers(c)
	})

	app.Post("/api/users", func(c *zinc.Context) error {
		service, ok := zinc.ServiceOf[*UserService](app)
		if !ok {
			return c.Status(zinc.StatusInternalServerError).String("UserService not available")
		}
		return service.CreateUser(c)
	})

	// Routes using type-safe service access with ContextServiceOf
	app.Get("/api/products", func(c *zinc.Context) error {
		service, ok := zinc.ContextServiceOf[*ProductService](c)
		if !ok {
			return c.Status(zinc.StatusInternalServerError).String("ProductService not available")
		}
		return service.GetProducts(c)
	})

	// Example of middleware that uses typed services
	productMiddleware := func(c *zinc.Context) error {
		// This middleware can access services without any app reference
		// The context already has proper access to the app instance
		service, ok := zinc.ContextServiceOf[*ProductService](c)
		if !ok {
			return c.Status(zinc.StatusInternalServerError).String("ProductService not available in middleware")
		}

		// Do something with the service in the middleware
		// For example, log the number of products
		log.Printf("Middleware: Found %d products", len(service.products))

		// Continue with the next handler
		return c.Next()
	}

	// Route with middleware that uses services
	app.Get("/api/middleware-products", productMiddleware, func(c *zinc.Context) error {
		// The handler can also access services
		service, ok := zinc.ContextServiceOf[*ProductService](c)
		if !ok {
			return c.Status(zinc.StatusInternalServerError).String("ProductService not available")
		}
		return c.JSON(service.products)
	})

	// Example of service-to-service communication
	app.Get("/api/user-with-products/:email", func(c *zinc.Context) error {
		email := c.Param("email")

		// Get both services
		userService, ok1 := zinc.ContextServiceOf[*UserService](c)
		productService, ok2 := zinc.ContextServiceOf[*ProductService](c)

		if !ok1 || !ok2 {
			return c.Status(zinc.StatusInternalServerError).String("Required services not available")
		}

		// Use services together
		user := userService.GetUserByEmail(email)
		if user == nil {
			return c.Status(zinc.StatusNotFound).String("User not found")
		}

		// Create a response with user and products
		response := struct {
			User     User      `json:"user"`
			Products []Product `json:"products"`
		}{
			User:     *user,
			Products: productService.products,
		}

		return c.JSON(response)
	})

	app.Serve()
}
