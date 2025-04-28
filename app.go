package zinc

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"
)

type App struct {
	config         *Config
	router         *Router
	middleware     []Middleware
	services       map[reflect.Type]any
	cronScheduler  *CronScheduler
	templateEngine *TemplateEngine
	wsHandler      *WebSocketHandler
	uploader       *FileUpload
	validator      *Validator
	db             *sql.DB
}

type Map map[string]any

// Use adds middleware to the application
func (a *App) Use(middleware ...Middleware) {
	a.middleware = append(a.middleware, middleware...)
}

// Register registers a service with the application by its concrete type
// This allows for type-safe retrieval using GetService() later
func (a *App) Injectable(service interface{}) {
	typ := reflect.TypeOf(service)
	a.services[typ] = service
}

// GetService retrieves a service by its concrete type
// Returns the service or nil if not found
func (a *App) GetService(serviceType reflect.Type) interface{} {
	return a.services[serviceType]
}

// ServiceOf is a convenience function to retrieve a service by type T
// Usage example: service := app.ServiceOf[*UserService]()
func ServiceOf[T any](a *App) (service T, ok bool) {
	typ := reflect.TypeOf((*T)(nil)).Elem()
	if s := a.services[typ]; s != nil {
		if svc, isOk := s.(T); isOk {
			return svc, true
		}
	}
	var zero T
	return zero, false
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
			fmt.Printf("Error during server shutdown: %v\n", err)
		}

		// Close database connection if it exists
		if a.db != nil {
			fmt.Println("Closing database connection...")
			if err := a.db.Close(); err != nil {
				fmt.Printf("Error closing database connection: %v\n", err)
			}
		}

		// Stop cron scheduler
		a.cronScheduler.Stop()

		serverStopCtx()
	}()

	// Print startup message if not disabled
	if !a.config.DisableStartupMessage {
		appName := "Zinc"
		appVersion := Version
		if a.config.AppName != "" {
			appName = a.config.AppName
		}
		if a.config.AppVersion != "" {
			appVersion = a.config.AppVersion
		}
		fmt.Printf("%s v%s server starting on %s...\n", appName, appVersion, serverPort)
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

// RemoveCRON removes a scheduled job by ID
func (a *App) RemoveCRON(id string) {
	a.cronScheduler.RemoveJob(id)
}

// StopScheduler stops all scheduled jobs
func (a *App) StopCRON() {
	a.cronScheduler.Stop()
}

// StartCRON starts the scheduler
func (a *App) StartCRON() {
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
func (a *App) Validate(s any) ValidationErrors {
	return a.validator.Validate(s)
}

// SetConfig sets the application configuration
func (a *App) SetConfig(config *Config) {
	a.config = config
}

// Version returns the current version of the Zinc framework.
func (a *App) Version() string {
	return Version
}

// Group creates a new route group with a specified prefix
func (a *App) Group(prefix string) *Group {
	return NewGroup(a, prefix)
}

// ConnectDB establishes a connection to the database and stores the pool.
// It requires the driver name (e.g., "postgres", "sqlite3") and the DSN.
// Remember to import the specific database driver in your main package (e.g., _ "github.com/lib/pq").
func (a *App) ConnectDB(driverName, dataSourceName string) error {
	// Sanitise and allow sqlite string
	if driverName == "sqlite" {
		driverName = "sqlite3"
	}
	// Run DB in memory by default, if no source provided
	if dataSourceName == "" {
		dataSourceName = ":memory:"
	}

	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database connection for driver %s: %w", driverName, err)
	}

	// Ping the database to verify the connection.
	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return fmt.Errorf("failed to connect to database (driver: %s): %w", driverName, err)
	}

	a.db = db
	// TODO: Add configuration options for connection pool settings (MaxOpenConns, MaxIdleConns, ConnMaxLifetime) via Config struct.
	fmt.Printf("Successfully connected to database using driver: %s\n", driverName)
	return nil
}

// GetDB retrieves the configured database connection pool.
// Returns nil if no database connection has been established.
func (a *App) GetDB() *sql.DB {
	return a.db
}
