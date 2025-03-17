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
	config         *Config
	router         *Router
	middleware     []Middleware
	services       map[string]any
	cronScheduler  *CronScheduler
	templateEngine *TemplateEngine
	wsHandler      *WebSocketHandler
	uploader       *FileUpload
	validator      *Validator
}

type RouteHandler func(c *Context) error

type Map map[string]any

// Use adds middleware to the application
func (a *App) Use(middleware ...Middleware) {
	a.middleware = append(a.middleware, middleware...)
}

// Service registers a service with the application
func (a *App) Service(name string, service interface{}) {
	a.services[name] = service
}

// Serve starts the HTTP server on the specified port
func (a *App) Serve(port ...string) error {
	// Start cron scheduler
	a.cronScheduler.Start()

	// Apply provided port or use config default
	serverPort := a.config.DefaultAddr
	if len(port) > 0 && port[0] != "" {
		serverPort = port[0]
	}
	addr := serverPort
	if !strings.Contains(serverPort, ":") {
		addr = ":" + serverPort
	}

	// Create a new server with timeouts and other config options
	server := &http.Server{
		Addr:         addr,
		Handler:      a,
		ReadTimeout:  a.config.ReadTimeout,
		WriteTimeout: a.config.WriteTimeout,
		IdleTimeout:  a.config.IdleTimeout,
		// Set max concurrent connections if configured
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	// Set TCP keep-alive settings
	if a.config.DisableKeepalive {
		server.SetKeepAlivesEnabled(false)
	}

	// Optionally print routes
	if a.config.EnablePrintRoutes {
		a.printRoutes()
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

	// Print startup message if not disabled
	if !a.config.DisableStartupMessage {
		appName := "Zinc"
		if a.config.AppName != "" {
			appName = a.config.AppName
		}
		fmt.Printf("%s server starting on %s...\n", appName, serverPort)
	}

	// Start server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
	return nil
}

// printRoutes prints all registered routes
func (a *App) printRoutes() {
	fmt.Println("Registered Routes:")
	fmt.Println("┌───────┬─────────────────────────────┐")
	fmt.Println("│ METHOD │ PATH                       │")
	fmt.Println("├───────┼─────────────────────────────┤")

	if a.router != nil && a.router.routes != nil {
		for method, routes := range a.router.routes {
			for path := range routes {
				fmt.Printf("│ %-5s │ %-27s │\n", method, path)
			}
		}
	}

	fmt.Println("└───────┴─────────────────────────────┘")
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
	a.uploader = upload
}

// Validate validates a struct using the validator
func (a *App) Validate(s interface{}) ValidationErrors {
	return a.validator.Validate(s)
}

// SetConfig sets the application configuration
func (a *App) SetConfig(config *Config) {
	a.config = config
}
