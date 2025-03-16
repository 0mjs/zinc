package zinc

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
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

	// Fast clear params - zero all at once using a single zero value
	for i := range c.PathParams {
		c.PathParams[i] = emptyParam
	}

	// Fast clear store - only if it has entries
	if len(c.Store) > 0 {
		// Clear map in one operation for small maps
		if len(c.Store) < 32 {
			for k := range c.Store {
				delete(c.Store, k)
			}
		} else {
			// For larger maps, replace entirely
			c.Store = make(map[string]interface{}, 16)
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
	if service, exists := c.services[name]; exists {
		return service
	}
	panic("Service '" + name + "' not found")
}

// Next calls the next middleware in the chain.
func (c *Context) Next() {
	c.index++
	if c.index < len(c.handlers) {
		c.handlers[c.index](c)
	}
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

// Body decodes the request body into the provided interface.
func (c *Context) Body(v interface{}) error {
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}
	defer c.Request.Body.Close()

	return json.NewDecoder(c.Request.Body).Decode(v)
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
		// Get app from context and validate
		if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
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
		// Get app from context and validate
		if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
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
		// Get app from context and validate
		if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
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
		// Get app from context and validate
		if app, ok := c.Store["app"].(*App); ok && app != nil && app.validator != nil {
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
