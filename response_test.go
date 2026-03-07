package zinc

import (
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

type rendererStub struct{}

type xmlPayload struct {
	XMLName xml.Name `xml:"response"`
	Value   string   `xml:"value"`
}

func (rendererStub) Render(w io.Writer, name string, data any, c *Context) error {
	_, err := io.WriteString(w, name+":ok")
	return err
}

func TestResponseHelpers(t *testing.T) {
	t.Run("headers cookies and content helpers", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/headers", func(c *Context) error {
			c.SetHeader("X-One", "1").AppendHeader("X-Many", "a", "b").Type("json").Location("/next").Vary("Origin", "Accept")
			c.SetCookie(&http.Cookie{Name: "session", Value: "abc", Path: "/"})
			c.ClearCookie("stale")
			return c.String("ok")
		}))
		resp := performRequest(t, app, http.MethodGet, "/headers", nil, nil)
		if resp.Header().Get("X-One") != "1" || resp.Header().Get(HeaderLocation) != "/next" {
			t.Fatal("header helpers failed")
		}
		if got := resp.Header().Values("X-Many"); len(got) != 2 {
			t.Fatalf("append header=%v", got)
		}
		if ct := resp.Header().Get(HeaderContentType); ct != "application/json" {
			t.Fatalf("content type=%q", ct)
		}
		if vary := resp.Header().Values(HeaderVary); len(vary) != 2 {
			t.Fatalf("vary=%v", vary)
		}
		cookies := resp.Result().Cookies()
		if len(cookies) < 2 {
			t.Fatalf("cookies=%v", cookies)
		}
	})

	t.Run("data encoders and streams", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/data", func(c *Context) error { return c.Data("application/custom", []byte("data")) }))
		mustDo(t, app.Get("/json", func(c *Context) error { return c.JSONPretty(Map{"ok": true}, "  ") }))
		mustDo(t, app.Get("/xml", func(c *Context) error { return c.XML(xmlPayload{Value: "x"}) }))
		mustDo(t, app.Get("/html", func(c *Context) error { return c.HTML("<p>x</p>") }))
		mustDo(t, app.Get("/stream", func(c *Context) error { return c.Stream("text/plain", strings.NewReader("stream")) }))
		mustDo(t, app.Get("/send-nil", func(c *Context) error { return c.Send(nil) }))
		mustDo(t, app.Get("/send-bytes", func(c *Context) error { return c.Send([]byte("bytes")) }))

		cases := map[string]string{"/data": "data", "/xml": "<value>x</value>", "/html": "<p>x</p>", "/stream": "stream", "/send-bytes": "bytes"}
		for path, expected := range cases {
			resp := performRequest(t, app, http.MethodGet, path, nil, nil)
			if !strings.Contains(resp.Body.String(), expected) {
				t.Fatalf("%s body=%q", path, resp.Body.String())
			}
		}
		jsonResp := performRequest(t, app, http.MethodGet, "/json", nil, nil)
		if !strings.Contains(jsonResp.Body.String(), "\n") {
			t.Fatalf("json pretty=%q", jsonResp.Body.String())
		}
		nilResp := performRequest(t, app, http.MethodGet, "/send-nil", nil, nil)
		if nilResp.Body.String() != "null" {
			t.Fatalf("nil body=%q", nilResp.Body.String())
		}
	})

	t.Run("no content redirect render and repeated write", func(t *testing.T) {
		app := NewWithConfig(Config{Renderer: rendererStub{}})
		mustDo(t, app.Get("/nocontent", func(c *Context) error { return c.NoContent() }))
		mustDo(t, app.Get("/redirect", func(c *Context) error { return c.Redirect(http.StatusMovedPermanently, "/to") }))
		mustDo(t, app.Get("/render", func(c *Context) error { return c.Render("home", nil) }))
		mustDo(t, app.Get("/double", func(c *Context) error {
			mustDo(t, c.String("once"))
			return c.String("twice")
		}))

		noContent := performRequest(t, app, http.MethodGet, "/nocontent", nil, nil)
		if noContent.Code != http.StatusNoContent || noContent.Body.Len() != 0 {
			t.Fatalf("nocontent=%d %q", noContent.Code, noContent.Body.String())
		}
		redirect := performRequest(t, app, http.MethodGet, "/redirect", nil, nil)
		if redirect.Code != http.StatusMovedPermanently || redirect.Header().Get(HeaderLocation) != "/to" {
			t.Fatalf("redirect=%d %q", redirect.Code, redirect.Header().Get(HeaderLocation))
		}
		render := performRequest(t, app, http.MethodGet, "/render", nil, nil)
		if render.Body.String() != "home:ok" {
			t.Fatalf("render=%q", render.Body.String())
		}
		double := performRequest(t, app, http.MethodGet, "/double", nil, nil)
		if double.Code != http.StatusOK || double.Body.String() != "once" {
			t.Fatalf("double=%d %q", double.Code, double.Body.String())
		}
	})

	t.Run("file helpers", func(t *testing.T) {
		filesystem := fstest.MapFS{"hello.txt": &fstest.MapFile{Data: []byte("world")}}
		app := New()
		mustDo(t, app.Get("/filefs", func(c *Context) error { return c.FileFS("hello.txt", filesystem) }))
		mustDo(t, app.Get("/attach", func(c *Context) error { return c.Attachment("testdata.txt", "download.txt") }))

		fileResp := performRequest(t, app, http.MethodGet, "/filefs", nil, nil)
		if fileResp.Body.String() != "world" {
			t.Fatalf("filefs body=%q", fileResp.Body.String())
		}

		dir := t.TempDir()
		path := filepath.Join(dir, "testdata.txt")
		mustDo(t, os.WriteFile(path, []byte("attachment"), 0o644))
		mustDo(t, app.Get("/file", func(c *Context) error { return c.File(path) }))
		mustDo(t, app.Get("/download", func(c *Context) error { return c.Download(path) }))
		attachApp := New()
		mustDo(t, attachApp.Get("/attach", func(c *Context) error { return c.Attachment(path, "download.txt") }))

		file := performRequest(t, app, http.MethodGet, "/file", nil, nil)
		if file.Body.String() != "attachment" {
			t.Fatalf("file body=%q", file.Body.String())
		}
		download := performRequest(t, app, http.MethodGet, "/download", nil, nil)
		if disposition := download.Header().Get(HeaderContentDisposition); disposition == "" {
			t.Fatal("download disposition missing")
		}
		attach := performRequest(t, attachApp, http.MethodGet, "/attach", nil, nil)
		if disposition := attach.Header().Get(HeaderContentDisposition); !strings.Contains(disposition, "download.txt") {
			t.Fatalf("attachment disposition=%q", disposition)
		}
	})

	t.Run("head semantics", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/head", func(c *Context) error { return c.JSON(Map{"ok": true}) }))
		resp := performRequest(t, app, http.MethodHead, "/head", nil, nil)
		if resp.Body.Len() != 0 {
			t.Fatalf("head body=%q", resp.Body.String())
		}
	})
}

func TestDirectResponseErrors(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, _ := newRecorderContext(t, req)
	defer ctx.release()
	mustDo(t, ctx.String("once"))
	if err := ctx.String("twice"); !errors.Is(err, ErrResponseAlreadySent) {
		t.Fatalf("err=%v", err)
	}
}
