package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0mjs/zinc"
)

func TestGzipCompressesAcceptedResponse(t *testing.T) {
	app := zinc.New()
	app.Use(Gzip())
	mustNoErrGzip(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("hello zinc")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "br, gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "gzip" {
		t.Fatalf("content-encoding=%q", got)
	}
	if got := rec.Header().Get(zinc.HeaderVary); got != zinc.HeaderAcceptEncoding {
		t.Fatalf("vary=%q", got)
	}
	if body := gunzipResponse(t, rec.Body.Bytes()); body != "hello zinc" {
		t.Fatalf("body=%q", body)
	}
}

func TestGzipSkipsWhenNotAccepted(t *testing.T) {
	app := zinc.New()
	app.Use(Gzip())
	mustNoErrGzip(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("plain")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "" {
		t.Fatalf("content-encoding=%q", got)
	}
	if rec.Body.String() != "plain" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestGzipSkipsNoBodyResponses(t *testing.T) {
	app := zinc.New()
	app.Use(Gzip())
	mustNoErrGzip(t, app.Get("/", func(c *zinc.Context) error {
		return c.Status(http.StatusNoContent).NoContent()
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "" {
		t.Fatalf("content-encoding=%q", got)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestGzipSkipsExistingContentEncoding(t *testing.T) {
	app := zinc.New()
	app.Use(Gzip())
	mustNoErrGzip(t, app.Get("/", func(c *zinc.Context) error {
		c.SetHeader(zinc.HeaderContentEncoding, "br")
		return c.String("encoded elsewhere")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "br" {
		t.Fatalf("content-encoding=%q", got)
	}
	if rec.Body.String() != "encoded elsewhere" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestGzipMinLength(t *testing.T) {
	app := zinc.New()
	app.Use(GzipWithConfig(GzipConfig{MinLength: 8}))
	mustNoErrGzip(t, app.Get("/small", func(c *zinc.Context) error {
		return c.String("small")
	}))
	mustNoErrGzip(t, app.Get("/large", func(c *zinc.Context) error {
		if _, err := c.Writer().Write([]byte("hello")); err != nil {
			return err
		}
		_, err := c.Writer().Write([]byte(" zinc"))
		return err
	}))

	req := httptest.NewRequest(http.MethodGet, "/small", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "" {
		t.Fatalf("small content-encoding=%q", got)
	}
	if rec.Body.String() != "small" {
		t.Fatalf("small body=%q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/large", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec = httptest.NewRecorder()
	app.ServeHTTP(rec, req)
	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "gzip" {
		t.Fatalf("large content-encoding=%q", got)
	}
	if body := gunzipResponse(t, rec.Body.Bytes()); body != "hello zinc" {
		t.Fatalf("large body=%q", body)
	}
}

func TestGzipHandlesErrorResponse(t *testing.T) {
	app := zinc.New()
	app.Use(Gzip())
	mustNoErrGzip(t, app.Get("/", func(*zinc.Context) error {
		return zinc.ErrForbidden
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%q", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "gzip" {
		t.Fatalf("content-encoding=%q", got)
	}
	if body := gunzipResponse(t, rec.Body.Bytes()); body != "Forbidden" {
		t.Fatalf("body=%q", body)
	}
}

func TestGzipSkipper(t *testing.T) {
	app := zinc.New()
	app.Use(GzipWithConfig(GzipConfig{
		Skipper: func(*zinc.Context) bool { return true },
	}))
	mustNoErrGzip(t, app.Get("/", func(c *zinc.Context) error {
		return c.String("plain")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(zinc.HeaderAcceptEncoding, "gzip")
	rec := httptest.NewRecorder()
	app.ServeHTTP(rec, req)

	if got := rec.Header().Get(zinc.HeaderContentEncoding); got != "" {
		t.Fatalf("content-encoding=%q", got)
	}
	if rec.Body.String() != "plain" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func gunzipResponse(t *testing.T, body []byte) string {
	t.Helper()

	reader, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new gzip reader: %v", err)
	}
	defer reader.Close()

	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read gzip response: %v", err)
	}
	return string(out)
}

func mustNoErrGzip(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
