package zinc

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCustomConfig(t *testing.T) {
	// Create app with custom config
	app := New(Config{
		DefaultAddr:               "127.0.0.1:5000",
		ServerHeader:              "TestServer",
		AppName:                   "TestApp",
		CaseSensitive:             true,
		StrictRouting:             true,
		BodyLimit:                 1024 * 1024, // 1MB
		DisableStartupMessage:     true,
		DisableDefaultContentType: true,
	})

	// Check if the config was set correctly
	if app.config.DefaultAddr != "127.0.0.1:5000" {
		t.Errorf("DefaultAddr = %s, want %s", app.config.DefaultAddr, "127.0.0.1:5000")
	}

	if app.config.ServerHeader != "TestServer" {
		t.Errorf("ServerHeader = %s, want %s", app.config.ServerHeader, "TestServer")
	}

	if app.config.AppName != "TestApp" {
		t.Errorf("AppName = %s, want %s", app.config.AppName, "TestApp")
	}

	if !app.config.CaseSensitive {
		t.Errorf("CaseSensitive = %v, want %v", app.config.CaseSensitive, true)
	}

	if !app.config.StrictRouting {
		t.Errorf("StrictRouting = %v, want %v", app.config.StrictRouting, true)
	}

	if app.config.BodyLimit != 1024*1024 {
		t.Errorf("BodyLimit = %d, want %d", app.config.BodyLimit, 1024*1024)
	}
}

func TestServerHeader(t *testing.T) {
	// Create app with custom server header
	app := New(Config{
		ServerHeader: "CustomServer",
	})

	// Register a route
	app.Get("/header-test", func(c *Context) error {
		return c.Send("OK")
	})

	// Make a request
	req := httptest.NewRequest(http.MethodGet, "/header-test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Check if the server header was set
	if w.Header().Get("Server") != "CustomServer" {
		t.Errorf("Server header = %s, want %s", w.Header().Get("Server"), "CustomServer")
	}
}

func TestCaseSensitiveRouting(t *testing.T) {
	tests := []struct {
		name          string
		caseSensitive bool
		definedPath   string
		requestPath   string
		shouldMatch   bool
	}{
		{
			name:          "Case sensitive - same case",
			caseSensitive: true,
			definedPath:   "/Users",
			requestPath:   "/Users",
			shouldMatch:   true,
		},
		{
			name:          "Case sensitive - different case",
			caseSensitive: true,
			definedPath:   "/Users",
			requestPath:   "/users",
			shouldMatch:   false,
		},
		{
			name:          "Case insensitive - different case",
			caseSensitive: false,
			definedPath:   "/Users",
			requestPath:   "/users",
			shouldMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create app with custom config
			app := New(Config{
				CaseSensitive: tt.caseSensitive,
			})

			// Register a route
			app.Get(tt.definedPath, func(c *Context) error {
				return c.Send("OK")
			})

			// Make a request
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			// Check if it matched
			if tt.shouldMatch && w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if !tt.shouldMatch && w.Code == http.StatusOK {
				t.Errorf("Expected status 404, got %d", w.Code)
			}
		})
	}
}

func TestStrictRouting(t *testing.T) {
	tests := []struct {
		name          string
		strictRouting bool
		definedPath   string
		requestPath   string
		shouldMatch   bool
	}{
		{
			name:          "Strict routing - same path",
			strictRouting: true,
			definedPath:   "/users",
			requestPath:   "/users",
			shouldMatch:   true,
		},
		{
			name:          "Strict routing - trailing slash",
			strictRouting: true,
			definedPath:   "/users",
			requestPath:   "/users/",
			shouldMatch:   false,
		},
		{
			name:          "Non-strict routing - trailing slash",
			strictRouting: false,
			definedPath:   "/users",
			requestPath:   "/users/",
			shouldMatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create app with custom config
			app := New(Config{
				StrictRouting: tt.strictRouting,
			})

			// Register a route
			app.Get(tt.definedPath, func(c *Context) error {
				return c.Send("OK")
			})

			// Make a request
			req := httptest.NewRequest(http.MethodGet, tt.requestPath, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			// Check if it matched
			if tt.shouldMatch && w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			if !tt.shouldMatch && w.Code == http.StatusOK {
				t.Errorf("Expected status 404, got %d", w.Code)
			}
		})
	}
}

func TestBodyLimit(t *testing.T) {
	// Create a test request with a large body
	largeBody := make([]byte, 2*1024) // 2KB
	for i := range largeBody {
		largeBody[i] = 'a'
	}

	// Create app with small body limit
	app := New(Config{
		BodyLimit: 1 * 1024, // 1KB
	})

	// Register a route that reads the body
	app.Post("/body-test", func(c *Context) error {
		body, err := c.Body()
		if err != nil {
			return c.Status(http.StatusBadRequest).Send(err.Error())
		}
		return c.Send(body)
	})

	// Make a request with a body larger than the limit
	req := httptest.NewRequest(http.MethodPost, "/body-test", strings.NewReader(string(largeBody)))
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Should get an error
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "exceeds the limit") {
		t.Errorf("Expected error message about body size limit, got %s", w.Body.String())
	}
}

func TestDisableDefaultContentType(t *testing.T) {
	tests := []struct {
		name                      string
		disableDefaultContentType bool
		expectContentType         bool
	}{
		{
			name:                      "Default content type enabled",
			disableDefaultContentType: false,
			expectContentType:         true,
		},
		{
			name:                      "Default content type disabled",
			disableDefaultContentType: true,
			expectContentType:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create app with custom config
			app := New(Config{
				DisableDefaultContentType: tt.disableDefaultContentType,
			})

			// Register a route that just sends a string (no content type set)
			app.Get("/content-type-test", func(c *Context) error {
				return c.Send("OK")
			})

			// Make a request
			req := httptest.NewRequest(http.MethodGet, "/content-type-test", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			// Check content type
			contentType := w.Header().Get("Content-Type")
			hasContentType := contentType != ""

			if tt.expectContentType && !hasContentType {
				t.Errorf("Expected Content-Type header, but none was set")
			}

			if !tt.expectContentType && hasContentType {
				t.Errorf("Expected no Content-Type header, but got %q", contentType)
			}
		})
	}
}
