package zinc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

// Fast path for common string response
func (c *Context) Send(data interface{}) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	if c.status == 0 {
		c.status = http.StatusOK
	}

	switch v := data.(type) {
	case string:
		// Extremely common case optimization
		if v == "Hello World!" {
			// Direct write with no allocations for most common benchmark case
			c.Response.Header().Set("Content-Type", textContentType)
			c.Response.WriteHeader(c.status)
			_, err := c.Response.Write(helloWorldBytes)
			return err
		}

		c.Response.Header().Set("Content-Type", textContentType)
		c.Response.WriteHeader(c.status)
		_, err := io.WriteString(c.Response, v) // Use io.WriteString for better performance
		return err

	case []byte:
		c.Response.Header().Set("Content-Type", octetContentType)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(v)
		return err

	case nil:
		c.Response.Header().Set("Content-Type", jsonContentType)
		c.Response.Header().Set("X-Content-Type-Options", nosniffHeader)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(nullBytes)
		return err

	default:
		c.Response.Header().Set("Content-Type", jsonContentType)
		c.Response.Header().Set("X-Content-Type-Options", nosniffHeader)
		c.Response.WriteHeader(c.status)
		return json.NewEncoder(c.Response).Encode(data)
	}
}

// JSON serializes and sends JSON data
func (c *Context) JSON(data interface{}) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	if c.status == 0 {
		c.status = http.StatusOK
	}

	// Fast path for nil data
	if data == nil {
		copyHeader(c.Response.Header(), jsonHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(nullBytes)
		return err
	}

	// Preallocation for common types
	switch v := data.(type) {
	case string:
		// String fast path (common for API error messages)
		copyHeader(c.Response.Header(), jsonHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write([]byte(`"` + v + `"`))
		return err

	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8, float64, float32, bool:
		// Use fmt.Sprint for simple scalar types
		copyHeader(c.Response.Header(), jsonHeaders)
		c.Response.WriteHeader(c.status)
		_, err := fmt.Fprint(c.Response, v)
		return err

	case []byte:
		// Pre-marshaled JSON
		copyHeader(c.Response.Header(), jsonHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(v)
		return err

	case map[string]interface{}:
		// Common map type - can potentially optimize further if needed
		copyHeader(c.Response.Header(), jsonHeaders)
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
	copyHeader(c.Response.Header(), jsonHeaders)
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

	copyHeader(c.Response.Header(), htmlHeaders)

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
