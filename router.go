package zinc

import (
	"fmt"
	"strings"
	"sync"
)

const (
	MethodGet     = "GET"     // RFC 7231, 4.3.1
	MethodHead    = "HEAD"    // RFC 7231, 4.3.2
	MethodPost    = "POST"    // RFC 7231, 4.3.3
	MethodPut     = "PUT"     // RFC 7231, 4.3.4
	MethodPatch   = "PATCH"   // RFC 5789
	MethodDelete  = "DELETE"  // RFC 7231, 4.3.5
	MethodConnect = "CONNECT" // RFC 7231, 4.3.6
	MethodOptions = "OPTIONS" // RFC 7231, 4.3.7
	MethodTrace   = "TRACE"   // RFC 7231, 4.3.8
	methodUse     = "USE"
)

const (
	MIMETextXML               = "text/xml"
	MIMETextHTML              = "text/html"
	MIMETextPlain             = "text/plain"
	MIMEApplicationXML        = "application/xml"
	MIMEApplicationJSON       = "application/json"
	MIMEApplicationJavaScript = "application/javascript"
	MIMEApplicationForm       = "application/x-www-form-urlencoded"
	MIMEOctetStream           = "application/octet-stream"
	MIMEMultipartForm         = "multipart/form-data"

	MIMETextXMLCharsetUTF8               = "text/xml; charset=utf-8"
	MIMETextHTMLCharsetUTF8              = "text/html; charset=utf-8"
	MIMETextPlainCharsetUTF8             = "text/plain; charset=utf-8"
	MIMEApplicationXMLCharsetUTF8        = "application/xml; charset=utf-8"
	MIMEApplicationJSONCharsetUTF8       = "application/json; charset=utf-8"
	MIMEApplicationJavaScriptCharsetUTF8 = "application/javascript; charset=utf-8"
)

const (
	paramIdentifier    = ':'
	wildcardIdentifier = '*'
)

type RouteNode struct {
	path     string
	part     string
	children []*RouteNode
	handler  RouteHandler
	handlers map[string]RouteHandler // Map of handlers by method
	isParam  bool
	isWild   bool
	method   string
}

type Route struct {
	path    string
	handler RouteHandler
	method  string
	parts   []string
}

type Middleware func(c *Context) error

type Router struct {
	routes     map[string]map[string]*Route
	router     *RouteNode
	middleware []Middleware
	cache      *RouteCache
	config     *Config // Reference to app config
}

// Use a fixed array to avoid allocations for common path lengths
func getPathParts(path string) []string {
	// Use a local fixed-size array for most paths (which are short)
	var fixedParts [8]string
	parts := fixedParts[:0]

	// Fast path for empty and root paths
	if path == "" || path == "/" {
		return parts
	}

	// Fast path for single component path (common case)
	if path[0] == '/' && !strings.ContainsRune(path[1:], '/') {
		if len(path) > 1 {
			parts = append(parts, path[1:])
		}
		return parts
	}

	// Skip the leading slash
	if path[0] == '/' {
		path = path[1:]
	}

	// Optimization: use pre-allocations for common path patterns
	pathLen := len(path)

	// Fast path for 1-2 slashes (most common case)
	slashCount := 0
	for i := range path {
		if path[i] == '/' {
			slashCount++
		}
	}

	// Preallocate exact capacity
	if cap(parts) < slashCount+1 {
		// Rare case for extremely deep paths
		parts = make([]string, 0, slashCount+1)
	}

	// Fast split without regexp
	start := 0
	for i := 0; i < pathLen; i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}

	// Add final part if path doesn't end with slash
	if start < pathLen {
		parts = append(parts, path[start:])
	}

	return parts
}

// Add adds a route to the router
func (r *Router) Add(method, path string, handlers ...RouteHandler) error {
	// Ensure we have a handler
	if len(handlers) == 0 {
		return fmt.Errorf("no handler provided for %s %s", method, path)
	}

	// Normalize path
	path = r.normalizePath(path)

	// Store original path for case-sensitive routing
	originalPath := path

	// Apply case insensitivity if configured
	lowerPath := path
	if r.config != nil && !r.config.CaseSensitive {
		lowerPath = strings.ToLower(path)
	}

	// Handle strict routing - store both with and without trailing slash if needed
	pathWithoutSlash := path
	if len(path) > 1 && path[len(path)-1] == '/' {
		pathWithoutSlash = path[:len(path)-1]
	}

	pathWithSlash := path
	if len(path) > 0 && path[len(path)-1] != '/' {
		pathWithSlash = path + "/"
	}

	// Initialize method map if it doesn't exist
	if r.routes == nil {
		r.routes = make(map[string]map[string]*Route)
	}

	// Initialize routes for this method if they don't exist
	if _, ok := r.routes[method]; !ok {
		r.routes[method] = make(map[string]*Route)
	}

	// Parse path into parts
	parts := getPathParts(path)

	// Get the main handler
	mainHandler := handlers[len(handlers)-1]

	// Handle middleware chain if there are multiple handlers
	if len(handlers) > 1 {
		// Convert RouteHandlers to Middleware
		middleware := make([]Middleware, len(handlers)-1)
		for i, h := range handlers[:len(handlers)-1] {
			middleware[i] = Middleware(h)
		}

		// Create a chain handler that runs middleware then the main handler
		chainHandler := mainHandler
		mainHandler = func(c *Context) error {
			c.setHandlers(append(middleware, func(c *Context) error {
				return chainHandler(c)
			}))
			return c.Next()
		}
	}

	// Create route object
	route := &Route{
		path:    path,
		handler: mainHandler,
		method:  method,
		parts:   parts,
	}

	// Add to static routes map
	if strings.IndexByte(path, ':') < 0 && strings.IndexByte(path, '*') < 0 {
		// Always add the original path
		r.routes[method][originalPath] = route

		// If case insensitive, add lowercase version too (if different)
		if r.config != nil && !r.config.CaseSensitive && lowerPath != originalPath {
			r.routes[method][lowerPath] = route
		}

		// If not strict routing, add both with and without trailing slash
		if r.config != nil && !r.config.StrictRouting {
			if pathWithoutSlash != originalPath {
				r.routes[method][pathWithoutSlash] = route
			}
			if pathWithSlash != originalPath {
				r.routes[method][pathWithSlash] = route
			}
		}
	}

	// Add to the trie for dynamic route matching
	if r.cache == nil || strings.IndexByte(path, ':') >= 0 || strings.IndexByte(path, '*') >= 0 {
		current := r.router
		if current == nil {
			current = &RouteNode{
				method: "", // Root node has no method
			}
			r.router = current
		}

		for i, part := range parts {
			isParam := false
			isWild := false

			if len(part) > 0 {
				switch part[0] {
				case paramIdentifier:
					isParam = true
					part = part[1:]
				case wildcardIdentifier:
					isWild = true
					part = "*"
				}
			}

			child := current.findChild(part, isParam, isWild)
			if child == nil {
				// Preallocate handlers map with capacity 1-4 based on likely needs
				handlerCapacity := 1
				if method == MethodGet || method == MethodPost {
					handlerCapacity = 4 // More common to have multiple handlers for these
				}

				child = &RouteNode{
					part:     part,
					isParam:  isParam,
					isWild:   isWild,
					method:   "",                                             // Intermediate nodes have no method
					handlers: make(map[string]RouteHandler, handlerCapacity), // Initialize handlers map
				}
				current.children = append(current.children, child)
			}

			if i == len(parts)-1 {
				// This is a leaf node - store the handler for this method
				if child.handlers == nil {
					// Preallocate with capacity 1-4 based on likely needs
					handlerCapacity := 1
					if method == MethodGet || method == MethodPost {
						handlerCapacity = 4 // More common to have multiple handlers for these
					}
					child.handlers = make(map[string]RouteHandler, handlerCapacity)
				}
				child.handlers[method] = mainHandler
				child.path = path
				child.method = method       // Keep this for backward compatibility
				child.handler = mainHandler // Keep this for backward compatibility
			}

			current = child
		}
	}

	return nil
}

type routeCacheKey struct {
	method string
	path   string
}

type routeCacheEntry struct {
	handler RouteHandler
	context *Context
}

// Add LRU cache for routes
type RouteCache struct {
	cache map[routeCacheKey]routeCacheEntry
	mu    sync.RWMutex
	size  int
	keys  []routeCacheKey // Track keys for simple LRU eviction
}

func NewRouteCache(size int) *RouteCache {
	return &RouteCache{
		cache: make(map[routeCacheKey]routeCacheEntry, size),
		size:  size,
		keys:  make([]routeCacheKey, 0, size),
	}
}

func (rc *RouteCache) get(key routeCacheKey) (routeCacheEntry, bool) {
	rc.mu.RLock()
	entry, ok := rc.cache[key]
	rc.mu.RUnlock()
	return entry, ok
}

func (rc *RouteCache) set(key routeCacheKey, entry routeCacheEntry) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// If already exists, just update
	if _, exists := rc.cache[key]; exists {
		rc.cache[key] = entry
		return
	}

	// If cache is full, evict one entry
	if len(rc.cache) >= rc.size {
		// Simple FIFO eviction strategy
		if len(rc.keys) > 0 {
			// Evict oldest entry
			oldKey := rc.keys[0]
			delete(rc.cache, oldKey)
			// Remove the key
			rc.keys = rc.keys[1:]
		}
	}

	// Add new entry
	rc.cache[key] = entry
	rc.keys = append(rc.keys, key)
}

func (r *Router) Find(method, path string) (RouteHandler, *Context) {
	// Apply config settings to path if needed
	originalPath := path

	// Handle case sensitivity
	if r.config != nil && !r.config.CaseSensitive {
		path = strings.ToLower(path)
	}

	// Handle strict routing - only modify path if strict routing is disabled
	trailingSlashModified := false
	if r.config != nil && !r.config.StrictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
		trailingSlashModified = true
	}

	// Static route fast path - most common case first
	if routes, ok := r.routes[method]; ok {
		// Try with original path first
		if route, ok := routes[originalPath]; ok {
			// Create a context with empty params - most common case for APIs
			ctx := &Context{PathParams: params{}}
			return route.handler, ctx
		}

		// If original path didn't match and we modified the path, try with modified path
		// But only if strict routing is disabled or the modification wasn't due to trailing slash
		if originalPath != path && (!r.config.StrictRouting || !trailingSlashModified) {
			if route, ok := routes[path]; ok {
				// Create a context with empty params - most common case for APIs
				ctx := &Context{PathParams: params{}}
				return route.handler, ctx
			}
		}
	}

	// Cache lookup
	key := routeCacheKey{method, path}
	if r.cache != nil {
		if entry, ok := r.cache.get(key); ok {
			return entry.handler, entry.context
		}
	}

	// Perform dynamic route matching
	// Create a context only when needed
	ctx := &Context{PathParams: params{}}

	// In-place dynamic route matching
	// Fast path for methods with no dynamic routes
	current := r.router
	if current == nil {
		return nil, nil
	}

	// Quick exit for empty trees
	if len(current.children) == 0 {
		return nil, nil
	}

	// Parse path parts only if needed for dynamic matching
	parts := getPathParts(path)

	// First try fast exact match at root level (most common)
	for i := 0; i < len(current.children); i++ {
		child := current.children[i]
		if !child.isParam && !child.isWild && len(parts) > 0 && child.part == parts[0] {
			remaining := parts[1:]
			if match := child.find(remaining, ctx, method); match != nil {
				// Get the handler for this method
				var handler RouteHandler
				if match.handlers != nil {
					if h, ok := match.handlers[method]; ok {
						handler = h
					}
				}
				// Fallback to the old handler field
				if handler == nil && match.handler != nil && (match.method == method || match.method == "") {
					handler = match.handler
				}

				if handler != nil {
					// Cache the result - but only allocate a new context for the cache
					if r.cache != nil {
						// Create a copy of the context's path params for caching
						cacheCtx := &Context{PathParams: ctx.PathParams}
						r.cache.set(key, routeCacheEntry{handler, cacheCtx})
					}
					return handler, ctx
				}
			}
		}
	}

	// Then check all routes (including params/wildcards)
	if match := current.find(parts, ctx, method); match != nil {
		// Get the handler for this method
		var handler RouteHandler
		if match.handlers != nil {
			if h, ok := match.handlers[method]; ok {
				handler = h
			}
		}
		// Fallback to the old handler field
		if handler == nil && match.handler != nil && (match.method == method || match.method == "") {
			handler = match.handler
		}

		if handler != nil {
			// Cache the result - but only allocate a new context for the cache
			if r.cache != nil {
				// Create a copy of the context's path params for caching
				cacheCtx := &Context{PathParams: ctx.PathParams}
				r.cache.set(key, routeCacheEntry{handler, cacheCtx})
			}
			return handler, ctx
		}
	}

	return nil, nil
}

func (r *Router) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

func chain(handlers []RouteHandler) RouteHandler {
	// Optimization for single handler case (common)
	if len(handlers) == 1 {
		return handlers[0]
	}

	// Optimization for two handlers case (also common)
	if len(handlers) == 2 {
		h1, h2 := handlers[0], handlers[1]
		return func(c *Context) error {
			if err := h1(c); err != nil {
				return err
			}
			if !c.written {
				return h2(c)
			}
			return nil
		}
	}

	// For 3+ handlers, use the general case
	return func(c *Context) error {
		for i, handler := range handlers {
			if err := handler(c); err != nil {
				return err
			}

			if c.written {
				// Fast return
				return nil
			}

			// Add performance hint for the runtime
			// Using simple check to help branch prediction
			if i >= len(handlers)-2 {
				// Last two handlers, no need for complex checks
				break
			}
		}

		// Handle the last handler(s) directly to avoid loop checks
		if len(handlers) > 0 && !c.written {
			return handlers[len(handlers)-1](c)
		}
		return nil
	}
}

func (n *RouteNode) find(parts []string, ctx *Context, method string) *RouteNode {
	if len(parts) == 0 {
		// Check if this node has a handler for the requested method
		if n.handlers != nil {
			if handler, ok := n.handlers[method]; ok && handler != nil {
				// Found a handler for this method
				return n
			}
		}
		// Fallback to the old method for backward compatibility
		if n.handler != nil && (n.method == method || n.method == "") {
			return n
		}
		return nil
	}

	part := parts[0]
	remaining := parts[1:]

	// Optimized search strategy:
	// 1. Use length-based child categorization to quickly find candidates
	childLen := len(n.children)

	// Fast path for single child case
	if childLen == 1 {
		child := n.children[0]
		// Either exact match or param/wildcard
		if !child.isParam && !child.isWild {
			if child.part == part {
				if match := child.find(remaining, ctx, method); match != nil {
					return match
				}
			}
		} else if child.isParam {
			ctx.setParam(child.part, part)
			if match := child.find(remaining, ctx, method); match != nil {
				return match
			}
		} else if child.isWild {
			ctx.setParam("*", strings.Join(append([]string{part}, remaining...), "/"))
			if child.handlers != nil {
				if handler, ok := child.handlers[method]; ok && handler != nil {
					return child
				}
			}
			if child.handler != nil && (child.method == method || child.method == "") {
				return child
			}
		}
		return nil
	}

	// Fast path for 2-3 children (common case)
	if childLen <= 3 {
		// First check exact matches (most common case)
		for i := 0; i < childLen; i++ {
			child := n.children[i]
			if !child.isParam && !child.isWild && child.part == part {
				if match := child.find(remaining, ctx, method); match != nil {
					return match
				}
			}
		}

		// Then check params
		for i := 0; i < childLen; i++ {
			child := n.children[i]
			if child.isParam {
				ctx.setParam(child.part, part)
				if match := child.find(remaining, ctx, method); match != nil {
					return match
				}
			}
		}

		// Finally check wildcards
		for i := 0; i < childLen; i++ {
			child := n.children[i]
			if child.isWild {
				ctx.setParam("*", strings.Join(append([]string{part}, remaining...), "/"))
				if child.handlers != nil {
					if handler, ok := child.handlers[method]; ok && handler != nil {
						return child
					}
				}
				if child.handler != nil && (child.method == method || child.method == "") {
					return child
				}
			}
		}

		return nil
	}

	// For larger numbers of children, use the original algorithm
	// 1. Check exact matches first (most common case)
	for _, child := range n.children {
		if !child.isParam && !child.isWild && child.part == part {
			if match := child.find(remaining, ctx, method); match != nil {
				return match
			}
		}
	}

	// 2. Check parameter matches
	for _, child := range n.children {
		if child.isParam {
			ctx.setParam(child.part, part)
			if match := child.find(remaining, ctx, method); match != nil {
				return match
			}
		}
	}

	// 3. Check wildcard matches last
	for _, child := range n.children {
		if child.isWild {
			ctx.setParam("*", strings.Join(append([]string{part}, remaining...), "/"))
			// Check if this wildcard node has a handler for the requested method
			if child.handlers != nil {
				if handler, ok := child.handlers[method]; ok && handler != nil {
					return child
				}
			}
			// Fallback to the old method for backward compatibility
			if child.handler != nil && (child.method == method || child.method == "") {
				return child
			}
		}
	}

	return nil
}

func (n *RouteNode) findChild(part string, isParam, isWild bool) *RouteNode {
	for _, child := range n.children {
		if child.part == part && child.isParam == isParam && child.isWild == isWild {
			return child
		}
	}
	return nil
}

var pathBuilderPool = sync.Pool{
	New: func() interface{} {
		return new(strings.Builder)
	},
}

func (r *Router) normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] == '/' {
		return path
	}

	builder := pathBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	builder.WriteByte('/')
	builder.WriteString(path)
	result := builder.String()
	pathBuilderPool.Put(builder)
	return result
}

// Get registers a route for the GET HTTP method with RouteHandler
func (a *App) Get(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodGet, path, handlers...)
}

// Post registers a route for the POST HTTP method with RouteHandler
func (a *App) Post(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodPost, path, handlers...)
}

// Put registers a route for the PUT HTTP method with RouteHandler
func (a *App) Put(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodPut, path, handlers...)
}

// Delete registers a route for the DELETE HTTP method with RouteHandler
func (a *App) Delete(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodDelete, path, handlers...)
}

// Patch registers a route for the PATCH HTTP method with RouteHandler
func (a *App) Patch(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodPatch, path, handlers...)
}

// Head registers a route for the HEAD HTTP method with RouteHandler
func (a *App) Head(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodHead, path, handlers...)
}

// Options registers a route for the OPTIONS HTTP method with RouteHandler
func (a *App) Options(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodOptions, path, handlers...)
}

// Connect registers a route for the CONNECT HTTP method with RouteHandler
func (a *App) Connect(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodConnect, path, handlers...)
}

// Trace registers a route for the TRACE HTTP method with RouteHandler
func (a *App) Trace(path string, handlers ...RouteHandler) error {
	return a.router.Add(MethodTrace, path, handlers...)
}

// StringHandler creates a RouteHandler that returns the provided string
func StringHandler(str string) RouteHandler {
	return func(c *Context) error {
		return c.Send(str)
	}
}
