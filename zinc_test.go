package zinc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// TestRouting tests basic routing functionality
func TestRouting(t *testing.T) {
	app := New()
	app.Get("/", func(c *Context) error {
		return c.Send("root")
	})
	app.Get("/hello", func(c *Context) error {
		return c.Send("Hello World!")
	})
	app.Post("/post", func(c *Context) error {
		return c.Send("post")
	})
	app.Put("/put", func(c *Context) error {
		return c.Send("put")
	})
	app.Delete("/delete", func(c *Context) error {
		return c.Send("delete")
	})
	app.Patch("/patch", func(c *Context) error {
		return c.Send("patch")
	})
	app.Head("/head", func(c *Context) error {
		// HEAD requests typically don't have a body
		return nil
	})
	app.Options("/options", func(c *Context) error {
		return c.Send("options")
	})

	tests := []struct {
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{http.MethodGet, "/", 200, "root"},
		{http.MethodGet, "/hello", 200, "Hello World!"},
		{http.MethodPost, "/post", 200, "post"},
		{http.MethodPut, "/put", 200, "put"},
		{http.MethodDelete, "/delete", 200, "delete"},
		{http.MethodPatch, "/patch", 200, "patch"},
		{http.MethodHead, "/head", 200, ""},
		{http.MethodOptions, "/options", 200, "options"},
		{http.MethodGet, "/notfound", 404, "404 page not found\n"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.method, tt.path), func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestRouteParams tests route parameter handling
func TestRouteParams(t *testing.T) {
	app := New()
	app.Get("/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.Send(fmt.Sprintf("User ID: %s", id))
	})
	app.Get("/posts/:postID/comments/:commentID", func(c *Context) error {
		postID := c.Param("postID")
		commentID := c.Param("commentID")
		return c.Send(fmt.Sprintf("Post %s, Comment %s", postID, commentID))
	})
	app.Get("/files/*path", func(c *Context) error {
		path := c.Param("*")
		return c.Send(fmt.Sprintf("File path: %s", path))
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/users/123", "User ID: 123"},
		{"/users/abc", "User ID: abc"},
		{"/posts/123/comments/456", "Post 123, Comment 456"},
		{"/files/images/avatar.png", "File path: images/avatar.png"},
		{"/files/documents/report.pdf", "File path: documents/report.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("want status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestRouteGroups tests route group functionality
func TestRouteGroups(t *testing.T) {
	app := New()

	api := app.Group("api")
	api.Get("/status", func(c *Context) error {
		return c.Send("API Status: OK")
	})

	v1 := api.Group("v1")
	v1.Get("/users", func(c *Context) error {
		return c.Send("API v1 Users")
	})

	v2 := api.Group("v2")
	v2.Get("/users", func(c *Context) error {
		return c.Send("API v2 Users")
	})

	admin := app.Group("admin")
	admin.Get("/dashboard", func(c *Context) error {
		return c.Send("Admin Dashboard")
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/api/status", "API Status: OK"},
		{"/api/v1/users", "API v1 Users"},
		{"/api/v2/users", "API v2 Users"},
		{"/admin/dashboard", "Admin Dashboard"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("want status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestQueryParams tests query parameter handling
func TestQueryParams(t *testing.T) {
	app := New()
	app.Get("/search", func(c *Context) error {
		q := c.Query("q")
		limit := c.Query("limit")
		return c.Send(fmt.Sprintf("Query: %s, Limit: %s", q, limit))
	})

	app.Get("/hasquery", func(c *Context) error {
		if c.HasQuery("exists") {
			return c.Send("Has Query")
		} else {
			return c.Send("Missing Query")
		}
	})

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/search?q=golang&limit=10", "Query: golang, Limit: 10"},
		{"/search?q=zinc", "Query: zinc, Limit: "},
		{"/search", "Query: , Limit: "},
		{"/hasquery?exists=true", "Has Query"},
		{"/hasquery", "Missing Query"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("want status 200, got %d", w.Code)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestResponseTypes tests different response types
func TestResponseTypes(t *testing.T) {
	app := New()
	app.Get("/text", func(c *Context) error {
		return c.Send("Plain text response")
	})
	app.Get("/bytes", func(c *Context) error {
		return c.Send([]byte("Bytes response"))
	})
	app.Get("/json", func(c *Context) error {
		return c.JSON(Map{"message": "JSON response", "status": "success"})
	})
	app.Get("/html", func(c *Context) error {
		return c.HTML("<h1>HTML response</h1>")
	})
	app.Get("/nil", func(c *Context) error {
		return c.Send(nil)
	})
	app.Get("/status", func(c *Context) error {
		return c.Status(http.StatusCreated).Send("Created")
	})

	tests := []struct {
		path            string
		wantStatus      int
		wantBody        string
		wantContains    string
		wantContentType string
	}{
		{"/text", 200, "Plain text response", "", "text/plain; charset=utf-8"},
		{"/bytes", 200, "Bytes response", "", "application/octet-stream"},
		{"/json", 200, "", `{"message":"JSON response","status":"success"}`, "application/json; charset=utf-8"},
		{"/html", 200, "<h1>HTML response</h1>", "", "text/html; charset=utf-8"},
		{"/nil", 200, "null", "", "application/json; charset=utf-8"},
		{"/status", 201, "Created", "", "text/plain; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantBody != "" && w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}

			if tt.wantContains != "" && !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("want body containing %q, got %q", tt.wantContains, w.Body.String())
			}

			if contentType := w.Header().Get("Content-Type"); contentType != tt.wantContentType {
				t.Errorf("want Content-Type %q, got %q", tt.wantContentType, contentType)
			}
		})
	}
}

// TestMiddleware tests middleware functionality
func TestMiddleware(t *testing.T) {
	app := New()

	// Add global middleware
	app.Use(func(c *Context) error {
		c.Set("global", "set")
		return c.Next()
	})

	// Route with route-specific middleware
	app.Get("/middleware", func(c *Context) error {
		c.Set("route", "set")
		return c.Next()
	}, func(c *Context) error {
		globalVal := c.Get("global")
		routeVal := c.Get("route")
		return c.Send(fmt.Sprintf("Global: %v, Route: %v", globalVal, routeVal))
	})

	// Route that tests middleware termination
	app.Get("/terminate", func(c *Context) error {
		return c.Send("Early response")
		// Not calling Next() should terminate the chain
	}, func(c *Context) error {
		// This should never be called
		return c.Send("Should not reach here")
	})

	// Test middleware order
	var order []string
	app.Get("/order",
		func(c *Context) error {
			order = append(order, "first")
			return c.Next()
		},
		func(c *Context) error {
			order = append(order, "second")
			return c.Next()
		},
		func(c *Context) error {
			order = append(order, "third")
			return c.Send(strings.Join(order, ","))
		},
	)

	tests := []struct {
		path     string
		wantBody string
	}{
		{"/middleware", "Global: set, Route: set"},
		{"/terminate", "Early response"},
		{"/order", "first,second,third"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			// Reset order slice for the /order test
			if tt.path == "/order" {
				order = []string{}
			}

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Body.String() != tt.wantBody {
				t.Errorf("want body %q, got %q", tt.wantBody, w.Body.String())
			}
		})
	}
}

// TestRequestBody tests request body handling
func TestRequestBody(t *testing.T) {
	app := New()
	app.Post("/json", func(c *Context) error {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		var user User
		if err := c.BodyParser(&user); err != nil {
			return c.Status(http.StatusBadRequest).Send(err.Error())
		}
		return c.JSON(Map{"received": user})
	})

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid JSON",
			body:       `{"name":"John","email":"john@example.com"}`,
			wantStatus: 200,
			wantBody:   `{"received":{"name":"John","email":"john@example.com"}}`,
		},
		{
			name:       "invalid JSON",
			body:       `{"name":"John", INVALID}`,
			wantStatus: 400,
			wantBody:   "invalid character 'I' looking for beginning of object key string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/json", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			// Normalize response by removing whitespace and newlines for comparison
			gotBody := strings.TrimSpace(w.Body.String())
			wantBody := strings.TrimSpace(tt.wantBody)
			if !strings.Contains(gotBody, wantBody) {
				t.Errorf("want body containing %q, got %q", wantBody, gotBody)
			}
		})
	}
}

// TestContextStore tests the context store functionality
func TestContextStore(t *testing.T) {
	app := New()
	app.Get("/store", func(c *Context) error {
		c.Set("string", "value")
		c.Set("number", 123)
		c.Set("bool", true)

		stringVal := c.Get("string")
		numberVal := c.Get("number")
		boolVal := c.Get("bool")

		return c.JSON(Map{
			"string": stringVal,
			"number": numberVal,
			"bool":   boolVal,
			"nil":    c.Get("nonexistent"),
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/store", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if result["string"] != "value" {
		t.Errorf("want string value %q, got %v", "value", result["string"])
	}

	// JSON numbers are floats in Go's map[string]interface{}
	if result["number"] != float64(123) {
		t.Errorf("want number value %v, got %v", float64(123), result["number"])
	}

	if result["bool"] != true {
		t.Errorf("want bool value %v, got %v", true, result["bool"])
	}

	if result["nil"] != nil {
		t.Errorf("want nil value, got %v", result["nil"])
	}
}

// TestMultipleResponses tests that multiple responses are handled correctly
func TestMultipleResponses(t *testing.T) {
	app := New()
	app.Get("/multiple", func(c *Context) error {
		// First response should work
		c.Send("First response")

		// Second response should be ignored and return an error
		err := c.Send("Second response")
		if err == nil {
			t.Error("Expected error for second response, got nil")
		}

		// Check that the error is as expected
		if err != ErrResponseAlreadySent {
			t.Errorf("Expected ErrResponseAlreadySent, got %v", err)
		}
		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/multiple", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "First response" {
		t.Errorf("want body %q, got %q", "First response", w.Body.String())
	}
}

// TestAppConfig tests app configuration
func TestAppConfig(t *testing.T) {
	// Default configuration
	app1 := New()
	if app1.config.DefaultAddr != "0.0.0.0:6538" {
		t.Errorf("want default address %q, got %q", "0.0.0.0:6538", app1.config.DefaultAddr)
	}

	// Custom configuration
	customConfig := Config{
		DefaultAddr: "127.0.0.1:3000",
	}
	app2 := New()
	app2.config = &customConfig

	if app2.config.DefaultAddr != "127.0.0.1:3000" {
		t.Errorf("want custom address %q, got %q", "127.0.0.1:3000", app2.config.DefaultAddr)
	}
}

// TestServices tests service registration and retrieval
func TestServices(t *testing.T) {
	app := New()

	// Register a service
	type UserService struct {
		GetUser func(id string) string
	}

	userService := &UserService{
		GetUser: func(id string) string {
			return "User " + id
		},
	}

	app.Injectable(userService)

	// Route that uses the service
	app.Get("/users/:id", func(c *Context) error {
		service := c.Service((*UserService)(nil)).(*UserService)
		userId := c.Param("id")
		return c.Send(service.GetUser(userId))
	})

	// Route that uses a non-existent service
	app.Get("/unknown-service", func(c *Context) error {
		type NonExistentService struct{}
		service := c.Service((*NonExistentService)(nil))
		if service == nil {
			return c.Status(http.StatusInternalServerError).Send("Service not found")
		}
		return c.Send("Should not reach here")
	})

	tests := []struct {
		path         string
		wantStatus   int
		wantContains string
	}{
		{"/users/123", 200, "User 123"},
		{"/unknown-service", 500, "Service not found"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
			}

			if !strings.Contains(w.Body.String(), tt.wantContains) {
				t.Errorf("want body containing %q, got %q", tt.wantContains, w.Body.String())
			}
		})
	}
}

// TestGETRequest tests a form data handling
func TestFormData(t *testing.T) {
	app := New()
	app.Post("/form", func(c *Context) error {
		if err := c.Request.ParseForm(); err != nil {
			return c.Status(http.StatusBadRequest).Send(err.Error())
		}

		name := c.Request.FormValue("name")
		email := c.Request.FormValue("email")
		return c.Send(fmt.Sprintf("Name: %s, Email: %s", name, email))
	})

	formData := "name=John+Doe&email=john@example.com"
	req := httptest.NewRequest(http.MethodPost, "/form", bytes.NewBufferString(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("want status 200, got %d", w.Code)
	}

	if w.Body.String() != "Name: John Doe, Email: john@example.com" {
		t.Errorf("want %q, got %q", "Name: John Doe, Email: john@example.com", w.Body.String())
	}
}

// TestStaticFiles tests the static file serving capability
func TestStaticFiles(t *testing.T) {
	// Create a temporary test file
	tempFile := t.TempDir() + "/test.txt"
	if err := os.WriteFile(tempFile, []byte("Static file content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	app := New()
	app.Get("/static", func(c *Context) error {
		return c.Static(tempFile)
	})

	req := httptest.NewRequest(http.MethodGet, "/static", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("want status 200, got %d", w.Code)
	}

	if w.Body.String() != "Static file content" {
		t.Errorf("want %q, got %q", "Static file content", w.Body.String())
	}
}

// TestPerformance is a simple performance test
func TestPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	app := New()
	app.Get("/perf", func(c *Context) error {
		return c.Send("Hello World!")
	})

	req := httptest.NewRequest(http.MethodGet, "/perf", nil)

	// Warm up
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}

	// Actual test
	iterations := 10000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
	}
	elapsed := time.Since(start)

	// Log performance metrics
	t.Logf("Performed %d requests in %v", iterations, elapsed)
	t.Logf("Average request time: %v", elapsed/time.Duration(iterations))
}

// TestDefaultAddrEmptyConfig tests the fix for the DefaultAddr port bug, when a config is passed with an empty DefaultAddr.
func TestDefaultAddrEmptyConfig(t *testing.T) {
	// Test with an empty DefaultAddr in the config
	emptyAddrConfig := Config{
		// Leave DefaultAddr empty but set some other option
		ServerHeader: "TestServer",
	}

	app := New(emptyAddrConfig)

	if app.config.DefaultAddr != "0.0.0.0:6538" {
		t.Errorf("want default address %q when DefaultAddr is empty, got %q",
			"0.0.0.0:6538", app.config.DefaultAddr)
	}

	// Verify other settings were preserved
	if app.config.ServerHeader != "TestServer" {
		t.Errorf("want ServerHeader %q, got %q", "TestServer", app.config.ServerHeader)
	}
}
