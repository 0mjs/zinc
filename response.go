package zinc

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"time"
)

var ErrResponseAlreadySent = errors.New("response already sent")

const (
	contentType = "Content-Type"
	jsonType    = "application/json; charset=utf-8"
	xmlType     = "application/xml; charset=utf-8"
	plainText   = "text/plain; charset=utf-8"
	htmlType    = "text/html; charset=utf-8"
	octetStream = "application/octet-stream"
)

var nullBytes = []byte("null")

var (
	plainTextHeader             = []string{plainText}
	jsonHeader                  = []string{jsonType}
	xmlHeader                   = []string{xmlType}
	htmlHeader                  = []string{htmlType}
	statusNotFoundBytes         = []byte(http.StatusText(http.StatusNotFound))
	statusMethodNotAllowedBytes = []byte(http.StatusText(http.StatusMethodNotAllowed))
)

func bodyAllowed(method string, status int) bool {
	if method == http.MethodHead {
		return false
	}
	if status >= 100 && status < 200 {
		return false
	}
	return status != http.StatusNoContent && status != http.StatusNotModified
}

func (c *Context) responseStatus() int {
	if c.status == 0 {
		return http.StatusOK
	}
	return c.status
}

func (c *Context) SetHeader(key, value string) *Context {
	c.Writer().Header().Set(key, value)
	return c
}

func (c *Context) AppendHeader(key string, values ...string) *Context {
	for _, value := range values {
		c.Writer().Header().Add(key, value)
	}
	return c
}

func (c *Context) Type(ext string) *Context {
	if ext == "" {
		return c
	}
	if ext[0] != '.' {
		ext = "." + ext
	}
	if contentType := mime.TypeByExtension(ext); contentType != "" {
		c.SetHeader(HeaderContentType, contentType)
	}
	return c
}

func (c *Context) Location(location string) *Context {
	return c.SetHeader(HeaderLocation, location)
}

func (c *Context) Vary(fields ...string) *Context {
	return c.AppendHeader(HeaderVary, fields...)
}

func (c *Context) String(data string) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	writer := c.Writer()
	header := writer.Header()
	if len(header[contentType]) == 0 {
		header[contentType] = plainTextHeader
	}

	status := c.status
	if status == 0 {
		status = http.StatusOK
	}
	method := ""
	if c.request != nil {
		method = c.request.Method
	}
	if !bodyAllowed(method, status) {
		if status != http.StatusOK {
			writer.WriteHeader(status)
		}
		return nil
	}
	if status != http.StatusOK {
		writer.WriteHeader(status)
	}

	_, err := io.WriteString(writer, data)
	return err
}

func (c *Context) Send(data any) error {
	switch value := data.(type) {
	case nil:
		return c.writeResponse(jsonType, func() error {
			_, err := c.Writer().Write(nullBytes)
			return err
		})
	case string:
		return c.String(value)
	case []byte:
		return c.Data(octetStream, value)
	default:
		return c.JSON(value)
	}
}

func (c *Context) Data(contentType string, b []byte) error {
	writer, writeBody, err := c.prepareResponse(contentType)
	if err != nil || !writeBody {
		return err
	}
	_, err = writer.Write(b)
	return err
}

func (c *Context) JSON(v any) error {
	return c.writeJSON(v, "")
}

func (c *Context) JSONPretty(v any, indent string) error {
	return c.writeJSON(v, indent)
}

func (c *Context) writeJSON(v any, indent string) error {
	writer, writeBody, err := c.prepareResponse(jsonType)
	if err != nil || !writeBody {
		return err
	}

	if v == nil {
		_, err = writer.Write(nullBytes)
		return err
	}

	if err := c.app.config.JSONCodec.Encode(writer, v, indent); err != nil {
		return err
	}
	return nil
}

func (c *Context) XML(v any) error {
	if v == nil {
		return c.writeResponse(xmlType, func() error {
			_, err := c.Writer().Write(nullBytes)
			return err
		})
	}

	var buf bytes.Buffer
	if err := xml.NewEncoder(&buf).Encode(v); err != nil {
		return err
	}
	return c.writeResponse(xmlType, func() error {
		_, err := c.Writer().Write(buf.Bytes())
		return err
	})
}

func (c *Context) HTML(data string) error {
	return c.writeResponse(htmlType, func() error {
		_, err := io.WriteString(c.Writer(), data)
		return err
	})
}

func (c *Context) Stream(contentType string, r io.Reader) error {
	return c.writeResponse(contentType, func() error {
		_, err := io.Copy(c.Writer(), r)
		return err
	})
}

func (c *Context) NoContent() error {
	if c.status == 0 || c.status == http.StatusOK {
		c.status = http.StatusNoContent
	}
	_, _, err := c.prepareResponse("")
	return err
}

func (c *Context) writeDefaultErrorResponse(status int, allowHeader string) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	writer := c.Writer()
	header := writer.Header()
	if len(allowHeader) != 0 {
		header.Set(HeaderAllow, allowHeader)
	}
	if len(header[contentType]) == 0 {
		header[contentType] = plainTextHeader
	}
	if !bodyAllowed(c.Method(), status) {
		writer.WriteHeader(status)
		return nil
	}

	writer.WriteHeader(status)
	switch status {
	case http.StatusNotFound:
		_, err := writer.Write(statusNotFoundBytes)
		return err
	case http.StatusMethodNotAllowed:
		_, err := writer.Write(statusMethodNotAllowedBytes)
		return err
	default:
		_, err := io.WriteString(writer, http.StatusText(status))
		return err
	}
}

func (c *Context) Redirect(code int, location string) error {
	if code == 0 {
		code = http.StatusFound
	}
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true
	c.Location(location)
	c.Writer().WriteHeader(code)
	return nil
}

func (c *Context) File(filePath string) error {
	return c.serveFile(filePath, nil, "")
}

func (c *Context) FileFS(filePath string, filesystem fs.FS) error {
	return c.serveFile(filePath, filesystem, "")
}

func (c *Context) Attachment(filePath string, name ...string) error {
	downloadName := filepath.Base(filePath)
	if len(name) > 0 && name[0] != "" {
		downloadName = name[0]
	}
	c.SetHeader(HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", downloadName))
	return c.File(filePath)
}

func (c *Context) Download(filePath string, name ...string) error {
	return c.Attachment(filePath, name...)
}

func (c *Context) Render(name string, data any) error {
	if c.app == nil || c.app.config.Renderer == nil {
		return errors.New("renderer is not configured")
	}
	var buf bytes.Buffer
	if err := c.app.config.Renderer.Render(&buf, name, data, c); err != nil {
		return err
	}
	return c.writeResponse(htmlType, func() error {
		_, err := c.Writer().Write(buf.Bytes())
		return err
	})
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer(), cookie)
}

func (c *Context) ClearCookie(names ...string) {
	expires := time.Unix(1, 0).UTC()
	for _, name := range names {
		http.SetCookie(c.Writer(), &http.Cookie{
			Name:    name,
			Value:   "",
			Path:    "/",
			MaxAge:  -1,
			Expires: expires,
		})
	}
}

func (c *Context) writeResponse(ct string, writeBody func() error) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true
	if ct != "" && c.Writer().Header().Get(contentType) == "" {
		c.Writer().Header().Set(contentType, ct)
	}
	status := c.responseStatus()
	if !bodyAllowed(c.Method(), status) {
		if status != http.StatusOK {
			c.Writer().WriteHeader(status)
		}
		return nil
	}
	if status != http.StatusOK {
		c.Writer().WriteHeader(status)
	}
	if writeBody != nil {
		return writeBody()
	}
	return nil
}

func (c *Context) prepareResponse(ct string) (http.ResponseWriter, bool, error) {
	if c.written {
		return nil, false, ErrResponseAlreadySent
	}
	c.written = true

	writer := c.Writer()
	if ct != "" {
		header := writer.Header()
		if len(header[contentType]) == 0 {
			switch ct {
			case plainText:
				header[contentType] = plainTextHeader
			case jsonType:
				header[contentType] = jsonHeader
			case xmlType:
				header[contentType] = xmlHeader
			case htmlType:
				header[contentType] = htmlHeader
			default:
				header.Set(contentType, ct)
			}
		}
	}

	status := c.responseStatus()
	method := ""
	if c.request != nil {
		method = c.request.Method
	}
	if !bodyAllowed(method, status) {
		if status != http.StatusOK {
			writer.WriteHeader(status)
		}
		return writer, false, nil
	}

	if status != http.StatusOK {
		writer.WriteHeader(status)
	}
	return writer, true, nil
}

func (c *Context) serveFile(filePath string, filesystem fs.FS, downloadName string) error {
	if c.written {
		return ErrResponseAlreadySent
	}
	c.written = true

	if filesystem == nil {
		http.ServeFile(c.Writer(), c.Request(), filePath)
		return nil
	}

	file, err := filesystem.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return fs.ErrInvalid
	}

	if rs, ok := file.(io.ReadSeeker); ok {
		http.ServeContent(c.Writer(), c.Request(), stat.Name(), stat.ModTime(), rs)
		return nil
	}

	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(data)
	http.ServeContent(c.Writer(), c.Request(), stat.Name(), stat.ModTime(), reader)
	return nil
}
