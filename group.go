package zinc

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

type Group struct {
	app        *App
	prefix     string
	middleware []HandlerFunc
}

func NewGroup(app *App, prefix string, handlers ...HandlerFunc) *Group {
	prefix = normalizeRegisteredPrefix(prefix)
	if prefix == "/" {
		prefix = ""
	}
	return &Group{
		app:        app,
		prefix:     prefix,
		middleware: append([]HandlerFunc(nil), handlers...),
	}
}

func (g *Group) Use(handlers ...HandlerFunc) *Group {
	g.middleware = append(g.middleware, handlers...)
	return g
}

func (g *Group) Group(prefix string, handlers ...HandlerFunc) *Group {
	fullPrefix := joinPaths(g.prefix, prefix)
	sub := NewGroup(g.app, fullPrefix)
	sub.middleware = append(sub.middleware, g.middleware...)
	sub.middleware = append(sub.middleware, handlers...)
	return sub
}

func (g *Group) Route(prefix string, fn func(*Group), handlers ...HandlerFunc) *Group {
	sub := g.Group(prefix, handlers...)
	if fn != nil {
		fn(sub)
	}
	return sub
}

func (g *Group) Mount(prefix string, h http.Handler) {
	g.app.Mount(joinPaths(g.prefix, prefix), h)
}

func (g *Group) Add(method, routePath string, handlers ...HandlerFunc) {
	mustRegister(g.add(method, routePath, "", handlers...))
}

func (g *Group) add(method, routePath, name string, handlers ...HandlerFunc) error {
	fullPath := joinPaths(g.prefix, routePath)
	allHandlers := make([]HandlerFunc, 0, len(g.middleware)+len(handlers))
	allHandlers = append(allHandlers, g.middleware...)
	allHandlers = append(allHandlers, handlers...)
	return g.app.router.AddNamed(method, fullPath, name, allHandlers...)
}

// Handle registers a source-defined route below the group prefix and panics
// when its declaration is invalid.
func (g *Group) Handle(spec RouteSpec) {
	mustRegister(g.TryHandle(spec))
}

// TryHandle registers a dynamically defined route below the group prefix.
func (g *Group) TryHandle(spec RouteSpec) error {
	if spec.Handler == nil {
		return errors.New("route handler is nil")
	}
	return g.add(spec.Method, spec.Path, spec.Name, spec.Handler)
}

// HandleHTTP registers a standard net/http handler below the group prefix.
func (g *Group) HandleHTTP(pattern string, handler http.Handler) {
	if handler == nil {
		panic("zinc: HTTP handler is nil")
	}
	method, routePath, err := parseHTTPRoutePattern(pattern)
	if err != nil {
		panic(err)
	}
	g.Add(method, routePath, Wrap(handler))
}

func (g *Group) RouteNotFound(routePath string, handlers ...HandlerFunc) {
	fullPath := joinPaths(g.prefix, routePath)
	allHandlers := make([]HandlerFunc, 0, len(g.middleware)+len(handlers))
	allHandlers = append(allHandlers, g.middleware...)
	allHandlers = append(allHandlers, handlers...)
	g.app.RouteNotFound(fullPath, allHandlers...)
}

func (g *Group) Get(path string, handlers ...HandlerFunc) {
	g.Add(MethodGet, path, handlers...)
}
func (g *Group) Post(path string, handlers ...HandlerFunc) {
	g.Add(MethodPost, path, handlers...)
}
func (g *Group) Put(path string, handlers ...HandlerFunc) {
	g.Add(MethodPut, path, handlers...)
}
func (g *Group) Delete(path string, handlers ...HandlerFunc) {
	g.Add(MethodDelete, path, handlers...)
}
func (g *Group) Patch(path string, handlers ...HandlerFunc) {
	g.Add(MethodPatch, path, handlers...)
}
func (g *Group) Head(path string, handlers ...HandlerFunc) {
	g.Add(MethodHead, path, handlers...)
}
func (g *Group) Options(path string, handlers ...HandlerFunc) {
	g.Add(MethodOptions, path, handlers...)
}
func (g *Group) Connect(path string, handlers ...HandlerFunc) {
	g.Add(MethodConnect, path, handlers...)
}
func (g *Group) Trace(path string, handlers ...HandlerFunc) {
	g.Add(MethodTrace, path, handlers...)
}

func (g *Group) Match(methods []string, routePath string, handlers ...HandlerFunc) {
	for _, method := range methods {
		g.Add(method, routePath, handlers...)
	}
}

func (g *Group) All(path string, handlers ...HandlerFunc) {
	g.Match(routeMethods, path, handlers...)
}

func (g *Group) Any(path string, handlers ...HandlerFunc) {
	g.All(path, handlers...)
}

func (g *Group) Static(prefix, root string, opts ...StaticOption) error {
	return g.app.Static(joinPaths(g.prefix, prefix), root, opts...)
}

func (g *Group) StaticFS(prefix string, filesystem fs.FS, opts ...StaticOption) error {
	return g.app.StaticFS(joinPaths(g.prefix, prefix), filesystem, opts...)
}

func (g *Group) File(routePath, file string) error {
	return g.app.File(joinPaths(g.prefix, routePath), file)
}

func (g *Group) FileFS(routePath, file string, filesystem fs.FS) error {
	return g.app.FileFS(joinPaths(g.prefix, routePath), file, filesystem)
}

func joinPaths(a, b string) string {
	if b == "" {
		if a == "" {
			return "/"
		}
		return a
	}
	if !strings.HasPrefix(b, "/") {
		b = "/" + b
	}
	if a == "" || a == "/" {
		return path.Clean(b)
	}
	joined := path.Join(a, b)
	if joined == "." {
		return "/"
	}
	if !strings.HasPrefix(joined, "/") {
		joined = "/" + joined
	}
	return joined
}
