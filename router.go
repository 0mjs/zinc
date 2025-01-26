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
	routes     map[string]map[string]*Route // method -> path -> route
	router     *RouteNode
	middleware []Middleware
}

var pathPartsCache = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 8) // Pre-allocate for common path lengths
	},
}

func getPathParts(path string) []string {
	parts := pathPartsCache.Get().([]string)
	parts = parts[:0] // Reset slice but keep capacity

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

func (r *Router) Find(method, path string) (RouteHandler, map[string]string) {
	// Try direct lookup first
	if methodRoutes, ok := r.routes[method]; ok {
		if route, ok := methodRoutes[path]; ok {
			return route.handler, nil // Don't allocate params map for static routes
		}
	}

	// Fall back to trie search for parameterized routes
	parts := getPathParts(path)
	params := make(map[string]string)
	node := r.router.find(parts, params)

	if node != nil && node.handler != nil {
		if matchedRoute := r.routes[method][node.path]; matchedRoute != nil {
			pathPartsCache.Put(parts)
			return matchedRoute.handler, params
		}
	}

	pathPartsCache.Put(parts)
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

func (n *RouteNode) find(parts []string, params map[string]string) *RouteNode {
	if len(parts) == 0 {
		return n
	}

	part := parts[0]
	parts = parts[1:]

	for _, child := range n.children {
		if child.isWild {
			params["*"] = strings.Join(append([]string{part}, parts...), "/")
			return child
		}

		if child.isParam {
			params[child.part] = part
			if matchChild := child.find(parts, params); matchChild != nil {
				return matchChild
			}
		} else if child.part == part {
			if matchChild := child.find(parts, params); matchChild != nil {
				return matchChild
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

func (r *Router) normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
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

func (r *Router) findRoute(path string, method string) *Route {
	for _, route := range r.routes[method] {
		if route.path == path {
			return route
		}
	}
	return nil
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
