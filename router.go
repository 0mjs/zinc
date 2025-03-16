package zinc

import (
	"strings"
	"sync"
)

const (
	MethodGet          = "GET"
	MethodPost         = "POST"
	MethodPut          = "PUT"
	MethodPatch        = "PATCH"
	MethodDelete       = "DELETE"
	MethodHead         = "HEAD"
	MethodOptions      = "OPTIONS"
	MethodConnect      = "CONNECT"
	MethodTrace        = "TRACE"
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

type Middleware func(c *Context)

type Router struct {
	routes     map[string]map[string]*Route
	router     *RouteNode
	middleware []Middleware
	cache      *RouteCache
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

func (r *Router) Add(method, path string, handlers ...interface{}) {
	// Initialize maps if needed
	if r.routes == nil {
		r.routes = make(map[string]map[string]*Route)
	}

	methodRoutes, ok := r.routes[method]
	if !ok {
		methodRoutes = make(map[string]*Route)
		r.routes[method] = methodRoutes
	}

	// Fast path for single handler (most common case)
	var mainHandler RouteHandler
	if len(handlers) == 1 {
		mainHandler = convertToRouteHandler(handlers[0])
	} else if len(handlers) > 1 {
		// Pre-allocate routeHandlers slice with exact capacity
		routeHandlers := make([]RouteHandler, 0, len(r.middleware)+len(handlers))

		// Add middleware handlers if any
		if len(r.middleware) > 0 {
			for _, mw := range r.middleware {
				routeHandlers = append(routeHandlers, RouteHandler(mw))
			}
		}

		// Add route handlers
		for _, handler := range handlers {
			rh := convertToRouteHandler(handler)
			routeHandlers = append(routeHandlers, rh)
		}

		mainHandler = chain(routeHandlers)
	}

	path = r.normalizePath(path)

	// Reuse path parts for static routes from a pool
	var parts []string
	if strings.IndexByte(path, ':') >= 0 || strings.IndexByte(path, '*') >= 0 {
		// Dynamic route - need to parse and process parts
		parts = getPathParts(path)
	} else {
		// Static route - use minimal parts array just to keep the API consistent
		parts = make([]string, 0, 1)
		if path != "/" && len(path) > 0 {
			if path[0] == '/' {
				parts = append(parts, path[1:])
			} else {
				parts = append(parts, path)
			}
		}
	}

	// Store in routes map
	methodRoutes[path] = &Route{
		path:    path,
		handler: mainHandler,
		method:  method,
		parts:   parts,
	}

	// Update trie storage - only needed for dynamic routes or if no cache exists
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
	// Static route fast path - most common case first
	if routes, ok := r.routes[method]; ok {
		if route, ok := routes[path]; ok {
			// Create a context with empty params - most common case for APIs
			ctx := &Context{PathParams: params{}}
			return route.handler, ctx
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
	found := current.find(parts, ctx, method)
	if found != nil {
		// Get the handler for this method
		var handler RouteHandler
		if found.handlers != nil {
			if h, ok := found.handlers[method]; ok {
				handler = h
			}
		}
		// Fallback to the old handler field
		if handler == nil && found.handler != nil && (found.method == method || found.method == "") {
			handler = found.handler
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

	// Return nil handler for 404
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
		return func(c *Context) {
			h1(c)
			if !c.written {
				h2(c)
			}
		}
	}

	// For 3+ handlers, use the general case
	return func(c *Context) {
		for i, handler := range handlers {
			handler(c)
			if c.written {
				// Fast return
				return
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
			handlers[len(handlers)-1](c)
		}
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

func (a *App) Get(path string, handlers ...interface{}) {
	a.router.Add(MethodGet, path, handlers...)
}

func (a *App) Post(path string, handlers ...interface{}) {
	a.router.Add(MethodPost, path, handlers...)
}

func (a *App) Put(path string, handlers ...interface{}) {
	a.router.Add(MethodPut, path, handlers...)
}

func (a *App) Delete(path string, handlers ...interface{}) {
	a.router.Add(MethodDelete, path, handlers...)
}

func (a *App) Patch(path string, handlers ...interface{}) {
	a.router.Add(MethodPatch, path, handlers...)
}

func (a *App) Head(path string, handlers ...interface{}) {
	a.router.Add(MethodHead, path, handlers...)
}

func (a *App) Options(path string, handlers ...interface{}) {
	a.router.Add(MethodOptions, path, handlers...)
}

func (a *App) Connect(path string, handlers ...interface{}) {
	a.router.Add(MethodConnect, path, handlers...)
}

func (a *App) Trace(path string, handlers ...interface{}) {
	a.router.Add(MethodTrace, path, handlers...)
}

func (r *Router) findRoute(method string, path string) (RouteHandler, *Context) {
	if route := r.routes[method][path]; route != nil {
		ctx := &Context{}
		return route.handler, ctx
	}
	return nil, nil
}

// Helper function to convert middleware slice to RouteHandler slice
func (r *Router) middlewareToHandlers() []RouteHandler {
	handlers := make([]RouteHandler, len(r.middleware))
	for i, m := range r.middleware {
		handlers[i] = RouteHandler(m)
	}
	return handlers
}

// Helper function to convert interface{} to RouteHandler
func convertToRouteHandler(handler interface{}) RouteHandler {
	switch v := handler.(type) {
	case string:
		return func(c *Context) {
			c.Send(v)
		}
	case RouteHandler:
		return v
	case func(*Context):
		return v
	case Middleware:
		return RouteHandler(v)
	default:
		panic("handler must be either a string, RouteHandler, or Middleware")
	}
}
