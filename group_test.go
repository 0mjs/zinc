package zinc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGroupBasic(t *testing.T) {
	app := New()

	// Create a group
	api := app.Group("/api")

	// Add a route to the group
	api.Get("/test", func(c *Context) error {
		return c.Send("API test")
	})

	// Test the route
	req := httptest.NewRequest("GET", "/api/test", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.Code)
	}

	if resp.Body.String() != "API test" {
		t.Errorf("Expected body %q but got %q", "API test", resp.Body.String())
	}
}

func TestGroupNested(t *testing.T) {
	app := New()

	// Create nested groups
	api := app.Group("/api")
	v1 := api.Group("/v1")
	users := v1.Group("/users")

	// Add routes to different levels
	api.Get("/status", func(c *Context) error {
		return c.Send("API status")
	})

	v1.Get("/info", func(c *Context) error {
		return c.Send("API v1 info")
	})

	users.Get("/:id", func(c *Context) error {
		return c.Send("User " + c.Param("id"))
	})

	// Test routes
	tests := []struct {
		path     string
		method   string
		status   int
		expected string
	}{
		{"/api/status", "GET", http.StatusOK, "API status"},
		{"/api/v1/info", "GET", http.StatusOK, "API v1 info"},
		{"/api/v1/users/123", "GET", http.StatusOK, "User 123"},
		{"/api/v2/info", "GET", http.StatusNotFound, ""},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, test.path, nil)
		resp := httptest.NewRecorder()
		app.ServeHTTP(resp, req)

		if resp.Code != test.status {
			t.Errorf("Path %s: Expected status %d but got %d", test.path, test.status, resp.Code)
		}

		if test.status == http.StatusOK && resp.Body.String() != test.expected {
			t.Errorf("Path %s: Expected body %q but got %q", test.path, test.expected, resp.Body.String())
		}
	}
}

func TestGroupMiddleware(t *testing.T) {
	app := New()

	// Create a group with middleware
	api := app.Group("/api")
	api.Use(func(c *Context) error {
		c.Set("test", "middleware-value")
		return c.Next()
	})

	// Add routes
	api.Get("/test", func(c *Context) error {
		value := c.Get("test").(string)
		return c.Send(value)
	})

	// Normal route without middleware
	app.Get("/test", func(c *Context) error {
		return c.Send("no-middleware")
	})

	// Test middleware propagation
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/test", "middleware-value"},
		{"/test", "no-middleware"},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", test.path, nil)
		resp := httptest.NewRecorder()
		app.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("Path %s: Expected status %d but got %d", test.path, http.StatusOK, resp.Code)
		}

		if resp.Body.String() != test.expected {
			t.Errorf("Path %s: Expected body %q but got %q", test.path, test.expected, resp.Body.String())
		}
	}
}

func TestGroupMiddlewareOrder(t *testing.T) {
	app := New()

	// Create nested groups with middleware
	api := app.Group("/api")
	api.Use(func(c *Context) error {
		c.Set("order", "1")
		return c.Next()
	})

	v1 := api.Group("/v1")
	v1.Use(func(c *Context) error {
		c.Set("order", c.Get("order").(string)+"2")
		return c.Next()
	})

	// Add a route that uses the middleware
	v1.Get("/test", func(c *Context) error {
		c.Set("order", c.Get("order").(string)+"3")
		return c.Send(c.Get("order").(string))
	})

	// Test middleware order
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	resp := httptest.NewRecorder()
	app.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status %d but got %d", http.StatusOK, resp.Code)
	}

	if resp.Body.String() != "123" {
		t.Errorf("Expected body %q but got %q", "123", resp.Body.String())
	}
}
