package zinc

import (
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestRouterBasic(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add some routes
	router.Add(MethodGet, "/", func(c *Context) error {
		return c.Send("root")
	})
	router.Add(MethodGet, "/users", func(c *Context) error {
		return c.Send("users")
	})
	router.Add(MethodGet, "/users/:id", func(c *Context) error {
		id := c.Param("id")
		return c.Send("user:" + id)
	})
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) error {
		id := c.Param("id")
		return c.Send("posts for user:" + id)
	})
	router.Add(MethodGet, "/files/*path", func(c *Context) error {
		path := c.Param("*")
		return c.Send("file:" + path)
	})

	tests := []struct {
		path     string
		wantSend string
	}{
		{"/", "root"},
		{"/users", "users"},
		{"/users/123", "user:123"},
		{"/users/456/posts", "posts for user:456"},
		{"/files/images/logo.png", "file:images/logo.png"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler, ctx := router.Find(MethodGet, tt.path)
			if handler == nil {
				t.Fatalf("Failed to find route for path: %s", tt.path)
			}

			w := httptest.NewRecorder()
			ctx.Response = w
			ctx.Request = httptest.NewRequest(MethodGet, tt.path, nil)
			handler(ctx)

			if got := w.Body.String(); got != tt.wantSend {
				t.Errorf("Expected body %q, got %q", tt.wantSend, got)
			}
		})
	}
}

func TestRouterParamPriority(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add routes with parameter and static paths
	router.Add(MethodGet, "/users/:id", func(c *Context) error {
		return c.Send("user:" + c.Param("id"))
	})
	router.Add(MethodGet, "/users/new", func(c *Context) error {
		return c.Send("new user form")
	})
	router.Add(MethodGet, "/users/admin", func(c *Context) error {
		return c.Send("admin user")
	})

	tests := []struct {
		path     string
		wantSend string
	}{
		{"/users/new", "new user form"},        // Should match exact path, not param
		{"/users/admin", "admin user"},         // Should match exact path, not param
		{"/users/123", "user:123"},             // Should match param route
		{"/users/something", "user:something"}, // Should match param route
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler, ctx := router.Find(MethodGet, tt.path)
			if handler == nil {
				t.Fatalf("Failed to find route for path: %s", tt.path)
			}

			w := httptest.NewRecorder()
			ctx.Response = w
			ctx.Request = httptest.NewRequest(MethodGet, tt.path, nil)
			handler(ctx)

			if got := w.Body.String(); got != tt.wantSend {
				t.Errorf("Expected body %q, got %q", tt.wantSend, got)
			}
		})
	}
}

func TestRouterWildcard(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add wildcard routes
	router.Add(MethodGet, "/static/*filepath", func(c *Context) error {
		return c.Send("static:" + c.Param("*"))
	})
	router.Add(MethodGet, "/download/*filepath", func(c *Context) error {
		return c.Send("download:" + c.Param("*"))
	})

	tests := []struct {
		path     string
		wantSend string
	}{
		{"/static/css/style.css", "static:css/style.css"},
		{"/static/js/app.js", "static:js/app.js"},
		{"/static/images/logo.png", "static:images/logo.png"},
		{"/download/files/document.pdf", "download:files/document.pdf"},
		{"/download/music/song.mp3", "download:music/song.mp3"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			handler, ctx := router.Find(MethodGet, tt.path)
			if handler == nil {
				t.Fatalf("Failed to find route for path: %s", tt.path)
			}

			w := httptest.NewRecorder()
			ctx.Response = w
			ctx.Request = httptest.NewRequest(MethodGet, tt.path, nil)
			handler(ctx)

			if got := w.Body.String(); got != tt.wantSend {
				t.Errorf("Expected body %q, got %q", tt.wantSend, got)
			}
		})
	}
}

func TestRouterMethodsAndNoMatch(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add routes with different methods
	router.Add(MethodGet, "/users", func(c *Context) error {
		return c.Send("get users")
	})
	router.Add(MethodPost, "/users", func(c *Context) error {
		return c.Send("create user")
	})
	router.Add(MethodPut, "/users/:id", func(c *Context) error {
		return c.Send("update user:" + c.Param("id"))
	})
	router.Add(MethodDelete, "/users/:id", func(c *Context) error {
		return c.Send("delete user:" + c.Param("id"))
	})

	// Test finding routes with correct HTTP method
	methods := []struct {
		method   string
		path     string
		wantSend string
	}{
		{MethodGet, "/users", "get users"},
		{MethodPost, "/users", "create user"},
		{MethodPut, "/users/123", "update user:123"},
		{MethodDelete, "/users/123", "delete user:123"},
	}

	for _, tt := range methods {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			handler, ctx := router.Find(tt.method, tt.path)
			if handler == nil {
				t.Fatalf("Failed to find route for method %s and path: %s", tt.method, tt.path)
			}

			w := httptest.NewRecorder()
			ctx.Response = w
			ctx.Request = httptest.NewRequest(tt.method, tt.path, nil)
			handler(ctx)

			if got := w.Body.String(); got != tt.wantSend {
				t.Errorf("Expected body %q, got %q", tt.wantSend, got)
			}
		})
	}

	// Test routes that don't match
	noMatches := []struct {
		method string
		path   string
	}{
		{MethodGet, "/not/found"},
		{MethodPost, "/users/123"}, // No POST route for /users/:id
		{MethodPut, "/invalid"},
		{MethodGet, "/users/123/nonexistent"},
	}

	for _, tt := range noMatches {
		t.Run("NoMatch_"+tt.method+"_"+tt.path, func(t *testing.T) {
			handler, _ := router.Find(tt.method, tt.path)
			if handler != nil {
				t.Errorf("Expected no handler for %s %s, but got one", tt.method, tt.path)
			}
		})
	}
}

func TestRouteNodeFind(t *testing.T) {
	// Test the RouteNode's find method directly
	root := &RouteNode{}

	// Add some children to the root
	exactChild := &RouteNode{part: "exact", isParam: false, isWild: false, method: MethodGet}
	paramChild := &RouteNode{part: "param", isParam: true, isWild: false, method: MethodGet}
	wildcardChild := &RouteNode{part: "*", isParam: false, isWild: true, method: MethodGet}

	root.children = append(root.children, exactChild, paramChild, wildcardChild)

	// Add a handler to each child
	exactChild.handler = func(c *Context) error { return c.Send("exact") }
	paramChild.handler = func(c *Context) error { return c.Send("param:" + c.Param("param")) }
	wildcardChild.handler = func(c *Context) error { return c.Send("wildcard:" + c.Param("*")) }

	// Test cases
	tests := []struct {
		parts       []string
		method      string
		wantHandler bool
		wantParams  map[string]string
	}{
		{[]string{"exact"}, MethodGet, true, map[string]string{}},
		{[]string{"something"}, MethodGet, true, map[string]string{"param": "something"}},
		{[]string{"wildcard", "path"}, MethodGet, true, map[string]string{"*": "wildcard/path"}},
		{[]string{}, MethodGet, false, map[string]string{}}, // Empty path should not match
		// Test different method
		{[]string{"exact"}, MethodPost, false, map[string]string{}},
	}

	for i, tt := range tests {
		ctx := &Context{PathParams: params{}}
		result := root.find(tt.parts, ctx, tt.method)

		if (result != nil && result.handler != nil) != tt.wantHandler {
			t.Errorf("Test %d: find(%v, %s) handler existence = %v, want %v",
				i, tt.parts, tt.method, result != nil && result.handler != nil, tt.wantHandler)
		}

		// Check params
		for k, v := range tt.wantParams {
			got := ctx.Param(k)
			if got != v {
				t.Errorf("Test %d: Param(%q) = %q, want %q", i, k, got, v)
			}
		}
	}
}

func TestRouterCache(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add a route with a parameter
	router.Add(MethodGet, "/users/:id", func(c *Context) error {
		return c.Send("user:" + c.Param("id"))
	})

	// First call should not use cache
	handler1, ctx1 := router.Find(MethodGet, "/users/123")
	if handler1 == nil {
		t.Fatal("Failed to find route for path: /users/123")
	}

	// Get cache entry
	key := routeCacheKey{MethodGet, "/users/123"}
	entry, exists := router.cache.get(key)

	if !exists {
		t.Error("Route should be cached but wasn't")
	}

	// Verify that ctx1 has the correct params
	if ctx1.Param("id") != "123" {
		t.Errorf("Expected param id=123, got %q", ctx1.Param("id"))
	}

	// Verify that cached entry has the correct params
	if entry.context.Param("id") != "123" {
		t.Errorf("Cached context parameter incorrect. Expected param id=123, got %q",
			entry.context.Param("id"))
	}

	// Second call should use cache
	handler2, ctx2 := router.Find(MethodGet, "/users/123")
	if handler2 == nil {
		t.Fatal("Failed to find route for path: /users/123 on second call")
	}

	// Verify that ctx2 has the correct params from the cache
	if ctx2.Param("id") != "123" {
		t.Errorf("Expected cached param id=123, got %q", ctx2.Param("id"))
	}

	// Verify that the handlers match
	if fmt.Sprintf("%p", handler1) != fmt.Sprintf("%p", handler2) {
		t.Error("Handlers don't match - cache might not be working correctly")
	}
}

func TestGetPathParts(t *testing.T) {
	tests := []struct {
		path  string
		parts []string
	}{
		{"", []string{}},
		{"/", []string{}},
		{"/users", []string{"users"}},
		{"/users/", []string{"users"}},
		{"/users/123", []string{"users", "123"}},
		{"/api/v1/users/123/posts", []string{"api", "v1", "users", "123", "posts"}},
		{"users/123", []string{"users", "123"}}, // No leading slash
	}

	for i, tt := range tests {
		parts := getPathParts(tt.path)

		if len(parts) != len(tt.parts) {
			t.Errorf("Test %d: getPathParts(%q) returned %d parts, want %d parts",
				i, tt.path, len(parts), len(tt.parts))
			continue
		}

		for j, part := range parts {
			if part != tt.parts[j] {
				t.Errorf("Test %d: getPathParts(%q)[%d] = %q, want %q",
					i, tt.path, j, part, tt.parts[j])
			}
		}
	}
}

func TestNormalizePath(t *testing.T) {
	router := &Router{}

	tests := []struct {
		path string
		want string
	}{
		{"", "/"},
		{"/", "/"},
		{"users", "/users"},
		{"/users", "/users"},
		{"/users/", "/users/"},
		{"//users", "//users"}, // Edge case - doesn't normalize double slashes
	}

	for i, tt := range tests {
		got := router.normalizePath(tt.path)
		if got != tt.want {
			t.Errorf("Test %d: normalizePath(%q) = %q, want %q", i, tt.path, got, tt.want)
		}
	}
}

func BenchmarkRouterFind(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(1000),
	}

	// Add routes for benchmarking
	router.Add(MethodGet, "/", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id/posts/:postId", func(c *Context) error { return nil })
	router.Add(MethodGet, "/files/*filepath", func(c *Context) error { return nil })
	router.Add(MethodGet, "/static/*filepath", func(c *Context) error { return nil })
	router.Add(MethodGet, "/api/v1/users", func(c *Context) error { return nil })
	router.Add(MethodGet, "/api/v1/users/:id", func(c *Context) error { return nil })

	// Define test paths
	paths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/123/posts",
		"/users/123/posts/456",
		"/files/images/logo.png",
		"/static/css/style.css",
		"/api/v1/users",
		"/api/v1/users/123",
	}

	// Benchmark finding each path
	for _, p := range paths {
		b.Run(p, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				router.Find(MethodGet, p)
			}
		})
	}
}

func BenchmarkRouterFindWithoutCache(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
	}

	// Add routes for benchmarking
	router.Add(MethodGet, "/", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) error { return nil })
	router.Add(MethodGet, "/users/:id/posts/:postId", func(c *Context) error { return nil })
	router.Add(MethodGet, "/files/*filepath", func(c *Context) error { return nil })
	router.Add(MethodGet, "/static/*filepath", func(c *Context) error { return nil })
	router.Add(MethodGet, "/api/v1/users", func(c *Context) error { return nil })
	router.Add(MethodGet, "/api/v1/users/:id", func(c *Context) error { return nil })

	// Define test paths
	paths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/123/posts",
		"/users/123/posts/456",
		"/files/images/logo.png",
		"/static/css/style.css",
		"/api/v1/users",
		"/api/v1/users/123",
	}

	// Benchmark finding each path
	for _, p := range paths {
		b.Run(p, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				router.Find(MethodGet, p)
			}
		})
	}
}

func TestRouterMethodHandling(t *testing.T) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Add routes with different methods for the same path
	router.Add(MethodPut, "/users/:id", func(c *Context) error {
		return c.Send("put user:" + c.Param("id"))
	})
	router.Add(MethodDelete, "/users/:id", func(c *Context) error {
		return c.Send("delete user:" + c.Param("id"))
	})
	router.Add(MethodPatch, "/users/:id", func(c *Context) error {
		return c.Send("patch user:" + c.Param("id"))
	})

	// Test each method separately
	methods := []struct {
		method   string
		path     string
		wantSend string
	}{
		{MethodPut, "/users/123", "put user:123"},
		{MethodDelete, "/users/123", "delete user:123"},
		{MethodPatch, "/users/123", "patch user:123"},
	}

	for _, tt := range methods {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			handler, ctx := router.Find(tt.method, tt.path)
			if handler == nil {
				t.Fatalf("Failed to find route for method %s and path: %s", tt.method, tt.path)
			}

			w := httptest.NewRecorder()
			ctx.Response = w
			ctx.Request = httptest.NewRequest(tt.method, tt.path, nil)
			handler(ctx)

			if got := w.Body.String(); got != tt.wantSend {
				t.Errorf("Expected body %q, got %q", tt.wantSend, got)
			}
		})
	}
}
