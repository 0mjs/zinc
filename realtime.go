package zinc

import (
	"bytes"
	stdctx "context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const sseContentType = "text/event-stream; charset=utf-8"

const (
	// WebSocketTextMessage is a text data message.
	WebSocketTextMessage = websocket.TextMessage
	// WebSocketBinaryMessage is a binary data message.
	WebSocketBinaryMessage = websocket.BinaryMessage
	// WebSocketCloseMessage is a close control message.
	WebSocketCloseMessage = websocket.CloseMessage
	// WebSocketPingMessage is a ping control message.
	WebSocketPingMessage = websocket.PingMessage
	// WebSocketPongMessage is a pong control message.
	WebSocketPongMessage = websocket.PongMessage
)

var (
	// ErrStreamingNotSupported indicates the response writer cannot flush streamed data.
	ErrStreamingNotSupported = errors.New("response writer does not support streaming")
	// ErrWebSocketHandlerNil indicates a nil websocket handler was provided.
	ErrWebSocketHandlerNil = errors.New("websocket handler is nil")
	// ErrSSEHandlerNil indicates a nil SSE handler was provided.
	ErrSSEHandlerNil = errors.New("sse handler is nil")
)

// SSEHandler handles a server-sent events stream.
type SSEHandler func(*Context, *SSEStream) error

// WebSocketConn aliases Gorilla's websocket connection for Zinc server handlers.
type WebSocketConn = websocket.Conn

// WebSocketHandler handles a websocket connection lifecycle.
type WebSocketHandler func(*Context, *WebSocketConn) error

// SSEEvent represents one server-sent event frame.
type SSEEvent struct {
	ID      string
	Event   string
	Data    any
	Retry   time.Duration
	Comment string
}

// WebSocketConfig controls websocket upgrades.
type WebSocketConfig struct {
	HandshakeTimeout  time.Duration
	ReadBufferSize    int
	WriteBufferSize   int
	Subprotocols      []string
	EnableCompression bool
	CheckOrigin       func(*Context) bool
	ResponseHeader    http.Header
}

// SSEStream writes event frames to the client.
type SSEStream struct {
	ctx     *Context
	writer  io.Writer
	flusher http.Flusher
	codec   JSONCodec
	mu      sync.Mutex
}

// SSE starts an event-stream response and invokes handler to write events.
func (c *Context) SSE(handler SSEHandler) error {
	if handler == nil {
		return ErrSSEHandlerNil
	}
	if c.written {
		return ErrResponseAlreadySent
	}

	flusher, ok := c.Writer().(http.Flusher)
	if !ok {
		return ErrStreamingNotSupported
	}

	header := c.Writer().Header()
	setHeaderIfEmpty(header, HeaderContentType, sseContentType)
	setHeaderIfEmpty(header, HeaderCacheControl, "no-cache")
	setHeaderIfEmpty(header, "X-Accel-Buffering", "no")

	status := c.responseStatus()
	c.written = true
	if !bodyAllowed(c.Method(), status) {
		if status != http.StatusOK {
			c.Writer().WriteHeader(status)
		}
		return nil
	}
	if status != http.StatusOK {
		c.Writer().WriteHeader(status)
	}

	stream := &SSEStream{
		ctx:     c,
		writer:  c.Writer(),
		flusher: flusher,
		codec:   c.jsonCodec(),
	}

	flusher.Flush()
	return handler(c, stream)
}

// LastEventID returns the Last-Event-ID request header used by SSE reconnects.
func (c *Context) LastEventID() string {
	return c.GetHeader(HeaderLastEventID)
}

// Context returns the request context for cancellation and deadlines.
func (s *SSEStream) Context() stdctx.Context {
	if s == nil || s.ctx == nil {
		return stdctx.Background()
	}
	return s.ctx.Context()
}

// Done returns the stream cancellation channel.
func (s *SSEStream) Done() <-chan struct{} {
	return s.Context().Done()
}

// Send writes an SSE frame and flushes it to the client.
func (s *SSEStream) Send(event SSEEvent) error {
	if s == nil || s.ctx == nil {
		return nil
	}

	var data string
	if event.Data != nil {
		encoded, err := s.encodeData(event.Data)
		if err != nil {
			return err
		}
		data = encoded
	}

	var b strings.Builder
	if event.Comment != "" {
		for _, line := range splitSSELines(event.Comment) {
			b.WriteString(": ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	if event.ID != "" {
		b.WriteString("id: ")
		b.WriteString(cleanSSEField(event.ID))
		b.WriteByte('\n')
	}
	if event.Event != "" {
		b.WriteString("event: ")
		b.WriteString(cleanSSEField(event.Event))
		b.WriteByte('\n')
	}
	if event.Retry > 0 {
		retryMS := event.Retry / time.Millisecond
		if retryMS < 1 {
			retryMS = 1
		}
		b.WriteString("retry: ")
		b.WriteString(strconv.FormatInt(int64(retryMS), 10))
		b.WriteByte('\n')
	}
	if event.Data != nil {
		for _, line := range splitSSELines(data) {
			b.WriteString("data: ")
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	b.WriteByte('\n')

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, err := io.WriteString(s.writer, b.String()); err != nil {
		return err
	}
	s.flusher.Flush()
	return nil
}

// Data writes an SSE event with only data.
func (s *SSEStream) Data(data any) error {
	return s.Send(SSEEvent{Data: data})
}

// Comment writes an SSE comment frame.
func (s *SSEStream) Comment(comment string) error {
	return s.Send(SSEEvent{Comment: comment})
}

// Retry writes an SSE retry directive in milliseconds.
func (s *SSEStream) Retry(duration time.Duration) error {
	return s.Send(SSEEvent{Retry: duration})
}

// UpgradeWebSocket upgrades the request to a websocket connection.
func (c *Context) UpgradeWebSocket(config ...WebSocketConfig) (*WebSocketConn, error) {
	if c.written {
		return nil, ErrResponseAlreadySent
	}

	cfg := firstWebSocketConfig(config)
	upgrader := websocket.Upgrader{
		HandshakeTimeout:  cfg.HandshakeTimeout,
		ReadBufferSize:    cfg.ReadBufferSize,
		WriteBufferSize:   cfg.WriteBufferSize,
		Subprotocols:      append([]string(nil), cfg.Subprotocols...),
		EnableCompression: cfg.EnableCompression,
	}
	if cfg.CheckOrigin != nil {
		upgrader.CheckOrigin = func(_ *http.Request) bool {
			return cfg.CheckOrigin(c)
		}
	}

	conn, err := upgrader.Upgrade(c.Writer(), c.Request(), cloneHeader(cfg.ResponseHeader))
	if err != nil {
		c.written = true
		return nil, err
	}
	c.written = true
	return conn, nil
}

// WebSocket upgrades and runs a websocket handler. The connection is closed when handler returns.
func (c *Context) WebSocket(handler WebSocketHandler, config ...WebSocketConfig) error {
	if handler == nil {
		return ErrWebSocketHandlerNil
	}
	conn, err := c.UpgradeWebSocket(config...)
	if err != nil {
		return err
	}
	defer conn.Close()
	return handler(c, conn)
}

func (s *SSEStream) encodeData(data any) (string, error) {
	switch value := data.(type) {
	case string:
		return value, nil
	case []byte:
		return string(value), nil
	}

	var b bytes.Buffer
	if err := s.codec.Encode(&b, data, ""); err != nil {
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}

func (c *Context) jsonCodec() JSONCodec {
	if c != nil && c.app != nil && c.app.config.JSONCodec != nil {
		return c.app.config.JSONCodec
	}
	return defaultJSONCodec{}
}

func setHeaderIfEmpty(header http.Header, key, value string) {
	if header.Get(key) == "" {
		header.Set(key, value)
	}
}

func splitSSELines(value string) []string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.Split(value, "\n")
}

func cleanSSEField(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	if idx := strings.IndexByte(value, '\n'); idx >= 0 {
		value = value[:idx]
	}
	return value
}

func firstWebSocketConfig(config []WebSocketConfig) WebSocketConfig {
	if len(config) > 0 {
		return config[0]
	}
	return WebSocketConfig{}
}

func cloneHeader(header http.Header) http.Header {
	if len(header) == 0 {
		return nil
	}
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
