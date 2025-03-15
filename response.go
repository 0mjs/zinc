package zinc

import (
	"encoding/json"
	"errors"
	"net/http"
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
	textHeaders = http.Header{
		"Content-Type": []string{textContentType},
	}
	htmlHeaders = http.Header{
		"Content-Type": []string{htmlContentType},
	}
	octetHeaders = http.Header{
		"Content-Type": []string{octetContentType},
	}
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
			copyHeader(c.Response.Header(), textHeaders)
			c.Response.WriteHeader(c.status)
			_, err := c.Response.Write(helloWorldBytes)
			return err
		}

		copyHeader(c.Response.Header(), textHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write([]byte(v))
		return err

	case []byte:
		copyHeader(c.Response.Header(), octetHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(v)
		return err

	case nil:
		copyHeader(c.Response.Header(), jsonHeaders)
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(nullBytes)
		return err

	default:
		copyHeader(c.Response.Header(), jsonHeaders)
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

	copyHeader(c.Response.Header(), jsonHeaders)
	c.Response.WriteHeader(c.status)

	if data == nil {
		_, err := c.Response.Write(nullBytes)
		return err
	}

	return json.NewEncoder(c.Response).Encode(data)
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
