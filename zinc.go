package zinc

import (
	"flag"
	"fmt"
	"net/http"
)

type App struct {
	router     *Router
	middleware []Middleware
	services   map[string]interface{}
	config     *Config
}

type RouteHandler func(c *Context)

type Map map[string]interface{}

func New() *App {
	return &App{
		router:     &Router{},
		middleware: make([]Middleware, 0),
		services:   make(map[string]interface{}),
		config:     &DefaultConfig,
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(w, r)
	defer ctx.release()

	ctx.services = a.services

	if len(a.middleware) > 0 {
		ctx.setHandlers(a.middleware)
		ctx.Next()
		if ctx.written {
			return
		}
	}

	handler, foundCtx := a.router.Find(r.Method, r.URL.Path)
	if handler != nil {
		if foundCtx != nil {
			ctx.PathParams = foundCtx.PathParams
		}
		handler(ctx)
		return
	}

	http.NotFound(w, r)
}

func (a *App) Use(middleware ...Middleware) {
	a.middleware = append(a.middleware, middleware...)
}

func (a *App) Service(name string, service interface{}) {
	a.services[name] = service
}

func (a *App) Serve(port ...string) error {
	serverPort := parseArgs(a)
	if len(port) > 0 && port[0] != "" {
		serverPort = port[0]
	}
	addr := serverPort
	if serverPort[0] != ':' {
		addr = ":" + serverPort
	}
	fmt.Printf("Server starting on port %s...\n", serverPort)
	return http.ListenAndServe(addr, a)
}

func parseArgs(a *App) string {
	port := flag.String("port", a.config.DefaultAddr, "port number for the server")
	flag.Parse()
	return *port
}
