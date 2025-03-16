package zinc

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type App struct {
	router         *Router
	middleware     []Middleware
	services       map[string]any
	config         *Config
	cronScheduler  *CronScheduler
	templateEngine *TemplateEngine
	wsHandler      *WebSocketHandler
	fileUpload     *FileUpload
	validator      *Validator
}

type RouteHandler func(c *Context)

type Map map[string]any

func New() *App {
	return &App{
		router: &Router{
			cache: NewRouteCache(1000), // Cache size of 1000 entries
		},
		middleware:    make([]Middleware, 0),
		services:      make(map[string]any),
		config:        &DefaultConfig,
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
				route.handler(ctx)
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
			entry.handler(ctx)
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

				route.handler(ctx)
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

	// Create a new server with timeouts
	server := &http.Server{
		Addr:         addr,
		Handler:      a,
		ReadTimeout:  a.config.ReadTimeout,
		WriteTimeout: a.config.WriteTimeout,
		IdleTimeout:  a.config.IdleTimeout,
	}

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal received, graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(serverCtx, a.config.ShutdownTimeout)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				fmt.Println("Graceful shutdown timed out... forcing exit")
				os.Exit(1)
			}
		}()

		// Trigger graceful shutdown
		fmt.Println("Shutting down server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Printf("Error during shutdown: %v\n", err)
		}

		// Stop cron scheduler
		a.cronScheduler.Stop()

		serverStopCtx()
	}()

	// Start server
	fmt.Printf("Server starting on port %s...\n", serverPort)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	return nil
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

// SetConfig sets the application configuration
func (a *App) SetConfig(config *Config) {
	a.config = config
}
