package zinc

import (
	"slices"
	"strings"
)

// Group represents a group of routes with a common prefix.
type Group struct {
	prefix     string
	app        *App
	middleware []Middleware
}

// Group creates a new group with a given prefix.
func (g *Group) Group(prefix string) *Group {
	fullPrefix := g.prefix + "/" + strings.Trim(prefix, "/")
	return &Group{
		prefix:     fullPrefix,
		app:        g.app,
		middleware: slices.Clone(g.middleware),
	}
}

// Group creates a new group with a given prefix.
func (a *App) Group(prefix string) *Group {
	return &Group{
		prefix:     strings.Trim(prefix, "/"),
		app:        a,
		middleware: []Middleware{},
	}
}

// Use adds middleware to the group
func (g *Group) Use(middleware ...Middleware) *Group {
	g.middleware = append(g.middleware, middleware...)
	return g
}

// combineHandlers combines the group middleware with the route handlers
func (g *Group) combineHandlers(handlers ...RouteHandler) []RouteHandler {
	// Convert middleware to RouteHandler for combination
	middlewareHandlers := make([]RouteHandler, len(g.middleware))
	for i, mw := range g.middleware {
		middlewareHandlers[i] = RouteHandler(mw)
	}

	// Combine middleware with handlers
	combined := make([]RouteHandler, len(middlewareHandlers)+len(handlers))
	copy(combined, middlewareHandlers)
	copy(combined[len(middlewareHandlers):], handlers)

	return combined
}

func (g *Group) Get(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Get(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Post(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Post(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Put(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Put(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Delete(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Delete(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Patch(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Patch(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Head(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Head(fullPath, g.combineHandlers(handlers...)...)
}

func (g *Group) Options(path string, handlers ...RouteHandler) error {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	return g.app.Options(fullPath, g.combineHandlers(handlers...)...)
}
