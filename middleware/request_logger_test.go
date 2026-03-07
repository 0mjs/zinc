package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/0mjs/zinc"
)

func TestRequestLoggerCapturesValues(t *testing.T) {
	app := zinc.New()

	var called int
	var got RequestLoggerValues
	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		LogLatency:       true,
		LogMethod:        true,
		LogURI:           true,
		LogRoutePath:     true,
		LogStatus:        true,
		LogError:         true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogUserAgent:     true,
		LogRequestID:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogHeaders:       []string{"X-Test"},
		LogQueryParams:   []string{"q"},
		LogValuesFunc: func(c *zinc.Context, v RequestLoggerValues) error {
			called++
			got = v
			return nil
		},
	}))
	mustNoErr(t, app.Get("/users/:id", func(c *zinc.Context) error {
		return c.String("ok")
	}))

	req := httptest.NewRequest(http.MethodGet, "http://example.com/users/42?q=one", nil)
	req.RemoteAddr = "10.2.3.4:12345"
	req.Header.Set("User-Agent", "zinc-test")
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("X-Test", "value")
	req.Header.Set("Content-Length", "123")
	expectedURI := req.RequestURI

	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if body := rec.Body.String(); body != "ok" {
		t.Fatalf("body=%q", body)
	}
	if called != 1 {
		t.Fatalf("log callback calls=%d", called)
	}
	if got.StartTime.IsZero() {
		t.Fatal("start time should be set")
	}
	if got.Method != http.MethodGet {
		t.Fatalf("method=%q", got.Method)
	}
	if got.URI != expectedURI {
		t.Fatalf("uri=%q", got.URI)
	}
	if got.RoutePath != "/users/:id" {
		t.Fatalf("route path=%q", got.RoutePath)
	}
	if got.Status != http.StatusOK {
		t.Fatalf("logged status=%d", got.Status)
	}
	if got.Error != nil {
		t.Fatalf("error=%v", got.Error)
	}
	if got.RemoteIP != "10.2.3.4" {
		t.Fatalf("remote ip=%q", got.RemoteIP)
	}
	if got.Host != "example.com" {
		t.Fatalf("host=%q", got.Host)
	}
	if got.UserAgent != "zinc-test" {
		t.Fatalf("user agent=%q", got.UserAgent)
	}
	if got.RequestID != "req-123" {
		t.Fatalf("request id=%q", got.RequestID)
	}
	if got.ContentLength != "123" {
		t.Fatalf("content-length=%q", got.ContentLength)
	}
	if got.ResponseSize != 2 {
		t.Fatalf("response size=%d", got.ResponseSize)
	}
	if values := got.Headers["X-Test"]; len(values) != 1 || values[0] != "value" {
		t.Fatalf("headers=%v", got.Headers)
	}
	if values := got.QueryParams["q"]; len(values) != 1 || values[0] != "one" {
		t.Fatalf("query params=%v", got.QueryParams)
	}
}

func TestRequestLoggerCapturesRouteErrorStatus(t *testing.T) {
	app := zinc.New()

	var got RequestLoggerValues
	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		LogStatus: true,
		LogError:  true,
		LogValuesFunc: func(c *zinc.Context, v RequestLoggerValues) error {
			got = v
			return nil
		},
	}))
	mustNoErr(t, app.Get("/boom", func(c *zinc.Context) error {
		return zinc.ErrForbidden
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d", rec.Code)
	}
	if got.Status != http.StatusForbidden {
		t.Fatalf("logged status=%d", got.Status)
	}
	if !errors.Is(got.Error, zinc.ErrForbidden) {
		t.Fatalf("logged error=%v", got.Error)
	}
}

func TestRequestLoggerHandleErrorForMiddlewareError(t *testing.T) {
	app := zinc.New()

	var got RequestLoggerValues
	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		HandleError: true,
		LogStatus:   true,
		LogError:    true,
		LogValuesFunc: func(c *zinc.Context, v RequestLoggerValues) error {
			got = v
			return nil
		},
	}))
	app.Use(func(c *zinc.Context) error {
		return zinc.ErrUnauthorized
	})
	mustNoErr(t, app.Get("/protected", func(c *zinc.Context) error {
		return c.String("ok")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d", rec.Code)
	}
	if got.Status != http.StatusUnauthorized {
		t.Fatalf("logged status=%d", got.Status)
	}
	if !errors.Is(got.Error, zinc.ErrUnauthorized) {
		t.Fatalf("logged error=%v", got.Error)
	}
}

func TestRequestLoggerSkipper(t *testing.T) {
	app := zinc.New()

	calls := 0
	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		Skipper:   func(*zinc.Context) bool { return true },
		LogStatus: true,
		LogValuesFunc: func(c *zinc.Context, v RequestLoggerValues) error {
			calls++
			return nil
		},
	}))
	mustNoErr(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("ok")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if calls != 0 {
		t.Fatalf("log callback calls=%d", calls)
	}
}

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDefaultRequestLoggerConfig(t *testing.T) {
	cfg := DefaultRequestLoggerConfig()
	if !cfg.HandleError {
		t.Fatal("HandleError should default to true")
	}
	if !cfg.LogLatency || !cfg.LogMethod || !cfg.LogURI || !cfg.LogStatus || !cfg.LogError {
		t.Fatal("core request logging flags should default to true")
	}
	if !cfg.LogRemoteIP || !cfg.LogHost || !cfg.LogUserAgent || !cfg.LogRequestID {
		t.Fatal("request metadata flags should default to true")
	}
	if !cfg.LogContentLength || !cfg.LogResponseSize {
		t.Fatal("payload size flags should default to true")
	}
}

func TestRequestLoggerAliases(t *testing.T) {
	app := zinc.New()
	app.Use(RequestLogger())
	app.Use(Logger())

	calls := 0
	app.Use(LoggerWithConfig(RequestLoggerConfig{
		LogStatus: true,
		LogValuesFunc: func(*zinc.Context, RequestLoggerValues) error {
			calls++
			return nil
		},
	}))

	mustNoErr(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("ok")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if calls != 1 {
		t.Fatalf("logger with config callback calls=%d", calls)
	}
}

func TestRequestLoggerLogValuesErrorIsReturned(t *testing.T) {
	app := zinc.New()
	wantErr := errors.New("log failed")

	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		LogStatus: true,
		LogValuesFunc: func(*zinc.Context, RequestLoggerValues) error {
			return wantErr
		},
	}))
	mustNoErr(t, app.Get("/", func(c *zinc.Context) error {
		return nil
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestRequestLoggerUsesBeforeNextFunc(t *testing.T) {
	app := zinc.New()
	beforeCalls := 0
	logCalls := 0

	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		BeforeNextFunc: func(*zinc.Context) {
			beforeCalls++
		},
		LogStatus: true,
		LogValuesFunc: func(*zinc.Context, RequestLoggerValues) error {
			logCalls++
			return nil
		},
	}))
	mustNoErr(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("ok")
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d", rec.Code)
	}
	if beforeCalls != 1 {
		t.Fatalf("before next calls=%d", beforeCalls)
	}
	if logCalls != 1 {
		t.Fatalf("log callback calls=%d", logCalls)
	}
}

func TestRequestLoggerRequestIDFallsBackToResponseHeader(t *testing.T) {
	app := zinc.New()

	var got RequestLoggerValues
	app.Use(RequestLoggerWithConfig(RequestLoggerConfig{
		LogRequestID: true,
		LogStatus:    true, // ensures response writer wrapping is enabled
		LogValuesFunc: func(_ *zinc.Context, v RequestLoggerValues) error {
			got = v
			return nil
		},
	}))
	mustNoErr(t, app.Get("/", func(c *zinc.Context) error {
		c.SetHeader(zinc.HeaderXRequestID, "resp-123")
		return c.NoContent()
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
	if got.RequestID != "resp-123" {
		t.Fatalf("request id=%q", got.RequestID)
	}
}

func TestDefaultRequestLogValuesFunc(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	logFn := defaultRequestLogValuesFunc(logger)

	cases := []RequestLoggerValues{
		{Method: http.MethodGet, URI: "/ok", RoutePath: "/ok", Status: http.StatusOK},
		{Method: http.MethodGet, URI: "/ok", Status: http.StatusOK},
		{Method: http.MethodGet, URI: "/boom", RoutePath: "/boom", Status: http.StatusInternalServerError, Error: errors.New("boom")},
		{Method: http.MethodGet, URI: "/boom", Status: http.StatusInternalServerError, Error: errors.New("boom")},
	}
	for _, tc := range cases {
		if err := logFn(nil, tc); err != nil {
			t.Fatalf("unexpected log function error: %v", err)
		}
	}

	// nil logger should fall back to slog.Default()
	if err := defaultRequestLogValuesFunc(nil)(nil, RequestLoggerValues{Method: http.MethodGet}); err != nil {
		t.Fatalf("unexpected fallback logger error: %v", err)
	}
}

func TestResolveRequestLogStatus(t *testing.T) {
	rw := newRequestLoggerResponseWriter(httptest.NewRecorder())
	if got := resolveRequestLogStatus(rw, nil); got != http.StatusOK {
		t.Fatalf("status with unwritten writer=%d", got)
	}

	rw.WriteHeader(http.StatusAccepted)
	if got := resolveRequestLogStatus(rw, errors.New("ignored")); got != http.StatusAccepted {
		t.Fatalf("status with written writer=%d", got)
	}

	wrappedHTTPError := fmt.Errorf("wrapped: %w", zinc.NewError(http.StatusTeapot))
	if got := resolveRequestLogStatus(nil, wrappedHTTPError); got != http.StatusTeapot {
		t.Fatalf("status from wrapped HTTPError=%d", got)
	}

	if got := resolveRequestLogStatus(nil, errors.New("boom")); got != http.StatusInternalServerError {
		t.Fatalf("status from non-http error=%d", got)
	}
}

func TestRequestLoggerResponseWriterWritePaths(t *testing.T) {
	base := &basicResponseWriter{}
	rw := newRequestLoggerResponseWriter(base)

	if rw.Written() {
		t.Fatal("writer should not be marked written initially")
	}
	if got := rw.Status(); got != http.StatusOK {
		t.Fatalf("default status=%d", got)
	}

	n, err := rw.Write([]byte("abc"))
	if err != nil {
		t.Fatalf("write err=%v", err)
	}
	if n != 3 {
		t.Fatalf("write bytes=%d", n)
	}
	if got := rw.Status(); got != http.StatusOK {
		t.Fatalf("status after write=%d", got)
	}
	if got := rw.Size(); got != 3 {
		t.Fatalf("size after write=%d", got)
	}
	if !rw.Written() {
		t.Fatal("writer should be marked written after write")
	}

	stringBase := &stringResponseWriter{}
	rwString := newRequestLoggerResponseWriter(stringBase)
	n, err = rwString.WriteString("xy")
	if err != nil {
		t.Fatalf("write string err=%v", err)
	}
	if n != 2 {
		t.Fatalf("write string bytes=%d", n)
	}
	if got := rwString.Size(); got != 2 {
		t.Fatalf("size after write string=%d", got)
	}
	if stringBase.body.String() != "xy" {
		t.Fatalf("unexpected body=%q", stringBase.body.String())
	}

	// Fallback path for writers without io.StringWriter.
	fallback := &basicResponseWriter{}
	rwFallback := newRequestLoggerResponseWriter(fallback)
	n, err = rwFallback.WriteString("z")
	if err != nil {
		t.Fatalf("write string fallback err=%v", err)
	}
	if n != 1 {
		t.Fatalf("write string fallback bytes=%d", n)
	}
	if got := rwFallback.Size(); got != 1 {
		t.Fatalf("size after fallback write string=%d", got)
	}

	headerOnly := newRequestLoggerResponseWriter(&basicResponseWriter{})
	headerOnly.WriteHeader(http.StatusCreated)
	headerOnly.WriteHeader(http.StatusInternalServerError)
	if got := headerOnly.Status(); got != http.StatusCreated {
		t.Fatalf("status should not be overwritten, got=%d", got)
	}
}

func TestRequestLoggerResponseWriterReadFromPaths(t *testing.T) {
	rfBase := &readerFromResponseWriter{}
	rw := newRequestLoggerResponseWriter(rfBase)
	n, err := rw.ReadFrom(strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("readfrom err=%v", err)
	}
	if n != 5 {
		t.Fatalf("readfrom bytes=%d", n)
	}
	if got := rw.Size(); got != 5 {
		t.Fatalf("size after readfrom=%d", got)
	}
	if got := rw.Status(); got != http.StatusOK {
		t.Fatalf("status after readfrom=%d", got)
	}
	if rfBase.body.String() != "hello" {
		t.Fatalf("unexpected body=%q", rfBase.body.String())
	}

	fallback := &basicResponseWriter{}
	rwFallback := newRequestLoggerResponseWriter(fallback)
	n, err = rwFallback.ReadFrom(strings.NewReader("copy"))
	if err != nil {
		t.Fatalf("readfrom fallback err=%v", err)
	}
	if n != 4 {
		t.Fatalf("readfrom fallback bytes=%d", n)
	}
	if got := rwFallback.Size(); got != 4 {
		t.Fatalf("size after fallback readfrom=%d", got)
	}
	if fallback.body.String() != "copy" {
		t.Fatalf("unexpected fallback body=%q", fallback.body.String())
	}
}

func TestRequestLoggerResponseWriterOptionalInterfaces(t *testing.T) {
	flushBase := &flushResponseWriter{}
	rwFlush := newRequestLoggerResponseWriter(flushBase)
	rwFlush.Flush()
	if !flushBase.flushed {
		t.Fatal("flush should be delegated")
	}

	plain := newRequestLoggerResponseWriter(&basicResponseWriter{})
	plain.Flush() // no-op path without http.Flusher

	_, _, err := plain.Hijack()
	if err == nil {
		t.Fatal("hijack should fail when unsupported")
	}

	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	hijackBase := &hijackResponseWriter{
		conn: serverConn,
		rw: bufio.NewReadWriter(
			bufio.NewReader(strings.NewReader("")),
			bufio.NewWriter(io.Discard),
		),
	}
	rwHijack := newRequestLoggerResponseWriter(hijackBase)
	conn, brw, err := rwHijack.Hijack()
	if err != nil {
		t.Fatalf("hijack err=%v", err)
	}
	if conn != serverConn {
		t.Fatal("unexpected hijacked conn")
	}
	if brw == nil {
		t.Fatal("expected readwriter from hijack")
	}

	if err := plain.Push("/asset.js", nil); !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("push err=%v", err)
	}

	pushBase := &pushResponseWriter{}
	rwPush := newRequestLoggerResponseWriter(pushBase)
	opts := &http.PushOptions{Method: http.MethodGet}
	if err := rwPush.Push("/asset.js", opts); err != nil {
		t.Fatalf("push err=%v", err)
	}
	if pushBase.target != "/asset.js" {
		t.Fatalf("push target=%q", pushBase.target)
	}
	if pushBase.opts != opts {
		t.Fatal("push options were not passed through")
	}

	if rwPush.Unwrap() != pushBase {
		t.Fatal("unwrap should return original writer")
	}
}

type basicResponseWriter struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (w *basicResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *basicResponseWriter) Write(p []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *basicResponseWriter) WriteHeader(code int) {
	w.status = code
}

type stringResponseWriter struct {
	basicResponseWriter
}

func (w *stringResponseWriter) WriteString(s string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.WriteString(s)
}

type readerFromResponseWriter struct {
	basicResponseWriter
}

func (w *readerFromResponseWriter) ReadFrom(r io.Reader) (int64, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.ReadFrom(r)
}

type flushResponseWriter struct {
	basicResponseWriter
	flushed bool
}

func (w *flushResponseWriter) Flush() {
	w.flushed = true
}

type hijackResponseWriter struct {
	basicResponseWriter
	conn net.Conn
	rw   *bufio.ReadWriter
	err  error
}

func (w *hijackResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.conn, w.rw, w.err
}

type pushResponseWriter struct {
	basicResponseWriter
	target string
	opts   *http.PushOptions
	err    error
}

func (w *pushResponseWriter) Push(target string, opts *http.PushOptions) error {
	w.target = target
	w.opts = opts
	return w.err
}
