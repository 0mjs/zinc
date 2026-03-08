package zinc

import (
	stdctx "context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

type noFlushResponseWriter struct {
	header http.Header
	code   int
	body   strings.Builder
}

func (w *noFlushResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *noFlushResponseWriter) Write(p []byte) (int, error) {
	if w.code == 0 {
		w.code = http.StatusOK
	}
	return w.body.Write(p)
}

func (w *noFlushResponseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func TestSSEHelpers(t *testing.T) {
	t.Run("sse route helper and framing", func(t *testing.T) {
		app := New()
		mustDo(t, app.SSE("/events", func(c *Context, stream *SSEStream) error {
			if err := stream.Comment("connected"); err != nil {
				return err
			}
			if err := stream.Send(SSEEvent{
				ID:    "1\nignored",
				Event: "tick",
				Data:  Map{"ok": true},
				Retry: 1500 * time.Millisecond,
			}); err != nil {
				return err
			}
			if err := stream.Send(SSEEvent{
				Event: "resume",
				Data:  c.LastEventID(),
			}); err != nil {
				return err
			}
			return stream.Data("line-1\nline-2")
		}))

		resp := performRequest(t, app, http.MethodGet, "/events", nil, map[string]string{
			HeaderLastEventID: "42",
		})
		if resp.Code != http.StatusOK {
			t.Fatalf("status=%d", resp.Code)
		}
		if got := resp.Header().Get(HeaderContentType); got != sseContentType {
			t.Fatalf("content-type=%q", got)
		}
		if got := resp.Header().Get(HeaderCacheControl); got != "no-cache" {
			t.Fatalf("cache-control=%q", got)
		}
		if got := resp.Header().Get("X-Accel-Buffering"); got != "no" {
			t.Fatalf("x-accel-buffering=%q", got)
		}
		body := resp.Body.String()
		for _, fragment := range []string{
			": connected\n",
			"id: 1\n",
			"event: tick\n",
			"retry: 1500\n",
			"data: {\"ok\":true}\n",
			"event: resume\n",
			"data: 42\n",
			"data: line-1\n",
			"data: line-2\n",
		} {
			if !strings.Contains(body, fragment) {
				t.Fatalf("expected fragment %q in body %q", fragment, body)
			}
		}
	})

	t.Run("group sse helper", func(t *testing.T) {
		app := New()
		group := app.Group("/rt")
		mustDo(t, group.SSE("/events", func(c *Context, stream *SSEStream) error {
			return stream.Data("ok")
		}))

		resp := performRequest(t, app, http.MethodGet, "/rt/events", nil, nil)
		if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), "data: ok\n") {
			t.Fatalf("response=%d %q", resp.Code, resp.Body.String())
		}
	})

	t.Run("sse error cases", func(t *testing.T) {
		app := New()
		if err := app.SSE("/events", nil); !errors.Is(err, ErrSSEHandlerNil) {
			t.Fatalf("err=%v", err)
		}
		group := app.Group("/rt")
		if err := group.SSE("/events", nil); !errors.Is(err, ErrSSEHandlerNil) {
			t.Fatalf("err=%v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/events", nil)

		ctxNoFlush := NewContext(&noFlushResponseWriter{}, req)
		defer ctxNoFlush.release()
		ctxNoFlush.app = New()
		if err := ctxNoFlush.SSE(func(*Context, *SSEStream) error { return nil }); !errors.Is(err, ErrStreamingNotSupported) {
			t.Fatalf("err=%v", err)
		}

		ctxNilHandler, _ := newRecorderContext(t, req)
		defer ctxNilHandler.release()
		if err := ctxNilHandler.SSE(nil); !errors.Is(err, ErrSSEHandlerNil) {
			t.Fatalf("err=%v", err)
		}

		ctxWritten, _ := newRecorderContext(t, req)
		defer ctxWritten.release()
		mustDo(t, ctxWritten.String("done"))
		if err := ctxWritten.SSE(func(*Context, *SSEStream) error { return nil }); !errors.Is(err, ErrResponseAlreadySent) {
			t.Fatalf("err=%v", err)
		}
	})
}

func TestWebSocketHelpers(t *testing.T) {
	t.Run("ws route helper echoes messages", func(t *testing.T) {
		app := New()
		mustDo(t, app.WS("/ws", func(c *Context, conn *WebSocketConn) error {
			msgType, payload, err := conn.ReadMessage()
			if err != nil {
				return err
			}
			return conn.WriteMessage(msgType, append([]byte("echo:"), payload...))
		}))

		server := httptest.NewServer(app)
		defer server.Close()

		conn, _, err := websocket.DefaultDialer.Dial(wsURL(server.URL, "/ws"), nil)
		mustDo(t, err)
		defer conn.Close()

		mustDo(t, conn.WriteMessage(WebSocketTextMessage, []byte("hello")))
		msgType, payload, err := conn.ReadMessage()
		mustDo(t, err)
		if msgType != WebSocketTextMessage || string(payload) != "echo:hello" {
			t.Fatalf("message=%d %q", msgType, string(payload))
		}
	})

	t.Run("group ws helper and config", func(t *testing.T) {
		app := New()
		group := app.Group("/rt")
		mustDo(t, group.WS("/ws", func(c *Context, conn *WebSocketConn) error {
			return conn.WriteMessage(WebSocketTextMessage, []byte(conn.Subprotocol()))
		}, WebSocketConfig{
			Subprotocols: []string{"chat.v2", "chat.v1"},
			CheckOrigin: func(c *Context) bool {
				return c.GetHeader(HeaderOrigin) == "https://client.example"
			},
			ResponseHeader: http.Header{
				"X-WS": []string{"ok"},
			},
		}))

		server := httptest.NewServer(app)
		defer server.Close()

		dialer := websocket.Dialer{Subprotocols: []string{"chat.v1"}}
		header := http.Header{}
		header.Set(HeaderOrigin, "https://client.example")
		conn, resp, err := dialer.Dial(wsURL(server.URL, "/rt/ws"), header)
		mustDo(t, err)
		defer conn.Close()

		if conn.Subprotocol() != "chat.v1" {
			t.Fatalf("subprotocol=%q", conn.Subprotocol())
		}
		if resp.Header.Get("X-WS") != "ok" {
			t.Fatalf("response header=%q", resp.Header.Get("X-WS"))
		}

		_, payload, err := conn.ReadMessage()
		mustDo(t, err)
		if string(payload) != "chat.v1" {
			t.Fatalf("payload=%q", string(payload))
		}
	})

	t.Run("upgrade failure does not get double-written", func(t *testing.T) {
		app := New()
		mustDo(t, app.WS("/ws", func(c *Context, conn *WebSocketConn) error {
			return nil
		}))

		resp := performRequest(t, app, http.MethodGet, "/ws", nil, nil)
		if resp.Code != http.StatusBadRequest {
			t.Fatalf("status=%d body=%q", resp.Code, resp.Body.String())
		}
		if body := resp.Body.String(); body != "Bad Request\n" {
			t.Fatalf("body=%q", resp.Body.String())
		}
	})

	t.Run("ws error cases", func(t *testing.T) {
		app := New()
		if err := app.WS("/ws", nil); !errors.Is(err, ErrWebSocketHandlerNil) {
			t.Fatalf("err=%v", err)
		}
		group := app.Group("/rt")
		if err := group.WS("/ws", nil); !errors.Is(err, ErrWebSocketHandlerNil) {
			t.Fatalf("err=%v", err)
		}

		req := httptest.NewRequest(http.MethodGet, "/ws", nil)
		ctx, _ := newRecorderContext(t, req)
		defer ctx.release()

		if err := ctx.WebSocket(nil); !errors.Is(err, ErrWebSocketHandlerNil) {
			t.Fatalf("err=%v", err)
		}

		mustDo(t, ctx.String("done"))
		if _, err := ctx.UpgradeWebSocket(); !errors.Is(err, ErrResponseAlreadySent) {
			t.Fatalf("err=%v", err)
		}
	})
}

func wsURL(serverURL, path string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http") + path
}

func TestSSEStreamContextDoneAndRetry(t *testing.T) {
	var nilStream *SSEStream
	if nilStream.Context() == nil {
		t.Fatal("nil stream context should fall back to background")
	}
	select {
	case <-nilStream.Done():
		t.Fatal("nil stream done channel should not be closed")
	default:
	}
	mustDo(t, nilStream.Retry(5*time.Millisecond))

	app := New()
	mustDo(t, app.SSE("/events", func(c *Context, stream *SSEStream) error {
		if stream.Context() == nil {
			t.Fatal("stream context must not be nil")
		}
		// Ensure Done delegates to the request context channel.
		if stream.Done() != c.Context().Done() {
			t.Fatal("stream done channel does not match request context")
		}
		return stream.Retry(750 * time.Millisecond)
	}))

	resp := performRequest(t, app, http.MethodGet, "/events", nil, nil)
	if resp.Code != http.StatusOK {
		t.Fatalf("status=%d", resp.Code)
	}
	if !strings.Contains(resp.Body.String(), "retry: 750\n") {
		t.Fatalf("body=%q", resp.Body.String())
	}
}

func TestSSEStreamContextFallbackForMissingContext(t *testing.T) {
	stream := &SSEStream{}
	if stream.Context() == nil {
		t.Fatal("context should not be nil")
	}
	if stream.Context() != stdctx.Background() {
		t.Fatal("expected background context fallback")
	}
}

func TestSSEAdditionalBranches(t *testing.T) {
	t.Run("SSE with non-OK and no-body statuses", func(t *testing.T) {
		app := New()
		calledCreated := false
		mustDo(t, app.Get("/created", func(c *Context) error {
			c.Status(http.StatusCreated)
			return c.SSE(func(_ *Context, stream *SSEStream) error {
				calledCreated = true
				return stream.Data("ok")
			})
		}))
		calledNoContent := false
		mustDo(t, app.Get("/nocontent", func(c *Context) error {
			c.Status(http.StatusNoContent)
			return c.SSE(func(_ *Context, stream *SSEStream) error {
				calledNoContent = true
				return stream.Data("should-not-write")
			})
		}))

		created := performRequest(t, app, http.MethodGet, "/created", nil, nil)
		if created.Code != http.StatusCreated || !calledCreated {
			t.Fatalf("created response=%d called=%v", created.Code, calledCreated)
		}
		if !strings.Contains(created.Body.String(), "data: ok\n") {
			t.Fatalf("body=%q", created.Body.String())
		}

		noContent := performRequest(t, app, http.MethodGet, "/nocontent", nil, nil)
		if noContent.Code != http.StatusNoContent {
			t.Fatalf("status=%d", noContent.Code)
		}
		if noContent.Body.Len() != 0 {
			t.Fatalf("body=%q", noContent.Body.String())
		}
		if calledNoContent {
			t.Fatal("handler should not run for 204 SSE response")
		}
	})

	t.Run("stream send encode and write errors", func(t *testing.T) {
		writer := &strings.Builder{}
		flusher := &flushRecorder{}
		stream := &SSEStream{
			ctx:     &Context{app: New()},
			writer:  writer,
			flusher: flusher,
			codec:   defaultJSONCodec{},
		}

		mustDo(t, stream.Send(SSEEvent{Data: []byte("raw-bytes"), Retry: time.Nanosecond}))
		body := writer.String()
		if !strings.Contains(body, "retry: 1\n") || !strings.Contains(body, "data: raw-bytes\n") {
			t.Fatalf("body=%q", body)
		}
		if flusher.calls == 0 {
			t.Fatal("expected flush call")
		}

		errStream := &SSEStream{
			ctx:     &Context{app: New()},
			writer:  errWriteCloser{err: errors.New("write failed")},
			flusher: &flushRecorder{},
			codec:   defaultJSONCodec{},
		}
		if err := errStream.Send(SSEEvent{Data: "x"}); err == nil || !strings.Contains(err.Error(), "write failed") {
			t.Fatalf("err=%v", err)
		}

		codecErrStream := &SSEStream{
			ctx:     &Context{app: New()},
			writer:  &strings.Builder{},
			flusher: &flushRecorder{},
			codec:   codecError{err: errors.New("encode failed")},
		}
		if err := codecErrStream.Send(SSEEvent{Data: Map{"x": 1}}); err == nil || !strings.Contains(err.Error(), "encode failed") {
			t.Fatalf("err=%v", err)
		}
	})

	t.Run("jsonCodec fallback", func(t *testing.T) {
		var nilCtx *Context
		if err := nilCtx.jsonCodec().Encode(io.Discard, Map{"ok": true}, ""); err != nil {
			t.Fatalf("err=%v", err)
		}
		plainCtx := &Context{}
		if err := plainCtx.jsonCodec().Encode(io.Discard, Map{"ok": true}, ""); err != nil {
			t.Fatalf("err=%v", err)
		}
	})
}

type flushRecorder struct {
	calls int
}

func (f *flushRecorder) Flush() {
	f.calls++
}

type errWriteCloser struct {
	err error
}

func (w errWriteCloser) Write([]byte) (int, error) {
	return 0, w.err
}

type codecError struct {
	err error
}

func (c codecError) Encode(io.Writer, any, string) error {
	return c.err
}

func (c codecError) Decode(io.Reader, any) error {
	return c.err
}
