package zinc

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBasicRouting(t *testing.T) {
	app := New()

	app.Get("/", func(c *Context) {
		c.Send("Hello World")
	})

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "GET root path",
			method:         "GET",
			path:           "/",
			expectedStatus: 200,
			expectedBody:   "Hello World",
		},
		{
			name:           "Non-existent path",
			method:         "GET",
			path:           "/notfound",
			expectedStatus: 404,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q; got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestParameterizedRoutes(t *testing.T) {
	app := New()

	app.Get("/users/:id", func(c *Context) {
		c.JSON(Map{"id": c.Param("id")})
	})

	app.Get("/files/*path", func(c *Context) {
		c.JSON(Map{"path": c.Param("*")})
	})

	tests := []struct {
		name          string
		path          string
		expectedParam string
		expectedKey   string
	}{
		{
			name:          "Named parameter",
			path:          "/users/123",
			expectedParam: "123",
			expectedKey:   "id",
		},
		{
			name:          "Wildcard parameter",
			path:          "/files/docs/report.pdf",
			expectedParam: "docs/report.pdf",
			expectedKey:   "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("expected status 200; got %d", w.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			paramKey := tt.expectedKey
			if tt.expectedKey == "*" {
				paramKey = "path"
			}

			if response[paramKey] != tt.expectedParam {
				t.Errorf("expected param %q; got %q", tt.expectedParam, response[paramKey])
			}
		})
	}
}

func TestMiddleware(t *testing.T) {
	app := New()

	// Add global middleware
	app.Use(func(c *Context) {
		c.Set("global", true)
		c.Next()
	})

	// Add route with middleware
	app.Get("/protected", func(c *Context) {
		c.Set("route", true)
		c.Next()
	}, func(c *Context) {
		globalVal := c.Get("global").(bool)
		routeVal := c.Get("route").(bool)
		c.JSON(Map{
			"global": globalVal,
			"route":  routeVal,
		})
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200; got %d", w.Code)
	}

	var response map[string]bool
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !response["global"] {
		t.Error("expected global middleware to be executed")
	}
	if !response["route"] {
		t.Error("expected route middleware to be executed")
	}
}

func TestHTTPMethods(t *testing.T) {
	app := New()

	methods := []string{
		MethodGet,
		MethodPost,
		MethodPut,
		MethodPatch,
		MethodDelete,
		MethodHead,
		MethodOptions,
	}

	// Register routes for all methods using App's method helpers
	for _, m := range methods {
		// Remove the method shadowing that was causing the issue
		switch m {
		case MethodGet:
			app.Get("/test", func(c *Context) {
				c.JSON(Map{"method": MethodGet})
			})
		case MethodPost:
			app.Post("/test", func(c *Context) {
				c.JSON(Map{"method": MethodPost})
			})
		case MethodPut:
			app.Put("/test", func(c *Context) {
				c.JSON(Map{"method": MethodPut})
			})
		case MethodPatch:
			app.Patch("/test", func(c *Context) {
				c.JSON(Map{"method": MethodPatch})
			})
		case MethodDelete:
			app.Delete("/test", func(c *Context) {
				c.JSON(Map{"method": MethodDelete})
			})
		case MethodHead:
			app.Head("/test", func(c *Context) {
				c.JSON(Map{"method": MethodHead})
			})
		case MethodOptions:
			app.Options("/test", func(c *Context) {
				c.JSON(Map{"method": MethodOptions})
			})
		}
	}

	// Rest of the test remains the same
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/test", nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("expected status 200; got %d for method %s", w.Code, method)
			}

			if method != MethodHead {
				var response map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}

				if response["method"] != method {
					t.Errorf("expected method %q; got %q", method, response["method"])
				}
			}
		})
	}
}

func TestJSONHandling(t *testing.T) {
	app := New()

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	app.Post("/users", func(c *Context) {
		var user User
		if err := c.Body(&user); err != nil {
			c.Status(400).JSON(Map{"error": err.Error()})
			return
		}
		c.JSON(user)
	})

	user := User{Name: "John", Age: 30}
	body, _ := json.Marshal(user)

	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200; got %d", w.Code)
	}

	var response User
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Name != user.Name || response.Age != user.Age {
		t.Errorf("expected user %+v; got %+v", user, response)
	}
}

func TestQueryParams(t *testing.T) {
	app := New()

	app.Get("/search", func(c *Context) {
		query := c.Query("q")
		page := c.Query("page")
		c.JSON(Map{
			"query": query,
			"page":  page,
		})
	})

	req := httptest.NewRequest("GET", "/search?q=test&page=1", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200; got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["query"] != "test" {
		t.Errorf("expected query 'test'; got %q", response["query"])
	}
	if response["page"] != "1" {
		t.Errorf("expected page '1'; got %q", response["page"])
	}
}

func TestServices(t *testing.T) {
	app := New()

	type UserService struct {
		prefix string
	}

	userService := &UserService{prefix: "User-"}
	app.Service("users", userService)

	app.Get("/service-test", func(c *Context) {
		service := c.Service("users").(*UserService)
		c.JSON(Map{"prefix": service.prefix})
	})

	req := httptest.NewRequest("GET", "/service-test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200; got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["prefix"] != "User-" {
		t.Errorf("expected prefix 'User-'; got %q", response["prefix"])
	}
}

func TestResponseTypes(t *testing.T) {
	app := New()

	app.Get("/text", func(c *Context) {
		c.Send("Hello World")
	})

	app.Get("/json", func(c *Context) {
		c.JSON(Map{"message": "Hello World"})
	})

	app.Get("/html", func(c *Context) {
		c.HTML("<h1>Hello World</h1>")
	})

	tests := []struct {
		name         string
		path         string
		contentType  string
		expectedBody string
	}{
		{
			name:         "Plain text response",
			path:         "/text",
			contentType:  "text/plain; charset=utf-8",
			expectedBody: "Hello World",
		},
		{
			name:         "JSON response",
			path:         "/json",
			contentType:  "application/json; charset=utf-8",
			expectedBody: "{\"message\":\"Hello World\"}\n",
		},
		{
			name:         "HTML response",
			path:         "/html",
			contentType:  "text/html; charset=utf-8",
			expectedBody: "<h1>Hello World</h1>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("expected status 200; got %d", w.Code)
			}

			if w.Header().Get("Content-Type") != tt.contentType {
				t.Errorf("expected Content-Type %q; got %q", tt.contentType, w.Header().Get("Content-Type"))
			}

			body, _ := io.ReadAll(w.Body)
			if string(body) != tt.expectedBody {
				t.Errorf("expected body %q; got %q", tt.expectedBody, string(body))
			}
		})
	}
}

func TestMethodNotAllowed(t *testing.T) {
	app := New()

	app.Get("/test", func(c *Context) {
		c.Send("GET only")
	})

	req := httptest.NewRequest(MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404; got %d", w.Code)
	}
}

func TestNestedRoutes(t *testing.T) {
	app := New()

	app.Get("/api/v1/users/:id/posts/:postId", func(c *Context) {
		c.JSON(Map{
			"userId": c.Param("id"),
			"postId": c.Param("postId"),
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/users/123/posts/456", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected status 200; got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["userId"] != "123" {
		t.Errorf("expected userId '123'; got %q", response["userId"])
	}
	if response["postId"] != "456" {
		t.Errorf("expected postId '456'; got %q", response["postId"])
	}
}

func TestMiddlewareOrder(t *testing.T) {
	app := New()
	order := []string{}

	app.Use(func(c *Context) {
		order = append(order, "global1")
		c.Next()
	})

	app.Use(func(c *Context) {
		order = append(order, "global2")
		c.Next()
	})

	app.Get("/test", func(c *Context) {
		order = append(order, "handler1")
		c.Next()
	}, func(c *Context) {
		order = append(order, "handler2")
		c.JSON(Map{"order": order})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	expected := []string{"global1", "global2", "handler1", "handler2"}
	if !reflect.DeepEqual(order, expected) {
		t.Errorf("expected order %v; got %v", expected, order)
	}
}

func TestContextStore(t *testing.T) {
	app := New()

	app.Use(func(c *Context) {
		c.Set("key1", "value1")
		c.Next()
	})

	app.Get("/test", func(c *Context) {
		c.Set("key2", "value2")
		val1 := c.Get("key1").(string)
		val2 := c.Get("key2").(string)
		c.JSON(Map{
			"key1": val1,
			"key2": val2,
		})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["key1"] != "value1" {
		t.Errorf("expected key1 'value1'; got %q", response["key1"])
	}
	if response["key2"] != "value2" {
		t.Errorf("expected key2 'value2'; got %q", response["key2"])
	}
}

func TestWildcardRoutes(t *testing.T) {
	app := New()

	app.Get("/static/*filepath", func(c *Context) {
		c.JSON(Map{"filepath": c.Param("*")})
	})

	tests := []struct {
		path     string
		expected string
	}{
		{"/static/css/style.css", "css/style.css"},
		{"/static/js/app.js", "js/app.js"},
		{"/static/images/logo.png", "images/logo.png"},
		{"/static/deep/nested/file.txt", "deep/nested/file.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)

			if w.Code != 200 {
				t.Errorf("expected status 200; got %d", w.Code)
			}

			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if response["filepath"] != tt.expected {
				t.Errorf("expected filepath %q; got %q", tt.expected, response["filepath"])
			}
		})
	}
}

func TestMiddlewareTermination(t *testing.T) {
	app := New()

	app.Use(func(c *Context) {
		c.Status(403).Send("Forbidden")
		// Don't call Next()
	})

	app.Get("/test", func(c *Context) {
		c.Send("Should not reach here")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Errorf("expected status 403; got %d", w.Code)
	}

	if w.Body.String() != "Forbidden" {
		t.Errorf("expected body 'Forbidden'; got %q", w.Body.String())
	}
}

func TestErrorHandling(t *testing.T) {
	app := New()

	app.Get("/error", func(c *Context) {
		c.Status(500).JSON(Map{"error": "Internal Server Error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("expected status 500; got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["error"] != "Internal Server Error" {
		t.Errorf("expected error 'Internal Server Error'; got %q", response["error"])
	}
}
