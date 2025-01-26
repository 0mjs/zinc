package zinc

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/labstack/echo/v4"
)

func init() {
	// Disable Gin debug output during tests
	gin.SetMode(gin.ReleaseMode)
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

// Benchmark Hello World
func BenchmarkHelloWorld(b *testing.B) {
	// Zinc
	b.Run("Zinc", func(b *testing.B) {
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
