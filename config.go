package zinc

import "time"

// Config holds the server configuration parameters.
type Config struct {
	// DefaultAddr specifies the HTTP server address.
	// If not set, the server will listen on "0.0.0.0:8080".
	DefaultAddr string
	// ShutdownTimeout specifies the maximum duration to wait for server shutdown.
	ShutdownTimeout time.Duration
	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request.
	IdleTimeout time.Duration
}

// DefaultConfig provides the default server configuration.
// It can be used as a base configuration for the server initialisation.
var DefaultConfig = Config{
	DefaultAddr:     "0.0.0.0:8080",
	ShutdownTimeout: 10 * time.Second,
	ReadTimeout:     5 * time.Second,
	WriteTimeout:    10 * time.Second,
	IdleTimeout:     120 * time.Second,
}
