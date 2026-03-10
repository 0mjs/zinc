package zinc

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var zincBenchmarkBindJSONBody = []byte(`{"name":"alice","roles":["admin","editor"]}`)

type zincBenchmarkBindQueryInput struct {
	Page    int    `query:"page"`
	Search  string `query:"search"`
	Enabled bool   `query:"enabled"`
}

type zincBenchmarkBindAPIInput struct {
	TeamID  int      `path:"teamID"`
	UserID  int      `path:"userID"`
	Verbose bool     `query:"verbose"`
	Limit   int      `query:"limit"`
	Name    string   `json:"name"`
	Roles   []string `json:"roles"`
}

type zincBenchmarkBindFormInput struct {
	Name  string   `form:"name"`
	Count int      `form:"count"`
	Tags  []string `form:"tags"`
}

func resetZincBenchmarkBody(req *http.Request, body []byte) {
	if len(body) == 0 {
		req.Body = http.NoBody
		req.ContentLength = 0
		return
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	req.ContentLength = int64(len(body))
}

func resetZincBenchmarkFormRequest(req *http.Request, body []byte) {
	resetZincBenchmarkBody(req, body)
	req.Form = nil
	req.PostForm = nil
	req.MultipartForm = nil
}

func resetZincBindBenchmarkContext(ctx *Context, rw *zincBenchmarkDiscardWriter, req *http.Request, app *App) {
	rw.reset()
	ctx.reset(rw, req)
	ctx.app = app
}

func BenchmarkZincBindQueryOnly(b *testing.B) {
	app := New()
	req := httptest.NewRequest(MethodGet, "/search?page=2&search=alpha&enabled=true", nil)
	rw := newZincBenchmarkDiscardWriter()
	ctx := newBareBenchmarkContext()

	resetZincBindBenchmarkContext(ctx, rw, req, app)
	var proof zincBenchmarkBindQueryInput
	if err := ctx.BindQuery(&proof); err != nil {
		b.Fatalf("bind query proof failed: %v", err)
	}
	if proof.Page != 2 || proof.Search != "alpha" || !proof.Enabled {
		b.Fatalf("unexpected query bind proof: %+v", proof)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetZincBindBenchmarkContext(ctx, rw, req, app)
		var input zincBenchmarkBindQueryInput
		if err := ctx.BindQuery(&input); err != nil {
			b.Fatalf("bind query failed: %v", err)
		}
		zincBenchmarkSinkInt = input.Page
		zincBenchmarkSinkString = input.Search
	}
}

func BenchmarkZincBindPathQueryJSON(b *testing.B) {
	app := New()
	req := httptest.NewRequest(MethodPost, "/teams/42/users/7?verbose=true&limit=25", bytes.NewReader(zincBenchmarkBindJSONBody))
	req.Header.Set(HeaderContentType, "application/json")
	rw := newZincBenchmarkDiscardWriter()
	ctx := newBareBenchmarkContext()

	resetZincBenchmarkBody(req, zincBenchmarkBindJSONBody)
	resetZincBindBenchmarkContext(ctx, rw, req, app)
	ctx.setParam("teamID", "42")
	ctx.setParam("userID", "7")
	var proof zincBenchmarkBindAPIInput
	if err := ctx.Bind(&proof); err != nil {
		b.Fatalf("bind path/query/json proof failed: %v", err)
	}
	if proof.TeamID != 42 || proof.UserID != 7 || !proof.Verbose || proof.Limit != 25 || proof.Name != "alice" || len(proof.Roles) != 2 {
		b.Fatalf("unexpected bind proof: %+v", proof)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetZincBenchmarkBody(req, zincBenchmarkBindJSONBody)
		resetZincBindBenchmarkContext(ctx, rw, req, app)
		ctx.setParam("teamID", "42")
		ctx.setParam("userID", "7")
		var input zincBenchmarkBindAPIInput
		if err := ctx.Bind(&input); err != nil {
			b.Fatalf("bind path/query/json failed: %v", err)
		}
		zincBenchmarkSinkInt = input.TeamID + input.UserID + input.Limit + len(input.Roles)
		zincBenchmarkSinkString = input.Name
	}
}

func BenchmarkZincBindJSONCachedBody(b *testing.B) {
	app := New()
	req := httptest.NewRequest(MethodPost, "/payload", bytes.NewReader(zincBenchmarkBindJSONBody))
	req.Header.Set(HeaderContentType, "application/json")
	rw := newZincBenchmarkDiscardWriter()
	ctx := newBareBenchmarkContext()

	resetZincBindBenchmarkContext(ctx, rw, req, app)
	var proof zincBenchmarkBindAPIInput
	if err := ctx.BindJSON(&proof); err != nil {
		b.Fatalf("bind cached json proof failed: %v", err)
	}
	if proof.Name != "alice" || len(proof.Roles) != 2 || !ctx.bodyRead {
		b.Fatalf("unexpected cached json proof: %+v bodyRead=%v", proof, ctx.bodyRead)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var input zincBenchmarkBindAPIInput
		if err := ctx.BindJSON(&input); err != nil {
			b.Fatalf("bind cached json failed: %v", err)
		}
		zincBenchmarkSinkInt = len(input.Roles)
		zincBenchmarkSinkString = input.Name
	}
}

func BenchmarkZincBindForm(b *testing.B) {
	app := New()
	formBody := []byte(url.Values{
		"name":  {"mia"},
		"count": {"25"},
		"tags":  {"alpha", "beta"},
	}.Encode())
	req := httptest.NewRequest(MethodPost, "/submit", bytes.NewReader(formBody))
	req.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")
	rw := newZincBenchmarkDiscardWriter()
	ctx := newBareBenchmarkContext()

	resetZincBenchmarkFormRequest(req, formBody)
	resetZincBindBenchmarkContext(ctx, rw, req, app)
	var proof zincBenchmarkBindFormInput
	if err := ctx.BindForm(&proof); err != nil {
		b.Fatalf("bind form proof failed: %v", err)
	}
	if proof.Name != "mia" || proof.Count != 25 || len(proof.Tags) != 2 {
		b.Fatalf("unexpected form bind proof: %+v", proof)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resetZincBenchmarkFormRequest(req, formBody)
		resetZincBindBenchmarkContext(ctx, rw, req, app)
		var input zincBenchmarkBindFormInput
		if err := ctx.BindForm(&input); err != nil {
			b.Fatalf("bind form failed: %v", err)
		}
		zincBenchmarkSinkInt = input.Count + len(input.Tags)
		zincBenchmarkSinkString = input.Name
	}
}
