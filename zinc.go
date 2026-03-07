package zinc

import (
	"net/http"
	"strings"
)

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if len(a.serverHeader) > 0 {
		w.Header()[HeaderServer] = a.serverHeader
	}

	ctx := NewContext(w, r)
	defer ctx.release()
	ctx.app = a

	if len(a.middleware) > 0 || len(a.prefixMiddleware) > 0 {
		handlers := a.preHandlersForPath(r.URL.Path)
		if len(handlers) > 0 {
			ctx.setHandlers(handlers)
			if err := ctx.Next(); err != nil {
				a.handleError(ctx, err)
			}
			return
		}
	}

	a.dispatch(ctx)
}

func (a *App) preHandlersForPath(path string) []HandlerFunc {
	count := len(a.middleware)
	for _, entry := range a.prefixMiddleware {
		if pathHasPrefix(path, entry.prefix, a.config.CaseSensitive) {
			count += len(entry.handlers)
		}
	}
	if count == 0 {
		return nil
	}
	handlers := make([]HandlerFunc, 0, count+1)
	handlers = append(handlers, a.middleware...)
	for _, entry := range a.prefixMiddleware {
		if pathHasPrefix(path, entry.prefix, a.config.CaseSensitive) {
			handlers = append(handlers, entry.handlers...)
		}
	}
	handlers = append(handlers, appDispatchHandler)
	return handlers
}

func (a *App) dispatch(ctx *Context) {
	method := ctx.Method()
	path := ctx.Path()

	handler := a.router.findInto(method, path, ctx)
	if handler == nil && method == MethodHead && a.config.AutoHead {
		ctx.truncateParams(0)
		handler = a.router.findInto(MethodGet, path, ctx)
	}
	if handler != nil {
		a.execute(ctx, handler)
		return
	}

	if mount := a.matchMount(path); mount != nil {
		ctx.setRoute(mount.info)
		mount.serve(ctx)
		return
	}

	allowed := a.router.allowedMethods(path, a.config.AutoHead, a.config.AutoOptions)
	if method == MethodOptions && a.config.AutoOptions && len(allowed) > 0 {
		ctx.SetHeader(HeaderAllow, strings.Join(allowed, ", "))
		_ = ctx.Status(StatusNoContent).NoContent()
		return
	}

	if a.config.HandleMethodNotAllowed && len(allowed) > 0 {
		ctx.Status(StatusMethodNotAllowed)
		ctx.SetHeader(HeaderAllow, strings.Join(allowed, ", "))
		if a.methodNA != nil {
			a.execute(ctx, a.methodNA)
			if !ctx.written {
				_ = ctx.NoContent()
			}
			return
		}
		a.handleError(ctx, ErrMethodNotAllowed)
		return
	}

	ctx.Status(StatusNotFound)
	if a.notFound != nil {
		a.execute(ctx, a.notFound)
		if !ctx.written {
			_ = ctx.String(http.StatusText(StatusNotFound))
		}
		return
	}
	a.handleError(ctx, ErrNotFound)
}

func (a *App) execute(ctx *Context, handler HandlerFunc) {
	if err := handler(ctx); err != nil {
		a.handleError(ctx, err)
	}
}

func (a *App) handleError(ctx *Context, err error) {
	if err == nil {
		return
	}
	if ctx != nil {
		ctx.lastErr = err
	}
	a.config.ErrorHandler(ctx, err)
}

func appDispatchHandler(c *Context) error {
	if c != nil && c.app != nil {
		c.app.dispatch(c)
	}
	return nil
}

func (a *App) matchMount(path string) *mountedHandler {
	for i := range a.mounts {
		mount := &a.mounts[i]
		if pathHasPrefix(path, mount.prefix, a.config.CaseSensitive) {
			return mount
		}
	}
	return nil
}

func (m *mountedHandler) serve(c *Context) {
	if m == nil || m.handler == nil {
		return
	}
	request := c.Request()
	cloned := request.Clone(request.Context())
	cloned.URL = cloneURL(request.URL)
	cloned.RequestURI = cloneRequestURI(cloned.URL)
	cloned.URL.Path = stripMountPrefix(cloned.URL.Path, m.prefixPath)
	if cloned.URL.RawPath != "" {
		cloned.URL.RawPath = stripMountPrefix(cloned.URL.RawPath, m.prefixPath)
	}
	if cloned.URL.Path == "" {
		cloned.URL.Path = "/"
	}
	if cloned.URL.RawPath == "" && cloned.URL.Path != "" {
		cloned.URL.RawPath = cloned.URL.Path
	}
	m.handler.ServeHTTP(c.Writer(), cloned)
	c.written = true
}

func stripMountPrefix(path, prefix string) string {
	if prefix == "/" {
		return path
	}
	trimmed := strings.TrimPrefix(path, prefix)
	if trimmed == "" {
		return "/"
	}
	if trimmed[0] != '/' {
		return "/" + trimmed
	}
	return trimmed
}
