package zinc

import (
	"net/http"
	"testing"
)

func TestRouterDynamicRoutesAndHelpers(t *testing.T) {
	app := New()
	mustDo(t, app.Get("/files/*path", func(c *Context) error {
		return c.String(c.Param("*"))
	}))
	mustDo(t, app.Any("/any", func(c *Context) error {
		return c.String(c.Method())
	}))

	wild := performRequest(t, app, http.MethodGet, "/files/a/b/c.txt", nil, nil)
	if wild.Body.String() != "a/b/c.txt" {
		t.Fatalf("wildcard=%q", wild.Body.String())
	}

	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodOptions, http.MethodConnect, http.MethodTrace} {
		resp := performRequest(t, app, method, "/any", nil, nil)
		if method == http.MethodHead {
			if resp.Code != http.StatusOK {
				t.Fatalf("status=%d", resp.Code)
			}
			continue
		}
		if resp.Body.String() != method {
			t.Fatalf("method=%s body=%q", method, resp.Body.String())
		}
	}
}

func TestRouterConflictsAndNormalization(t *testing.T) {
	router := &Router{config: &DefaultConfig}
	mustDo(t, router.Add(MethodGet, "users/:id", func(c *Context) error { return nil }))
	if err := router.Add(MethodGet, "/users/:id", func(c *Context) error { return nil }); err == nil {
		t.Fatal("expected duplicate route error")
	}
	if got := router.normalizePath("users"); got != "/users" {
		t.Fatalf("normalized=%q", got)
	}
}

func TestGroupHelpers(t *testing.T) {
	app := New()
	api := app.Group("/api", func(c *Context) error {
		c.Set("group", true)
		return c.Next()
	})
	v1 := api.Route("/v1", nil)
	mustDo(t, v1.Get("/ping", func(c *Context) error {
		group, _ := c.Get("group")
		if group != true {
			t.Fatal("group middleware missing")
		}
		return c.String("pong")
	}))

	resp := performRequest(t, app, http.MethodGet, "/api/v1/ping", nil, nil)
	if resp.Body.String() != "pong" {
		t.Fatalf("body=%q", resp.Body.String())
	}
}
