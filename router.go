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

var pathPartsCache = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8)
	},
}

// Use a fixed array to avoid allocations for common path lengths
func getPathParts(path string) []string {
	// Use a local fixed-size array for most paths (which are short)
	var fixedParts [8]string
	parts := fixedParts[:0]

	if path == "" || path == "/" {
		return parts
	}

	// Fast path for common URL patterns
	if path[0] == '/' && len(path) > 1 {
		// Skip the leading slash
		path = path[1:]
	}

	// Fast split without regexp
	start := 0
	partCount := 0
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				if partCount < len(fixedParts) {
					parts = append(parts, path[start:i])
					partCount++
				}
			}
			start = i + 1
		}
	}

	if start < len(path) && partCount < len(fixedParts) {
		parts = append(parts, path[start:])
	}

	return parts
}

func (r *Router) Add(method, path string, handlers ...interface{}) {
	// Initialize maps if needed
	if r.routes == nil {
		r.routes = make(map[string]map[string]*Route)
	}
	if r.routes[method] == nil {
		r.routes[method] = make(map[string]*Route)
	}

	// Pre-allocate routeHandlers slice with exact capacity
	routeHandlers := make([]RouteHandler, 0, len(r.middleware)+len(handlers))
	routeHandlers = append(routeHandlers, r.middlewareToHandlers()...)

	for _, handler := range handlers {
		rh := convertToRouteHandler(handler)
		routeHandlers = append(routeHandlers, rh)
	}

	var mainHandler RouteHandler
	if len(routeHandlers) > 0 {
		mainHandler = chain(routeHandlers)
	}

	path = r.normalizePath(path)
	parts := getPathParts(path)

	// Store in routes map
	r.routes[method][path] = &Route{
		path:    path,
		handler: mainHandler,
		method:  method,
		parts:   parts,
	}

	// Update trie storage
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
			child = &RouteNode{
				part:     part,
				isParam:  isParam,
				isWild:   isWild,
				method:   "",                            // Intermediate nodes have no method
				handlers: make(map[string]RouteHandler), // Initialize handlers map
			}
			current.children = append(current.children, child)
		}

		if i == len(parts)-1 {
			// This is a leaf node - store the handler for this method
			if child.handlers == nil {
				child.handlers = make(map[string]RouteHandler)
			}
			child.handlers[method] = mainHandler
			child.path = path
			child.method = method       // Keep this for backward compatibility
			child.handler = mainHandler // Keep this for backward compatibility
		}

		current = child
	}

	pathPartsCache.Put(parts)
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
}

func NewRouteCache(size int) *RouteCache {
	return &RouteCache{
		cache: make(map[routeCacheKey]routeCacheEntry, size),
		size:  size,
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
	if len(rc.cache) >= rc.size {
		// Simple eviction: clear the cache when full
		rc.cache = make(map[routeCacheKey]routeCacheEntry, rc.size)
	}
	rc.cache[key] = entry
	rc.mu.Unlock()
}

func (r *Router) Find(method, path string) (RouteHandler, *Context) {
	// Static route fast path - most common case first
	if routes, ok := r.routes[method]; ok {
		if route, ok := routes[path]; ok {
			// Create a context with empty params
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

	// Parse path parts only if needed for dynamic matching
	parts := getPathParts(path)
	current := r.router

	if current != nil {
		// Optimize common cases first - check for exact match at root level
		for _, child := range current.children {
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
						// Cache the result
						if r.cache != nil {
							r.cache.set(key, routeCacheEntry{handler, ctx})
						}
						return handler, ctx
					}
				}
			}
		}

		// Then check for dynamic routes
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
				// Cache the result
				if r.cache != nil {
					r.cache.set(key, routeCacheEntry{handler, ctx})
				}
				return handler, ctx
			}
		}
	}

	// Return nil handler for 404
	return nil, nil
}

func (r *Router) Use(middleware ...Middleware) {
	r.middleware = append(r.middleware, middleware...)
}

func chain(handlers []RouteHandler) RouteHandler {
	return func(c *Context) {
		for _, handler := range handlers {
			handler(c)
			if c.written {
				return
			}
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
