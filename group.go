package zinc

import (
	"path"
	"strings"
)

// Group represents a group of routes with a common prefix
// and potentially shared middleware
type Group struct {
	app        *App
	prefix     string
	middleware []Middleware
}

// NewGroup creates a new route group
func NewGroup(app *App, prefix string) *Group {
	// Ensure prefix starts with /
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return &Group{
		app:        app,
		prefix:     prefix,
		middleware: make([]Middleware, 0),
	}
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...Middleware) *Group {
	g.middleware = append(g.middleware, middleware...)
	return g
}

// Group creates a new sub-group with an additional prefix
func (g *Group) Group(prefix string) *Group {
	// Ensure prefix starts with /
	if prefix != "" && !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	fullPrefix := path.Join(g.prefix, prefix)

	return &Group{
		app:        g.app,
		prefix:     fullPrefix,
		middleware: append([]Middleware{}, g.middleware...),
	}
}

// Get registers a route for the GET HTTP method
func (g *Group) Get(path string, handlers ...RouteHandler) error {
	return g.Add(MethodGet, path, handlers...)
}

// Post registers a route for the POST HTTP method
func (g *Group) Post(path string, handlers ...RouteHandler) error {
	return g.Add(MethodPost, path, handlers...)
}

// Put registers a route for the PUT HTTP method
func (g *Group) Put(path string, handlers ...RouteHandler) error {
	return g.Add(MethodPut, path, handlers...)
}

// Delete registers a route for the DELETE HTTP method
func (g *Group) Delete(path string, handlers ...RouteHandler) error {
	return g.Add(MethodDelete, path, handlers...)
}

// Patch registers a route for the PATCH HTTP method
func (g *Group) Patch(path string, handlers ...RouteHandler) error {
	return g.Add(MethodPatch, path, handlers...)
}

// Head registers a route for the HEAD HTTP method
func (g *Group) Head(path string, handlers ...RouteHandler) error {
	return g.Add(MethodHead, path, handlers...)
}

// Options registers a route for the OPTIONS HTTP method
func (g *Group) Options(path string, handlers ...RouteHandler) error {
	return g.Add(MethodOptions, path, handlers...)
}

// Connect registers a route for the CONNECT HTTP method
func (g *Group) Connect(path string, handlers ...RouteHandler) error {
	return g.Add(MethodConnect, path, handlers...)
}

// Trace registers a route for the TRACE HTTP method
func (g *Group) Trace(path string, handlers ...RouteHandler) error {
	return g.Add(MethodTrace, path, handlers...)
}

// Add adds a route to the group with the given method and path
func (g *Group) Add(method, path string, handlers ...RouteHandler) error {
	// Ensure path starts with /
	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	// Combine group prefix with route path
	fullPath := joinPaths(g.prefix, path)

	// Combine group middleware with route handlers
	allHandlers := make([]RouteHandler, len(g.middleware)+len(handlers))

	// Copy middleware as route handlers
	for i, mw := range g.middleware {
		middleware := mw // Create a local copy to avoid closure issues
		allHandlers[i] = func(c *Context) error {
			return middleware(c)
		}
	}

	// Copy route handlers
	copy(allHandlers[len(g.middleware):], handlers)

	// Add route to app
	return g.app.router.Add(method, fullPath, allHandlers...)
}

// Helper function to join two path segments correctly
func joinPaths(a, b string) string {
	if a == "" {
		return b
	}
	if b == "" {
		return a
	}

	aSlash := strings.HasSuffix(a, "/")
	bSlash := strings.HasPrefix(b, "/")

	if aSlash && bSlash {
		return a + b[1:]
	} else if !aSlash && !bSlash {
		return a + "/" + b
	}

	return a + b
}
