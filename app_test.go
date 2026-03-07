package zinc

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppRoutingMiddlewareAndFallbacks(t *testing.T) {
	t.Run("middleware cache and params", func(t *testing.T) {
		app := New()
		calls := 0
		app.Use(func(c *Context) error {
			calls++
			return c.Next()
		})
		mustDo(t, app.Get("/users/:id", func(c *Context) error {
			return c.String(c.Param("id"))
		}))

		for i := 0; i < 3; i++ {
			resp := performRequest(t, app, http.MethodGet, "/users/42", nil, nil)
			if resp.Code != http.StatusOK {
				t.Fatalf("status=%d", resp.Code)
			}
			if body := resp.Body.String(); body != "42" {
				t.Fatalf("body=%q", body)
			}
		}
		if calls != 3 {
			t.Fatalf("middleware calls=%d want 3", calls)
		}
	})

	t.Run("middleware post-next sees route metadata", func(t *testing.T) {
		app := New()
		fullPath := ""
		app.Use(func(c *Context) error {
			if err := c.Next(); err != nil {
				return err
			}
			fullPath = c.FullPath()
			return nil
		})
		mustDo(t, app.Get("/post-next", func(c *Context) error {
			return c.String("ok")
		}))

		resp := performRequest(t, app, http.MethodGet, "/post-next", nil, nil)
		if resp.Code != http.StatusOK {
			t.Fatalf("status=%d", resp.Code)
		}
		if fullPath != "/post-next" {
			t.Fatalf("fullPath=%q", fullPath)
		}
	})

	t.Run("prefix middleware and route metadata", func(t *testing.T) {
		app := New()
		app.UsePrefix("/api", func(c *Context) error {
			c.Set("prefix", true)
			return c.Next()
		})
		app.Route("/api", func(api *Group) {
			api.Use(func(c *Context) error {
				c.Set("group", "users")
				return c.Next()
			})
			mustDo(t, api.Get("/users/:id", func(c *Context) error {
				prefix, _ := c.Get("prefix")
				group, _ := c.Get("group")
				return c.JSON(Map{
					"prefix":    prefix,
					"group":     group,
					"full_path": c.FullPath(),
					"method":    c.Route().Method,
					"path":      c.Route().Path,
					"id":        c.Param("id"),
				})
			}))
		})

		resp := performRequest(t, app, http.MethodGet, "/api/users/99", nil, nil)
		body := resp.Body.String()
		for _, fragment := range []string{"\"prefix\":true", "\"group\":\"users\"", "\"full_path\":\"/api/users/:id\"", "\"method\":\"GET\"", "\"id\":\"99\""} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("body missing %q: %s", fragment, body)
			}
		}
	})

	t.Run("middleware stop semantics", func(t *testing.T) {
		app := New()
		called := false
		app.Use(func(c *Context) error {
			return nil
		})
		mustDo(t, app.Get("/blocked", func(c *Context) error {
			called = true
			return c.String("nope")
		}))

		resp := performRequest(t, app, http.MethodGet, "/blocked", nil, nil)
		if called {
			t.Fatal("handler should not have been called")
		}
		if resp.Code != http.StatusOK || resp.Body.Len() != 0 {
			t.Fatalf("unexpected response: code=%d body=%q", resp.Code, resp.Body.String())
		}
	})

	t.Run("not found and method not allowed", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/items", func(c *Context) error { return c.String("ok") }))
		app.NotFound(func(c *Context) error { return c.String("missing") })
		app.MethodNotAllowed(func(c *Context) error { return c.String("wrong method") })

		notFound := performRequest(t, app, http.MethodGet, "/missing", nil, nil)
		if notFound.Code != http.StatusNotFound || notFound.Body.String() != "missing" {
			t.Fatalf("not found response = %d %q", notFound.Code, notFound.Body.String())
		}

		methodNA := performRequest(t, app, http.MethodPost, "/items", nil, nil)
		if methodNA.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status=%d", methodNA.Code)
		}
		allow := methodNA.Header().Get(HeaderAllow)
		for _, method := range []string{MethodGet, MethodHead, MethodOptions} {
			if !strings.Contains(allow, method) {
				t.Fatalf("allow header missing %s: %q", method, allow)
			}
		}
		if methodNA.Body.String() != "wrong method" {
			t.Fatalf("body=%q", methodNA.Body.String())
		}
	})

	t.Run("auto head and auto options", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/health", func(c *Context) error {
			return c.String("ok")
		}))

		head := performRequest(t, app, http.MethodHead, "/health", nil, nil)
		if head.Code != http.StatusOK || head.Body.Len() != 0 {
			t.Fatalf("head response = %d %q", head.Code, head.Body.String())
		}

		options := performRequest(t, app, http.MethodOptions, "/health", nil, nil)
		if options.Code != http.StatusNoContent {
			t.Fatalf("status=%d", options.Code)
		}
		if allow := options.Header().Get(HeaderAllow); !strings.Contains(allow, MethodGet) {
			t.Fatalf("allow=%q", allow)
		}
	})

	t.Run("mount strips prefix", func(t *testing.T) {
		app := New()
		mux := http.NewServeMux()
		mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
			_, _ = io.WriteString(w, r.URL.Path)
		})
		app.Mount("/sub", mux)

		resp := performRequest(t, app, http.MethodGet, "/sub/hello", nil, nil)
		if resp.Body.String() != "/hello" {
			t.Fatalf("body=%q", resp.Body.String())
		}

		routes := app.Routes()
		foundMount := false
		for _, route := range routes {
			if route.Method == methodUse && route.Path == "/sub" {
				foundMount = true
			}
		}
		if !foundMount {
			t.Fatal("mount route not present in Routes()")
		}
	})
}

func TestAppConfigLifecycleAndErrors(t *testing.T) {
	t.Run("new with config and handler", func(t *testing.T) {
		app := NewWithConfig(Config{ServerHeader: "zinc-test", CaseSensitive: true})
		if got := app.Handler(); got != app {
			t.Fatal("Handler should return app")
		}
		if app.config.ServerHeader != "zinc-test" {
			t.Fatalf("server header=%q", app.config.ServerHeader)
		}
		if !app.config.CaseSensitive {
			t.Fatal("CaseSensitive should be true")
		}
	})

	t.Run("case sensitive and strict routing", func(t *testing.T) {
		app := NewWithConfig(Config{CaseSensitive: true, StrictRouting: true})
		mustDo(t, app.Get("/Hello", func(c *Context) error { return c.String("ok") }))

		lower := performRequest(t, app, http.MethodGet, "/hello", nil, nil)
		if lower.Code != http.StatusNotFound {
			t.Fatalf("status=%d", lower.Code)
		}

		slash := performRequest(t, app, http.MethodGet, "/Hello/", nil, nil)
		if slash.Code != http.StatusNotFound {
			t.Fatalf("status=%d", slash.Code)
		}
	})

	t.Run("custom error handler", func(t *testing.T) {
		app := NewWithConfig(Config{ErrorHandler: func(c *Context, err error) {
			_ = c.Status(http.StatusTeapot).String("handled")
		}})
		mustDo(t, app.Get("/boom", func(c *Context) error {
			return errors.New("boom")
		}))

		resp := performRequest(t, app, http.MethodGet, "/boom", nil, nil)
		if resp.Code != http.StatusTeapot || resp.Body.String() != "handled" {
			t.Fatalf("response=%d %q", resp.Code, resp.Body.String())
		}
	})

	t.Run("server header", func(t *testing.T) {
		app := NewWithConfig(Config{ServerHeader: "zinc/edge"})
		mustDo(t, app.Get("/", func(c *Context) error { return c.String("ok") }))
		resp := performRequest(t, app, http.MethodGet, "/", nil, nil)
		if got := resp.Header().Get(HeaderServer); got != "zinc/edge" {
			t.Fatalf("server header=%q", got)
		}
	})

	t.Run("serve and shutdown", func(t *testing.T) {
		app := New()
		mustDo(t, app.Get("/ping", func(c *Context) error { return c.String("pong") }))

		ln, err := net.Listen("tcp", "127.0.0.1:0")
		mustDo(t, err)
		defer ln.Close()

		errCh := make(chan error, 1)
		go func() { errCh <- app.Serve(ln) }()

		url := "http://" + ln.Addr().String() + "/ping"
		var resp *http.Response
		for i := 0; i < 20; i++ {
			resp, err = http.Get(url)
			if err == nil {
				break
			}
			time.Sleep(25 * time.Millisecond)
		}
		mustDo(t, err)
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		mustDo(t, err)
		if string(body) != "pong" {
			t.Fatalf("body=%q", string(body))
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		mustDo(t, app.Shutdown(ctx))
		select {
		case err := <-errCh:
			mustDo(t, err)
		case <-time.After(time.Second):
			t.Fatal("server did not shut down")
		}
	})

	t.Run("invalid lifecycle inputs", func(t *testing.T) {
		app := New()
		if err := app.Serve(nil); err == nil {
			t.Fatal("Serve(nil) should fail")
		}
		if err := app.Listen("bad-addr"); err == nil {
			t.Fatal("Listen should fail for invalid address")
		}
		if err := app.ListenTLS("bad-addr", "missing.crt", "missing.key"); err == nil {
			t.Fatal("ListenTLS should fail for invalid address")
		}
		if err := app.Shutdown(context.Background()); err != nil {
			t.Fatalf("Shutdown without server: %v", err)
		}
	})
}

func TestWrapAndWrapFunc(t *testing.T) {
	app := New()
	mustDo(t, app.Get("/h", Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "handler")
	}))))
	mustDo(t, app.Get("/f", WrapFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "func")
	})))

	resp1 := performRequest(t, app, http.MethodGet, "/h", nil, nil)
	if resp1.Body.String() != "handler" {
		t.Fatalf("body=%q", resp1.Body.String())
	}
	resp2 := performRequest(t, app, http.MethodGet, "/f", nil, nil)
	if resp2.Body.String() != "func" {
		t.Fatalf("body=%q", resp2.Body.String())
	}
}

func TestStaticAndFileRoutes(t *testing.T) {
	dir := t.TempDir()
	mustDo(t, os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("world"), 0o644))

	app := New()
	mustDo(t, app.Static("/assets", dir))
	mustDo(t, app.File("/single", filepath.Join(dir, "hello.txt")))

	assets := performRequest(t, app, http.MethodGet, "/assets/hello.txt", nil, nil)
	if assets.Body.String() != "world" {
		t.Fatalf("body=%q", assets.Body.String())
	}

	single := performRequest(t, app, http.MethodGet, "/single", nil, nil)
	if single.Body.String() != "world" {
		t.Fatalf("body=%q", single.Body.String())
	}
}
