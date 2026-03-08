package zinc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type Map map[string]any

type HandlerFunc func(*Context) error

type RouteHandler = HandlerFunc

type Middleware = HandlerFunc

type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}

type routeMeta struct {
	method  string
	path    string
	handler HandlerFunc
}

func (m routeMeta) export() RouteInfo {
	return RouteInfo{
		Method:  m.method,
		Path:    m.path,
		Handler: handlerName(m.handler),
	}
}

type prefixMiddleware struct {
	prefix   string
	handlers []HandlerFunc
}

type mountedHandler struct {
	prefix     string
	prefixPath string
	handler    http.Handler
	info       routeMeta
}

type App struct {
	config           Config
	router           *Router
	middleware       []HandlerFunc
	middlewareChain  []HandlerFunc
	prefixMiddleware []prefixMiddleware
	mounts           []mountedHandler
	notFound         HandlerFunc
	methodNA         HandlerFunc
	server           *http.Server
	serverMu         sync.Mutex
	serverHeader     []string
}

func New() *App {
	return NewWithConfig(DefaultConfig)
}

func NewWithConfig(cfg Config) *App {
	cfg = normalizeConfig(cfg)

	var cache *RouteCache
	if cfg.RouteCacheSize > 0 {
		cache = NewRouteCache(cfg.RouteCacheSize)
	}

	app := &App{
		config: cfg,
		router: &Router{
			cache:  cache,
			config: &cfg,
		},
		middleware: make([]HandlerFunc, 0),
	}
	if cfg.ServerHeader != "" {
		app.serverHeader = []string{cfg.ServerHeader}
	}
	return app
}

func normalizeConfig(cfg Config) Config {
	if cfg.BodyLimit == 0 {
		cfg.BodyLimit = DefaultConfig.BodyLimit
	}
	if cfg.ReadTimeout == 0 {
		cfg.ReadTimeout = DefaultConfig.ReadTimeout
	}
	if cfg.WriteTimeout == 0 {
		cfg.WriteTimeout = DefaultConfig.WriteTimeout
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = DefaultConfig.IdleTimeout
	}
	if cfg.ProxyHeader == "" {
		cfg.ProxyHeader = DefaultConfig.ProxyHeader
	}
	if cfg.JSONCodec == nil {
		cfg.JSONCodec = defaultJSONCodec{}
	}
	if cfg.Binder == nil {
		cfg.Binder = defaultBinder{codec: cfg.JSONCodec}
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = defaultErrorHandler
	}
	return cfg
}

func (a *App) Handler() http.Handler {
	return a
}

func (a *App) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return a.Serve(ln)
}

func (a *App) ListenTLS(addr, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return a.serveTLS(ln, certFile, keyFile)
}

func (a *App) Serve(ln net.Listener) error {
	if ln == nil {
		return errors.New("listener is nil")
	}
	return a.serve(ln, "", "")
}

func (a *App) serveTLS(ln net.Listener, certFile, keyFile string) error {
	if ln == nil {
		return errors.New("listener is nil")
	}
	return a.serve(ln, certFile, keyFile)
}

func (a *App) serve(ln net.Listener, certFile, keyFile string) error {
	srv := &http.Server{
		Handler:      a,
		ReadTimeout:  a.config.ReadTimeout,
		WriteTimeout: a.config.WriteTimeout,
		IdleTimeout:  a.config.IdleTimeout,
	}

	a.serverMu.Lock()
	a.server = srv
	a.serverMu.Unlock()

	var err error
	if certFile != "" || keyFile != "" {
		err = srv.ServeTLS(ln, certFile, keyFile)
	} else {
		err = srv.Serve(ln)
	}
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (a *App) Shutdown(ctx context.Context) error {
	a.serverMu.Lock()
	srv := a.server
	a.serverMu.Unlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

func (a *App) Use(handlers ...HandlerFunc) {
	a.middleware = append(a.middleware, handlers...)
	a.rebuildMiddlewareChain()
}

func (a *App) UsePrefix(prefix string, handlers ...HandlerFunc) {
	prefix = normalizeRegisteredPrefix(prefix)
	a.prefixMiddleware = append(a.prefixMiddleware, prefixMiddleware{prefix: prefix, handlers: append([]HandlerFunc(nil), handlers...)})
}

func (a *App) rebuildMiddlewareChain() {
	if len(a.middleware) == 0 {
		a.middlewareChain = nil
		return
	}
	chain := make([]HandlerFunc, len(a.middleware)+1)
	copy(chain, a.middleware)
	chain[len(chain)-1] = appDispatchHandler
	a.middlewareChain = chain
}

func (a *App) Group(prefix string, handlers ...HandlerFunc) *Group {
	return NewGroup(a, prefix, handlers...)
}

func (a *App) Route(prefix string, fn func(*Group), handlers ...HandlerFunc) *Group {
	group := a.Group(prefix, handlers...)
	if fn != nil {
		fn(group)
	}
	return group
}

func (a *App) Mount(prefix string, h http.Handler) {
	prefix = normalizeRegisteredPrefix(prefix)
	entry := mountedHandler{
		prefix:     storedPrefix(prefix, a.config.CaseSensitive),
		prefixPath: prefix,
		handler:    h,
		info:       routeMeta{method: methodUse, path: prefix, handler: Wrap(h)},
	}
	a.mounts = append(a.mounts, entry)
	sort.SliceStable(a.mounts, func(i, j int) bool {
		return len(a.mounts[i].prefixPath) > len(a.mounts[j].prefixPath)
	})
}

func (a *App) NotFound(handler HandlerFunc) {
	a.notFound = handler
}

func (a *App) MethodNotAllowed(handler HandlerFunc) {
	a.methodNA = handler
}

func (a *App) Routes() []RouteInfo {
	routes := a.router.Routes()
	if len(a.mounts) == 0 {
		return routes
	}
	out := make([]RouteInfo, 0, len(routes)+len(a.mounts))
	out = append(out, routes...)
	for _, mount := range a.mounts {
		out = append(out, mount.info.export())
	}
	return out
}

func Wrap(h http.Handler) HandlerFunc {
	return func(c *Context) error {
		h.ServeHTTP(c.Writer(), c.Request())
		c.written = true
		return nil
	}
}

func WrapFunc(fn http.HandlerFunc) HandlerFunc {
	return Wrap(fn)
}

func normalizeRegisteredPrefix(prefix string) string {
	if prefix == "" {
		return "/"
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if len(prefix) > 1 {
		prefix = strings.TrimSuffix(prefix, "/")
		if prefix == "" {
			return "/"
		}
	}
	return prefix
}

func storedPrefix(prefix string, caseSensitive bool) string {
	if caseSensitive {
		return prefix
	}
	return strings.ToLower(prefix)
}

func pathHasPrefix(path, prefix string, caseSensitive bool) bool {
	if !caseSensitive {
		path = strings.ToLower(path)
		prefix = strings.ToLower(prefix)
	}
	if prefix == "/" {
		return true
	}
	if path == prefix {
		return true
	}
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	if strings.HasSuffix(prefix, "/") {
		return true
	}
	return len(path) > len(prefix) && path[len(prefix)] == '/'
}
