package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	. "github.com/0mjs/zinc"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/labstack/echo/v5"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	os.Exit(m.Run())
}

const (
	benchmarkHelloResponse = "Hello World!"
	benchmarkOKResponse    = "OK"
	staticColdRouteCount   = 64
	largeStaticRouteCount  = 256
	largeParamRouteCount   = 128
	coldPathRequestCount   = 256
	throughputDuration     = 1500 * time.Millisecond
)

var throughputConcurrencyLevels = []int{1, 8, 32, 128}

var (
	benchmarkSinkString  string
	benchmarkSinkBool    bool
	benchmarkSinkHandler http.Handler
)

type benchmarkResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    struct {
		Items []string `json:"items"`
		Count int      `json:"count"`
	} `json:"data"`
}

var benchmarkJSONData = func() benchmarkResponse {
	resp := benchmarkResponse{
		Message: "Success",
		Status:  http.StatusOK,
	}
	resp.Data.Items = []string{"item1", "item2", "item3", "item4", "item5"}
	resp.Data.Count = len(resp.Data.Items)
	return resp
}()

type benchmarkAPIBindInput struct {
	TeamID  int      `path:"teamID" uri:"teamID" param:"teamID"`
	UserID  int      `path:"userID" uri:"userID" param:"userID"`
	Verbose bool     `query:"verbose" form:"verbose"`
	Limit   int      `query:"limit" form:"limit"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
}

type benchmarkAPIResponse struct {
	OK        bool   `json:"ok"`
	TeamID    int    `json:"team_id"`
	UserID    int    `json:"user_id"`
	Limit     int    `json:"limit"`
	RoleCount int    `json:"role_count"`
	Name      string `json:"name"`
}

var benchmarkAPIBindBody = []byte(`{"name":"alice","roles":["admin","editor"]}`)

type discardResponseWriter struct {
	header http.Header
	status int
	bytes  int
}

func newDiscardResponseWriter() *discardResponseWriter {
	return &discardResponseWriter{
		header: make(http.Header, 8),
	}
}

func (w *discardResponseWriter) Header() http.Header {
	return w.header
}

func (w *discardResponseWriter) WriteHeader(code int) {
	w.status = code
}

func (w *discardResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.bytes += len(p)
	return len(p), nil
}

func (w *discardResponseWriter) WriteString(s string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.bytes += len(s)
	return len(s), nil
}

func (w *discardResponseWriter) reset() {
	clear(w.header)
	w.status = 0
	w.bytes = 0
}

type benchmarkCase struct {
	name  string
	build func() http.Handler
}

type preparedBenchmarkRequest struct {
	request *http.Request
	body    []byte
}

type middlewareContextKey int

const (
	middlewareKey1 middlewareContextKey = iota
	middlewareKey2
	middlewareKey3
	middlewareKey4
	middlewareKey5
)

func mustNoErr(err error) {
	if err != nil {
		panic(err)
	}
}

func newGinBenchmarkRouter() *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	return r
}

func newPreparedBenchmarkRequest(method, target string, body []byte, headers http.Header) preparedBenchmarkRequest {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req := httptest.NewRequest(method, target, reader)
	for key, values := range headers {
		req.Header[key] = append([]string(nil), values...)
	}

	return preparedBenchmarkRequest{
		request: req,
		body:    append([]byte(nil), body...),
	}
}

func (r *preparedBenchmarkRequest) reset() *http.Request {
	if len(r.body) == 0 {
		return r.request
	}

	r.request.Body = io.NopCloser(bytes.NewReader(r.body))
	r.request.ContentLength = int64(len(r.body))
	r.request.Form = nil
	r.request.PostForm = nil
	r.request.MultipartForm = nil
	return r.request
}

func runServeHTTPBenchmarks(b *testing.B, method, target string, cases []benchmarkCase) {
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			handler := bc.build()
			req := httptest.NewRequest(method, target, nil)
			rw := newDiscardResponseWriter()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rw.reset()
				handler.ServeHTTP(rw, req)
			}
			benchmarkSinkInt = rw.status + rw.bytes
		})
	}
}

func runServeHTTPRequestSetBenchmarks(b *testing.B, cases []benchmarkCase, requests []*http.Request) {
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			handler := bc.build()
			rw := newDiscardResponseWriter()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rw.reset()
				handler.ServeHTTP(rw, requests[i%len(requests)])
			}
			benchmarkSinkInt = rw.status + rw.bytes
		})
	}
}

func runPreparedRequestBenchmarks(b *testing.B, cases []benchmarkCase, request preparedBenchmarkRequest) {
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			handler := bc.build()
			rw := newDiscardResponseWriter()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rw.reset()
				handler.ServeHTTP(rw, request.reset())
			}
			benchmarkSinkInt = rw.status + rw.bytes
		})
	}
}

func runPreparedRequestSetBenchmarks(b *testing.B, cases []benchmarkCase, requests []preparedBenchmarkRequest) {
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			handler := bc.build()
			rw := newDiscardResponseWriter()

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rw.reset()
				req := &requests[i%len(requests)]
				handler.ServeHTTP(rw, req.reset())
			}
			benchmarkSinkInt = rw.status + rw.bytes
		})
	}
}

func runRegistrationBenchmarks(b *testing.B, cases []benchmarkCase) {
	for _, bc := range cases {
		b.Run(bc.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchmarkSinkHandler = bc.build()
			}
		})
	}
}

func buildRequests(method string, targets []string) []*http.Request {
	requests := make([]*http.Request, len(targets))
	for i, target := range targets {
		requests[i] = httptest.NewRequest(method, target, nil)
	}
	return requests
}

func stdMiddleware(next http.Handler, key middlewareContextKey) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), key, true)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func chiMiddleware(key middlewareContextKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func echoMiddleware(key string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set(key, true)
			return next(c)
		}
	}
}

func zincMiddleware(key string) Middleware {
	return func(c *Context) error {
		c.Set(key, true)
		return c.Next()
	}
}

func ginMiddleware(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(key, true)
		c.Next()
	}
}

func largeStaticPath(i int) string {
	return "/static/" + strconv.Itoa(i)
}

func largeParamPath(i int) string {
	return "/teams/" + strconv.Itoa(i) + "/users/123"
}

func largeParamPatternColon(i int) string {
	return "/teams/" + strconv.Itoa(i) + "/users/:id"
}

func largeParamPatternBrace(i int) string {
	return "/teams/" + strconv.Itoa(i) + "/users/{id}"
}

func largeParamPatternServeMux(i int) string {
	return "GET " + largeParamPatternBrace(i)
}

func coldParamTargets(count int) []string {
	targets := make([]string, count)
	for i := 0; i < count; i++ {
		targets[i] = "/hello/name" + strconv.Itoa(i)
	}
	return targets
}

func staticColdPath(i int) string {
	return "/static-cold/" + strconv.Itoa(i)
}

func staticColdTargets(count int) []string {
	targets := make([]string, count)
	for i := 0; i < count; i++ {
		targets[i] = staticColdPath(i % staticColdRouteCount)
	}
	return targets
}

func mixedLargeStaticTargets(count int) []string {
	targets := make([]string, count)
	for i := 0; i < count; i++ {
		targets[i] = largeStaticPath(i % largeStaticRouteCount)
	}
	return targets
}

func mixedLargeParamTargets(count int) []string {
	targets := make([]string, count)
	for i := 0; i < count; i++ {
		targets[i] = "/teams/" + strconv.Itoa(i%largeParamRouteCount) + "/users/" + strconv.Itoa(1000+i)
	}
	return targets
}

func benchmarkAPIQueryTarget() string {
	return "/teams/42/users/7?verbose=true&limit=25"
}

func benchmarkAPIBindTarget() string {
	return "/teams/42/users/7?verbose=true&limit=25"
}

func benchmarkAPIBodyHeaders() http.Header {
	header := make(http.Header, 1)
	header.Set(HeaderContentType, "application/json")
	return header
}

func consumeBenchmarkAPIInput(input benchmarkAPIBindInput) {
	benchmarkSinkString = input.Name
	benchmarkSinkBool = input.Verbose
	benchmarkSinkInt = input.TeamID + input.UserID + input.Limit + len(input.Roles)
}

func benchmarkAPIResponseFrom(input benchmarkAPIBindInput) benchmarkAPIResponse {
	return benchmarkAPIResponse{
		OK:        true,
		TeamID:    input.TeamID,
		UserID:    input.UserID,
		Limit:     input.Limit,
		RoleCount: len(input.Roles),
		Name:      input.Name,
	}
}

func parseBenchmarkInt(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return n
}

func fillBenchmarkAPIQuery(input *benchmarkAPIBindInput, req *http.Request) {
	values := req.URL.Query()
	input.Verbose = values.Get("verbose") == "true"
	input.Limit = parseBenchmarkInt(values.Get("limit"))
}

func buildZincHelloHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/", func(c *Context) error {
		return c.String(benchmarkHelloResponse)
	}))
	return app
}

func buildServeMuxHelloHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return mux
}

func buildChiHelloHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildEchoHelloHandler() http.Handler {
	e := echo.New()
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return e
}

func buildGinHelloHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return r
}

func buildHttpRouterHelloHandler() http.Handler {
	r := httprouter.New()
	r.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildZincStaticHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/hello", func(c *Context) error {
		return c.String(benchmarkHelloResponse)
	}))
	return app
}

func buildServeMuxStaticHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return mux
}

func buildChiStaticHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildEchoStaticHandler() http.Handler {
	e := echo.New()
	e.GET("/hello", func(c *echo.Context) error {
		return c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return e
}

func buildGinStaticHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return r
}

func buildHttpRouterStaticHandler() http.Handler {
	r := httprouter.New()
	r.GET("/hello", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildZincStaticColdHandler() http.Handler {
	app := New()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		mustNoErr(app.Get(path, func(c *Context) error {
			return c.String(benchmarkOKResponse)
		}))
	}
	return app
}

func buildServeMuxStaticColdHandler() http.Handler {
	mux := http.NewServeMux()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return mux
}

func buildChiStaticColdHandler() http.Handler {
	r := chi.NewRouter()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildEchoStaticColdHandler() http.Handler {
	e := echo.New()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		e.GET(path, func(c *echo.Context) error {
			return c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return e
}

func buildGinStaticColdHandler() http.Handler {
	r := newGinBenchmarkRouter()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		r.GET(path, func(c *gin.Context) {
			c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return r
}

func buildHttpRouterStaticColdHandler() http.Handler {
	r := httprouter.New()
	for i := 0; i < staticColdRouteCount; i++ {
		path := staticColdPath(i)
		r.GET(path, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildZincParamHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/hello/:name", func(c *Context) error {
		benchmarkSinkString = c.Param("name")
		return c.String(benchmarkHelloResponse)
	}))
	return app
}

func buildServeMuxParamHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		benchmarkSinkString = r.PathValue("name")
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return mux
}

func buildChiParamHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/hello/{name}", func(w http.ResponseWriter, r *http.Request) {
		benchmarkSinkString = chi.URLParam(r, "name")
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildEchoParamHandler() http.Handler {
	e := echo.New()
	e.GET("/hello/:name", func(c *echo.Context) error {
		benchmarkSinkString = c.Param("name")
		return c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return e
}

func buildGinParamHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/hello/:name", func(c *gin.Context) {
		benchmarkSinkString = c.Param("name")
		c.String(http.StatusOK, benchmarkHelloResponse)
	})
	return r
}

func buildHttpRouterParamHandler() http.Handler {
	r := httprouter.New()
	r.GET("/hello/:name", func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		benchmarkSinkString = ps.ByName("name")
		_, _ = io.WriteString(w, benchmarkHelloResponse)
	})
	return r
}

func buildZincJSONHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/json", func(c *Context) error {
		return c.JSON(benchmarkJSONData)
	}))
	return app
}

func buildServeMuxJSONHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkJSONData)
	})
	return mux
}

func buildChiJSONHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkJSONData)
	})
	return r
}

func buildEchoJSONHandler() http.Handler {
	e := echo.New()
	e.GET("/json", func(c *echo.Context) error {
		return c.JSON(http.StatusOK, benchmarkJSONData)
	})
	return e
}

func buildGinJSONHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/json", func(c *gin.Context) {
		c.JSON(http.StatusOK, benchmarkJSONData)
	})
	return r
}

func buildHttpRouterJSONHandler() http.Handler {
	r := httprouter.New()
	r.GET("/json", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkJSONData)
	})
	return r
}

func buildZincQueryHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/query", func(c *Context) error {
		benchmarkSinkBool = c.Query("name") != "" && c.Query("age") != "" && c.Query("city") != ""
		return c.String(benchmarkOKResponse)
	}))
	return app
}

func buildServeMuxQueryHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /query", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		benchmarkSinkBool = values.Get("name") != "" && values.Get("age") != "" && values.Get("city") != ""
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return mux
}

func buildChiQueryHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/query", func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		benchmarkSinkBool = values.Get("name") != "" && values.Get("age") != "" && values.Get("city") != ""
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildEchoQueryHandler() http.Handler {
	e := echo.New()
	e.GET("/query", func(c *echo.Context) error {
		benchmarkSinkBool = c.QueryParam("name") != "" && c.QueryParam("age") != "" && c.QueryParam("city") != ""
		return c.String(http.StatusOK, benchmarkOKResponse)
	})
	return e
}

func buildGinQueryHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/query", func(c *gin.Context) {
		benchmarkSinkBool = c.Query("name") != "" && c.Query("age") != "" && c.Query("city") != ""
		c.String(http.StatusOK, benchmarkOKResponse)
	})
	return r
}

func buildHttpRouterQueryHandler() http.Handler {
	r := httprouter.New()
	r.GET("/query", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		values := r.URL.Query()
		benchmarkSinkBool = values.Get("name") != "" && values.Get("age") != "" && values.Get("city") != ""
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildZincMiddlewareHandler() http.Handler {
	app := New()
	app.Use(
		zincMiddleware("mw1"),
		zincMiddleware("mw2"),
		zincMiddleware("mw3"),
		zincMiddleware("mw4"),
		zincMiddleware("mw5"),
	)
	mustNoErr(app.Get("/middleware", func(c *Context) error {
		_, ok1 := c.Get("mw1")
		_, ok2 := c.Get("mw2")
		_, ok3 := c.Get("mw3")
		_, ok4 := c.Get("mw4")
		_, ok5 := c.Get("mw5")
		benchmarkSinkBool = ok1 && ok2 && ok3 && ok4 && ok5
		return c.String(benchmarkOKResponse)
	}))
	return app
}

func buildServeMuxMiddlewareHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /middleware", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		benchmarkSinkBool = ctx.Value(middlewareKey1) != nil &&
			ctx.Value(middlewareKey2) != nil &&
			ctx.Value(middlewareKey3) != nil &&
			ctx.Value(middlewareKey4) != nil &&
			ctx.Value(middlewareKey5) != nil
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})

	var handler http.Handler = mux
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func buildChiMiddlewareHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware(middlewareKey1))
	r.Use(chiMiddleware(middlewareKey2))
	r.Use(chiMiddleware(middlewareKey3))
	r.Use(chiMiddleware(middlewareKey4))
	r.Use(chiMiddleware(middlewareKey5))
	r.Get("/middleware", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		benchmarkSinkBool = ctx.Value(middlewareKey1) != nil &&
			ctx.Value(middlewareKey2) != nil &&
			ctx.Value(middlewareKey3) != nil &&
			ctx.Value(middlewareKey4) != nil &&
			ctx.Value(middlewareKey5) != nil
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildEchoMiddlewareHandler() http.Handler {
	e := echo.New()
	e.Use(
		echoMiddleware("mw1"),
		echoMiddleware("mw2"),
		echoMiddleware("mw3"),
		echoMiddleware("mw4"),
		echoMiddleware("mw5"),
	)
	e.GET("/middleware", func(c *echo.Context) error {
		benchmarkSinkBool = c.Get("mw1") != nil &&
			c.Get("mw2") != nil &&
			c.Get("mw3") != nil &&
			c.Get("mw4") != nil &&
			c.Get("mw5") != nil
		return c.String(http.StatusOK, benchmarkOKResponse)
	})
	return e
}

func buildGinMiddlewareHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.Use(
		ginMiddleware("mw1"),
		ginMiddleware("mw2"),
		ginMiddleware("mw3"),
		ginMiddleware("mw4"),
		ginMiddleware("mw5"),
	)
	r.GET("/middleware", func(c *gin.Context) {
		_, ok1 := c.Get("mw1")
		_, ok2 := c.Get("mw2")
		_, ok3 := c.Get("mw3")
		_, ok4 := c.Get("mw4")
		_, ok5 := c.Get("mw5")
		benchmarkSinkBool = ok1 && ok2 && ok3 && ok4 && ok5
		c.String(http.StatusOK, benchmarkOKResponse)
	})
	return r
}

func buildHttpRouterMiddlewareHandler() http.Handler {
	r := httprouter.New()
	r.GET("/middleware", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		ctx := r.Context()
		benchmarkSinkBool = ctx.Value(middlewareKey1) != nil &&
			ctx.Value(middlewareKey2) != nil &&
			ctx.Value(middlewareKey3) != nil &&
			ctx.Value(middlewareKey4) != nil &&
			ctx.Value(middlewareKey5) != nil
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})

	var handler http.Handler = r
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func buildZincNotFoundHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/found", func(c *Context) error {
		return c.String(benchmarkOKResponse)
	}))
	return app
}

func buildServeMuxNotFoundHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /found", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return mux
}

func buildChiNotFoundHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/found", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildEchoNotFoundHandler() http.Handler {
	e := echo.New()
	e.GET("/found", func(c *echo.Context) error {
		return c.String(http.StatusOK, benchmarkOKResponse)
	})
	return e
}

func buildGinNotFoundHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/found", func(c *gin.Context) {
		c.String(http.StatusOK, benchmarkOKResponse)
	})
	return r
}

func buildHttpRouterNotFoundHandler() http.Handler {
	r := httprouter.New()
	r.GET("/found", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildZincLargeStaticHandler() http.Handler {
	app := New()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		mustNoErr(app.Get(path, func(c *Context) error {
			return c.String(benchmarkOKResponse)
		}))
	}
	return app
}

func buildServeMuxLargeStaticHandler() http.Handler {
	mux := http.NewServeMux()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		mux.HandleFunc("GET "+path, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return mux
}

func buildChiLargeStaticHandler() http.Handler {
	r := chi.NewRouter()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildEchoLargeStaticHandler() http.Handler {
	e := echo.New()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		e.GET(path, func(c *echo.Context) error {
			return c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return e
}

func buildGinLargeStaticHandler() http.Handler {
	r := newGinBenchmarkRouter()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		r.GET(path, func(c *gin.Context) {
			c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return r
}

func buildHttpRouterLargeStaticHandler() http.Handler {
	r := httprouter.New()
	for i := 0; i < largeStaticRouteCount; i++ {
		path := largeStaticPath(i)
		r.GET(path, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildZincLargeParamHandler() http.Handler {
	app := New()
	for i := 0; i < largeParamRouteCount; i++ {
		path := largeParamPatternColon(i)
		mustNoErr(app.Get(path, func(c *Context) error {
			benchmarkSinkString = c.Param("id")
			return c.String(benchmarkOKResponse)
		}))
	}
	return app
}

func buildServeMuxLargeParamHandler() http.Handler {
	mux := http.NewServeMux()
	for i := 0; i < largeParamRouteCount; i++ {
		pattern := largeParamPatternServeMux(i)
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			benchmarkSinkString = r.PathValue("id")
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return mux
}

func buildChiLargeParamHandler() http.Handler {
	r := chi.NewRouter()
	for i := 0; i < largeParamRouteCount; i++ {
		path := largeParamPatternBrace(i)
		r.Get(path, func(w http.ResponseWriter, r *http.Request) {
			benchmarkSinkString = chi.URLParam(r, "id")
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildEchoLargeParamHandler() http.Handler {
	e := echo.New()
	for i := 0; i < largeParamRouteCount; i++ {
		path := largeParamPatternColon(i)
		e.GET(path, func(c *echo.Context) error {
			benchmarkSinkString = c.Param("id")
			return c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return e
}

func buildGinLargeParamHandler() http.Handler {
	r := newGinBenchmarkRouter()
	for i := 0; i < largeParamRouteCount; i++ {
		path := largeParamPatternColon(i)
		r.GET(path, func(c *gin.Context) {
			benchmarkSinkString = c.Param("id")
			c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return r
}

func buildHttpRouterLargeParamHandler() http.Handler {
	r := httprouter.New()
	for i := 0; i < largeParamRouteCount; i++ {
		path := largeParamPatternColon(i)
		r.GET(path, func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
			benchmarkSinkString = ps.ByName("id")
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildZincRPSHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/rps", func(c *Context) error {
		return c.String(benchmarkOKResponse)
	}))
	return app
}

func buildServeMuxRPSHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /rps", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return mux
}

func buildChiRPSHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/rps", func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func buildEchoRPSHandler() http.Handler {
	e := echo.New()
	e.GET("/rps", func(c *echo.Context) error {
		return c.String(http.StatusOK, benchmarkOKResponse)
	})
	return e
}

func buildGinRPSHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/rps", func(c *gin.Context) {
		c.String(http.StatusOK, benchmarkOKResponse)
	})
	return r
}

func buildHttpRouterRPSHandler() http.Handler {
	r := httprouter.New()
	r.GET("/rps", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		_, _ = io.WriteString(w, benchmarkOKResponse)
	})
	return r
}

func requestContextMiddlewareSatisfied(r *http.Request) bool {
	ctx := r.Context()
	return ctx.Value(middlewareKey1) != nil &&
		ctx.Value(middlewareKey2) != nil &&
		ctx.Value(middlewareKey3) != nil &&
		ctx.Value(middlewareKey4) != nil &&
		ctx.Value(middlewareKey5) != nil
}

func zincMiddlewareSatisfied(c *Context) bool {
	_, ok1 := c.Get("mw1")
	_, ok2 := c.Get("mw2")
	_, ok3 := c.Get("mw3")
	_, ok4 := c.Get("mw4")
	_, ok5 := c.Get("mw5")
	return ok1 && ok2 && ok3 && ok4 && ok5
}

func echoMiddlewareSatisfied(c *echo.Context) bool {
	return c.Get("mw1") != nil &&
		c.Get("mw2") != nil &&
		c.Get("mw3") != nil &&
		c.Get("mw4") != nil &&
		c.Get("mw5") != nil
}

func ginMiddlewareSatisfied(c *gin.Context) bool {
	_, ok1 := c.Get("mw1")
	_, ok2 := c.Get("mw2")
	_, ok3 := c.Get("mw3")
	_, ok4 := c.Get("mw4")
	_, ok5 := c.Get("mw5")
	return ok1 && ok2 && ok3 && ok4 && ok5
}

func buildZincAPIParamQueryJSONHandler() http.Handler {
	app := New()
	mustNoErr(app.Get("/teams/:teamID/users/:userID", func(c *Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		return c.JSON(benchmarkAPIResponseFrom(input))
	}))
	return app
}

func buildServeMuxAPIParamQueryJSONHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /teams/{teamID}/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(r.PathValue("teamID")),
			UserID: parseBenchmarkInt(r.PathValue("userID")),
		}
		fillBenchmarkAPIQuery(&input, r)
		consumeBenchmarkAPIInput(input)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})
	return mux
}

func buildChiAPIParamQueryJSONHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/teams/{teamID}/users/{userID}", func(w http.ResponseWriter, req *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(chi.URLParam(req, "teamID")),
			UserID: parseBenchmarkInt(chi.URLParam(req, "userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		consumeBenchmarkAPIInput(input)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildEchoAPIParamQueryJSONHandler() http.Handler {
	e := echo.New()
	e.GET("/teams/:teamID/users/:userID", func(c *echo.Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		return c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return e
}

func buildGinAPIParamQueryJSONHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.GET("/teams/:teamID/users/:userID", func(c *gin.Context) {
		var input benchmarkAPIBindInput
		if err := c.ShouldBindUri(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if err := c.ShouldBindQuery(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildHttpRouterAPIParamQueryJSONHandler() http.Handler {
	r := httprouter.New()
	r.GET("/teams/:teamID/users/:userID", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(ps.ByName("teamID")),
			UserID: parseBenchmarkInt(ps.ByName("userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		consumeBenchmarkAPIInput(input)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildZincAPIHappyPathHandler() http.Handler {
	app := New()
	app.Use(
		zincMiddleware("mw1"),
		zincMiddleware("mw2"),
		zincMiddleware("mw3"),
		zincMiddleware("mw4"),
		zincMiddleware("mw5"),
	)
	mustNoErr(app.Get("/teams/:teamID/users/:userID", func(c *Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && zincMiddlewareSatisfied(c)
		return c.JSON(benchmarkAPIResponseFrom(input))
	}))
	return app
}

func buildServeMuxAPIHappyPathHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /teams/{teamID}/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(r.PathValue("teamID")),
			UserID: parseBenchmarkInt(r.PathValue("userID")),
		}
		fillBenchmarkAPIQuery(&input, r)
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(r)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})

	var handler http.Handler = mux
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func buildChiAPIHappyPathHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware(middlewareKey1))
	r.Use(chiMiddleware(middlewareKey2))
	r.Use(chiMiddleware(middlewareKey3))
	r.Use(chiMiddleware(middlewareKey4))
	r.Use(chiMiddleware(middlewareKey5))
	r.Get("/teams/{teamID}/users/{userID}", func(w http.ResponseWriter, req *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(chi.URLParam(req, "teamID")),
			UserID: parseBenchmarkInt(chi.URLParam(req, "userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(req)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildEchoAPIHappyPathHandler() http.Handler {
	e := echo.New()
	e.Use(
		echoMiddleware("mw1"),
		echoMiddleware("mw2"),
		echoMiddleware("mw3"),
		echoMiddleware("mw4"),
		echoMiddleware("mw5"),
	)
	e.GET("/teams/:teamID/users/:userID", func(c *echo.Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && echoMiddlewareSatisfied(c)
		return c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return e
}

func buildGinAPIHappyPathHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.Use(
		ginMiddleware("mw1"),
		ginMiddleware("mw2"),
		ginMiddleware("mw3"),
		ginMiddleware("mw4"),
		ginMiddleware("mw5"),
	)
	r.GET("/teams/:teamID/users/:userID", func(c *gin.Context) {
		var input benchmarkAPIBindInput
		if err := c.ShouldBindUri(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if err := c.ShouldBindQuery(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && ginMiddlewareSatisfied(c)
		c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildHttpRouterAPIHappyPathHandler() http.Handler {
	r := httprouter.New()
	r.GET("/teams/:teamID/users/:userID", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(ps.ByName("teamID")),
			UserID: parseBenchmarkInt(ps.ByName("userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(req)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})

	var handler http.Handler = r
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func buildZincAPIBindJSONHappyPathHandler() http.Handler {
	app := New()
	app.Use(
		zincMiddleware("mw1"),
		zincMiddleware("mw2"),
		zincMiddleware("mw3"),
		zincMiddleware("mw4"),
		zincMiddleware("mw5"),
	)
	mustNoErr(app.Post("/teams/:teamID/users/:userID", func(c *Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && zincMiddlewareSatisfied(c)
		return c.JSON(benchmarkAPIResponseFrom(input))
	}))
	return app
}

func buildServeMuxAPIBindJSONHappyPathHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /teams/{teamID}/users/{userID}", func(w http.ResponseWriter, r *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(r.PathValue("teamID")),
			UserID: parseBenchmarkInt(r.PathValue("userID")),
		}
		fillBenchmarkAPIQuery(&input, r)
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(r)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})

	var handler http.Handler = mux
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func buildChiAPIBindJSONHappyPathHandler() http.Handler {
	r := chi.NewRouter()
	r.Use(chiMiddleware(middlewareKey1))
	r.Use(chiMiddleware(middlewareKey2))
	r.Use(chiMiddleware(middlewareKey3))
	r.Use(chiMiddleware(middlewareKey4))
	r.Use(chiMiddleware(middlewareKey5))
	r.Post("/teams/{teamID}/users/{userID}", func(w http.ResponseWriter, req *http.Request) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(chi.URLParam(req, "teamID")),
			UserID: parseBenchmarkInt(chi.URLParam(req, "userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(req)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildEchoAPIBindJSONHappyPathHandler() http.Handler {
	e := echo.New()
	e.Use(
		echoMiddleware("mw1"),
		echoMiddleware("mw2"),
		echoMiddleware("mw3"),
		echoMiddleware("mw4"),
		echoMiddleware("mw5"),
	)
	e.POST("/teams/:teamID/users/:userID", func(c *echo.Context) error {
		var input benchmarkAPIBindInput
		if err := c.Bind(&input); err != nil {
			return err
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && echoMiddlewareSatisfied(c)
		return c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return e
}

func buildGinAPIBindJSONHappyPathHandler() http.Handler {
	r := newGinBenchmarkRouter()
	r.Use(
		ginMiddleware("mw1"),
		ginMiddleware("mw2"),
		ginMiddleware("mw3"),
		ginMiddleware("mw4"),
		ginMiddleware("mw5"),
	)
	r.POST("/teams/:teamID/users/:userID", func(c *gin.Context) {
		var input benchmarkAPIBindInput
		if err := c.ShouldBindUri(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if err := c.ShouldBindQuery(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && ginMiddlewareSatisfied(c)
		c.JSON(http.StatusOK, benchmarkAPIResponseFrom(input))
	})
	return r
}

func buildHttpRouterAPIBindJSONHappyPathHandler() http.Handler {
	r := httprouter.New()
	r.POST("/teams/:teamID/users/:userID", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		input := benchmarkAPIBindInput{
			TeamID: parseBenchmarkInt(ps.ByName("teamID")),
			UserID: parseBenchmarkInt(ps.ByName("userID")),
		}
		fillBenchmarkAPIQuery(&input, req)
		if err := json.NewDecoder(req.Body).Decode(&input); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		consumeBenchmarkAPIInput(input)
		benchmarkSinkBool = benchmarkSinkBool && requestContextMiddlewareSatisfied(req)
		w.Header().Set(HeaderContentType, "application/json")
		_ = json.NewEncoder(w).Encode(benchmarkAPIResponseFrom(input))
	})

	var handler http.Handler = r
	handler = stdMiddleware(handler, middlewareKey5)
	handler = stdMiddleware(handler, middlewareKey4)
	handler = stdMiddleware(handler, middlewareKey3)
	handler = stdMiddleware(handler, middlewareKey2)
	handler = stdMiddleware(handler, middlewareKey1)
	return handler
}

func helloWorldCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincHelloHandler},
		{name: "ServeMux", build: buildServeMuxHelloHandler},
		{name: "HttpRouter", build: buildHttpRouterHelloHandler},
		{name: "Chi", build: buildChiHelloHandler},
		{name: "Echo", build: buildEchoHelloHandler},
		{name: "Gin", build: buildGinHelloHandler},
	}
}

func staticRouteCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincStaticHandler},
		{name: "ServeMux", build: buildServeMuxStaticHandler},
		{name: "HttpRouter", build: buildHttpRouterStaticHandler},
		{name: "Chi", build: buildChiStaticHandler},
		{name: "Echo", build: buildEchoStaticHandler},
		{name: "Gin", build: buildGinStaticHandler},
	}
}

func staticColdCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincStaticColdHandler},
		{name: "ServeMux", build: buildServeMuxStaticColdHandler},
		{name: "HttpRouter", build: buildHttpRouterStaticColdHandler},
		{name: "Chi", build: buildChiStaticColdHandler},
		{name: "Echo", build: buildEchoStaticColdHandler},
		{name: "Gin", build: buildGinStaticColdHandler},
	}
}

func paramCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincParamHandler},
		{name: "ServeMux", build: buildServeMuxParamHandler},
		{name: "HttpRouter", build: buildHttpRouterParamHandler},
		{name: "Chi", build: buildChiParamHandler},
		{name: "Echo", build: buildEchoParamHandler},
		{name: "Gin", build: buildGinParamHandler},
	}
}

func jsonCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincJSONHandler},
		{name: "ServeMux", build: buildServeMuxJSONHandler},
		{name: "HttpRouter", build: buildHttpRouterJSONHandler},
		{name: "Chi", build: buildChiJSONHandler},
		{name: "Echo", build: buildEchoJSONHandler},
		{name: "Gin", build: buildGinJSONHandler},
	}
}

func queryCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincQueryHandler},
		{name: "ServeMux", build: buildServeMuxQueryHandler},
		{name: "HttpRouter", build: buildHttpRouterQueryHandler},
		{name: "Chi", build: buildChiQueryHandler},
		{name: "Echo", build: buildEchoQueryHandler},
		{name: "Gin", build: buildGinQueryHandler},
	}
}

func middlewareCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincMiddlewareHandler},
		{name: "ServeMux", build: buildServeMuxMiddlewareHandler},
		{name: "HttpRouter", build: buildHttpRouterMiddlewareHandler},
		{name: "Chi", build: buildChiMiddlewareHandler},
		{name: "Echo", build: buildEchoMiddlewareHandler},
		{name: "Gin", build: buildGinMiddlewareHandler},
	}
}

func notFoundCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincNotFoundHandler},
		{name: "ServeMux", build: buildServeMuxNotFoundHandler},
		{name: "HttpRouter", build: buildHttpRouterNotFoundHandler},
		{name: "Chi", build: buildChiNotFoundHandler},
		{name: "Echo", build: buildEchoNotFoundHandler},
		{name: "Gin", build: buildGinNotFoundHandler},
	}
}

func largeStaticCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincLargeStaticHandler},
		{name: "ServeMux", build: buildServeMuxLargeStaticHandler},
		{name: "HttpRouter", build: buildHttpRouterLargeStaticHandler},
		{name: "Chi", build: buildChiLargeStaticHandler},
		{name: "Echo", build: buildEchoLargeStaticHandler},
		{name: "Gin", build: buildGinLargeStaticHandler},
	}
}

func largeParamCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincLargeParamHandler},
		{name: "ServeMux", build: buildServeMuxLargeParamHandler},
		{name: "HttpRouter", build: buildHttpRouterLargeParamHandler},
		{name: "Chi", build: buildChiLargeParamHandler},
		{name: "Echo", build: buildEchoLargeParamHandler},
		{name: "Gin", build: buildGinLargeParamHandler},
	}
}

func apiParamQueryJSONCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincAPIParamQueryJSONHandler},
		{name: "ServeMux", build: buildServeMuxAPIParamQueryJSONHandler},
		{name: "HttpRouter", build: buildHttpRouterAPIParamQueryJSONHandler},
		{name: "Chi", build: buildChiAPIParamQueryJSONHandler},
		{name: "Echo", build: buildEchoAPIParamQueryJSONHandler},
		{name: "Gin", build: buildGinAPIParamQueryJSONHandler},
	}
}

func apiHappyPathCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincAPIHappyPathHandler},
		{name: "ServeMux", build: buildServeMuxAPIHappyPathHandler},
		{name: "HttpRouter", build: buildHttpRouterAPIHappyPathHandler},
		{name: "Chi", build: buildChiAPIHappyPathHandler},
		{name: "Echo", build: buildEchoAPIHappyPathHandler},
		{name: "Gin", build: buildGinAPIHappyPathHandler},
	}
}

func apiBindJSONHappyPathCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincAPIBindJSONHappyPathHandler},
		{name: "ServeMux", build: buildServeMuxAPIBindJSONHappyPathHandler},
		{name: "HttpRouter", build: buildHttpRouterAPIBindJSONHappyPathHandler},
		{name: "Chi", build: buildChiAPIBindJSONHappyPathHandler},
		{name: "Echo", build: buildEchoAPIBindJSONHappyPathHandler},
		{name: "Gin", build: buildGinAPIBindJSONHappyPathHandler},
	}
}

func rpsCases() []benchmarkCase {
	return []benchmarkCase{
		{name: "Zinc", build: buildZincRPSHandler},
		{name: "ServeMux", build: buildServeMuxRPSHandler},
		{name: "HttpRouter", build: buildHttpRouterRPSHandler},
		{name: "Chi", build: buildChiRPSHandler},
		{name: "Echo", build: buildEchoRPSHandler},
		{name: "Gin", build: buildGinRPSHandler},
	}
}

func BenchmarkHelloWorld(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/", helloWorldCases())
}

func BenchmarkStaticRoute(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/hello", staticRouteCases())
}

func BenchmarkStaticRouteCold(b *testing.B) {
	runServeHTTPRequestSetBenchmarks(b, staticColdCases(), buildRequests(http.MethodGet, staticColdTargets(coldPathRequestCount)))
}

func BenchmarkRouterParam(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/hello/world", paramCases())
}

func BenchmarkRouterParamCold(b *testing.B) {
	runServeHTTPRequestSetBenchmarks(b, paramCases(), buildRequests(http.MethodGet, coldParamTargets(coldPathRequestCount)))
}

func BenchmarkJSONResponse(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/json", jsonCases())
}

func BenchmarkQueryParams(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/query?name=john&age=25&city=london", queryCases())
}

func BenchmarkMiddlewareChain(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/middleware", middlewareCases())
}

func BenchmarkNotFound(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/missing", notFoundCases())
}

func BenchmarkLargeRouteSetStatic(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, largeStaticPath(largeStaticRouteCount-1), largeStaticCases())
}

func BenchmarkLargeRouteSetStaticMixed(b *testing.B) {
	runServeHTTPRequestSetBenchmarks(b, largeStaticCases(), buildRequests(http.MethodGet, mixedLargeStaticTargets(coldPathRequestCount)))
}

func BenchmarkLargeRouteSetNotFound(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, "/static/missing", largeStaticCases())
}

func BenchmarkLargeRouteSetMethodMismatch(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodPost, largeStaticPath(largeStaticRouteCount-1), largeStaticCases())
}

func BenchmarkLargeRouteSetParam(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, largeParamPath(largeParamRouteCount-1), largeParamCases())
}

func BenchmarkLargeRouteSetParamMixed(b *testing.B) {
	runServeHTTPRequestSetBenchmarks(b, largeParamCases(), buildRequests(http.MethodGet, mixedLargeParamTargets(coldPathRequestCount)))
}

func BenchmarkRouteRegistrationStatic(b *testing.B) {
	runRegistrationBenchmarks(b, largeStaticCases())
}

func BenchmarkRouteRegistrationParam(b *testing.B) {
	runRegistrationBenchmarks(b, largeParamCases())
}

func BenchmarkAPIParamQueryJSON(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, benchmarkAPIQueryTarget(), apiParamQueryJSONCases())
}

func BenchmarkAPIHappyPath(b *testing.B) {
	runServeHTTPBenchmarks(b, http.MethodGet, benchmarkAPIQueryTarget(), apiHappyPathCases())
}

func BenchmarkAPIBindJSONHappyPath(b *testing.B) {
	runPreparedRequestBenchmarks(
		b,
		apiBindJSONHappyPathCases(),
		newPreparedBenchmarkRequest(http.MethodPost, benchmarkAPIBindTarget(), benchmarkAPIBindBody, benchmarkAPIBodyHeaders()),
	)
}

func BenchmarkRequestsPerSecond(b *testing.B) {
	for _, concurrency := range throughputConcurrencyLevels {
		concurrency := concurrency
		b.Run("Concurrency"+strconv.Itoa(concurrency), func(b *testing.B) {
			for _, bc := range rpsCases() {
				b.Run(bc.name, func(b *testing.B) {
					server := httptest.NewServer(bc.build())
					defer server.Close()
					measureRPS(b, server.URL+"/rps", concurrency, throughputDuration)
				})
			}
		})
	}
}

func measureRPS(b *testing.B, url string, concurrency int, duration time.Duration) {
	var (
		totalRequests int64
		wg            sync.WaitGroup
		client        = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        concurrency,
				MaxIdleConnsPerHost: concurrency,
				MaxConnsPerHost:     concurrency,
				DisableKeepAlives:   false,
			},
		}
	)
	baseReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		b.Fatalf("create base request: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					req := baseReq.Clone(ctx)
					resp, err := client.Do(req)
					if err != nil {
						if !errors.Is(err, context.DeadlineExceeded) {
							b.Logf("request error: %v", err)
						}
						continue
					}

					_, _ = io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					if resp.StatusCode == http.StatusOK {
						atomic.AddInt64(&totalRequests, 1)
					}
				}
			}
		}()
	}

	wg.Wait()

	requestsPerSec := float64(totalRequests) / duration.Seconds()
	b.ReportMetric(requestsPerSec, "reqs/s")
}

func TestRunBenchmarks(t *testing.T) {
	t.Skip(`
From benchmarks/:
    go test -run=^$ -bench 'BenchmarkHelloWorld|BenchmarkStaticRoute|BenchmarkStaticRouteCold|BenchmarkRouterParam|BenchmarkRouterParamCold|BenchmarkJSONResponse|BenchmarkQueryParams|BenchmarkMiddlewareChain|BenchmarkNotFound|BenchmarkLargeRouteSetStatic|BenchmarkLargeRouteSetStaticMixed|BenchmarkLargeRouteSetNotFound|BenchmarkLargeRouteSetMethodMismatch|BenchmarkLargeRouteSetParam|BenchmarkLargeRouteSetParamMixed|BenchmarkAPIParamQueryJSON|BenchmarkAPIHappyPath|BenchmarkAPIBindJSONHappyPath|BenchmarkRouteRegistrationStatic|BenchmarkRouteRegistrationParam' -benchmem

To run the comparison suite from the repo root:
    cd benchmarks && go test -run=^$ -bench 'BenchmarkHelloWorld|BenchmarkStaticRoute|BenchmarkStaticRouteCold|BenchmarkRouterParam|BenchmarkRouterParamCold|BenchmarkJSONResponse|BenchmarkQueryParams|BenchmarkMiddlewareChain|BenchmarkNotFound|BenchmarkLargeRouteSetStatic|BenchmarkLargeRouteSetStaticMixed|BenchmarkLargeRouteSetNotFound|BenchmarkLargeRouteSetMethodMismatch|BenchmarkLargeRouteSetParam|BenchmarkLargeRouteSetParamMixed|BenchmarkAPIParamQueryJSON|BenchmarkAPIHappyPath|BenchmarkAPIBindJSONHappyPath|BenchmarkRouteRegistrationStatic|BenchmarkRouteRegistrationParam' -benchmem

Dispatch-only slice:
    go test -run=^$ -bench 'BenchmarkHelloWorld|BenchmarkStaticRoute|BenchmarkStaticRouteCold|BenchmarkRouterParam|BenchmarkRouterParamCold|BenchmarkNotFound|BenchmarkLargeRouteSetStatic|BenchmarkLargeRouteSetStaticMixed|BenchmarkLargeRouteSetNotFound|BenchmarkLargeRouteSetMethodMismatch|BenchmarkLargeRouteSetParam|BenchmarkLargeRouteSetParamMixed' -benchmem

Idiomatic framework-path slice:
    go test -run=^$ -bench 'BenchmarkJSONResponse|BenchmarkQueryParams|BenchmarkMiddlewareChain|BenchmarkAPIParamQueryJSON|BenchmarkAPIHappyPath|BenchmarkAPIBindJSONHappyPath' -benchmem

To run the end-to-end throughput benchmark from benchmarks/:
    go test -run=^$ -bench BenchmarkRequestsPerSecond

Notes:
- The request/response harness now reuses requests and a discard response writer to reduce benchmark noise.
- Body-consuming endpoint benches reset request bodies between iterations instead of charging full request construction cost to the handler path.
- The suite includes ServeMux and HttpRouter as additional net/http-based baselines.
- BenchmarkHelloWorld exercises Zinc's special-case root fast path, while BenchmarkStaticRoute measures a normal non-root static route.
- The cold route benchmarks rotate request paths to avoid flattering Zinc's route cache.
- The throughput benchmark uses a real loopback listener with a concurrency sweep; run it with -count=3 or higher and compare reqs/s, not ns/op.
`)
}
