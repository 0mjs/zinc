package middleware

import (
	"errors"

	"github.com/0mjs/zinc"
)

type CasbinEnforcer interface {
	Enforce(args ...any) (bool, error)
}

type CasbinValueFunc func(*zinc.Context) any

type CasbinErrorHandler func(*zinc.Context, error) error

type CasbinAuthConfig struct {
	Skipper        func(*zinc.Context) bool
	Enforcer       CasbinEnforcer
	Subject        CasbinValueFunc
	Object         CasbinValueFunc
	Action         CasbinValueFunc
	SuccessHandler zinc.RouteHandler
	ErrorHandler   CasbinErrorHandler
}

var ErrCasbinAuthRejected = errors.New("zinccasbin: request rejected")

func CasbinAuth(enforcer CasbinEnforcer, subject CasbinValueFunc) zinc.Middleware {
	return CasbinAuthWithConfig(CasbinAuthConfig{
		Enforcer: enforcer,
		Subject:  subject,
	})
}

func CasbinAuthWithConfig(config CasbinAuthConfig) zinc.Middleware {
	cfg := resolveCasbinAuthConfig(config)

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		ok, err := cfg.Enforcer.Enforce(cfg.Subject(c), cfg.Object(c), cfg.Action(c))
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}
		if !ok {
			return cfg.ErrorHandler(c, ErrCasbinAuthRejected)
		}
		return cfg.SuccessHandler(c)
	}
}

func CasbinSubjectFromBasicAuth() CasbinValueFunc {
	return func(c *zinc.Context) any {
		username, _ := BasicAuthUsername(c)
		return username
	}
}

func CasbinSubjectFromKeyAuth() CasbinValueFunc {
	return func(c *zinc.Context) any {
		state, ok := KeyAuthCurrent(c)
		if !ok {
			return ""
		}
		return state.Key
	}
}

func CasbinSubjectFromContext(key any) CasbinValueFunc {
	return func(c *zinc.Context) any {
		value, _ := c.Get(key)
		return value
	}
}

func CasbinObjectPath() CasbinValueFunc {
	return func(c *zinc.Context) any {
		return c.Path()
	}
}

func CasbinActionMethod() CasbinValueFunc {
	return func(c *zinc.Context) any {
		return c.Method()
	}
}

func resolveCasbinAuthConfig(config CasbinAuthConfig) CasbinAuthConfig {
	cfg := CasbinAuthConfig{
		Object: CasbinObjectPath(),
		Action: CasbinActionMethod(),
		SuccessHandler: func(c *zinc.Context) error {
			return c.Next()
		},
		ErrorHandler: func(_ *zinc.Context, err error) error {
			if errors.Is(err, ErrCasbinAuthRejected) {
				return zinc.ErrForbidden
			}
			return err
		},
	}
	cfg.Skipper = config.Skipper
	if config.Enforcer == nil {
		panic("zinccasbin: Enforcer is required")
	}
	cfg.Enforcer = config.Enforcer
	if config.Subject == nil {
		panic("zinccasbin: Subject is required")
	}
	cfg.Subject = config.Subject
	if config.Object != nil {
		cfg.Object = config.Object
	}
	if config.Action != nil {
		cfg.Action = config.Action
	}
	if config.SuccessHandler != nil {
		cfg.SuccessHandler = config.SuccessHandler
	}
	if config.ErrorHandler != nil {
		cfg.ErrorHandler = config.ErrorHandler
	}
	return cfg
}
