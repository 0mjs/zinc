package cors

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/0mjs/zinc"
)

// Config defines the configuration options for CORS middleware.
type Config struct {
	// AllowOrigins defines a list of origins that may access the resource.
	// Default value is ["*"] which allows all origins.
	AllowOrigins []string

	// AllowMethods defines a list of methods allowed when accessing the resource.
	// This is used in response to a preflight request.
	// Default value is [GET, POST, HEAD, PUT, DELETE, PATCH]
	AllowMethods []string

	// AllowHeaders defines a list of request headers that can be used when
	// making the actual request. This is in response to a preflight request.
	// Default value is [Origin, Content-Type, Accept, Authorization]
	AllowHeaders []string

	// ExposeHeaders defines a whitelist headers that clients are allowed to access.
	// Default value is []
	ExposeHeaders []string

	// AllowCredentials indicates whether the request can include user credentials like
	// cookies, HTTP authentication or client side SSL certificates.
	// Default value is false
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached.
	// Default value is 0 which means no max age (browser default)
	MaxAge int

	// Skipper defines a function to skip middleware.
	// Default value is nil
	Skipper func(c *zinc.Context) bool
}

// DefaultConfig returns a default configuration for CORS middleware.
func DefaultConfig() Config {
	return Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodHead,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           0,
	}
}

// New returns a new CORS middleware with default configuration.
func New() zinc.Middleware {
	return NewWithConfig(DefaultConfig())
}

// Option defines a function that modifies the CORS configuration.
type Option func(*Config)

// WithAllowOrigins sets the allowed origins.
func WithAllowOrigins(origins ...string) Option {
	return func(c *Config) {
		c.AllowOrigins = origins
	}
}

// WithAllowMethods sets the allowed HTTP methods.
func WithAllowMethods(methods ...string) Option {
	return func(c *Config) {
		c.AllowMethods = methods
	}
}

// WithAllowHeaders sets the allowed headers.
func WithAllowHeaders(headers ...string) Option {
	return func(c *Config) {
		c.AllowHeaders = headers
	}
}

// WithExposeHeaders sets the exposed headers.
func WithExposeHeaders(headers ...string) Option {
	return func(c *Config) {
		c.ExposeHeaders = headers
	}
}

// WithAllowCredentials sets the allow credentials flag.
func WithAllowCredentials(allow bool) Option {
	return func(c *Config) {
		c.AllowCredentials = allow
	}
}

// WithMaxAge sets the maximum age for preflight requests.
func WithMaxAge(maxAge time.Duration) Option {
	return func(c *Config) {
		c.MaxAge = int(maxAge.Seconds())
	}
}

// WithMaxAgeSeconds sets the maximum age for preflight requests in seconds.
func WithMaxAgeSeconds(seconds int) Option {
	return func(c *Config) {
		c.MaxAge = seconds
	}
}

// WithSkipper sets the skipper function.
func WithSkipper(skipper func(c *zinc.Context) bool) Option {
	return func(c *Config) {
		c.Skipper = skipper
	}
}

// NewWithOptions returns a new CORS middleware with the provided options.
func NewWithOptions(options ...Option) zinc.Middleware {
	config := DefaultConfig()
	for _, option := range options {
		option(&config)
	}
	return NewWithConfig(config)
}

// NewWithConfig returns a new CORS middleware with the provided config.
func NewWithConfig(config Config) zinc.Middleware {
	// Convert origins to map for faster lookup
	allowOriginsMap := make(map[string]bool, len(config.AllowOrigins))
	for _, origin := range config.AllowOrigins {
		allowOriginsMap[origin] = true
	}

	// Convert headers and methods to uppercase for consistent comparison
	allowMethodsStr := strings.Join(config.AllowMethods, ",")
	allowHeadersStr := strings.Join(config.AllowHeaders, ",")
	exposeHeadersStr := strings.Join(config.ExposeHeaders, ",")

	// Return the middleware function
	return func(c *zinc.Context) error {
		// Skip if requested
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		// Set vary header for caching
		c.Response.Header().Add("Vary", "Origin")
		c.Response.Header().Add("Vary", "Access-Control-Request-Method")
		c.Response.Header().Add("Vary", "Access-Control-Request-Headers")

		// Get origin from request
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			// Not a CORS request, continue with next handler
			return c.Next()
		}

		// Check if origin is allowed
		allowOrigin := ""
		if config.AllowOrigins[0] == "*" && !config.AllowCredentials {
			// Allow all origins if wildcard is set and credentials aren't allowed
			allowOrigin = "*"
		} else if allowOriginsMap[origin] {
			// Origin is specifically allowed
			allowOrigin = origin
		} else if allowOriginsMap["*"] {
			// Wildcard is allowed, but credentials are enabled, so echo the origin
			allowOrigin = origin
		} else {
			// Origin not allowed
			// For security and privacy reasons, don't provide specific error
			// Continue with next handler without setting CORS headers
			return c.Next()
		}

		// Set allow origin header
		c.Response.Header().Set("Access-Control-Allow-Origin", allowOrigin)

		// Set allow credentials header if enabled
		if config.AllowCredentials {
			c.Response.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Handle preflight requests
		if c.Method == http.MethodOptions {
			requestMethod := c.Request.Header.Get("Access-Control-Request-Method")
			if requestMethod != "" {
				// Set allowed methods
				c.Response.Header().Set("Access-Control-Allow-Methods", allowMethodsStr)

				// Set allowed headers
				requestHeaders := c.Request.Header.Get("Access-Control-Request-Headers")
				if requestHeaders != "" {
					c.Response.Header().Set("Access-Control-Allow-Headers", allowHeadersStr)
				}

				// Set max age if configured - always set it for preflight requests
				if config.MaxAge > 0 {
					c.Response.Header().Set("Access-Control-Max-Age", strconv.Itoa(config.MaxAge))
				}

				// Return OK for preflight requests
				return c.Status(http.StatusNoContent).Send("")
			}
		}

		// For actual requests, set exposed headers if configured
		if len(config.ExposeHeaders) > 0 {
			c.Response.Header().Set("Access-Control-Expose-Headers", exposeHeadersStr)
		}

		// Continue with next handler
		return c.Next()
	}
}
