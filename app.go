package zinc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
)

type Map map[string]any

type HandlerFunc func(*Context) error

type RouteHandler = HandlerFunc

type Middleware = HandlerFunc

type RouteInfo struct {
	Name    string
	Method  string
	Path    string
	Params  []string
	Mounted bool
	Handler string
}

type RouteSpec struct {
	Name    string
	Method  string
	Path    string
	Handler HandlerFunc
}

type routeMeta struct {
	name      string
	method    string
	path      string
	params    []string
	mounted   bool
	handlerPC uintptr
}

func (m routeMeta) export() RouteInfo {
	return RouteInfo{
		Name:    m.name,
		Method:  m.method,
		Path:    m.path,
		Params:  append([]string(nil), m.params...),
		Mounted: m.mounted,
		Handler: handlerNameFromPC(m.handlerPC),
	}
}

func (m routeMeta) url(values []string) (string, error) {
	if len(m.params) != len(values) {
		return "", fmt.Errorf("route %q expects %d params, got %d", m.name, len(m.params), len(values))
	}
	if len(m.params) == 0 {
		return m.path, nil
	}

	var builder strings.Builder
	valueIndex := 0
	for i := 0; i < len(m.path); i++ {
		switch m.path[i] {
		case ParamIdentifier:
			start := i + 1
			end := start
			for end < len(m.path) && m.path[end] != '/' && m.path[end] != '<' {
				end++
			}
			builder.WriteString(url.PathEscape(values[valueIndex]))
			valueIndex++
			if end < len(m.path) && m.path[end] == '<' {
				constraintEnd, _, err := parseParamConstraint(m.path, end)
				if err != nil {
					return "", err
				}
				i = constraintEnd - 1
			} else {
				i = end - 1
			}
		case WildcardIdentifier:
			start := i + 1
			end := start
			for end < len(m.path) && m.path[end] != '/' {
				end++
			}
			builder.WriteString(values[valueIndex])
			valueIndex++
			i = end - 1
		default:
			builder.WriteByte(m.path[i])
		}
	}
	return builder.String(), nil
}

func newRouteMeta(method, path, name string, handler HandlerFunc, params []string, mounted bool) routeMeta {
	return routeMeta{
		name:      name,
		method:    method,
		path:      path,
		params:    append([]string(nil), params...),
		mounted:   mounted,
		handlerPC: handlerPC(handler),
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
	notFoundRoutes   *Router
	middleware       []HandlerFunc
	middlewareChain  []HandlerFunc
	prefixMiddleware []prefixMiddleware
	mounts           []mountedHandler
	notFound         HandlerFunc
	methodNA         HandlerFunc
	server           *http.Server
	serverMu         sync.Mutex
	serverHeader     []string
	defaultErrors    bool
}

func New() *App {
	return NewWithConfig(DefaultConfig)
}

func NewWithConfig(cfg Config) *App {
	defaultErrors := cfg.ErrorHandler == nil
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
		middleware:    make([]HandlerFunc, 0),
		defaultErrors: defaultErrors,
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
	if cfg.RequestBinder == nil {
		cfg.RequestBinder = defaultBinder{codec: cfg.JSONCodec}
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = defaultErrorHandler
	}
	return cfg
}

func (a *App) Handler() http.Handler {
	return a
}

// Listen starts the app on addr, defaulting to :8080 when addr is omitted.
func (a *App) Listen(addr ...string) error {
	listenAddr, err := resolveListenAddr(addr...)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	return a.Serve(ln)
}

func resolveListenAddr(addr ...string) (string, error) {
	switch len(addr) {
	case 0:
		return DefaultListenAddr, nil
	case 1:
		if addr[0] == "" {
			return DefaultListenAddr, nil
		}
		return addr[0], nil
	default:
		return "", fmt.Errorf("expected at most one listen address, got %d", len(addr))
	}
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
		info:       newRouteMeta(methodUse, prefix, "", Wrap(h), nil, true),
	}
	a.mounts = append(a.mounts, entry)
	sort.SliceStable(a.mounts, func(i, j int) bool {
		return len(a.mounts[i].prefixPath) > len(a.mounts[j].prefixPath)
	})
}

func (a *App) AcquireContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := NewContext(w, r)
	ctx.app = a
	return ctx
}

func (a *App) ReleaseContext(c *Context) {
	if c == nil {
		return
	}
	c.release()
}

func (a *App) NotFound(handler HandlerFunc) {
	a.notFound = handler
}

func (a *App) RouteNotFound(path string, handlers ...HandlerFunc) error {
	if len(handlers) == 0 {
		return errors.New("route handler is nil")
	}
	if a.notFoundRoutes == nil {
		a.notFoundRoutes = &Router{config: &a.config}
	}
	return a.notFoundRoutes.Add(MethodGet, path, handlers...)
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

func (a *App) Handle(spec RouteSpec) error {
	if spec.Handler == nil {
		return errors.New("route handler is nil")
	}
	return a.router.AddNamed(spec.Method, spec.Path, spec.Name, spec.Handler)
}

func (a *App) RouteByName(name string) (RouteInfo, bool) {
	meta, ok := a.router.routeMetaByName(name)
	if !ok {
		return RouteInfo{}, false
	}
	return meta.export(), true
}

func (a *App) URL(name string, params ...string) (string, error) {
	meta, ok := a.router.routeMetaByName(name)
	if !ok {
		return "", fmt.Errorf("route %q not found", name)
	}
	return meta.url(params)
}

func (a *App) FindRoute(method, path string) (RouteInfo, bool) {
	_, ctx := a.router.Find(method, path)
	if ctx != nil {
		return ctx.Route(), true
	}
	if mount := a.matchMount(path); mount != nil {
		return mount.info.export(), true
	}
	return RouteInfo{}, false
}

func (a *App) RoutesByMethod(method string) []RouteInfo {
	routes := a.Routes()
	out := make([]RouteInfo, 0, len(routes))
	for _, route := range routes {
		if route.Method == method {
			out = append(out, route)
		}
	}
	return out
}

func (a *App) RoutesByPrefix(prefix string) []RouteInfo {
	prefix = normalizeRegisteredPrefix(prefix)
	routes := a.Routes()
	out := make([]RouteInfo, 0, len(routes))
	for _, route := range routes {
		if strings.HasPrefix(route.Path, prefix) {
			out = append(out, route)
		}
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
