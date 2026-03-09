package zinc

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
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

func TestNormalizeGetHandlersBranches(t *testing.T) {
	handlers, err := normalizeGetHandlers()
	mustDo(t, err)
	if handlers != nil {
		t.Fatalf("handlers=%v", handlers)
	}

	var nilHandler HandlerFunc
	if _, err := normalizeGetHandlers(nilHandler); err == nil || !strings.Contains(err.Error(), "handler at index 0 is nil") {
		t.Fatalf("err=%v", err)
	}

	var nilFunc func(*Context) error
	if _, err := normalizeGetHandlers(nilFunc); err == nil || !strings.Contains(err.Error(), "handler at index 0 is nil") {
		t.Fatalf("err=%v", err)
	}

	if _, err := normalizeGetHandlers(nil); err == nil || !strings.Contains(err.Error(), "handler at index 0 is nil") {
		t.Fatalf("err=%v", err)
	}

	if _, err := normalizeGetHandlers(123); err == nil || !strings.Contains(err.Error(), "unsupported GET handler type") {
		t.Fatalf("err=%v", err)
	}

	okHandlers, err := normalizeGetHandlers(
		"hello",
		HandlerFunc(func(*Context) error { return nil }),
		func(*Context) error { return nil },
	)
	mustDo(t, err)
	if len(okHandlers) != 3 {
		t.Fatalf("handlers len=%d", len(okHandlers))
	}

	ctx, rec := newRecorderContext(t, httptest.NewRequest(http.MethodGet, "/", nil))
	defer ctx.release()
	mustDo(t, okHandlers[0](ctx))
	if rec.Body.String() != "hello" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestLowercasePathAndHandlerNameBranches(t *testing.T) {
	if got, changed := lowercasePath("/abc"); got != "/abc" || changed {
		t.Fatalf("got=%q changed=%v", got, changed)
	}
	if got, changed := lowercasePath("/ABC"); got != "/abc" || !changed {
		t.Fatalf("got=%q changed=%v", got, changed)
	}
	if got, changed := lowercasePath("/ÄBC"); got != "/äbc" || !changed {
		t.Fatalf("got=%q changed=%v", got, changed)
	}

	if name := handlerName(nil); name != "" {
		t.Fatalf("name=%q", name)
	}
	handler := HandlerFunc(func(*Context) error { return nil })
	name := handlerName(handler)
	if name == "" {
		t.Fatal("expected handler name")
	}
	if again := handlerName(handler); again != name {
		t.Fatalf("cached name=%q first=%q", again, name)
	}
}

func TestRouteCacheSetUpdateAndEviction(t *testing.T) {
	cache := NewRouteCache(1)

	key1 := routeCacheKey{method: MethodGet, path: "/one"}
	key2 := routeCacheKey{method: MethodGet, path: "/two"}
	route1 := &radixRoute{infoIndex: 1}
	route1Updated := &radixRoute{infoIndex: 2}
	route2 := &radixRoute{infoIndex: 3}
	entry1 := routeCacheEntry{route: route1}
	entry1Updated := routeCacheEntry{route: route1Updated}
	entry2 := routeCacheEntry{route: route2}

	cache.set(key1, entry1)
	cache.set(key1, entry1Updated) // update existing key branch
	if got, ok := cache.get(key1); !ok || got.route != route1Updated {
		t.Fatalf("updated entry=%v ok=%v", got, ok)
	}

	cache.set(key2, entry2) // eviction branch for size=1
	if _, ok := cache.get(key1); ok {
		t.Fatal("expected key1 to be evicted")
	}
	if got, ok := cache.get(key2); !ok || got.route != route2 {
		t.Fatalf("entry2=%v ok=%v", got, ok)
	}
}

func TestRadixNodeBranches(t *testing.T) {
	root := &radixNode{kind: radixRoot}
	mustDo(t, root.add("/foo", &radixRoute{}))
	mustDo(t, root.add("/fob", &radixRoute{})) // triggers static-node split branch

	if err := root.add("/foo", &radixRoute{}); err == nil {
		t.Fatal("expected duplicate route error")
	}
	if err := root.add("/files/*path/more", &radixRoute{}); err == nil || !strings.Contains(err.Error(), "wildcard must be final") {
		t.Fatalf("err=%v", err)
	}

	paramHolder := &radixNode{}
	if first, second := paramHolder.addParamChild(), paramHolder.addParamChild(); first != second {
		t.Fatal("param child should be reused")
	}
	catchAllHolder := &radixNode{}
	if first, second := catchAllHolder.addCatchAllChild(), catchAllHolder.addCatchAllChild(); first != second {
		t.Fatal("catch-all child should be reused")
	}

	static := &radixNode{kind: radixStatic, prefix: "abc"}
	if matched := static.lookup("ab", 0, &[8]paramRange{}, 0); matched != nil {
		t.Fatalf("matched=%v", matched)
	}

	param := &radixNode{kind: radixParam}
	if matched := param.lookup("/x", 0, &[8]paramRange{}, 0); matched != nil {
		t.Fatalf("matched=%v", matched)
	}

	mismatch := &radixNode{route: &radixRoute{paramCount: 1}}
	if matched := mismatch.matchRoute(0); matched != nil {
		t.Fatalf("matched=%v", matched)
	}
}

func TestRouterFindIntoAndDynamicCacheBranches(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(2),
		config: &DefaultConfig,
	}
	mustDo(t, router.Add(MethodGet, "/Case", func(*Context) error { return nil }))
	mustDo(t, router.Add(MethodGet, "/trim", func(*Context) error { return nil }))

	ctxCase := &Context{}
	if handler := router.findInto(MethodGet, "/CASE", ctxCase); handler == nil {
		t.Fatal("expected case-insensitive match")
	}
	if ctxCase.FullPath() != "/Case" {
		t.Fatalf("full path=%q", ctxCase.FullPath())
	}

	ctxTrim := &Context{}
	if handler := router.findInto(MethodGet, "/trim/", ctxTrim); handler == nil {
		t.Fatal("expected non-strict trailing slash match")
	}

	dynamic := &Router{
		cache: NewRouteCache(2),
		trees: map[string]*radixNode{
			MethodGet: {kind: radixRoot},
		},
	}
	cacheKey := routeCacheKey{method: MethodGet, path: "/cached"}
	dynamic.cache.set(cacheKey, routeCacheEntry{route: nil})
	if handler := dynamic.findDynamicInto(MethodGet, "/cached", &Context{}); handler != nil {
		t.Fatalf("handler=%v", handler)
	}
	if handler := dynamic.findDynamicInto(MethodGet, "/missing", &Context{}); handler != nil {
		t.Fatalf("handler=%v", handler)
	}
}

func TestRouterAddAndMatchErrorBranches(t *testing.T) {
	router := &Router{config: &DefaultConfig}
	if err := router.Add(MethodGet, "/users"); err == nil || !strings.Contains(err.Error(), "no handler provided") {
		t.Fatalf("err=%v", err)
	}

	app := New()
	if err := app.Match([]string{MethodGet}, "/users"); err == nil {
		t.Fatal("expected Match to fail when handlers are missing")
	}
}

func TestRouterAllowedMethodsSharedPathIndex(t *testing.T) {
	router := &Router{config: &DefaultConfig}
	mustDo(t, router.Add(MethodGet, "/shared/one", func(*Context) error { return nil }))
	mustDo(t, router.Add(MethodPost, "/shared/two", func(*Context) error { return nil }))
	mustDo(t, router.Add(MethodGet, "/users/:id", func(*Context) error { return nil }))
	mustDo(t, router.Add(MethodPost, "/users/me", func(*Context) error { return nil }))

	shared := router.allowedMethods("/shared/one", true, true)
	if want := []string{MethodGet, MethodHead, MethodOptions}; !reflect.DeepEqual(shared, want) {
		t.Fatalf("shared allow = %v want %v", shared, want)
	}

	overlap := router.allowedMethods("/users/me", true, true)
	if want := []string{MethodGet, MethodHead, MethodPost, MethodOptions}; !reflect.DeepEqual(overlap, want) {
		t.Fatalf("overlap allow = %v want %v", overlap, want)
	}

	if header := router.allowedMethodHeader("/users/me", true, true); header != "GET, HEAD, POST, OPTIONS" {
		t.Fatalf("allow header = %q", header)
	}
}
