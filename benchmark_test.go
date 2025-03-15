package zinc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

// BenchmarkRouterStatic benchmarks static route matching
func BenchmarkRouterStatic(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Setup some static routes
	router.Add(MethodGet, "/", func(c *Context) {})
	router.Add(MethodGet, "/users", func(c *Context) {})
	router.Add(MethodGet, "/users/settings", func(c *Context) {})
	router.Add(MethodGet, "/users/settings/profile", func(c *Context) {})
	router.Add(MethodGet, "/products", func(c *Context) {})
	router.Add(MethodGet, "/products/categories", func(c *Context) {})

	// Run the benchmark
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.Find(MethodGet, "/users")
	}
}

// BenchmarkRouterParameter benchmarks parameter route matching
func BenchmarkRouterParameter(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Setup parameter routes
	router.Add(MethodGet, "/users/:id", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts/:postID", func(c *Context) {})
	router.Add(MethodGet, "/products/:category/:id", func(c *Context) {})

	// Run the benchmark
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.Find(MethodGet, "/users/123")
	}
}

// BenchmarkRouterWildcard benchmarks wildcard route matching
func BenchmarkRouterWildcard(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Setup wildcard routes
	router.Add(MethodGet, "/files/*path", func(c *Context) {})
	router.Add(MethodGet, "/static/*filepath", func(c *Context) {})

	// Run the benchmark
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		router.Find(MethodGet, "/static/css/site.css")
	}
}

// BenchmarkRouterMixed benchmarks a mix of static, parameter, and wildcard routes
func BenchmarkRouterMixed(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Setup a mix of routes
	router.Add(MethodGet, "/", func(c *Context) {})
	router.Add(MethodGet, "/users", func(c *Context) {})
	router.Add(MethodGet, "/users/:id", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts/:postID", func(c *Context) {})
	router.Add(MethodGet, "/files/*path", func(c *Context) {})
	router.Add(MethodGet, "/static/*filepath", func(c *Context) {})

	paths := []string{
		"/",
		"/users",
		"/users/123",
		"/users/123/posts",
		"/users/123/posts/456",
		"/files/document.pdf",
		"/static/css/style.css",
	}

	// Run the benchmark for each path
	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				router.Find(MethodGet, path)
			}
		})
	}
}

// BenchmarkRouterNoCache benchmarks route matching without using the cache
func BenchmarkRouterNoCache(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
	}

	// Setup a mix of routes
	router.Add(MethodGet, "/", func(c *Context) {})
	router.Add(MethodGet, "/users", func(c *Context) {})
	router.Add(MethodGet, "/users/:id", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) {})

	// Run the benchmark for different paths
	paths := []string{
		"/users",     // Static path
		"/users/123", // Parameter path
	}

	// Run the benchmark for each path
	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				router.Find(MethodGet, path)
			}
		})
	}
}

// BenchmarkRouterWithCache benchmarks route matching with cache enabled
func BenchmarkRouterWithCache(b *testing.B) {
	router := &Router{
		routes: make(map[string]map[string]*Route),
		cache:  NewRouteCache(100),
	}

	// Setup the same routes as in the no-cache benchmark
	router.Add(MethodGet, "/", func(c *Context) {})
	router.Add(MethodGet, "/users", func(c *Context) {})
	router.Add(MethodGet, "/users/:id", func(c *Context) {})
	router.Add(MethodGet, "/users/:id/posts", func(c *Context) {})

	// Same paths as in the no-cache benchmark
	paths := []string{
		"/users",     // Static path
		"/users/123", // Parameter path
	}

	// Run the benchmark for each path
	for _, path := range paths {
		b.Run(path, func(b *testing.B) {
			// Warm up the cache with an initial request
			router.Find(MethodGet, path)

			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				router.Find(MethodGet, path)
			}
		})
	}
}

// BenchmarkContextParam benchmarks the Context.Param method
func BenchmarkContextParamPerformance(b *testing.B) {
	// Create a context with parameters
	ctx := &Context{
		PathParams: params{},
	}

	// Set up path parameters
	ctx.PathParams[0] = param{key: "id", value: "123"}
	ctx.PathParams[1] = param{key: "name", value: "test"}
	ctx.PathParams[2] = param{key: "a", value: "1"}
	ctx.PathParams[3] = param{key: "longparametername", value: "value"}

	// Different parameter retrieval scenarios
	cases := []struct {
		name  string
		param string
	}{
		{"ExistingParam", "id"},
		{"ShortParam", "a"},
		{"LongParam", "longparametername"},
		{"NonExistentParam", "notfound"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				ctx.Param(tc.param)
			}
		})
	}
}

// BenchmarkFullRequest benchmarks a complete request through the App
func BenchmarkFullRequest(b *testing.B) {
	app := New()

	// Add some routes
	app.Get("/", func(c *Context) {
		c.Send("Hello World!")
	})

	app.Get("/json", func(c *Context) {
		c.JSON(Map{"message": "Hello World!"})
	})

	app.Get("/users/:id", func(c *Context) {
		id := c.Param("id")
		c.JSON(Map{"id": id, "name": "User " + id})
	})

	// Benchmark different request types
	cases := []struct {
		name string
		path string
	}{
		{"StaticText", "/"},
		{"JSONResponse", "/json"},
		{"ParameterRoute", "/users/123"},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkSendMethods benchmarks different response sending methods
func BenchmarkSendMethods(b *testing.B) {
	cases := []struct {
		name string
		fn   func(w http.ResponseWriter, c *Context)
	}{
		{
			name: "Text",
			fn: func(w http.ResponseWriter, c *Context) {
				c.Send("Hello World!")
			},
		},
		{
			name: "JSON",
			fn: func(w http.ResponseWriter, c *Context) {
				c.JSON(Map{"message": "Hello World!"})
			},
		},
		{
			name: "HTML",
			fn: func(w http.ResponseWriter, c *Context) {
				c.HTML("<h1>Hello World!</h1>")
			},
		},
		{
			name: "Bytes",
			fn: func(w http.ResponseWriter, c *Context) {
				c.Send([]byte("Hello World!"))
			},
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				c := &Context{
					Response: w,
					status:   http.StatusOK,
				}
				tc.fn(w, c)
			}
		})
	}
}

// BenchmarkMiddleware benchmarks middleware execution
func BenchmarkMiddleware(b *testing.B) {
	// Different middleware chain lengths
	cases := []struct {
		name  string
		count int
	}{
		{"NoMiddleware", 0},
		{"OneMiddleware", 1},
		{"ThreeMiddleware", 3},
		{"FiveMiddleware", 5},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			app := New()

			// Add middleware based on count
			for i := 0; i < tc.count; i++ {
				app.Use(func(c *Context) {
					c.Next()
				})
			}

			// Add a simple handler
			app.Get("/", func(c *Context) {
				c.Send("Hello World!")
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)
			}
		})
	}
}

// BenchmarkBodyParsing benchmarks request body parsing
func BenchmarkBodyParsing(b *testing.B) {
	// Different JSON payload sizes
	cases := []struct {
		name    string
		payload string
	}{
		{
			name:    "SmallJSON",
			payload: `{"name":"John","email":"john@example.com"}`,
		},
		{
			name:    "MediumJSON",
			payload: `{"id":123,"name":"John Doe","email":"john@example.com","age":30,"roles":["user","admin"],"address":{"street":"123 Main St","city":"New York","zip":"10001"}}`,
		},
	}

	for _, tc := range cases {
		b.Run(tc.name, func(b *testing.B) {
			app := New()

			type User struct {
				ID      int      `json:"id,omitempty"`
				Name    string   `json:"name"`
				Email   string   `json:"email"`
				Age     int      `json:"age,omitempty"`
				Roles   []string `json:"roles,omitempty"`
				Address struct {
					Street string `json:"street,omitempty"`
					City   string `json:"city,omitempty"`
					Zip    string `json:"zip,omitempty"`
				} `json:"address,omitempty"`
			}

			app.Post("/", func(c *Context) {
				var user User
				if err := c.Body(&user); err != nil {
					c.Status(400).Send(err.Error())
					return
				}
				c.JSON(user)
			})

			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.payload))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				app.ServeHTTP(w, req)
			}
		})
	}
}
