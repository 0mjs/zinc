package zinc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

type benchmarkResult struct {
	name       string
	framework  string
	opsPerSec  float64
	nsPerOp    float64
	bytesPerOp int
}

func TestMain(m *testing.M) {
	// Disable Gin debug output during tests
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

// PrintResults formats and prints benchmark results in a nice table
func PrintResults(results []benchmarkResult) {
	// Group results by test name
	resultsByTest := make(map[string][]benchmarkResult)
	var testNames []string

	for _, result := range results {
		if _, exists := resultsByTest[result.name]; !exists {
			testNames = append(testNames, result.name)
		}
		resultsByTest[result.name] = append(resultsByTest[result.name], result)
	}

	sort.Strings(testNames)

	// Print results
	for _, testName := range testNames {
		testResults := resultsByTest[testName]

		// Sort by ops/sec (higher is better)
		sort.Slice(testResults, func(i, j int) bool {
			return testResults[i].opsPerSec > testResults[j].opsPerSec
		})

		fmt.Printf("\n%s====== %s ======%s\n", colorYellow, testName, colorReset)
		fmt.Printf("%-10s %-15s %-15s %-15s\n", "Framework", "Ops/sec", "ns/op", "B/op")

		// Calculate the highest ops/sec for highlighting the winner
		highestOps := testResults[0].opsPerSec

		for i, result := range testResults {
			color := colorReset
			if i == 0 {
				color = colorGreen // Highlight the winner
			} else if result.opsPerSec >= highestOps*0.95 {
				color = colorCyan // Highlight very close (within 5%)
			}

			fmt.Printf("%s%-10s %-15.2f %-15.2f %-15d%s\n",
				color,
				result.framework,
				result.opsPerSec,
				result.nsPerOp,
				result.bytesPerOp,
				colorReset)
		}
	}
}

// Collect results from a benchmark
func collectResults(name string, b *testing.B, results *[]benchmarkResult) {
	b.Helper()

	b.Run("Collect", func(b *testing.B) {
		b.ReportAllocs()
		b.SkipNow() // Skip actual execution

		// Extract framework name (assuming format "Framework Description")
		parts := strings.SplitN(b.Name(), " ", 2)
		framework := parts[0]

		// Record result
		*results = append(*results, benchmarkResult{
			name:       name,
			framework:  framework,
			opsPerSec:  float64(b.N) * float64(time.Second) / float64(b.Elapsed().Nanoseconds()),
			nsPerOp:    float64(b.Elapsed().Nanoseconds()) / float64(b.N),
			bytesPerOp: int(testing.AllocsPerRun(1, func() {})), // This will be updated with real values during test run
		})
	})
}

// Zinc handlers
func zincHelloHandler(c *Context) {
	c.Send("Hello World!")
}

func zincParamHandler(c *Context) {
	c.Send(fmt.Sprintf("Hello, %s!", c.Param("name")))
}

// Chi handlers
func chiHelloHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World!"))
}

func chiParamHandler(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	w.Write([]byte(fmt.Sprintf("Hello, %s!", name)))
}

// Echo handlers
func echoHelloHandler(c echo.Context) error {
	return c.String(http.StatusOK, "Hello World!")
}

func echoParamHandler(c echo.Context) error {
	name := c.Param("name")
	return c.String(http.StatusOK, fmt.Sprintf("Hello, %s!", name))
}

// Gin handlers
func ginHelloHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

func ginParamHandler(c *gin.Context) {
	name := c.Param("name")
	c.String(http.StatusOK, fmt.Sprintf("Hello, %s!", name))
}

// Test struct for JSON benchmarks
type testResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    struct {
		Items []string `json:"items"`
		Count int      `json:"count"`
	} `json:"data"`
}

// Prepare a test response to avoid allocations in the benchmarks
var testJSONData = func() testResponse {
	r := testResponse{
		Message: "Success",
		Status:  200,
	}
	r.Data.Items = []string{"item1", "item2", "item3", "item4", "item5"}
	r.Data.Count = len(r.Data.Items)
	return r
}()

// Zinc JSON handler
func zincJSONHandler(c *Context) {
	c.JSON(testJSONData)
}

// Chi JSON handler
func chiJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testJSONData)
}

// Echo JSON handler
func echoJSONHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, testJSONData)
}

// Gin JSON handler
func ginJSONHandler(c *gin.Context) {
	c.JSON(http.StatusOK, testJSONData)
}

// Zinc query params handler
func zincQueryHandler(c *Context) {
	name := c.Query("name")
	age := c.Query("age")
	city := c.Query("city")
	c.Send(fmt.Sprintf("Hello, %s! You are %s years old and from %s.", name, age, city))
}

// Chi query params handler
func chiQueryHandler(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	age := r.URL.Query().Get("age")
	city := r.URL.Query().Get("city")
	w.Write([]byte(fmt.Sprintf("Hello, %s! You are %s years old and from %s.", name, age, city)))
}

// Echo query params handler
func echoQueryHandler(c echo.Context) error {
	name := c.QueryParam("name")
	age := c.QueryParam("age")
	city := c.QueryParam("city")
	return c.String(http.StatusOK, fmt.Sprintf("Hello, %s! You are %s years old and from %s.", name, age, city))
}

// Gin query params handler
func ginQueryHandler(c *gin.Context) {
	name := c.Query("name")
	age := c.Query("age")
	city := c.Query("city")
	c.String(http.StatusOK, fmt.Sprintf("Hello, %s! You are %s years old and from %s.", name, age, city))
}

// Middleware handlers for Zinc
func zincMiddleware1(c *Context) {
	c.Set("middleware1", true)
	c.Next()
}

func zincMiddleware2(c *Context) {
	c.Set("middleware2", true)
	c.Next()
}

func zincMiddleware3(c *Context) {
	c.Set("middleware3", true)
	c.Next()
}

func zincMiddleware4(c *Context) {
	c.Set("middleware4", true)
	c.Next()
}

func zincMiddleware5(c *Context) {
	c.Set("middleware5", true)
	c.Next()
}

func zincMiddlewareHandler(c *Context) {
	// Access all middleware values to ensure they're used
	v1 := c.Get("middleware1")
	v2 := c.Get("middleware2")
	v3 := c.Get("middleware3")
	v4 := c.Get("middleware4")
	v5 := c.Get("middleware5")

	c.Send(fmt.Sprintf("Middleware chain complete: %v %v %v %v %v", v1, v2, v3, v4, v5))
}

// Middleware handlers for Chi
func chiMiddleware1(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "middleware1", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "middleware2", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddleware3(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "middleware3", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddleware4(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "middleware4", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddleware5(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), "middleware5", true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddlewareHandler(w http.ResponseWriter, r *http.Request) {
	v1 := r.Context().Value("middleware1")
	v2 := r.Context().Value("middleware2")
	v3 := r.Context().Value("middleware3")
	v4 := r.Context().Value("middleware4")
	v5 := r.Context().Value("middleware5")

	w.Write([]byte(fmt.Sprintf("Middleware chain complete: %v %v %v %v %v", v1, v2, v3, v4, v5)))
}

// Middleware handlers for Echo
func echoMiddleware1(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware1", true)
		return next(c)
	}
}

func echoMiddleware2(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware2", true)
		return next(c)
	}
}

func echoMiddleware3(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware3", true)
		return next(c)
	}
}

func echoMiddleware4(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware4", true)
		return next(c)
	}
}

func echoMiddleware5(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware5", true)
		return next(c)
	}
}

// Proper Echo middleware
func echoMW1(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware1", true)
		return next(c)
	}
}

func echoMW2(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware2", true)
		return next(c)
	}
}

func echoMW3(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware3", true)
		return next(c)
	}
}

func echoMW4(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware4", true)
		return next(c)
	}
}

func echoMW5(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("middleware5", true)
		return next(c)
	}
}

func echoMiddlewareHandler(c echo.Context) error {
	v1 := c.Get("middleware1")
	v2 := c.Get("middleware2")
	v3 := c.Get("middleware3")
	v4 := c.Get("middleware4")
	v5 := c.Get("middleware5")

	return c.String(http.StatusOK, fmt.Sprintf("Middleware chain complete: %v %v %v %v %v", v1, v2, v3, v4, v5))
}

// Middleware handlers for Gin
func ginMiddleware1(c *gin.Context) {
	c.Set("middleware1", true)
	c.Next()
}

func ginMiddleware2(c *gin.Context) {
	c.Set("middleware2", true)
	c.Next()
}

func ginMiddleware3(c *gin.Context) {
	c.Set("middleware3", true)
	c.Next()
}

func ginMiddleware4(c *gin.Context) {
	c.Set("middleware4", true)
	c.Next()
}

func ginMiddleware5(c *gin.Context) {
	c.Set("middleware5", true)
	c.Next()
}

func ginMiddlewareHandler(c *gin.Context) {
	v1, _ := c.Get("middleware1")
	v2, _ := c.Get("middleware2")
	v3, _ := c.Get("middleware3")
	v4, _ := c.Get("middleware4")
	v5, _ := c.Get("middleware5")

	c.String(http.StatusOK, fmt.Sprintf("Middleware chain complete: %v %v %v %v %v", v1, v2, v3, v4, v5))
}

// Benchmark Hello World
func BenchmarkHelloWorld(b *testing.B) {
	// Zinc
	b.Run("Zinc 🪙", func(b *testing.B) {
		app := New()
		app.Get("/", zincHelloHandler)
		req := httptest.NewRequest("GET", "/", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Get("/", chiHelloHandler)
		req := httptest.NewRequest("GET", "/", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.GET("/", echoHelloHandler)
		req := httptest.NewRequest("GET", "/", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/", ginHelloHandler)
		req := httptest.NewRequest("GET", "/", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Benchmark Router with Parameter
func BenchmarkRouterParam(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		app.Get("/hello/:name", zincParamHandler)
		req := httptest.NewRequest("GET", "/hello/world", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Get("/hello/{name}", chiParamHandler)
		req := httptest.NewRequest("GET", "/hello/world", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.GET("/hello/:name", echoParamHandler)
		req := httptest.NewRequest("GET", "/hello/world", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/hello/:name", ginParamHandler)
		req := httptest.NewRequest("GET", "/hello/world", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Benchmark JSON Response
func BenchmarkJSONResponse(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		app.Get("/json", zincJSONHandler)
		req := httptest.NewRequest("GET", "/json", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Get("/json", chiJSONHandler)
		req := httptest.NewRequest("GET", "/json", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.GET("/json", echoJSONHandler)
		req := httptest.NewRequest("GET", "/json", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/json", ginJSONHandler)
		req := httptest.NewRequest("GET", "/json", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Benchmark Query Parameters
func BenchmarkQueryParams(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		app.Get("/query", zincQueryHandler)
		req := httptest.NewRequest("GET", "/query?name=john&age=25&city=newyork", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Get("/query", chiQueryHandler)
		req := httptest.NewRequest("GET", "/query?name=john&age=25&city=newyork", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.GET("/query", echoQueryHandler)
		req := httptest.NewRequest("GET", "/query?name=john&age=25&city=newyork", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/query", ginQueryHandler)
		req := httptest.NewRequest("GET", "/query?name=john&age=25&city=newyork", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Benchmark Middleware Chain
func BenchmarkMiddlewareChain(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		app.Use(zincMiddleware1, zincMiddleware2, zincMiddleware3, zincMiddleware4, zincMiddleware5)
		app.Get("/middleware", zincMiddlewareHandler)
		req := httptest.NewRequest("GET", "/middleware", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Use(chiMiddleware1)
		r.Use(chiMiddleware2)
		r.Use(chiMiddleware3)
		r.Use(chiMiddleware4)
		r.Use(chiMiddleware5)
		r.Get("/middleware", chiMiddlewareHandler)
		req := httptest.NewRequest("GET", "/middleware", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()

		// Convert our middleware functions to Echo middleware
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return echoMW1(echoMW2(echoMW3(echoMW4(echoMW5(next)))))
		})

		e.GET("/middleware", echoMiddlewareHandler)
		req := httptest.NewRequest("GET", "/middleware", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.Use(ginMiddleware1, ginMiddleware2, ginMiddleware3, ginMiddleware4, ginMiddleware5)
		r.GET("/middleware", ginMiddlewareHandler)
		req := httptest.NewRequest("GET", "/middleware", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Zinc nested routes handler
func zincNestedHandler(c *Context) {
	category := c.Param("category")
	id := c.Param("id")
	subresource := c.Param("subresource")
	c.Send(fmt.Sprintf("Resource: category=%s, id=%s, subresource=%s", category, id, subresource))
}

// Chi nested routes handler
func chiNestedHandler(w http.ResponseWriter, r *http.Request) {
	category := chi.URLParam(r, "category")
	id := chi.URLParam(r, "id")
	subresource := chi.URLParam(r, "subresource")
	w.Write([]byte(fmt.Sprintf("Resource: category=%s, id=%s, subresource=%s", category, id, subresource)))
}

// Echo nested routes handler
func echoNestedHandler(c echo.Context) error {
	category := c.Param("category")
	id := c.Param("id")
	subresource := c.Param("subresource")
	return c.String(http.StatusOK, fmt.Sprintf("Resource: category=%s, id=%s, subresource=%s", category, id, subresource))
}

// Gin nested routes handler
func ginNestedHandler(c *gin.Context) {
	category := c.Param("category")
	id := c.Param("id")
	subresource := c.Param("subresource")
	c.String(http.StatusOK, fmt.Sprintf("Resource: category=%s, id=%s, subresource=%s", category, id, subresource))
}

// Benchmark Nested Routes
func BenchmarkNestedRoutes(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		app.Get("/api/:category/:id/:subresource", zincNestedHandler)
		req := httptest.NewRequest("GET", "/api/products/123/details", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Get("/api/{category}/{id}/{subresource}", chiNestedHandler)
		req := httptest.NewRequest("GET", "/api/products/123/details", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		e.GET("/api/:category/:id/:subresource", echoNestedHandler)
		req := httptest.NewRequest("GET", "/api/products/123/details", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		r.GET("/api/:category/:id/:subresource", ginNestedHandler)
		req := httptest.NewRequest("GET", "/api/products/123/details", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}

// Zinc group handler
func zincGroupHandler(c *Context) {
	resource := c.Param("resource")
	action := c.Param("action")
	c.Send(fmt.Sprintf("API Resource: %s, Action: %s", resource, action))
}

// Chi group handler
func chiGroupHandler(w http.ResponseWriter, r *http.Request) {
	resource := chi.URLParam(r, "resource")
	action := chi.URLParam(r, "action")
	w.Write([]byte(fmt.Sprintf("API Resource: %s, Action: %s", resource, action)))
}

// Echo group handler
func echoGroupHandler(c echo.Context) error {
	resource := c.Param("resource")
	action := c.Param("action")
	return c.String(http.StatusOK, fmt.Sprintf("API Resource: %s, Action: %s", resource, action))
}

// Gin group handler
func ginGroupHandler(c *gin.Context) {
	resource := c.Param("resource")
	action := c.Param("action")
	c.String(http.StatusOK, fmt.Sprintf("API Resource: %s, Action: %s", resource, action))
}

// Benchmark Route Groups
func BenchmarkRouteGroups(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
		app := New()
		apiGroup := app.Group("api")
		v1Group := apiGroup.Group("v1")
		v1Group.Get("/:resource/:action", zincGroupHandler)

		req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			app.ServeHTTP(w, req)
		}
	})

	// Chi
	b.Run("Chi", func(b *testing.B) {
		r := chi.NewRouter()
		r.Route("/api", func(r chi.Router) {
			r.Route("/v1", func(r chi.Router) {
				r.Get("/{resource}/{action}", chiGroupHandler)
			})
		})
		req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})

	// Echo
	b.Run("Echo", func(b *testing.B) {
		e := echo.New()
		apiGroup := e.Group("/api")
		v1Group := apiGroup.Group("/v1")
		v1Group.GET("/:resource/:action", echoGroupHandler)

		req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			e.ServeHTTP(w, req)
		}
	})

	// Gin
	b.Run("Gin", func(b *testing.B) {
		gin.SetMode(gin.ReleaseMode)
		r := gin.New()
		apiGroup := r.Group("/api")
		v1Group := apiGroup.Group("/v1")
		v1Group.GET("/:resource/:action", ginGroupHandler)

		req := httptest.NewRequest("GET", "/api/v1/users/list", nil)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
		}
	})
}
