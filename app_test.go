package zinc

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAppRouting(t *testing.T) {
	app := New()

	// Register some routes
	app.Get("/", func(c *Context) error {
		return c.Send("root")
	})

	app.Post("/users", func(c *Context) error {
		return c.Send("create user")
	})

	app.Put("/users/:id", func(c *Context) error {
		return c.Send("update user " + c.Param("id"))
	})

	app.Delete("/users/:id", func(c *Context) error {
		return c.Send("delete user " + c.Param("id"))
	})

	app.Patch("/users/:id", func(c *Context) error {
		return c.Send("patch user " + c.Param("id"))
	})

	app.Head("/users", func(c *Context) error {
		// HEAD response typically has no body
		return nil
	})

	app.Options("/users", func(c *Context) error {
		return c.Send("options for users")
	})

	// Test each route
	tests := []struct {
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{http.MethodGet, "/", 200, "root"},
		{http.MethodPost, "/users", 200, "create user"},
		{http.MethodPut, "/users/123", 200, "update user 123"},
		{http.MethodDelete, "/users/123", 200, "delete user 123"},
		{http.MethodPatch, "/users/123", 200, "patch user 123"},
		{http.MethodHead, "/users", 200, ""},
		{http.MethodOptions, "/users", 200, "options for users"},
		{http.MethodGet, "/not-found", 404, "404 page not found\n"},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			app.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status code = %d, want %d", w.Code, tt.wantStatus)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("Response body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestAppMiddleware(t *testing.T) {
	app := New()

	// Add global middleware
	app.Use(func(c *Context) error {
		c.Set("global", "middleware")
		return c.Next()
	})

	// Add route with middleware
	app.Get("/middleware", func(c *Context) error {
		c.Set("first", "middleware")
		return c.Next()
	}, func(c *Context) error {
		c.Set("second", "middleware")
		return c.Next()
	}, func(c *Context) error {
		global := c.Get("global")
		first := c.Get("first")
		second := c.Get("second")
		return c.Send(global.(string) + " " + first.(string) + " " + second.(string))
	})

	// Test middleware execution
	req := httptest.NewRequest(http.MethodGet, "/middleware", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	wantBody := "middleware middleware middleware"
	if w.Body.String() != wantBody {
		t.Errorf("Response body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestAppGroup(t *testing.T) {
	app := New()

	// Create route groups
	api := app.Group("/api")
	api.Get("/status", func(c *Context) error {
		return c.Send("API Status")
	})

	v1 := api.Group("/v1")
	v1.Get("/users", func(c *Context) error {
		return c.Send("API v1 Users")
	})

	v2 := api.Group("/v2")
	v2.Get("/users", func(c *Context) error {
		return c.Send("API v2 Users")
	})

	// Test grouped routes
	tests := []struct {
		path     string
		wantBody string
	}{
		{"/api/status", "API Status"},
		{"/api/v1/users", "API v1 Users"},
		{"/api/v2/users", "API v2 Users"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("Status code = %d, want %d", w.Code, 200)
			}

			if w.Body.String() != tt.wantBody {
				t.Errorf("Response body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestAppGroupMiddleware(t *testing.T) {
	app := New()

	// Create route group with middleware
	api := app.Group("/api")
	api.Use(func(c *Context) error {
		c.Set("group", "middleware")
		return c.Next()
	})

	api.Get("/test", func(c *Context) error {
		groupMw := c.Get("group")
		return c.Send(groupMw.(string))
	})

	// Test group middleware
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	wantBody := "middleware"
	if w.Body.String() != wantBody {
		t.Errorf("Response body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestAppServices(t *testing.T) {
	app := New()

	// Create a test service
	type TestService struct {
		GetValue func() string
	}

	service := &TestService{
		GetValue: func() string {
			return "service value"
		},
	}

	// Register the service
	app.Service("test", service)

	// Add a route that uses the service
	app.Get("/service", func(c *Context) error {
		service := c.Service("test").(*TestService)
		return c.Send(service.GetValue())
	})

	// Test accessing the service
	req := httptest.NewRequest(http.MethodGet, "/service", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	wantBody := "service value"
	if w.Body.String() != wantBody {
		t.Errorf("Response body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestAppStaticHandler(t *testing.T) {
	app := New()

	// String route handler
	app.Get("/static-string", "Hello from static string")

	// Test static string handler
	req := httptest.NewRequest(http.MethodGet, "/static-string", nil)
	w := httptest.NewRecorder()

	app.ServeHTTP(w, req)

	wantBody := "Hello from static string"
	if w.Body.String() != wantBody {
		t.Errorf("Response body = %q, want %q", w.Body.String(), wantBody)
	}
}

func TestAppTestConfig(t *testing.T) {
	// Test default config
	app1 := New()
	if app1.config.DefaultAddr != "0.0.0.0:6530" {
		t.Errorf("Default address = %q, want %q", app1.config.DefaultAddr, "0.0.0.0:6530")
	}

	// Test custom config
	customConfig := &Config{
		DefaultAddr: "127.0.0.1:3000",
	}

	app2 := New()
	app2.config = customConfig

	if app2.config.DefaultAddr != "127.0.0.1:3000" {
		t.Errorf("Custom address = %q, want %q", app2.config.DefaultAddr, "127.0.0.1:3000")
	}
}

func TestAppFastPath(t *testing.T) {
	app := New()

	// Add a simple static route for fast path testing
	app.Get("/fast", func(c *Context) error {
		return c.Send("Hello World!")
	})

	// Make multiple requests to the same route to trigger fast path
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/fast", nil)
		w := httptest.NewRecorder()

		app.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Errorf("Iteration %d: Status code = %d, want %d", i, w.Code, 200)
		}

		if w.Body.String() != "Hello World!" {
			t.Errorf("Iteration %d: Response body = %q, want %q", i, w.Body.String(), "Hello World!")
		}
	}
}

func benchmarkApp(b *testing.B, path string) {
	app := New()

	// Add routes
	app.Get("/", func(c *Context) error {
		return c.Send("root")
	})

	app.Get("/users", func(c *Context) error {
		return c.Send("users")
	})

	app.Get("/users/:id", func(c *Context) error {
		return c.Send("user " + c.Param("id"))
	})

	app.Get("/users/:id/posts", func(c *Context) error {
		return c.Send("user " + c.Param("id") + " posts")
	})

	// Run the benchmark
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, path, nil)
		app.ServeHTTP(w, req)
	}
}

func BenchmarkAppRootRoute(b *testing.B) {
	benchmarkApp(b, "/")
}

func BenchmarkAppStaticRoute(b *testing.B) {
	benchmarkApp(b, "/users")
}

func BenchmarkAppParameterRoute(b *testing.B) {
	benchmarkApp(b, "/users/123")
}

func BenchmarkAppNestedRoute(b *testing.B) {
	benchmarkApp(b, "/users/123/posts")
}
