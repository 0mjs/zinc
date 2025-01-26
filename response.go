package zinc

import (
	"encoding/json"
	"errors"
	"net/http"
)

var ErrResponseAlreadySent = errors.New("response already sent")

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
		c.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write([]byte(v))
		return err
	case []byte:
		c.Response.Header().Set("Content-Type", "application/octet-stream")
		c.Response.WriteHeader(c.status)
		_, err := c.Response.Write(v)
		return err
	case nil:
		c.Response.WriteHeader(c.status)
		return nil
	default:
		c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
		c.Response.WriteHeader(c.status)
		return json.NewEncoder(c.Response).Encode(data)
	}
}

func (c *Context) JSON(data interface{}) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.Response.Header().Set("Content-Type", "application/json; charset=utf-8")
	c.Response.Header().Set("X-Content-Type-Options", "nosniff")

	c.Response.WriteHeader(c.status)

	if data == nil {
		_, err := c.Response.Write([]byte("null"))
		return err
	}

	return json.NewEncoder(c.Response).Encode(data)
}

func (c *Context) HTML(data string) error {
	c.written = true
	c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")

	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.Response.WriteHeader(c.status)
	_, err := c.Response.Write([]byte(data))
	return err
}

func (c *Context) Static(filepath string) error {
	c.written = true

	http.ServeFile(c.Response, c.Request, filepath)
	return nil
}
