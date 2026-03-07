package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
