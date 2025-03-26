package zinc

import (
	"net/http"
	"reflect"
	"strings"
)

/*
New creates a new Zinc application instance with the specified configuration.
If no configuration is provided, the default configuration is used.
*/
func New(config ...Config) *App {
	cfg := DefaultConfig
	if len(config) > 0 {
		cfg = config[0]

		if cfg.DefaultAddr == "" {
			cfg.DefaultAddr = DefaultConfig.DefaultAddr
		}
	}

	var cache *RouteCache
	if cfg.RouteCacheSize > 0 {
		cache = NewRouteCache(cfg.RouteCacheSize)
	}

	return &App{
		config: &cfg,
		router: &Router{
			cache:  cache,
			config: &cfg,
		},
		middleware:    make([]Middleware, 0),
		services:      make(map[reflect.Type]any),
		cronScheduler: newCronScheduler(),
		validator:     NewValidator(),
	}
}

/*
ServeHTTP is the default HTTP handler for the Zinc application.
*/
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	// Try fast path first
	if a.handleFastPath(w, r, method, path) {
		return
	}

	// Prepare response headers
	a.prepareResponse(w)

	// Special handling for strict routing with trailing slash
	if a.config.StrictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		// In strict routing, paths with trailing slashes are distinct
		// Check if we have an exact match first
		if routes, ok := a.router.routes[method]; ok {
			if route, ok := routes[path]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()
				ctx.app = a
				if err := route.handler(ctx); err != nil && !ctx.written {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}
		}

		// If no exact match and strict routing, don't try without the slash
		if len(a.middleware) == 0 {
			http.NotFound(w, r)
			return
		}
	}

	// Normalize path based on configuration
	normalizedPath, pathModified, _ := a.normalizePath(path)

	// Try route cache if enabled
	if a.router.cache != nil {
		key := routeCacheKey{method, normalizedPath}
		if entry, ok := a.router.cache.get(key); ok {
			ctx := NewContext(w, r)
			defer ctx.release()
			ctx.app = a
			ctx.services = a.services
			ctx.PathParams = entry.context.PathParams

			if err := entry.handler(ctx); err != nil && !ctx.written {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// Try direct route lookup for static routes without middleware
	if len(a.middleware) == 0 {
		// With strict routing, we should only use the original path
		if a.config.StrictRouting {
			// Only try the original path
			if routes, ok := a.router.routes[method]; ok {
				if route, ok := routes[path]; ok {
					ctx := NewContext(w, r)
					defer ctx.release()
					ctx.app = a
					if err := route.handler(ctx); err != nil && !ctx.written {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}
			http.NotFound(w, r)
			return
		}

		// For non-strict routing, try both paths
		if a.tryDirectRoute(w, r, method, normalizedPath, path, pathModified) {
			return
		}
	}

	// Normal path with middleware and dynamic routes
	ctx := NewContext(w, r)
	defer ctx.release()
	ctx.app = a
	ctx.services = a.services

	// Handle middleware
	if len(a.middleware) > 0 {
		ctx.setHandlers(a.middleware)
		if err := ctx.Next(); err != nil {
			if !ctx.written {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		if ctx.written {
			return
		}
	}

	// Find and execute route handler
	var handler RouteHandler
	var foundCtx *Context

	// In strict routing mode, only try the original path
	if a.config.StrictRouting {
		handler, foundCtx = a.router.Find(method, path)
	} else {
		// Try with original path first if path was modified
		if pathModified {
			handler, foundCtx = a.router.Find(method, path)
		}

		// If not found with original path, try with normalized path
		if handler == nil {
			handler, foundCtx = a.router.Find(method, normalizedPath)
		}
	}

	if handler != nil {
		if foundCtx != nil {
			ctx.PathParams = foundCtx.PathParams
			ctx.app = a

			// Cache the route if caching is enabled
			if a.router.cache != nil {
				cachePath := normalizedPath
				if pathModified && handler != nil {
					_, testCtx := a.router.Find(method, path)
					if testCtx != nil {
						cachePath = path
					}
				}

				key := routeCacheKey{method, cachePath}
				a.router.cache.set(key, routeCacheEntry{
					handler: handler,
					context: &Context{PathParams: foundCtx.PathParams},
				})
			}
		}

		if err := handler(ctx); err != nil && !ctx.written {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	http.NotFound(w, r)
}

// prepareResponse sets up common response headers based on configuration
func (a *App) prepareResponse(w http.ResponseWriter) {
	if a.config.ServerHeader != "" {
		w.Header().Set("Server", a.config.ServerHeader)
	}

	// Only set Content-Type if not disabled and not already set
	if !a.config.DisableDefaultContentType && w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}
}

// normalizePath handles path modifications based on configuration
func (a *App) normalizePath(path string) (string, bool, bool) {
	pathModified := false
	trailingSlashModified := false

	// If strict routing is enabled, we should not normalize the path at all
	if a.config.StrictRouting {
		return path, false, false
	}

	// Handle case sensitivity setting
	if !a.config.CaseSensitive {
		path = strings.ToLower(path)
		pathModified = true
	}

	// Handle trailing slash - only remove if not in strict routing mode
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
		pathModified = true
		trailingSlashModified = true
	}

	return path, pathModified, trailingSlashModified
}

// handleFastPath attempts to handle the request via the fast path
func (a *App) handleFastPath(w http.ResponseWriter, r *http.Request, method, path string) bool {
	if method == MethodGet && path == "/" && len(a.middleware) == 0 {
		if routes, ok := a.router.routes[method]; ok {
			if route, ok := routes[path]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()
				ctx.app = a

				if err := route.handler(ctx); err != nil && !ctx.written {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return true
			}
		}
	}
	return false
}

// tryDirectRoute attempts to handle the request via direct route lookup
func (a *App) tryDirectRoute(w http.ResponseWriter, r *http.Request, method, path, originalPath string, pathModified bool) bool {
	if routes, ok := a.router.routes[method]; ok {
		// In strict routing mode, only try the exact original path
		if a.config.StrictRouting {
			// In strict mode, don't use the normalized path, only try the original
			if route, ok := routes[originalPath]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()
				ctx.app = a
				if err := route.handler(ctx); err != nil && !ctx.written {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return true
			}
			return false
		}

		// Non-strict routing: try both paths
		if pathModified {
			if route, ok := routes[originalPath]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()
				ctx.app = a
				if err := route.handler(ctx); err != nil && !ctx.written {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return true
			}
		}

		if route, ok := routes[path]; ok {
			ctx := NewContext(w, r)
			defer ctx.release()
			ctx.app = a
			if err := route.handler(ctx); err != nil && !ctx.written {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return true
		}
	}
	return false
}
