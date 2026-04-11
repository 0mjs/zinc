package middleware

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/0mjs/zinc"
)

var (
	ErrDecompressUnsupportedEncoding = errors.New("zincdecompress: unsupported content encoding")
	ErrDecompressInvalidBody         = errors.New("zincdecompress: invalid compressed body")
)

type DecompressConfig struct {
	Skipper func(*zinc.Context) bool
}

func Decompress() zinc.Middleware {
	return DecompressWithConfig(DecompressConfig{})
}

func DecompressWithConfig(config DecompressConfig) zinc.Middleware {
	return func(c *zinc.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		req := c.Request()
		if req == nil || req.Body == nil {
			return c.Next()
		}

		encoding := strings.TrimSpace(strings.ToLower(req.Header.Get(zinc.HeaderContentEncoding)))
		if encoding == "" || encoding == "identity" {
			return c.Next()
		}
		if encoding != "gzip" {
			return errors.Join(zinc.ErrUnsupportedMediaType, fmt.Errorf("%w: %s", ErrDecompressUnsupportedEncoding, encoding))
		}

		reader, err := gzip.NewReader(req.Body)
		if err != nil {
			return errors.Join(zinc.ErrBadRequest, fmt.Errorf("%w: %v", ErrDecompressInvalidBody, err))
		}

		req.Body = &gzipRequestBody{
			reader: reader,
			body:   req.Body,
		}
		req.Header.Del(zinc.HeaderContentEncoding)
		req.Header.Del(zinc.HeaderContentLength)
		req.ContentLength = -1

		return c.Next()
	}
}

type gzipRequestBody struct {
	reader *gzip.Reader
	body   io.Closer
}

func (b *gzipRequestBody) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

func (b *gzipRequestBody) Close() error {
	err := b.reader.Close()
	if closeErr := b.body.Close(); err == nil {
		err = closeErr
	}
	return err
}
