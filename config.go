package zinc

// Config holds the server configuration parameters.
type Config struct {
	// DefaultAddr specifies the HTTP server address.
	DefaultAddr string
}

// DefaultConfig provides the default server configuration.
// It can be used as a base configuration for the server initialisation.
var DefaultConfig = Config{
	DefaultAddr: "0.0.0.0:8080",
}
