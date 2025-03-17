package zinc

import "time"

// Config holds the server configuration parameters.
type Config struct {
	// Server configuration
	// DefaultAddr specifies the HTTP server address.
	// If not set, the server will listen on "0.0.0.0:8080".
	DefaultAddr string
	// ServerHeader sets the value of the Server HTTP header.
	ServerHeader string
	// AppName specifies the name of the application.
	AppName string

	// Timeout settings
	// ShutdownTimeout specifies the maximum duration to wait for server shutdown.
	ShutdownTimeout time.Duration
	// ReadTimeout is the maximum duration for reading the entire request.
	ReadTimeout time.Duration
	// WriteTimeout is the maximum duration before timing out writes of the response.
	WriteTimeout time.Duration
	// IdleTimeout is the maximum amount of time to wait for the next request.
	IdleTimeout time.Duration

	// Router settings
	// CaseSensitive determines if routes should be case-sensitive.
	// When false, /Foo and /foo are treated as the same route.
	CaseSensitive bool
	// StrictRouting determines if routes with trailing slashes are different than those without.
	// When false, /api and /api/ are treated as the same route.
	StrictRouting bool
	// BodyLimit sets the maximum allowed size for a request body in bytes.
	// Default is 4MB.
	BodyLimit int64
	// Concurrency sets the maximum number of concurrent connections.
	Concurrency int
	// EnableTrustedProxyCheck enables checking for trusted proxies
	// when determining client IP addresses.
	EnableTrustedProxyCheck bool
	// TrustedProxies is a list of IP addresses or CIDR blocks that are trusted.
	TrustedProxies []string
	// ProxyHeader specifies which header to use for client IP address.
	ProxyHeader string
	// DisableKeepalive disables keep-alive connections.
	DisableKeepalive bool
	// DisableDefaultContentType disables sending the Content-Type header in responses
	// when no content type is explicitly set.
	DisableDefaultContentType bool

	// Routing cache settings
	// RouteCacheSize determines how many routes are stored in the route cache.
	// Set to 0 to disable caching.
	RouteCacheSize int

	// Debug settings
	// DisableStartupMessage disables the startup message when the server starts.
	DisableStartupMessage bool
	// EnablePrintRoutes enables printing all routes on startup.
	EnablePrintRoutes bool
}

// DefaultConfig provides the default server configuration.
// It can be used as a base configuration for the server initialisation.
var DefaultConfig = Config{
	DefaultAddr:               "0.0.0.0:8080",
	ServerHeader:              "Zinc",
	AppName:                   "",
	ShutdownTimeout:           10 * time.Second,
	ReadTimeout:               5 * time.Second,
	WriteTimeout:              10 * time.Second,
	IdleTimeout:               120 * time.Second,
	CaseSensitive:             false,
	StrictRouting:             false,
	BodyLimit:                 4 * 1024 * 1024, // 4MB
	Concurrency:               256 * 1024,      // 256K concurrent connections
	EnableTrustedProxyCheck:   false,
	TrustedProxies:            []string{},
	ProxyHeader:               "X-Forwarded-For",
	DisableKeepalive:          false,
	DisableDefaultContentType: false,
	RouteCacheSize:            1000,
	DisableStartupMessage:     false,
	EnablePrintRoutes:         false,
}
