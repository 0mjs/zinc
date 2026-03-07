package zinc

import (
	"bytes"
	stdctx "context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type validatingStub struct{}

func (validatingStub) Validate(v any) error {
	if payload, ok := v.(*bindPayload); ok && payload.Name == "" {
		return ErrBadRequest.WithMessage("name required")
	}
	return nil
}

type bindPayload struct {
	ID    int      `path:"id"`
	Page  int      `query:"page"`
	Name  string   `json:"name" form:"name"`
	Auth  string   `header:"x-auth"`
	Tags  []string `form:"tags"`
	Ready bool     `query:"ready"`
}

func TestContextRequestHelpersAndMetadata(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "https://example.com/users/42?page=3&ready=true", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set(HeaderXForwardedFor, "203.0.113.8, 10.0.0.1")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set(HeaderUpgrade, "websocket")
	req.Header.Set(HeaderConnection, "keep-alive, Upgrade")
	req.Header.Set(HeaderXRequestID, "req-123")
	req.AddCookie(&http.Cookie{Name: "session", Value: "abc"})

	ctx, resp := newRecorderContext(t, req)
	defer ctx.release()
	ctx.app = NewWithConfig(Config{ProxyHeader: HeaderXForwardedFor, TrustedProxies: []string{"10.0.0.1"}})
	ctx.setParam("id", "42")

	ctx.Set("key", "value")
	if got, ok := ctx.Get("key"); !ok || got != "value" {
		t.Fatalf("Get = %v %v", got, ok)
	}
	if ctx.MustGet("key") != "value" {
		t.Fatal("MustGet returned wrong value")
	}

	if ctx.Method() != http.MethodGet {
		t.Fatalf("method=%q", ctx.Method())
	}
	if ctx.Path() != "/users/42" {
		t.Fatalf("path=%q", ctx.Path())
	}
	if ctx.OriginalURL() != "/users/42?page=3&ready=true" {
		t.Fatalf("original=%q", ctx.OriginalURL())
	}
	ctx.SetPath("/users/99")
	if ctx.Path() != "/users/99" {
		t.Fatalf("path=%q", ctx.Path())
	}
	if ctx.Param("id") != "42" || ctx.ParamOr("missing", "fallback") != "fallback" {
		t.Fatal("param helpers failed")
	}
	if ctx.Query("page") != "3" || ctx.QueryOr("missing", "fallback") != "fallback" {
		t.Fatal("query helpers failed")
	}
	if ctx.QueryValues().Encode() != (url.Values{"page": []string{"3"}, "ready": []string{"true"}}).Encode() {
		t.Fatalf("query values=%v", ctx.QueryValues())
	}
	if ctx.GetHeader(HeaderXRequestID) != "req-123" {
		t.Fatal("header lookup failed")
	}
	cookie, err := ctx.Cookie("session")
	mustDo(t, err)
	if cookie.Value != "abc" || len(ctx.Cookies()) != 1 {
		t.Fatal("cookie helpers failed")
	}
	if ctx.Scheme() != "https" || !ctx.Secure() {
		t.Fatal("scheme helpers failed")
	}
	if ctx.IP() != "203.0.113.8" {
		t.Fatalf("ip=%q", ctx.IP())
	}
	if len(ctx.IPs()) != 2 || ctx.RemoteIP() != "10.0.0.1" {
		t.Fatalf("ips=%v remote=%q", ctx.IPs(), ctx.RemoteIP())
	}
	if !ctx.IsWebSocket() {
		t.Fatal("expected websocket upgrade")
	}
	preflightReq := httptest.NewRequest(http.MethodOptions, "http://example.com", nil)
	preflightReq.Header.Set(HeaderAccessControlRequestMethod, http.MethodPost)
	preflightCtx, _ := newRecorderContext(t, preflightReq)
	defer preflightCtx.release()
	if !preflightCtx.IsPreflight() {
		t.Fatal("expected preflight request")
	}
	if ctx.RequestID() != "req-123" {
		t.Fatal("request id helper failed")
	}

	valueCtx := stdctx.WithValue(ctx.Context(), "key", "value")
	ctx.SetContext(valueCtx)
	if got := ctx.Context().Value("key"); got != "value" {
		t.Fatalf("context value=%v", got)
	}
	ctx.SetWriter(resp)
	ctx.SetRequest(req)
	if ctx.Writer() != resp || ctx.Request() != req {
		t.Fatal("raw writer/request helpers failed")
	}
}

func TestContextBodyBindingAndUploads(t *testing.T) {
	t.Run("body bytes and bind", func(t *testing.T) {
		app := NewWithConfig(Config{Validator: validatingStub{}})
		body := strings.NewReader(`{"name":"matt"}`)
		req := httptest.NewRequest(http.MethodPost, "/users/7?page=2&ready=true", body)
		req.Header.Set(HeaderContentType, "application/json")
		req.Header.Set("X-Auth", "secret")
		ctx, _ := newRecorderContext(t, req)
		defer ctx.release()
		ctx.app = app
		ctx.setParam("id", "7")

		bytesBody, err := ctx.BodyBytes()
		mustDo(t, err)
		if string(bytesBody) != `{"name":"matt"}` {
			t.Fatalf("body=%q", string(bytesBody))
		}
		stringBody, err := ctx.BodyString()
		mustDo(t, err)
		if stringBody != `{"name":"matt"}` {
			t.Fatalf("body=%q", stringBody)
		}

		var payload bindPayload
		mustDo(t, ctx.Bind(&payload))
		mustDo(t, ctx.BindHeader(&payload))
		if payload.ID != 7 || payload.Page != 2 || payload.Name != "matt" || payload.Auth != "secret" || !payload.Ready {
			t.Fatalf("payload=%+v", payload)
		}
	})

	t.Run("xml and form binding", func(t *testing.T) {
		app := NewWithConfig(Config{Validator: validatingStub{}})

		xmlReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`<bindPayload><Name>zoe</Name></bindPayload>`))
		xmlReq.Header.Set(HeaderContentType, "application/xml")
		xmlCtx, _ := newRecorderContext(t, xmlReq)
		defer xmlCtx.release()
		xmlCtx.app = app
		var xmlPayload struct {
			Name string `xml:"Name"`
		}
		mustDo(t, xmlCtx.BindXML(&xmlPayload))
		if xmlPayload.Name != "zoe" {
			t.Fatalf("xml payload=%+v", xmlPayload)
		}

		form := url.Values{"name": {"mia"}, "tags": {"a", "b"}}
		formReq := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
		formReq.Header.Set(HeaderContentType, "application/x-www-form-urlencoded")
		formCtx, _ := newRecorderContext(t, formReq)
		defer formCtx.release()
		formCtx.app = app
		var formPayload bindPayload
		mustDo(t, formCtx.BindForm(&formPayload))
		if formPayload.Name != "mia" || len(formPayload.Tags) != 2 {
			t.Fatalf("form payload=%+v", formPayload)
		}
	})

	t.Run("validator failure", func(t *testing.T) {
		app := NewWithConfig(Config{Validator: validatingStub{}})
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
		req.Header.Set(HeaderContentType, "application/json")
		ctx, _ := newRecorderContext(t, req)
		defer ctx.release()
		ctx.app = app
		var payload bindPayload
		if err := ctx.Bind(&payload); err == nil {
			t.Fatal("expected validation error")
		}
	})

	t.Run("multipart upload helpers", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "hello.txt")
		mustDo(t, err)
		_, err = part.Write([]byte("upload"))
		mustDo(t, err)
		mustDo(t, writer.WriteField("name", "doc"))
		mustDo(t, writer.Close())

		req := httptest.NewRequest(http.MethodPost, "/", &body)
		req.Header.Set(HeaderContentType, writer.FormDataContentType())
		ctx, _ := newRecorderContext(t, req)
		defer ctx.release()

		file, err := ctx.FormFile("file")
		mustDo(t, err)
		files, err := ctx.FormFiles("file")
		mustDo(t, err)
		form, err := ctx.MultipartForm()
		mustDo(t, err)
		if form.Value["name"][0] != "doc" || len(files) != 1 || file.Filename != "hello.txt" {
			t.Fatal("multipart helpers failed")
		}

		dst := filepath.Join(t.TempDir(), "saved.txt")
		mustDo(t, ctx.SaveFile(file, dst))
		data, err := os.ReadFile(dst)
		mustDo(t, err)
		if string(data) != "upload" {
			t.Fatalf("saved file=%q", string(data))
		}
	})
}

func TestContextRouteAndJSONEncodingHelpers(t *testing.T) {
	app := New()
	mustDo(t, app.Get("/meta/:id", func(c *Context) error {
		return c.JSON(Map{"path": c.FullPath(), "route": c.Route(), "id": c.Param("id")})
	}))
	resp := performRequest(t, app, http.MethodGet, "/meta/11", nil, nil)
	var payload map[string]any
	mustDo(t, json.Unmarshal(resp.Body.Bytes(), &payload))
	if payload["path"] != "/meta/:id" || payload["id"] != "11" {
		t.Fatalf("payload=%v", payload)
	}
}
