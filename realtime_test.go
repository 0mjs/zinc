package zinc

import (
	"errors"
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
		mustDo(t, app.WS("/ws", func(c *Context, conn *websocket.Conn) error {
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

		mustDo(t, conn.WriteMessage(websocket.TextMessage, []byte("hello")))
		msgType, payload, err := conn.ReadMessage()
		mustDo(t, err)
		if msgType != websocket.TextMessage || string(payload) != "echo:hello" {
			t.Fatalf("message=%d %q", msgType, string(payload))
		}
	})

	t.Run("group ws helper and config", func(t *testing.T) {
		app := New()
		group := app.Group("/rt")
		mustDo(t, group.WS("/ws", func(c *Context, conn *websocket.Conn) error {
			return conn.WriteMessage(websocket.TextMessage, []byte(conn.Subprotocol()))
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
		mustDo(t, app.WS("/ws", func(c *Context, conn *websocket.Conn) error {
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
