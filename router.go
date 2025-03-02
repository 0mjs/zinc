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
	isParam  bool
	isWild   bool
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

func getPathParts(path string) []string {
	parts := pathPartsCache.Get().([]string)
	parts = parts[:0]

	if path == "" || path == "/" {
		pathPartsCache.Put(parts)
		return parts
	}

	start := 0
	if path[0] == '/' {
		start = 1
	}

	for i := start; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}

	if start < len(path) {
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
		current = &RouteNode{}
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
				part:    part,
				isParam: isParam,
				isWild:  isWild,
			}
			current.children = append(current.children, child)
		}

		if i == len(parts)-1 {
			child.handler = mainHandler
			child.path = path
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

type pathParts struct {
	parts [8]string // Most paths won't exceed 8 parts
	count int
}

var pathPartsPool = sync.Pool{
	New: func() interface{} {
		return &pathParts{}
	},
}

func (r *Router) Find(method, path string) (RouteHandler, *Context) {
	// Fast path for exact matches
	if routes, ok := r.routes[method]; ok {
		if route, ok := routes[path]; ok {
			return route.handler, &Context{PathParams: params{}}
		}
	}

	// Then check cache
	key := routeCacheKey{method, path}
	if entry, ok := r.cache.get(key); ok {
		return entry.handler, entry.context
	}

	// Fall back to trie matching
	handler, ctx := r.findRoute(method, path)
	if handler != nil {
		r.cache.set(key, routeCacheEntry{handler, ctx})
	}
	return handler, ctx
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

func (n *RouteNode) find(parts []string, ctx *Context) *RouteNode {
	if len(parts) == 0 {
		return n
	}

	part := parts[0]
	remaining := parts[1:]

	// Try exact matches first
	for _, child := range n.children {
		if !child.isParam && !child.isWild && child.part == part {
			if match := child.find(remaining, ctx); match != nil {
				return match
			}
		}
	}

	// Then try parameter matches
	for _, child := range n.children {
		if child.isParam {
			// Save current param state in case we need to backtrack
			oldValue := ctx.Param(child.part)

			// Set new param
			ctx.setParam(child.part, part)

			if match := child.find(remaining, ctx); match != nil {
				return match
			}

			// Backtrack: restore old value if this path didn't work
			if oldValue != "" {
				ctx.setParam(child.part, oldValue)
			}
		}
	}

	// Finally try wildcards
	for _, child := range n.children {
		if child.isWild {
			ctx.setParam("*", strings.Join(append([]string{part}, remaining...), "/"))
			return child
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
