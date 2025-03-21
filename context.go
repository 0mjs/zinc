package zinc

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Context holds the context for a request.
// It is used to pass data between middleware and handlers.
type Context struct {
	Response    http.ResponseWriter
	Request     *http.Request
	PathParams  params
	QueryParams url.Values
	Method      string
	written     bool
	handlers    []Middleware
	index       int
	Store       map[string]interface{}
	status      int
	services    map[string]interface{}
	// Use preallocation for common per-request data
	paramKeys  [8]string // Cache parameter keys
	paramVals  [8]string // Cache parameter values
	paramCount int       // Number of parameters set
	// Direct reference to app to avoid map lookups
	app *App
}

type param struct {
	key   string
	value string
}

// Increase parameter storage to 8 to handle more complex routes without allocations
type params [8]param

var emptyParam param

// Add method to reset context state
var contextPool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with fixed-size maps to avoid dynamic resizing
		c := &Context{
			Store:    make(map[string]interface{}, 8), // Increased capacity
			services: make(map[string]interface{}, 4), // Increased capacity
			status:   http.StatusOK,
			index:    -1,
		}
		return c
	},
}

// NewContext creates a new context for a request.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.reset(w, r)
	return c
}

// Optimize reset to minimize operations
func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Response = w
	c.Request = r
	if r != nil {
		c.Method = r.Method
	}
	c.written = false
	c.index = -1
	c.status = http.StatusOK
	c.QueryParams = nil
	c.handlers = nil
	c.paramCount = 0
	c.app = nil

	// Fast clear params - zero all at once using a single zero value
	for i := range c.PathParams {
		c.PathParams[i] = emptyParam
	}

	// Only clear the store if it has entries
	// This optimization helps with the Hello World benchmark
	if len(c.Store) > 0 {
		// For Hello World case with just one entry, we can optimize
		if len(c.Store) == 1 {
			// Check if it's just the 'app' key which is common in the Hello World case
			if _, ok := c.Store["app"]; ok && len(c.Store) == 1 {
				delete(c.Store, "app")
				return
			}
		}

		// For other cases, clear the entire store
		for k := range c.Store {
			delete(c.Store, k)
		}
	}
}

// Add method to release context back to pool
func (c *Context) release() {
	if c == nil {
		return
	}
	// Clear references to allow GC
	c.Response = nil
	c.Request = nil
	c.handlers = nil
	c.QueryParams = nil
	c.written = false
	contextPool.Put(c)
}

// Service returns a service by name.
func (c *Context) Service(name string) interface{} {
	// First check local services map
	if service, exists := c.services[name]; exists {
		return service
	}

	// Then check app services if app reference is available
	if c.app != nil && c.app.services != nil {
		if service, exists := c.app.services[name]; exists {
			return service
		}
	}

	return nil
}

// Next calls the next middleware in the chain.
func (c *Context) Next() error {
	c.index++
	if c.index < len(c.handlers) {
		return c.handlers[c.index](c)
	}
	return nil
}

// setHandlers sets the handlers for the context.
func (c *Context) setHandlers(handlers []Middleware) {
	c.handlers = handlers
	c.index = -1
}

// Set stores a value in the context store.
func (c *Context) Set(key string, value interface{}) {
	c.Store[key] = value
}

// Get retrieves a value from the context store.
func (c *Context) Get(key string) interface{} {
	return c.Store[key]
}

// Status sets the status code for the response.
func (c *Context) Status(code int) *Context {
	c.status = code
	return c
}

// Param retrieves a path parameter by name.
func (c *Context) Param(name string) string {
	// Using a direct array access is faster than a loop for small arrays
	// Most routes have very few parameters
	if len(name) == 0 {
		return ""
	}

	// Enhanced algorithm with small string optimization
	// First check cached parameter key/value pairs
	for i := 0; i < c.paramCount; i++ {
		if c.paramKeys[i] == name {
			return c.paramVals[i]
		}
	}

	// Fast path for single-character parameter names (common in RESTful APIs)
	if len(name) == 1 {
		for i := range c.PathParams {
			if c.PathParams[i].key == name {
				// Cache for future lookups
				if c.paramCount < len(c.paramKeys) {
					c.paramKeys[c.paramCount] = name
					c.paramVals[c.paramCount] = c.PathParams[i].value
					c.paramCount++
				}
				return c.PathParams[i].value
			}
			if c.PathParams[i].key == "" {
				break // End of params
			}
		}
		return ""
	}

	// General case for multi-character names
	for i := range c.PathParams {
		if c.PathParams[i].key == name {
			// Cache for future lookups
			if c.paramCount < len(c.paramKeys) {
				c.paramKeys[c.paramCount] = name
				c.paramVals[c.paramCount] = c.PathParams[i].value
				c.paramCount++
			}
			return c.PathParams[i].value
		}
		if c.PathParams[i].key == "" {
			break // End of params
		}
	}
	return ""
}

// Query retrieves a query parameter by name (with lazy loading)
func (c *Context) Query(name string) string {
	// Lazy loading of query params with zero allocations for common cases
	if c.QueryParams == nil && c.Request != nil {
		if c.Request.URL.RawQuery == "" {
			return ""
		}
		c.QueryParams = c.Request.URL.Query()
	}
	if c.QueryParams != nil {
		return c.QueryParams.Get(name)
	}
	return ""
}

// Optimized method to check if a query parameter exists (avoids additional parsing)
func (c *Context) HasQuery(name string) bool {
	if c.QueryParams == nil && c.Request != nil {
		if c.Request.URL.RawQuery == "" {
			return false
		}
		c.QueryParams = c.Request.URL.Query()
	}
	if c.QueryParams != nil {
		_, exists := c.QueryParams[name]
		return exists
	}
	return false
}

// Body returns the request body as a string.
func (c *Context) Body() (string, error) {
	// Use direct app reference if available, otherwise fallback to Store
	var app *App
	if c.app != nil {
		app = c.app
	} else if a, ok := c.Get("app").(*App); ok {
		app = a
	} else {
		return "", errors.New("unable to get app reference")
	}

	// Check for body size limit
	if app.config.BodyLimit > 0 {
		r := io.LimitReader(c.Request.Body, app.config.BodyLimit)
		body, err := io.ReadAll(r)
		if err != nil {
			return "", err
		}

		// Check if the limit was exceeded
		if int64(len(body)) >= app.config.BodyLimit {
			return "", fmt.Errorf("request body size exceeds the limit of %d bytes", app.config.BodyLimit)
		}

		return string(body), nil
	}

	// No limit set, read entire body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// BodyParser parses the request body into a provided struct.
func (c *Context) BodyParser(out interface{}) error {
	// Use direct app reference if available, otherwise fallback to Store
	var app *App
	if c.app != nil {
		app = c.app
	} else if a, ok := c.Get("app").(*App); ok {
		app = a
	} else {
		return errors.New("unable to get app reference")
	}

	// Determine content type
	contentType := c.Request.Header.Get("Content-Type")

	// Check for body size limit
	var body []byte
	var err error

	if app.config.BodyLimit > 0 {
		r := io.LimitReader(c.Request.Body, app.config.BodyLimit+1) // +1 to detect if limit is exceeded
		body, err = io.ReadAll(r)
		if err != nil {
			return err
		}

		// Check if the limit was exceeded
		if int64(len(body)) > app.config.BodyLimit {
			return fmt.Errorf("request body size exceeds the limit of %d bytes", app.config.BodyLimit)
		}
	} else {
		// No limit set, read entire body
		body, err = io.ReadAll(c.Request.Body)
		if err != nil {
			return err
		}
	}

	// Parse based on content type
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		if err := json.Unmarshal(body, out); err != nil {
			return err
		}
	case strings.HasPrefix(contentType, "application/xml"):
		if err := xml.Unmarshal(body, out); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Validate the struct after parsing
	if app.validator != nil {
		if errors := app.validator.Validate(out); len(errors) > 0 {
			return errors
		}
	}

	return nil
}

// setParam sets a path parameter with optimized allocation
func (c *Context) setParam(key, value string) {
	// Fast path for the first parameters (most common case)
	for i := range c.PathParams {
		if c.PathParams[i].key == "" {
			c.PathParams[i] = param{key: key, value: value}

			// Cache the parameter for fast access
			if c.paramCount < len(c.paramKeys) {
				c.paramKeys[c.paramCount] = key
				c.paramVals[c.paramCount] = value
				c.paramCount++
			}
			return
		}
	}
}

// BindOptions holds options for binding data
type BindOptions struct {
	// DisableValidation disables validation after binding
	DisableValidation bool
	// DisableUnknownFields disables errors for unknown fields
	DisableUnknownFields bool
}

// BindJSON binds the JSON request body to the provided struct and validates it
func (c *Context) BindJSON(v interface{}, opts ...*BindOptions) error {
	if c.Request.Body == nil {
		return errors.New("request body is empty")
	}
	defer c.Request.Body.Close()

	options := getBindOptions(opts)
	decoder := json.NewDecoder(c.Request.Body)
	if !options.DisableUnknownFields {
		decoder.DisallowUnknownFields()
	}

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to bind JSON: %w", err)
	}

	if !options.DisableValidation {
		// Use direct app reference if available, otherwise fallback to Store
		if c.app != nil && c.app.validator != nil {
			if errors := c.app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		} else if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
			if errors := app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		}
	}

	return nil
}

// BindXML binds the XML request body to the provided struct and validates it
func (c *Context) BindXML(v interface{}, opts ...*BindOptions) error {
	if c.Request.Body == nil {
		return errors.New("request body is empty")
	}
	defer c.Request.Body.Close()

	options := getBindOptions(opts)
	if err := xml.NewDecoder(c.Request.Body).Decode(v); err != nil {
		return fmt.Errorf("failed to bind XML: %w", err)
	}

	if !options.DisableValidation {
		// Use direct app reference if available, otherwise fallback to Store
		if c.app != nil && c.app.validator != nil {
			if errors := c.app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		} else if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
			if errors := app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		}
	}

	return nil
}

// BindForm binds form data to the provided struct and validates it
func (c *Context) BindForm(v interface{}, opts ...*BindOptions) error {
	options := getBindOptions(opts)

	// Parse form if not already parsed
	if c.Request.Form == nil {
		if err := c.Request.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form: %w", err)
		}
	}

	if err := bindData(v, c.Request.Form, "form"); err != nil {
		return err
	}

	if !options.DisableValidation {
		// Use direct app reference if available, otherwise fallback to Store
		if c.app != nil && c.app.validator != nil {
			if errors := c.app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		} else if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
			if errors := app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		}
	}

	return nil
}

// BindQuery binds query parameters to the provided struct and validates it
func (c *Context) BindQuery(v interface{}, opts ...*BindOptions) error {
	options := getBindOptions(opts)

	// Ensure query params are parsed
	if c.QueryParams == nil {
		c.QueryParams = c.Request.URL.Query()
	}

	if err := bindData(v, c.QueryParams, "query"); err != nil {
		return err
	}

	if !options.DisableValidation {
		// Use direct app reference if available, otherwise fallback to Store
		if c.app != nil && c.app.validator != nil {
			if errors := c.app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		} else if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
			if errors := app.validator.Validate(v); len(errors) > 0 {
				return errors
			}
		}
	}

	return nil
}

// Bind automatically binds request data based on Content-Type
func (c *Context) Bind(v interface{}, opts ...*BindOptions) error {
	contentType := c.Request.Header.Get("Content-Type")

	// Extract the MIME type
	if idx := strings.IndexByte(contentType, ';'); idx >= 0 {
		contentType = contentType[0:idx]
	}
	contentType = strings.TrimSpace(contentType)

	// Bind based on Content-Type
	switch contentType {
	case "application/json":
		return c.BindJSON(v, opts...)
	case "application/xml", "text/xml":
		return c.BindXML(v, opts...)
	case "application/x-www-form-urlencoded", "multipart/form-data":
		return c.BindForm(v, opts...)
	default:
		// Default to JSON for common API usage
		return c.BindJSON(v, opts...)
	}
}

// FormFile returns the first file from the multipart form
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	file, header, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	file.Close()
	return header, nil
}

// FormFiles returns all files with the given field name
func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}

	fhs := c.Request.MultipartForm.File[name]
	if len(fhs) == 0 {
		return nil, errors.New("no files with specified name found")
	}

	return fhs, nil
}

// SaveFile saves a file from a multipart form to the specified destination
func (c *Context) SaveFile(fileHeader *multipart.FileHeader, dst string) error {
	src, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// Helper function to get bind options with defaults
func getBindOptions(opts []*BindOptions) *BindOptions {
	if len(opts) > 0 && opts[0] != nil {
		return opts[0]
	}
	return &BindOptions{}
}

// Helper function to bind data to struct from form or query values
func bindData(ptr interface{}, data map[string][]string, tag string) error {
	if ptr == nil {
		return errors.New("binding element must be a pointer")
	}

	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("binding element must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get the tag value
		tagValue := field.Tag.Get(tag)
		if tagValue == "" {
			// Use lowercase field name if no tag
			tagValue = strings.ToLower(field.Name)
		} else if tagValue == "-" {
			// Skip this field if tag is "-"
			continue
		}

		// Extract the field name (before any comma)
		if idx := strings.Index(tagValue, ","); idx >= 0 {
			tagValue = tagValue[:idx]
		}

		// Check if the field exists in the data
		inputValue, exists := data[tagValue]
		if !exists || len(inputValue) == 0 {
			continue
		}

		// Set the field value based on its type
		if err := setFieldValue(fieldValue, inputValue[0]); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.Name, err)
		}
	}

	return nil
}

// Helper function to set a field value based on its type
func setFieldValue(value reflect.Value, input string) error {
	if !value.CanSet() {
		return nil
	}

	switch value.Kind() {
	case reflect.String:
		value.SetString(input)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(input, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(input, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return err
		}
		value.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(input)
		if err != nil {
			return err
		}
		value.SetBool(val)
	case reflect.Slice:
		// Only support []string for now
		if value.Type().Elem().Kind() == reflect.String {
			value.Set(reflect.ValueOf([]string{input}))
		}
	default:
		return fmt.Errorf("unsupported field type: %s", value.Kind())
	}

	return nil
}

// IP returns the client's IP address.
func (c *Context) IP() string {
	// Get app reference for accessing config
	app, ok := c.Get("app").(*App)
	if !ok {
		return c.RemoteIP()
	}

	// If trusted proxy checking is enabled
	if app.config.EnableTrustedProxyCheck {
		// Get the remote IP from the request
		remoteIP := c.RemoteIP()

		// Check if the remote IP is in the trusted proxies list
		if isTrustedProxy(remoteIP, app.config.TrustedProxies) {
			// If it's trusted, get the client IP from the proxy header
			if proxyHeader := app.config.ProxyHeader; proxyHeader != "" {
				if clientIP := c.Request.Header.Get(proxyHeader); clientIP != "" {
					// The header might contain multiple IPs (e.g. "client, proxy1, proxy2")
					// We need to get the first one which is the client IP
					if commaIndex := strings.Index(clientIP, ","); commaIndex > 0 {
						return strings.TrimSpace(clientIP[:commaIndex])
					}
					return strings.TrimSpace(clientIP)
				}
			}
		}

		// Fallback to remote IP if not trusted or no proxy header
		return remoteIP
	}

	// If proxy checking is disabled, return the remote IP
	return c.RemoteIP()
}

// RemoteIP returns the remote IP address of the request.
func (c *Context) RemoteIP() string {
	// Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// If we can't split it, just return it as is
		return c.Request.RemoteAddr
	}
	return ip
}

// isTrustedProxy checks if the given IP is in the list of trusted proxies.
func isTrustedProxy(ip string, trustedProxies []string) bool {
	if len(trustedProxies) == 0 {
		return false
	}

	for _, trustedIP := range trustedProxies {
		if strings.Contains(trustedIP, "/") {
			// It's a CIDR block
			_, ipNet, err := net.ParseCIDR(trustedIP)
			if err != nil {
				continue
			}

			parsedIP := net.ParseIP(ip)
			if parsedIP != nil && ipNet.Contains(parsedIP) {
				return true
			}
		} else {
			// It's a single IP
			if ip == trustedIP {
				return true
			}
		}
	}

	return false
}

// Version returns the current version of the Zinc framework.
func (c *Context) Version() string {
	return Version
}

// ContextServiceOf is a helper function to retrieve a service by type from a Context
// Usage example: service, ok := zinc.ContextServiceOf[*UserService](c)
func ContextServiceOf[T any](c *Context) (service T, ok bool) {
	// Check if we have direct app access
	if c.app != nil {
		// Get the type we're looking for
		typ := reflect.TypeOf((*T)(nil)).Elem()

		// Try to get the service directly from the app's typedServices map
		if s := c.app.typedServices[typ]; s != nil {
			if svc, isOk := s.(T); isOk {
				return svc, true
			}
		}

		// If not found by type, try using type name string lookup
		if s := c.app.services[typ.String()]; s != nil {
			if svc, isOk := s.(T); isOk {
				return svc, true
			}
		}
	}

	// As a fallback, try to get the app from the Store if it's not directly set
	if c.app == nil {
		if app, exists := c.Store["app"].(*App); exists && app != nil {
			// Same lookup logic but with the app from Store
			typ := reflect.TypeOf((*T)(nil)).Elem()

			if s := app.typedServices[typ]; s != nil {
				if svc, isOk := s.(T); isOk {
					return svc, true
				}
			}

			if s := app.services[typ.String()]; s != nil {
				if svc, isOk := s.(T); isOk {
					return svc, true
				}
			}
		}
	}

	var zero T
	return zero, false
}
