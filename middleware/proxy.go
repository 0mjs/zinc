package middleware

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/0mjs/zinc"
)

type ProxyConfig struct {
	Skipper      func(*zinc.Context) bool
	Target       string
	Director     func(*http.Request)
	Modify       func(*http.Response) error
	ErrorHandler func(http.ResponseWriter, *http.Request, error)
}

func Proxy(target string) zinc.Middleware {
	return ProxyWithConfig(ProxyConfig{Target: target})
}

func ProxyWithConfig(config ProxyConfig) zinc.Middleware {
	proxy := newReverseProxy(config)

	return func(c *zinc.Context) error {
		if config.Skipper != nil && config.Skipper(c) {
			return c.Next()
		}
		proxy.ServeHTTP(c.Writer(), c.Request())
		return nil
	}
}

func newReverseProxy(config ProxyConfig) *httputil.ReverseProxy {
	if config.Target == "" {
		panic("zincproxy: Target is required")
	}
	target, err := url.Parse(config.Target)
	if err != nil {
		panic("zincproxy: Target must be a valid URL")
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	if config.Director != nil {
		defaultDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			defaultDirector(req)
			config.Director(req)
		}
	}
	if config.Modify != nil {
		proxy.ModifyResponse = config.Modify
	}
	if config.ErrorHandler != nil {
		proxy.ErrorHandler = config.ErrorHandler
	}
	return proxy
}
