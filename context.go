package zinc

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"sync"
)

// Context holds the context for a request.
// It is used to pass data between middleware and handlers.
type Context struct {
	Response    http.ResponseWriter
	Request     *http.Request
	PathParams  map[string]string
	QueryParams url.Values
	Method      string
	written     bool
	handlers    []Middleware
	index       int
	Store       map[string]interface{}
	status      int
	services    map[string]interface{}
}

// Pool of contexts to reduce allocations
var contextPool = sync.Pool{
	New: func() interface{} {
		return &Context{
			PathParams: make(map[string]string, 4), // Pre-allocate with reasonable size
			Store:      make(map[string]interface{}, 4),
		}
	},
}

// NewContext creates a new context for a request.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.reset(w, r)
	return c
}

// Add method to reset context state
func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.Response = w
	c.Request = r
	c.QueryParams = r.URL.Query()
	c.Method = r.Method
	c.written = false
	c.index = -1
	c.status = http.StatusOK

	// Clear maps instead of reallocating
	for k := range c.PathParams {
		delete(c.PathParams, k)
	}
	for k := range c.Store {
		delete(c.Store, k)
	}
}

// Add method to release context back to pool
func (c *Context) release() {
	c.Response = nil
	c.Request = nil
	c.handlers = nil
	c.QueryParams = nil
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
	return c.PathParams[name]
}

// Query retrieves a query parameter by name.
func (c *Context) Query(name string) string {
	return c.QueryParams.Get(name)
}

// Body decodes the request body into the provided interface.
func (c *Context) Body(v interface{}) error {
	if c.Request.Body == nil {
		return errors.New("request body is nil")
	}
	defer c.Request.Body.Close()

	return json.NewDecoder(c.Request.Body).Decode(v)
}
