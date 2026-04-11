package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/0mjs/zinc"
)

type RequestIDGenerator func(*zinc.Context) (string, error)

type RequestIDConfig struct {
	Skipper   func(*zinc.Context) bool
	Header    string
	Generator RequestIDGenerator
}

type RequestIDState struct {
	ID        string
	Header    string
	Generated bool
}

type requestIDContextKey int

const requestIDStateContextKey requestIDContextKey = iota

func DefaultRequestIDConfig() RequestIDConfig {
	return RequestIDConfig{
		Header:    zinc.HeaderXRequestID,
		Generator: RandomRequestID,
	}
}

func RequestID() zinc.Middleware {
	return RequestIDWithConfig(DefaultRequestIDConfig())
}

func RequestIDWithConfig(config RequestIDConfig) zinc.Middleware {
	cfg := resolveRequestIDConfig(config)

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		req := c.Request()
		id := ""
		if req != nil {
			id = req.Header.Get(cfg.Header)
		}

		generated := false
		if id == "" {
			var err error
			id, err = cfg.Generator(c)
			if err != nil {
				return err
			}
			generated = true
			if req != nil {
				req.Header.Set(cfg.Header, id)
			}
		}

		c.SetHeader(cfg.Header, id)
		c.Set(requestIDStateContextKey, RequestIDState{
			ID:        id,
			Header:    cfg.Header,
			Generated: generated,
		})

		return c.Next()
	}
}

func RequestIDCurrent(c *zinc.Context) (RequestIDState, bool) {
	if c == nil {
		return RequestIDState{}, false
	}
	value, ok := c.Get(requestIDStateContextKey)
	if !ok {
		return RequestIDState{}, false
	}
	state, ok := value.(RequestIDState)
	return state, ok
}

func MustRequestIDCurrent(c *zinc.Context) RequestIDState {
	state, ok := RequestIDCurrent(c)
	if !ok {
		panic("zincrequestid: request id not found")
	}
	return state
}

func RequestIDValue(c *zinc.Context) string {
	state, ok := RequestIDCurrent(c)
	if ok {
		return state.ID
	}
	if c == nil {
		return ""
	}
	return c.RequestID()
}

func RandomRequestID(*zinc.Context) (string, error) {
	var buf [16]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf[:]), nil
}

func StaticRequestID(id string) RequestIDGenerator {
	return func(*zinc.Context) (string, error) {
		return id, nil
	}
}

func resolveRequestIDConfig(config RequestIDConfig) RequestIDConfig {
	cfg := DefaultRequestIDConfig()
	cfg.Skipper = config.Skipper
	if config.Header != "" {
		cfg.Header = config.Header
	}
	if config.Generator != nil {
		cfg.Generator = config.Generator
	}
	if cfg.Header == "" {
		panic("zincrequestid: Header is required")
	}
	if cfg.Generator == nil {
		panic("zincrequestid: Generator is required")
	}
	return cfg
}
