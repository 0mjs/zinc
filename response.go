package zinc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
)

var ErrResponseAlreadySent = errors.New("response already sent")

// Pre-allocated constant byte slices for common responses
var (
	nullBytes        = []byte("null")
	helloWorldBytes  = []byte("Hello World!")
	textContentType  = "text/plain; charset=utf-8"
	jsonContentType  = "application/json; charset=utf-8"
	htmlContentType  = "text/html; charset=utf-8"
	octetContentType = "application/octet-stream"
	nosniffHeader    = "nosniff"

	// Pre-allocated common headers
	jsonHeaders = http.Header{
		"Content-Type":           []string{jsonContentType},
		"X-Content-Type-Options": []string{nosniffHeader},
	}
	// textHeaders = http.Header{
	// 	"Content-Type": []string{textContentType},
	// }
	htmlHeaders = http.Header{
		"Content-Type": []string{htmlContentType},
	}
	// octetHeaders = http.Header{
	// 	"Content-Type": []string{octetContentType},
	// }
)

// Send sends a response with the appropriate content type.
func (c *Context) Send(data interface{}) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	// Get app config to check if default content type is disabled
	app, ok := c.Get("app").(*App)
	disableDefaultContentType := false
	if ok && app != nil && app.config != nil {
		disableDefaultContentType = app.config.DisableDefaultContentType
	}

	// Set content type based on data type (if not disabled)
	if !disableDefaultContentType {
		switch data.(type) {
		case string:
			c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		case []byte:
			c.Response.Header().Set("Content-Type", "application/octet-stream")
		case nil:
			c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		default:
			c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
	}

	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.Response.WriteHeader(c.status)

	switch d := data.(type) {
	case nil:
		_, err := c.Response.Write([]byte("null"))
		return err
	case string:
		_, err := c.Response.Write([]byte(d))
		return err
	case []byte:
		_, err := c.Response.Write(d)
		return err
	default:
		// For JSON responses, ensure the content type is set correctly
		if !disableDefaultContentType {
			c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		}
		return json.NewEncoder(c.Response).Encode(data)
	}
}

// JSON serializes and sends JSON data
func (c *Context) JSON(data interface{}) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	// Always set JSON content type since this is an explicit JSON method
	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")

	if c.status == 0 {
		c.status = http.StatusOK
	}

	// Fast path for nil data
	if data == nil {
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(nullBytes)
		return err
	}

	// Preallocation for common types
	switch v := data.(type) {
	case string:
		// String fast path (common for API error messages)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write([]byte(`"` + v + `"`))
		return err

	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8, float64, float32, bool:
		// Use fmt.Sprint for simple scalar types
		c.Response.WriteHeader(c.status)
		_, err := fmt.Fprint(c.Response, v)
		return err

	case []byte:
		// Pre-marshaled JSON
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(v)
		return err

	case map[string]interface{}:
		// Common map type - can potentially optimize further if needed
		c.Response.WriteHeader(c.status)

		// Use a buffer pool for marshaling to avoid GC pressure
		buf := getJSONBuffer()
		defer putJSONBuffer(buf)

		encoder := json.NewEncoder(buf)
		if err := encoder.Encode(v); err != nil {
			return err
		}

		_, err := c.Response.Write(buf.Bytes())
		return err
	}

	// Default path for complex types
	c.Response.WriteHeader(c.status)
	return json.NewEncoder(c.Response).Encode(data)
}

// Pool of buffers for JSON marshaling
var jsonBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func getJSONBuffer() *bytes.Buffer {
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func putJSONBuffer(buf *bytes.Buffer) {
	jsonBufferPool.Put(buf)
}

// Fast copying of header values without allocations
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (c *Context) HTML(data string) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	// Always set HTML content type since this is an explicit HTML method
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")

	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.Response.WriteHeader(c.status)
	_, err := c.Response.Write([]byte(data))
	return err
}

func (c *Context) Static(filepath string) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	http.ServeFile(c.Response, c.Request, filepath)
	return nil
}
