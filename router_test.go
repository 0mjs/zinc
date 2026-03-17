package zinc

import (
	"fmt"
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

func TestRouteCacheInvalidateClearsLazilyOnNextAccess(t *testing.T) {
	cache := NewRouteCache(2)
	key := routeCacheKey{method: MethodGet, path: "/one"}
	entry := routeCacheEntry{route: &radixRoute{infoIndex: 1}}

	cache.set(key, entry)
	cache.invalidate()

	if _, ok := cache.get(key); ok {
		t.Fatal("expected first cache access after invalidation to clear stale entries")
	}

	cache.set(key, entry)
	if got, ok := cache.get(key); !ok || got.route == nil || got.route.infoIndex != 1 {
		t.Fatalf("entry=%+v ok=%v", got, ok)
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
	if matched := static.lookup("ab", 0, &paramRanges{}, 0); matched != nil {
		t.Fatalf("matched=%v", matched)
	}

	param := &radixNode{kind: radixParam}
	if matched := param.lookup("/x", 0, &paramRanges{}, 0); matched != nil {
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

func TestRouterDispatchIntoCachesMissResults(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(routeCacheMinRoutes + 8),
		config: &DefaultConfig,
	}
	for i := 0; i < routeCacheMinRoutes; i++ {
		path := fmt.Sprintf("/bulk/%d", i)
		mustDo(t, router.Add(MethodGet, path, func(*Context) error { return nil }))
	}
	mustDo(t, router.Add(MethodPost, "/only-post", func(*Context) error { return nil }))

	handled, allowed, err := router.dispatchInto(MethodPut, "/only-post", true, &Context{})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if handled {
		t.Fatal("expected method mismatch to miss the handler")
	}
	if allowed.mask != methodMaskPost || len(allowed.extra) != 0 {
		t.Fatalf("allowed=%v", allowed)
	}

	key := routeCacheKey{method: MethodPut, path: "/only-post"}
	entry, ok := router.cache.get(key)
	if !ok {
		t.Fatal("expected dispatch miss to be cached")
	}
	if entry.route != nil {
		t.Fatalf("cached route=%v", entry.route)
	}
	if entry.allowed.mask != methodMaskPost || len(entry.allowed.extra) != 0 {
		t.Fatalf("cached allowed=%v", entry.allowed)
	}

	mustDo(t, router.Add(MethodPut, "/only-post", func(*Context) error { return nil }))

	handled, allowed, err = router.dispatchInto(MethodPut, "/only-post", true, &Context{})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected registered method to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}

	entry, ok = router.cache.get(key)
	if ok {
		t.Fatalf("expected stale miss cache entry to clear on first access, got=%+v", entry)
	}
}

func TestRouterSupportsMoreThanInlinePathParams(t *testing.T) {
	pattern, path, names, _ := buildSequentialParamRoute(10)

	app := New()
	mustDo(t, app.Get(pattern, func(c *Context) error {
		return c.String(c.Param(names[len(names)-1]))
	}))

	resp := performRequest(t, app, http.MethodGet, path, nil, nil)
	if resp.Body.String() != "10" {
		t.Fatalf("body=%q", resp.Body.String())
	}
}

func TestRouterDispatchIntoCachesManyParams(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(routeCacheMinRoutes + 8),
		config: &DefaultConfig,
	}
	for i := 0; i < routeCacheMinRoutes; i++ {
		path := fmt.Sprintf("/bulk/%d", i)
		mustDo(t, router.Add(MethodGet, path, func(*Context) error { return nil }))
	}

	pattern, path, names, _ := buildSequentialParamRoute(10)
	mustDo(t, router.Add(MethodGet, pattern, func(*Context) error { return nil }))

	ctx := &Context{}
	handled, allowed, err := router.dispatchInto(MethodGet, path, false, ctx)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected dynamic route to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}
	if got := ctx.Param(names[8]); got != "9" {
		t.Fatalf("param9=%q", got)
	}
	if got := ctx.Param(names[9]); got != "10" {
		t.Fatalf("param10=%q", got)
	}
	if len(ctx.PathParams) < len(names) || ctx.PathParams[9].key != names[9] {
		t.Fatalf("path params not expanded: len=%d last=%+v", len(ctx.PathParams), ctx.PathParams[9])
	}

	key := routeCacheKey{method: MethodGet, path: path}
	if entry, ok := router.cache.get(key); !ok || entry.route == nil {
		t.Fatalf("expected cached hit entry, got=%+v ok=%v", entry, ok)
	}

	ctxCached := &Context{}
	handled, allowed, err = router.dispatchInto(MethodGet, path, false, ctxCached)
	if err != nil {
		t.Fatalf("cached err=%v", err)
	}
	if !handled {
		t.Fatal("expected cached dispatch to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("cached allowed=%v", allowed)
	}
	if got := ctxCached.Param(names[9]); got != "10" {
		t.Fatalf("cached param10=%q", got)
	}
}

func TestRouterDispatchIntoCachesSmallDynamicRouteSets(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(8),
		config: &DefaultConfig,
	}
	mustDo(t, router.Add(MethodGet, "/items/:id", func(*Context) error { return nil }))

	ctx := &Context{}
	handled, allowed, err := router.dispatchInto(MethodGet, "/items/42", false, ctx)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected dynamic route to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}

	key := routeCacheKey{method: MethodGet, path: "/items/42"}
	entry, ok := router.cache.get(key)
	if !ok {
		t.Fatal("expected small dynamic route set to populate the dispatch cache")
	}
	if entry.route == nil {
		t.Fatalf("cached entry=%+v", entry)
	}
}

func TestRouterDispatchIntoRefreshesCachedDynamicHitAfterAdd(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(8),
		config: &DefaultConfig,
	}
	mustDo(t, router.Add(MethodGet, "/items/:id", func(*Context) error { return nil }))

	ctxParam := &Context{}
	handled, allowed, err := router.dispatchInto(MethodGet, "/items/new", false, ctxParam)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected param route to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}
	if ctxParam.routeIndex != 0 {
		t.Fatalf("route index=%d", ctxParam.routeIndex)
	}

	key := routeCacheKey{method: MethodGet, path: "/items/new"}
	entry, ok := router.cache.get(key)
	if !ok || entry.route == nil || entry.route.infoIndex != 0 {
		t.Fatalf("cached entry=%+v ok=%v", entry, ok)
	}

	mustDo(t, router.Add(MethodGet, "/items/new", func(*Context) error { return nil }))
	if entry, ok := router.cache.get(key); ok {
		t.Fatalf("expected stale cached hit to clear on first access after Add, got=%+v", entry)
	}

	ctxStatic := &Context{}
	handled, allowed, err = router.dispatchInto(MethodGet, "/items/new", false, ctxStatic)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected static route to handle the request")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}
	if ctxStatic.routeIndex != 1 {
		t.Fatalf("route index=%d", ctxStatic.routeIndex)
	}
}

func TestRouterDispatchIntoCachedManyParamsIsolation(t *testing.T) {
	router := &Router{
		cache:  NewRouteCache(routeCacheMinRoutes + 8),
		config: &DefaultConfig,
	}
	for i := 0; i < routeCacheMinRoutes; i++ {
		path := fmt.Sprintf("/bulk/%d", i)
		mustDo(t, router.Add(MethodGet, path, func(*Context) error { return nil }))
	}

	pattern, pathShort, names, _ := buildSequentialParamRoute(10)
	var pathLong strings.Builder
	for i := 0; i < 10; i++ {
		pathLong.WriteByte('/')
		pathLong.WriteString(fmt.Sprintf("%d", 1000+i))
	}

	mustDo(t, router.Add(MethodGet, pattern, func(*Context) error { return nil }))

	shortCtx := &Context{}
	handled, _, err := router.dispatchInto(MethodGet, pathShort, false, shortCtx)
	if err != nil {
		t.Fatalf("short err=%v", err)
	}
	if !handled {
		t.Fatal("expected short path to match")
	}
	if got := shortCtx.Param(names[9]); got != "10" {
		t.Fatalf("short param10=%q", got)
	}

	longCtx := &Context{}
	handled, _, err = router.dispatchInto(MethodGet, pathLong.String(), false, longCtx)
	if err != nil {
		t.Fatalf("long err=%v", err)
	}
	if !handled {
		t.Fatal("expected long path to match")
	}
	if got := longCtx.Param(names[9]); got != "1009" {
		t.Fatalf("long param10=%q", got)
	}

	shortCachedCtx := &Context{}
	handled, _, err = router.dispatchInto(MethodGet, pathShort, false, shortCachedCtx)
	if err != nil {
		t.Fatalf("cached err=%v", err)
	}
	if !handled {
		t.Fatal("expected cached short path to match")
	}
	if got := shortCachedCtx.Param(names[8]); got != "9" {
		t.Fatalf("cached param9=%q", got)
	}
	if got := shortCachedCtx.Param(names[9]); got != "10" {
		t.Fatalf("cached param10=%q", got)
	}
}

func TestRouterSupportsCustomDynamicMethods(t *testing.T) {
	const methodPurge = "PURGE"

	router := &Router{config: &DefaultConfig}
	mustDo(t, router.Add(methodPurge, "/items/:id", func(*Context) error { return nil }))

	ctx := &Context{}
	handled, allowed, err := router.dispatchInto(methodPurge, "/items/42", false, ctx)
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	if !handled {
		t.Fatal("expected custom dynamic method to match")
	}
	if !allowed.empty() {
		t.Fatalf("allowed=%v", allowed)
	}
	if got := ctx.Param("id"); got != "42" {
		t.Fatalf("id=%q", got)
	}

	handled, allowed, err = router.dispatchInto(MethodGet, "/items/42", true, &Context{})
	if err != nil {
		t.Fatalf("method mismatch err=%v", err)
	}
	if handled {
		t.Fatal("expected GET to miss the PURGE route")
	}
	if allowed.mask != 0 {
		t.Fatalf("mask=%v", allowed.mask)
	}
	if !reflect.DeepEqual(allowed.extra, []string{methodPurge}) {
		t.Fatalf("extra=%v", allowed.extra)
	}
	if header := allowed.header(true, true); header != "OPTIONS, PURGE" {
		t.Fatalf("allow header=%q", header)
	}
}

func buildSequentialParamRoute(count int) (string, string, []string, paramRanges) {
	var (
		pattern strings.Builder
		path    strings.Builder
		names   = make([]string, count)
		ranges  paramRanges
	)
	for i := 0; i < count; i++ {
		name := fmt.Sprintf("p%d", i+1)
		value := fmt.Sprintf("%d", i+1)
		names[i] = name

		pattern.WriteByte('/')
		pattern.WriteByte(':')
		pattern.WriteString(name)

		path.WriteByte('/')
		start := path.Len()
		path.WriteString(value)
		ranges.set(i, paramRange{
			start: uint32(start),
			end:   uint32(path.Len()),
		})
	}
	return pattern.String(), path.String(), names, ranges
}
