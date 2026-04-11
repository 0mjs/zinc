package middleware

import (
	"net/http"

	"github.com/0mjs/zinc"
)

type RedirectConfig struct {
	Skipper    func(*zinc.Context) bool
	Rules      map[string]string
	StatusCode int
}

func Redirect(from, to string, statusCode ...int) zinc.Middleware {
	code := http.StatusMovedPermanently
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return RedirectWithConfig(RedirectConfig{
		Rules:      map[string]string{from: to},
		StatusCode: code,
	})
}

func RedirectWithRules(rules map[string]string, statusCode ...int) zinc.Middleware {
	code := http.StatusMovedPermanently
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	return RedirectWithConfig(RedirectConfig{
		Rules:      rules,
		StatusCode: code,
	})
}

func RedirectWithConfig(config RedirectConfig) zinc.Middleware {
	rules := cloneRewriteRules(config.Rules)
	statusCode := config.StatusCode
	if statusCode == 0 {
		statusCode = http.StatusMovedPermanently
	}

	return func(c *zinc.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		req := c.Request()
		if req == nil || req.URL == nil {
			return c.Next()
		}

		target, ok := rewriteTarget(req.URL.Path, rules)
		if !ok {
			return c.Next()
		}
		return c.Redirect(statusCode, pathWithRawQuery(target, req.URL.RawQuery))
	}
}
