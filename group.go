package zinc

import (
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

func (g *Group) Add(method, routePath string, handlers ...HandlerFunc) error {
	fullPath := joinPaths(g.prefix, routePath)
	allHandlers := make([]HandlerFunc, 0, len(g.middleware)+len(handlers))
	allHandlers = append(allHandlers, g.middleware...)
	allHandlers = append(allHandlers, handlers...)
	return g.app.Add(method, fullPath, allHandlers...)
}

func (g *Group) Get(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodGet, path, handlers...)
}
func (g *Group) Post(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodPost, path, handlers...)
}
func (g *Group) Put(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodPut, path, handlers...)
}
func (g *Group) Delete(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodDelete, path, handlers...)
}
func (g *Group) Patch(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodPatch, path, handlers...)
}
func (g *Group) Head(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodHead, path, handlers...)
}
func (g *Group) Options(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodOptions, path, handlers...)
}
func (g *Group) Connect(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodConnect, path, handlers...)
}
func (g *Group) Trace(path string, handlers ...HandlerFunc) error {
	return g.Add(MethodTrace, path, handlers...)
}

func (g *Group) Match(methods []string, routePath string, handlers ...HandlerFunc) error {
	for _, method := range methods {
		if err := g.Add(method, routePath, handlers...); err != nil {
			return err
		}
	}
	return nil
}

func (g *Group) All(path string, handlers ...HandlerFunc) error {
	return g.Match(routeMethods, path, handlers...)
}

func (g *Group) Any(path string, handlers ...HandlerFunc) error {
	return g.All(path, handlers...)
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
