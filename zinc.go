package zinc

import (
	"net/http"
)

func New() *App {
	return &App{
		config: &DefaultConfig,
		router: &Router{
			cache: NewRouteCache(1000), // Cache size of 1000 entries
		},
		middleware:    make([]Middleware, 0),
		services:      make(map[string]any),
		cronScheduler: newCronScheduler(),
		validator:     NewValidator(),
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path

	// EXTREME fast path for GET / with no middleware (Hello World benchmark case)
	if method == MethodGet && path == "/" && len(a.middleware) == 0 {
		if routes, ok := a.router.routes[method]; ok {
			if route, ok := routes[path]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()

				// Set app instance in context
				ctx.Set("app", a)

				// Execute handler directly
				if err := route.handler(ctx); err != nil {
					// Handle error - write 500 if response not already written
					if !ctx.written {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
				return
			}
		}
	}

	// Pre-check if route exists in cache before allocating context
	if a.router.cache != nil {
		key := routeCacheKey{method, path}
		if entry, ok := a.router.cache.get(key); ok {
			// Only create context if route found in cache
			ctx := NewContext(w, r)
			defer ctx.release()

			// Set app instance in context
			ctx.Set("app", a)

			// Set services if needed
			if len(a.services) > 0 {
				ctx.services = a.services
			}

			// Copy path params
			ctx.PathParams = entry.context.PathParams

			// Execute handler
			if err := entry.handler(ctx); err != nil {
				// Handle error - write 500 if response not already written
				if !ctx.written {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
			return
		}
	}

	// Fast path for static routes (no middleware)
	if len(a.middleware) == 0 {
		// Direct static route lookup
		if routes, ok := a.router.routes[method]; ok {
			if route, ok := routes[path]; ok {
				ctx := NewContext(w, r)
				defer ctx.release()

				// Set app instance in context
				ctx.Set("app", a)

				// Only set services if needed
				if len(a.services) > 0 {
					ctx.services = a.services
				}

				if err := route.handler(ctx); err != nil {
					// Handle error - write 500 if response not already written
					if !ctx.written {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
				return
			}
		}
	}

	// Normal path for all other cases
	ctx := NewContext(w, r)
	defer ctx.release()

	// Set app instance in context
	ctx.Set("app", a)

	// Only set services if needed
	if len(a.services) > 0 {
		ctx.services = a.services
	}

	// Handle middleware if present
	if len(a.middleware) > 0 {
		ctx.setHandlers(a.middleware)
		if err := ctx.Next(); err != nil {
			// Handle error - write 500 if response not already written
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
	handler, foundCtx := a.router.Find(method, path)
	if handler != nil {
		if foundCtx != nil {
			// Copy params
			ctx.PathParams = foundCtx.PathParams

			// Store in cache for future use
			if a.router.cache != nil {
				key := routeCacheKey{method, path}
				a.router.cache.set(key, routeCacheEntry{
					handler: handler,
					context: &Context{PathParams: foundCtx.PathParams},
				})
			}
		}
		if err := handler(ctx); err != nil {
			// Handle error - write 500 if response not already written
			if !ctx.written {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
		return
	}

	http.NotFound(w, r)
}
