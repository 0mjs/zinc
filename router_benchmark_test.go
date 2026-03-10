package zinc

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

const zincBenchmarkOK = "OK"

var (
	zincBenchmarkSinkHandler HandlerFunc
	zincBenchmarkSinkString  string
	zincBenchmarkSinkInt     int
)

type zincBenchmarkDiscardWriter struct {
	header http.Header
	status int
	bytes  int
}

func newZincBenchmarkDiscardWriter() *zincBenchmarkDiscardWriter {
	return &zincBenchmarkDiscardWriter{
		header: make(http.Header, 8),
	}
}

func (w *zincBenchmarkDiscardWriter) Header() http.Header {
	return w.header
}

func (w *zincBenchmarkDiscardWriter) WriteHeader(code int) {
	w.status = code
}

func (w *zincBenchmarkDiscardWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.bytes += len(p)
	return len(p), nil
}

func (w *zincBenchmarkDiscardWriter) WriteString(s string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	w.bytes += len(s)
	return len(s), nil
}

func (w *zincBenchmarkDiscardWriter) reset() {
	clear(w.header)
	w.status = 0
	w.bytes = 0
}

func mustNoErrBenchmark(err error) {
	if err != nil {
		panic(err)
	}
}

func newBareBenchmarkContext() *Context {
	return &Context{
		status:     http.StatusOK,
		index:      -1,
		routeIndex: -1,
	}
}

func resetBareBenchmarkContext(c *Context) {
	c.initPathParams()
	c.writer = nil
	c.request = nil
	c.queryParams = nil
	c.written = false
	c.handlers = nil
	c.index = -1
	c.status = http.StatusOK
	c.app = nil
	c.routeInfo = routeMeta{}
	c.routeIndex = -1
	c.routeIndexed = false
	c.lastErr = nil
	c.body = nil
	c.bodyRead = false
	c.bodyErr = nil
	c.paramPath = ""
	for i := 0; i < c.paramCount; i++ {
		c.PathParams[i] = emptyParam
	}
	c.paramCount = 0
	if len(c.store) > 0 {
		for key := range c.store {
			delete(c.store, key)
		}
	}
}

func newDispatchBenchmarkContext(method, target string) (*Context, *zincBenchmarkDiscardWriter, *http.Request) {
	rw := newZincBenchmarkDiscardWriter()
	req := httptest.NewRequest(method, target, nil)
	ctx := newBareBenchmarkContext()
	ctx.reset(rw, req)
	return ctx, rw, req
}

func resetDispatchBenchmarkContext(ctx *Context, rw *zincBenchmarkDiscardWriter, req *http.Request) {
	rw.reset()
	ctx.reset(rw, req)
}

func newDiagnosticRouterApp(cfg Config) *App {
	app := NewWithConfig(cfg)
	mustNoErrBenchmark(app.Get("/hello", func(c *Context) error {
		return c.String(zincBenchmarkOK)
	}))
	mustNoErrBenchmark(app.Get("/teams/:teamId/users/:userId", func(c *Context) error {
		zincBenchmarkSinkString = c.Param("teamId") + "|" + c.Param("userId")
		return c.String(zincBenchmarkOK)
	}))
	mustNoErrBenchmark(app.Get("/teams/:teamId/users/:userId/preferences", func(c *Context) error {
		return c.String(zincBenchmarkOK)
	}))
	mustNoErrBenchmark(app.Get("/items/:id", func(c *Context) error {
		zincBenchmarkSinkString = c.Param("id")
		return c.String(zincBenchmarkOK)
	}))
	return app
}

func newCacheBenchmarkApp() *App {
	cfg := DefaultConfig
	cfg.RouteCacheSize = 64
	app := NewWithConfig(cfg)
	for i := 0; i < 64; i++ {
		path := "/cache/" + strconv.Itoa(i) + "/items/:id"
		mustNoErrBenchmark(app.Get(path, func(c *Context) error {
			zincBenchmarkSinkString = c.Param("id")
			return c.String(zincBenchmarkOK)
		}))
	}
	return app
}

func newCaseInsensitiveBenchmarkApp() *App {
	cfg := DefaultConfig
	cfg.CaseSensitive = false
	app := NewWithConfig(cfg)
	mustNoErrBenchmark(app.Get("/Reports/Daily", func(c *Context) error {
		return c.String(zincBenchmarkOK)
	}))
	return app
}

func newStrictRoutingBenchmarkApp() *App {
	cfg := DefaultConfig
	cfg.StrictRouting = true
	cfg.RouteCacheSize = 0
	app := NewWithConfig(cfg)
	mustNoErrBenchmark(app.Get("/teams/:teamId", func(c *Context) error {
		return c.String(zincBenchmarkOK)
	}))
	return app
}

func newMountBenchmarkApp() *App {
	app := New()
	app.Mount("/assets", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			_, _ = io.WriteString(w, zincBenchmarkOK)
		case "/css/app/site.css":
			_, _ = io.WriteString(w, zincBenchmarkOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	return app
}

func buildColdCacheTargets(count int) []string {
	targets := make([]string, count)
	for i := 0; i < count; i++ {
		targets[i] = "/cache/" + strconv.Itoa(i%64) + "/items/" + strconv.Itoa(1000+i)
	}
	return targets
}

func BenchmarkZincRouterFindStatic(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx := newBareBenchmarkContext()
	resetBareBenchmarkContext(ctx)
	handler := router.findInto(MethodGet, "/hello", ctx)
	if handler == nil || ctx.routeInfo.path != "/hello" {
		b.Fatalf("unexpected static find result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetBareBenchmarkContext(ctx)
		zincBenchmarkSinkHandler = router.findInto(MethodGet, "/hello", ctx)
	}
}

func BenchmarkZincRouterFindParam(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx := newBareBenchmarkContext()
	resetBareBenchmarkContext(ctx)
	handler := router.findInto(MethodGet, "/teams/42/users/7", ctx)
	if handler == nil || ctx.Param("teamId") != "42" || ctx.Param("userId") != "7" {
		b.Fatalf("unexpected param find result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetBareBenchmarkContext(ctx)
		zincBenchmarkSinkHandler = router.findInto(MethodGet, "/teams/42/users/7", ctx)
		zincBenchmarkSinkString = ctx.Param("teamId") + "|" + ctx.Param("userId")
	}
}

func BenchmarkZincRouterFindNotFound(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx := newBareBenchmarkContext()
	resetBareBenchmarkContext(ctx)
	if handler := router.findInto(MethodGet, "/missing/path", ctx); handler != nil {
		b.Fatalf("expected miss")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetBareBenchmarkContext(ctx)
		zincBenchmarkSinkHandler = router.findInto(MethodGet, "/missing/path", ctx)
	}
}

func BenchmarkZincRouterDispatchStatic(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, "/hello")
	handled, allowed, err := router.dispatchInto(MethodGet, "/hello", false, ctx)
	if !handled || !allowed.empty() || err != nil || rw.status != http.StatusOK {
		b.Fatalf("unexpected static dispatch result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		handled, _, err = router.dispatchInto(MethodGet, "/hello", false, ctx)
		if !handled || err != nil {
			b.Fatalf("dispatch failed")
		}
	}
	zincBenchmarkSinkInt = rw.status + rw.bytes
}

func BenchmarkZincRouterDispatchParam(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, "/teams/42/users/7")
	zincBenchmarkSinkString = ""
	handled, allowed, err := router.dispatchInto(MethodGet, "/teams/42/users/7", false, ctx)
	if !handled || !allowed.empty() || err != nil || zincBenchmarkSinkString != "42|7" || rw.status != http.StatusOK {
		b.Fatalf("unexpected param dispatch result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		zincBenchmarkSinkString = ""
		handled, _, err = router.dispatchInto(MethodGet, "/teams/42/users/7", false, ctx)
		if !handled || err != nil {
			b.Fatalf("dispatch failed")
		}
	}
}

func BenchmarkZincRouterDispatchNotFound(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, "/missing/path")
	handled, allowed, err := router.dispatchInto(MethodGet, "/missing/path", true, ctx)
	if handled || !allowed.empty() || err != nil {
		b.Fatalf("unexpected not found result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		handled, allowed, err = router.dispatchInto(MethodGet, "/missing/path", true, ctx)
		if handled || err != nil || !allowed.empty() {
			b.Fatalf("dispatch not found failed")
		}
	}
}

func BenchmarkZincRouterDispatchMethodMismatch(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	target := "/teams/42/users/7/preferences"
	ctx, rw, req := newDispatchBenchmarkContext(MethodPost, target)
	handled, allowed, err := router.dispatchInto(MethodPost, target, true, ctx)
	header := allowed.header(true, true)
	if handled || err != nil || header == "" || !strings.Contains(header, http.MethodGet) {
		b.Fatalf("unexpected method mismatch result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		handled, allowed, err = router.dispatchInto(MethodPost, target, true, ctx)
		if handled || err != nil {
			b.Fatalf("dispatch mismatch failed")
		}
		zincBenchmarkSinkString = allowed.header(true, true)
	}
}

func BenchmarkZincRouterCacheHitParam(b *testing.B) {
	router := newCacheBenchmarkApp().router
	target := "/cache/63/items/999"
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, target)
	handled, _, err := router.dispatchInto(MethodGet, target, false, ctx)
	if !handled || err != nil || zincBenchmarkSinkString != "999" {
		b.Fatalf("unexpected cache warmup result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		zincBenchmarkSinkString = ""
		handled, _, err = router.dispatchInto(MethodGet, target, false, ctx)
		if !handled || err != nil {
			b.Fatalf("cache hit failed")
		}
	}
}

func BenchmarkZincRouterCacheColdParam(b *testing.B) {
	router := newCacheBenchmarkApp().router
	targets := buildColdCacheTargets(256)
	requests := make([]*http.Request, len(targets))
	for i, target := range targets {
		requests[i] = httptest.NewRequest(MethodGet, target, nil)
	}

	ctx := newBareBenchmarkContext()
	rw := newZincBenchmarkDiscardWriter()
	ctx.reset(rw, requests[0])
	handled, _, err := router.dispatchInto(MethodGet, targets[0], false, ctx)
	if !handled || err != nil {
		b.Fatalf("unexpected cold cache warmup result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		target := targets[i%len(targets)]
		req := requests[i%len(requests)]
		resetDispatchBenchmarkContext(ctx, rw, req)
		zincBenchmarkSinkString = ""
		handled, _, err = router.dispatchInto(MethodGet, target, false, ctx)
		if !handled || err != nil {
			b.Fatalf("cache cold dispatch failed")
		}
	}
}

func BenchmarkZincRouterCaseInsensitiveStatic(b *testing.B) {
	router := newCaseInsensitiveBenchmarkApp().router
	target := "/reports/daily"
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, target)
	handled, _, err := router.dispatchInto(MethodGet, target, false, ctx)
	if !handled || err != nil || rw.status != http.StatusOK {
		b.Fatalf("unexpected case insensitive result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		handled, _, err = router.dispatchInto(MethodGet, target, false, ctx)
		if !handled || err != nil {
			b.Fatalf("case insensitive dispatch failed")
		}
	}
}

func BenchmarkZincRouterStrictRoutingSlash(b *testing.B) {
	router := newStrictRoutingBenchmarkApp().router
	target := "/teams/42/"
	ctx, rw, req := newDispatchBenchmarkContext(MethodGet, target)
	handled, allowed, err := router.dispatchInto(MethodGet, target, true, ctx)
	if handled || err != nil || !allowed.empty() {
		b.Fatalf("unexpected strict routing result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetDispatchBenchmarkContext(ctx, rw, req)
		handled, allowed, err = router.dispatchInto(MethodGet, target, true, ctx)
		if handled || err != nil || !allowed.empty() {
			b.Fatalf("strict routing dispatch failed")
		}
	}
}

func BenchmarkZincRouterMethodNotAllowed(b *testing.B) {
	router := newDiagnosticRouterApp(DefaultConfig).router
	header := router.allowedMethodHeader("/items/42", true, true)
	if header == "" || !strings.Contains(header, http.MethodGet) {
		b.Fatalf("unexpected allow header")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		zincBenchmarkSinkString = router.allowedMethodHeader("/items/42", true, true)
	}
}

func BenchmarkZincMountStatic(b *testing.B) {
	app := newMountBenchmarkApp()
	req := httptest.NewRequest(MethodGet, "/assets/health", nil)
	rw := newZincBenchmarkDiscardWriter()
	app.ServeHTTP(rw, req)
	if rw.status != http.StatusOK {
		b.Fatalf("unexpected mount static result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		app.ServeHTTP(rw, req)
	}
	zincBenchmarkSinkInt = rw.status + rw.bytes
}

func BenchmarkZincMountDeepPath(b *testing.B) {
	app := newMountBenchmarkApp()
	req := httptest.NewRequest(MethodGet, "/assets/css/app/site.css", nil)
	rw := newZincBenchmarkDiscardWriter()
	app.ServeHTTP(rw, req)
	if rw.status != http.StatusOK {
		b.Fatalf("unexpected mount deep result")
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rw.reset()
		app.ServeHTTP(rw, req)
	}
	zincBenchmarkSinkInt = rw.status + rw.bytes
}

func TestRunZincDiagnostics(t *testing.T) {
	t.Skip(`
Run the Zinc diagnostic benchmark slice from the repo root:
    go test -run=^$ -bench '^BenchmarkZinc(Router|Mount|Bind)' -benchmem

Notes:
- These benches are intentionally Zinc-only and live in the root package so they can exercise internal router and mount behavior directly.
- The bind benches measure direct Context binding paths separately from the public cross-framework suite.
- They are regression guards, not public framework-comparison numbers.
`)
}
