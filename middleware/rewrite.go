package middleware

import (
	"strings"

	"github.com/0mjs/zinc"
)

type RewriteConfig struct {
	Skipper func(*zinc.Context) bool
	Rules   map[string]string
}

func Rewrite(from, to string) zinc.Middleware {
	return RewriteWithRules(map[string]string{from: to})
}

func RewriteWithRules(rules map[string]string) zinc.Middleware {
	return RewriteWithConfig(RewriteConfig{Rules: rules})
}

func RewriteWithConfig(config RewriteConfig) zinc.Middleware {
	rules := cloneRewriteRules(config.Rules)

	return func(c *zinc.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}

		req := c.Request()
		if req == nil || req.URL == nil {
			return c.Next()
		}

		if target, ok := rewriteTarget(req.URL.Path, rules); ok {
			c.SetPath(target)
		}
		return c.Next()
	}
}

func cloneRewriteRules(rules map[string]string) map[string]string {
	out := make(map[string]string, len(rules))
	for from, to := range rules {
		if from == "" {
			continue
		}
		out[from] = to
	}
	return out
}

func rewriteTarget(path string, rules map[string]string) (string, bool) {
	if len(rules) == 0 {
		return "", false
	}
	if target, ok := rules[path]; ok {
		return target, true
	}
	for from, to := range rules {
		if !strings.HasSuffix(from, "*") {
			continue
		}
		prefix := strings.TrimSuffix(from, "*")
		if !strings.HasPrefix(path, prefix) {
			continue
		}
		tail := strings.TrimPrefix(path, prefix)
		if strings.Contains(to, "*") {
			return strings.Replace(to, "*", tail, 1), true
		}
		return to + tail, true
	}
	return "", false
}
