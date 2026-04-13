package middleware

import (
	"net/http/pprof"
	"strings"

	"github.com/0mjs/zinc"
)

const defaultPprofPrefix = "/debug/pprof"

func Pprof() zinc.Middleware {
	return PprofWithPrefix(defaultPprofPrefix)
}

func PprofWithPrefix(prefix string) zinc.Middleware {
	prefix = strings.TrimRight(prefix, "/")
	if prefix == "" {
		prefix = defaultPprofPrefix
	}
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	return func(c *zinc.Context) error {
		switch c.Path() {
		case prefix, prefix + "/":
			pprof.Index(c.Writer(), c.Request())
			return nil
		case prefix + "/cmdline":
			pprof.Cmdline(c.Writer(), c.Request())
			return nil
		case prefix + "/profile":
			pprof.Profile(c.Writer(), c.Request())
			return nil
		case prefix + "/symbol":
			pprof.Symbol(c.Writer(), c.Request())
			return nil
		case prefix + "/trace":
			pprof.Trace(c.Writer(), c.Request())
			return nil
		default:
			if strings.HasPrefix(c.Path(), prefix+"/") {
				pprof.Index(c.Writer(), c.Request())
				return nil
			}
			return c.Next()
		}
	}
}
