package zinc

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
)

type App struct {
	router         *Router
	middleware     []Middleware
	services       map[string]interface{}
	config         *Config
	cronScheduler  *CronScheduler
	templateEngine *TemplateEngine
	wsHandler      *WebSocketHandler
	fileUpload     *FileUpload
	validator      *Validator
}

type RouteHandler func(c *Context)

type Map map[string]interface{}

func New() *App {
	return &App{
		router: &Router{
			cache: NewRouteCache(1000), // Cache size of 1000 entries
		},
		middleware:    make([]Middleware, 0),
		services:      make(map[string]interface{}),
		config:        &DefaultConfig,
		cronScheduler: newCronScheduler(),
		validator:     NewValidator(),
	}
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extreme Fast path for common case - no middleware, direct static routes
	// This optimization significantly improves performance for simple routes
	if len(a.middleware) == 0 {
		method := r.Method
		path := r.URL.Path

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

				route.handler(ctx)
				return
			}
		}

		// Try cached route lookup for common dynamic routes
		key := routeCacheKey{method, path}
		if a.router.cache != nil {
			if entry, ok := a.router.cache.get(key); ok {
				ctx := NewContext(w, r)
				defer ctx.release()

				// Set app instance in context
				ctx.Set("app", a)

				if len(a.services) > 0 {
					ctx.services = a.services
				}

				ctx.PathParams = entry.context.PathParams
				entry.handler(ctx)
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
		ctx.Next()
		if ctx.written {
			return
		}
	}

	// Find and execute route handler
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
	// Start cron scheduler
	a.cronScheduler.Start()

	serverPort := a.config.DefaultAddr
	if len(port) > 0 && port[0] != "" {
		serverPort = port[0]
	}
	addr := serverPort
	if !strings.Contains(serverPort, ":") {
		addr = ":" + serverPort
	}
	fmt.Printf("Server starting on port %s...\n", serverPort)
	return http.ListenAndServe(addr, a)
}

// Cron adds a cron job to be executed on the given schedule
// Supports both error-returning and non-error-returning handlers
func (a *App) Cron(id string, schedule string, handler interface{}) error {
	switch h := handler.(type) {
	case func() error:
		return a.cronScheduler.AddJob(id, schedule, h)
	case func():
		wrappedHandler := func() error {
			h()
			return nil
		}
		return a.cronScheduler.AddJob(id, schedule, wrappedHandler)
	default:
		return fmt.Errorf("unsupported handler type: handler must be func() or func() error")
	}
}

// For backward compatibility
// Schedule adds a cron job to be executed on the given schedule
func (a *App) Schedule(id string, schedule string, handler func() error) error {
	return a.Cron(id, schedule, handler)
}

// For backward compatibility
// ScheduleFunc adds a cron job that executes a function without returning an error
func (a *App) ScheduleFunc(id string, schedule string, handler func()) error {
	return a.Cron(id, schedule, handler)
}

// RemoveCron removes a scheduled job by ID
func (a *App) RemoveCron(id string) {
	a.cronScheduler.RemoveJob(id)
}

// For backward compatibility
// RemoveSchedule removes a scheduled job by ID
func (a *App) RemoveSchedule(id string) {
	a.RemoveCron(id)
}

// StopScheduler stops all scheduled jobs
func (a *App) StopScheduler() {
	a.cronScheduler.Stop()
}

// StartScheduler starts the scheduler
func (a *App) StartScheduler() {
	a.cronScheduler.Start()
}

// SetTemplateEngine sets the template engine
func (a *App) SetTemplateEngine(engine *TemplateEngine) {
	a.templateEngine = engine
}

// SetWebSocketHandler sets the WebSocket handler
func (a *App) SetWebSocketHandler(handler *WebSocketHandler) {
	a.wsHandler = handler
}

// SetFileUpload sets the file upload handler
func (a *App) SetFileUpload(upload *FileUpload) {
	a.fileUpload = upload
}

// Validate validates a struct using the validator
func (a *App) Validate(s interface{}) ValidationErrors {
	return a.validator.Validate(s)
}

func parseArgs(a *App) string {
	port := flag.String("port", a.config.DefaultAddr, "port number for the server")
	flag.Parse()
	return *port
}
