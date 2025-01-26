package zinc

import "strings"

// Group represents a group of routes with a common prefix.
type Group struct {
	prefix string
	app    *App
}

// Group creates a new group with a given prefix.
func (g *Group) Group(prefix string) *Group {
	fullPrefix := g.prefix + "/" + strings.Trim(prefix, "/")
	return &Group{
		prefix: fullPrefix,
		app:    g.app,
	}
}

// Group creates a new group with a given prefix.
func (a *App) Group(prefix string) *Group {
	return &Group{
		prefix: strings.Trim(prefix, "/"),
		app:    a,
	}
}

func (g *Group) Get(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Get(fullPath, handler)
}

func (g *Group) Post(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Post(fullPath, handler)
}

func (g *Group) Put(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Put(fullPath, handler)
}

func (g *Group) Delete(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Delete(fullPath, handler)
}

func (g *Group) Patch(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Patch(fullPath, handler)
}

func (g *Group) Head(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Head(fullPath, handler)
}

func (g *Group) Options(path string, handler interface{}) {
	fullPath := "/" + g.prefix + "/" + strings.Trim(path, "/")
	g.app.Options(fullPath, handler)
}
