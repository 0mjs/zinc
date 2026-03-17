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
	ctx.app = a

	if len(a.middlewareChain) > 0 && len(a.prefixMiddleware) == 0 {
		ctx.setHandlers(a.middlewareChain)
		if err := ctx.Next(); err != nil {
			a.handleError(ctx, err)
		}
		ctx.release()
		return
	}

	if len(a.middleware) > 0 || len(a.prefixMiddleware) > 0 {
		handlers := a.preHandlersForPath(r.URL.Path)
		if len(handlers) > 0 {
			ctx.setHandlers(handlers)
			if err := ctx.Next(); err != nil {
				a.handleError(ctx, err)
			}
			ctx.release()
			return
		}
	}

	if err := a.dispatch(ctx); err != nil {
		a.handleError(ctx, err)
	}
	ctx.release()
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

func (a *App) dispatch(ctx *Context) error {
	method := ctx.Method()
	path := ctx.Path()
	needsAllowScan := (method == MethodOptions && a.config.AutoOptions) || a.config.HandleMethodNotAllowed

	handled, allowed, err := a.router.dispatchInto(method, path, needsAllowScan, ctx)
	if !handled && method == MethodHead && a.config.AutoHead {
		ctx.truncateParams(0)
		handled, allowed, err = a.router.dispatchInto(MethodGet, path, needsAllowScan, ctx)
	}
	if handled {
		return err
	}

	if mount := a.matchMount(path); mount != nil {
		ctx.setRoute(mount.info)
		mount.serve(ctx)
		return nil
	}

	allowedHeader := allowed.header(a.config.AutoHead, a.config.AutoOptions)
	if method == MethodOptions && a.config.AutoOptions && allowedHeader != "" {
		ctx.SetHeader(HeaderAllow, allowedHeader)
		return ctx.Status(StatusNoContent).NoContent()
	}

	if a.config.HandleMethodNotAllowed && allowedHeader != "" {
		if a.methodNA == nil && a.defaultErrors {
			return ctx.writeDefaultErrorResponse(StatusMethodNotAllowed, allowedHeader)
		}
		ctx.Status(StatusMethodNotAllowed)
		ctx.SetHeader(HeaderAllow, allowedHeader)
		if a.methodNA != nil {
			if err := a.methodNA(ctx); err != nil {
				return err
			}
			if !ctx.written {
				return ctx.NoContent()
			}
			return nil
		}
		if a.defaultErrors {
			return ctx.String(http.StatusText(StatusMethodNotAllowed))
		}
		return ErrMethodNotAllowed
	}

	if a.notFound == nil && a.defaultErrors {
		return ctx.writeDefaultErrorResponse(StatusNotFound, "")
	}

	ctx.Status(StatusNotFound)
	if a.notFound != nil {
		if err := a.notFound(ctx); err != nil {
			return err
		}
		if !ctx.written {
			return ctx.String(http.StatusText(StatusNotFound))
		}
		return nil
	}
	if a.defaultErrors {
		return ctx.String(http.StatusText(StatusNotFound))
	}
	return ErrNotFound
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
		return c.app.dispatch(c)
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
