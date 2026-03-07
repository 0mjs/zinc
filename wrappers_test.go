package zinc

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestMethodWrapperCoverage(t *testing.T) {
	app := New()
	mustDo(t, app.Post("/post", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Put("/put", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Delete("/delete", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Patch("/patch", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Head("/head", func(c *Context) error { return c.String("head") }))
	mustDo(t, app.Options("/options", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Connect("/connect", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Trace("/trace", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Match([]string{MethodGet, MethodPost}, "/match", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, app.Get("/string", StringHandler("string")))

	cases := []struct {
		method string
		path   string
		body   string
	}{
		{MethodPost, "/post", MethodPost},
		{MethodPut, "/put", MethodPut},
		{MethodDelete, "/delete", MethodDelete},
		{MethodPatch, "/patch", MethodPatch},
		{MethodOptions, "/options", MethodOptions},
		{MethodConnect, "/connect", MethodConnect},
		{MethodTrace, "/trace", MethodTrace},
		{MethodGet, "/match", MethodGet},
		{MethodPost, "/match", MethodPost},
		{MethodGet, "/string", "string"},
	}
	for _, tc := range cases {
		resp := performRequest(t, app, tc.method, tc.path, nil, nil)
		if resp.Body.String() != tc.body {
			t.Fatalf("%s %s body=%q", tc.method, tc.path, resp.Body.String())
		}
	}
	head := performRequest(t, app, MethodHead, "/head", nil, nil)
	if head.Code != http.StatusOK || head.Body.Len() != 0 {
		t.Fatalf("head=%d %q", head.Code, head.Body.String())
	}
}

func TestGetShorthandStringHandler(t *testing.T) {
	app := New()
	mustDo(t, app.Get("/", "Hello, world!"))

	resp := performRequest(t, app, MethodGet, "/", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d", resp.Code)
	}
	if body := resp.Body.String(); body != "Hello, world!" {
		t.Fatalf("body=%q", body)
	}
}

func TestGroupGetShorthandStringHandler(t *testing.T) {
	app := New()
	group := app.Group("/api")
	mustDo(t, group.Get("/hello", "from group"))

	resp := performRequest(t, app, MethodGet, "/api/hello", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d", resp.Code)
	}
	if body := resp.Body.String(); body != "from group" {
		t.Fatalf("body=%q", body)
	}
}

func TestGetShorthandInvalidHandlerType(t *testing.T) {
	app := New()
	err := app.Get("/bad", 123)
	if err == nil {
		t.Fatal("expected error for invalid GET handler type")
	}
	if !strings.Contains(err.Error(), "unsupported GET handler type int") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGroupWrapperCoverage(t *testing.T) {
	app := New()
	group := app.Group("/g")
	mustDo(t, group.Post("/post", func(c *Context) error { return c.String("post") }))
	mustDo(t, group.Put("/put", func(c *Context) error { return c.String("put") }))
	mustDo(t, group.Delete("/delete", func(c *Context) error { return c.String("delete") }))
	mustDo(t, group.Patch("/patch", func(c *Context) error { return c.String("patch") }))
	mustDo(t, group.Head("/head", func(c *Context) error { return c.String("head") }))
	mustDo(t, group.Options("/options", func(c *Context) error { return c.String("options") }))
	mustDo(t, group.Connect("/connect", func(c *Context) error { return c.String("connect") }))
	mustDo(t, group.Trace("/trace", func(c *Context) error { return c.String("trace") }))
	mustDo(t, group.Match([]string{MethodGet, MethodPost}, "/match", func(c *Context) error { return c.String("match") }))
	mustDo(t, group.All("/all", func(c *Context) error { return c.String(c.Method()) }))
	mustDo(t, group.Any("/any", func(c *Context) error { return c.String(c.Method()) }))

	mux := http.NewServeMux()
	mux.HandleFunc("/mounted", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("mounted"))
	})
	group.Mount("/mount", mux)

	fsys := fstest.MapFS{"hello.txt": &fstest.MapFile{Data: []byte("hello")}}
	mustDo(t, group.StaticFS("/static", fsys, WithStaticBrowse(true), WithStaticIndex("index.html")))
	mustDo(t, group.FileFS("/file", "hello.txt", fsys))

	paths := map[string]string{
		"/g/post":             "post",
		"/g/put":              "put",
		"/g/delete":           "delete",
		"/g/patch":            "patch",
		"/g/options":          "options",
		"/g/connect":          "connect",
		"/g/trace":            "trace",
		"/g/match":            "match",
		"/g/static/hello.txt": "hello",
		"/g/file":             "hello",
		"/g/mount/mounted":    "mounted",
	}
	methods := map[string]string{
		"/g/post":    MethodPost,
		"/g/put":     MethodPut,
		"/g/delete":  MethodDelete,
		"/g/patch":   MethodPatch,
		"/g/options": MethodOptions,
		"/g/connect": MethodConnect,
		"/g/trace":   MethodTrace,
	}
	for path, expected := range paths {
		method := MethodGet
		if override, ok := methods[path]; ok {
			method = override
		}
		resp := performRequest(t, app, method, path, nil, nil)
		if resp.Body.String() != expected {
			t.Fatalf("%s %s body=%q", method, path, resp.Body.String())
		}
	}

	head := performRequest(t, app, MethodHead, "/g/head", nil, nil)
	if head.Code != http.StatusOK {
		t.Fatalf("head status=%d", head.Code)
	}
}

func TestBinderWrapperCoverage(t *testing.T) {
	app := NewWithConfig(Config{Validator: validatingStub{}})
	req := httptest.NewRequest(MethodPost, "/users/12?ready=true&page=5", strings.NewReader(`{"name":"lin"}`))
	req.Header.Set(HeaderContentType, "application/json")
	ctx, _ := newRecorderContext(t, req)
	defer ctx.release()
	ctx.app = app
	ctx.setParam("id", "12")

	var fromPath struct {
		ID int `path:"id"`
	}
	mustDo(t, ctx.BindPath(&fromPath))
	if fromPath.ID != 12 {
		t.Fatalf("path payload=%+v", fromPath)
	}

	var fromQuery struct {
		Page  int  `query:"page"`
		Ready bool `query:"ready"`
	}
	mustDo(t, ctx.BindQuery(&fromQuery))
	if fromQuery.Page != 5 || !fromQuery.Ready {
		t.Fatalf("query payload=%+v", fromQuery)
	}

	var fromBody bindPayload
	mustDo(t, ctx.BindJSON(&fromBody))
	if fromBody.Name != "lin" {
		t.Fatalf("json payload=%+v", fromBody)
	}

	var viaBinder bindPayload
	mustDo(t, app.config.Binder.BindBody(ctx, &viaBinder))
	if viaBinder.Name != "lin" {
		t.Fatalf("binder payload=%+v", viaBinder)
	}
}

func TestStaticOptionsAndServeTLSCoverage(t *testing.T) {
	cfg := StaticConfig{}
	WithStaticBrowse(true)(&cfg)
	WithStaticIndex("home.html")(&cfg)
	if !cfg.Browse || cfg.Index != "home.html" {
		t.Fatalf("static config=%+v", cfg)
	}

	app := New()
	if err := app.serveTLS(nil, "cert", "key"); err == nil {
		t.Fatal("serveTLS should fail for nil listener")
	}
}

func TestFileAndStaticWrapperCoverage(t *testing.T) {
	dir := t.TempDir()
	mustDo(t, os.WriteFile(filepath.Join(dir, "asset.txt"), []byte("asset"), 0o644))

	app := New()
	group := app.Group("/files")
	mustDo(t, group.Static("/static", dir))
	mustDo(t, group.File("/single", filepath.Join(dir, "asset.txt")))

	fsys := fstest.MapFS{"asset.txt": &fstest.MapFile{Data: []byte("asset-fs")}}
	mustDo(t, app.FileFS("/filefs", "asset.txt", fsys))

	staticResp := performRequest(t, app, MethodGet, "/files/static/asset.txt", nil, nil)
	if staticResp.Body.String() != "asset" {
		t.Fatalf("static body=%q", staticResp.Body.String())
	}

	fileResp := performRequest(t, app, MethodGet, "/files/single", nil, nil)
	if fileResp.Body.String() != "asset" {
		t.Fatalf("file body=%q", fileResp.Body.String())
	}

	fileFSResp := performRequest(t, app, MethodGet, "/filefs", nil, nil)
	if fileFSResp.Body.String() != "asset-fs" {
		t.Fatalf("filefs body=%q", fileFSResp.Body.String())
	}
}
