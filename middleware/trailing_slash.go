package middleware

import (
	"net/http"
	"strings"

	"github.com/0mjs/zinc"
)

type TrailingSlashConfig struct {
	Skipper    func(*zinc.Context) bool
	Add        bool
	Redirect   bool
	StatusCode int
}

func TrailingSlash() zinc.Middleware {
	return RemoveTrailingSlash()
}

func RemoveTrailingSlash() zinc.Middleware {
	return TrailingSlashWithConfig(TrailingSlashConfig{})
}

func AddTrailingSlash() zinc.Middleware {
	return TrailingSlashWithConfig(TrailingSlashConfig{Add: true})
}

func TrailingSlashWithConfig(config TrailingSlashConfig) zinc.Middleware {
	cfg := resolveTrailingSlashConfig(config)

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		req := c.Request()
		if req == nil || req.URL == nil {
			return c.Next()
		}

		nextPath := normalizeTrailingSlashPath(req.URL.Path, cfg.Add)
		if nextPath == req.URL.Path {
			return c.Next()
		}

		if cfg.Redirect {
			return c.Redirect(cfg.StatusCode, pathWithRawQuery(nextPath, req.URL.RawQuery))
		}

		c.SetPath(nextPath)
		return c.Next()
	}
}

func resolveTrailingSlashConfig(config TrailingSlashConfig) TrailingSlashConfig {
	if config.StatusCode == 0 {
		config.StatusCode = http.StatusMovedPermanently
	}
	return config
}

func normalizeTrailingSlashPath(path string, add bool) string {
	if path == "" {
		return "/"
	}
	if path == "/" {
		return path
	}
	if add {
		if strings.HasSuffix(path, "/") {
			return path
		}
		return path + "/"
	}
	return strings.TrimRight(path, "/")
}

func pathWithRawQuery(path, rawQuery string) string {
	if rawQuery == "" {
		return path
	}
	return path + "?" + rawQuery
}
