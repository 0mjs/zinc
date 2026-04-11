package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/0mjs/zinc"
)

const HeaderUberTraceID = "Uber-Trace-Id"

type JaegerObserver func(*zinc.Context, JaegerSpan) error

type JaegerOperationName func(*zinc.Context) string

type JaegerConfig struct {
	Skipper   func(*zinc.Context) bool
	Observe   JaegerObserver
	Operation JaegerOperationName
	Now       func() time.Time
}

type JaegerSpan struct {
	TraceID      string
	SpanID       string
	ParentSpanID string
	Flags        string
	Operation    string
	StartTime    time.Time
	Duration     time.Duration
	Error        error
}

type jaegerContextKey int

const jaegerSpanContextKey jaegerContextKey = iota

func Jaeger(observer JaegerObserver) zinc.Middleware {
	return JaegerWithConfig(JaegerConfig{Observe: observer})
}

func JaegerWithConfig(config JaegerConfig) zinc.Middleware {
	cfg := resolveJaegerConfig(config)
	now := cfg.Now

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		span := newJaegerSpan(c, cfg.Operation(c), now())
		c.Set(jaegerSpanContextKey, span)
		c.SetHeader(HeaderUberTraceID, span.headerValue())

		err := c.Next()
		span.Operation = cfg.Operation(c)
		span.Error = err
		span.Duration = now().Sub(span.StartTime)
		if cfg.Observe != nil {
			if observeErr := cfg.Observe(c, span); observeErr != nil {
				return observeErr
			}
		}
		return err
	}
}

func JaegerCurrent(c *zinc.Context) (JaegerSpan, bool) {
	if c == nil {
		return JaegerSpan{}, false
	}
	value, ok := c.Get(jaegerSpanContextKey)
	if !ok {
		return JaegerSpan{}, false
	}
	span, ok := value.(JaegerSpan)
	return span, ok
}

func resolveJaegerConfig(config JaegerConfig) JaegerConfig {
	if config.Operation == nil {
		config.Operation = func(c *zinc.Context) string {
			if route := c.FullPath(); route != "" {
				return c.Method() + " " + route
			}
			return c.Method() + " " + c.Path()
		}
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	return config
}

func newJaegerSpan(c *zinc.Context, operation string, start time.Time) JaegerSpan {
	traceID, parentID, flags := parseUberTraceID(c.GetHeader(HeaderUberTraceID))
	if traceID == "" {
		traceID = randomHex(16)
	}
	if flags == "" {
		flags = "1"
	}
	return JaegerSpan{
		TraceID:      traceID,
		SpanID:       randomHex(8),
		ParentSpanID: parentID,
		Flags:        flags,
		Operation:    operation,
		StartTime:    start,
	}
}

func parseUberTraceID(value string) (traceID, parentID, flags string) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 4 {
		return "", "", ""
	}
	return parts[0], parts[1], parts[3]
}

func (s JaegerSpan) headerValue() string {
	return s.TraceID + ":" + s.SpanID + ":" + s.ParentSpanID + ":" + s.Flags
}

func randomHex(bytesLen int) string {
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return strings.Repeat("0", bytesLen*2)
	}
	return hex.EncodeToString(buf)
}
