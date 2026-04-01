package zinc

import "time"

// Config holds the application and server configuration.
type Config struct {
	ServerHeader string

	CaseSensitive          bool
	StrictRouting          bool
	AutoHead               bool
	AutoOptions            bool
	HandleMethodNotAllowed bool

	BodyLimit    int64
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	ProxyHeader    string
	TrustedProxies []string

	RequestBinder RequestBinder
	Validator     Validator
	Renderer      Renderer
	JSONCodec     JSONCodec
	ErrorHandler  ErrorHandler

	RouteCacheSize int
}

// DefaultConfig is the baseline Zinc configuration.
var DefaultConfig = Config{
	ServerHeader:           "",
	CaseSensitive:          false,
	StrictRouting:          false,
	AutoHead:               true,
	AutoOptions:            true,
	HandleMethodNotAllowed: true,
	BodyLimit:              4 << 20,
	ReadTimeout:            5 * time.Second,
	WriteTimeout:           10 * time.Second,
	IdleTimeout:            120 * time.Second,
	ProxyHeader:            "X-Forwarded-For",
	TrustedProxies:         nil,
	RouteCacheSize:         1000,
}
