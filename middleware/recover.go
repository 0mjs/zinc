package middleware

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/0mjs/zinc"
)

type RecoverHandler func(*zinc.Context, *RecoverError) error

type RecoverConfig struct {
	Skipper      func(*zinc.Context) bool
	Handler      RecoverHandler
	StackSize    int
	DisableStack bool
}

type RecoverError struct {
	Value any
	Stack []byte
}

func (e *RecoverError) Error() string {
	if e == nil {
		return "zincrecover: panic recovered"
	}
	return fmt.Sprintf("zincrecover: panic recovered: %v", e.Value)
}

func DefaultRecoverConfig() RecoverConfig {
	return RecoverConfig{
		Handler:   defaultRecoverHandler,
		StackSize: 4 << 10,
	}
}

func Recover() zinc.Middleware {
	return RecoverWithConfig(DefaultRecoverConfig())
}

func RecoverWithConfig(config RecoverConfig) zinc.Middleware {
	cfg := resolveRecoverConfig(config)

	return func(c *zinc.Context) (err error) {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		defer func() {
			if value := recover(); value != nil {
				recoverErr := &RecoverError{
					Value: value,
					Stack: captureRecoverStack(cfg),
				}
				err = cfg.Handler(c, recoverErr)
			}
		}()

		return c.Next()
	}
}

func resolveRecoverConfig(config RecoverConfig) RecoverConfig {
	cfg := DefaultRecoverConfig()
	cfg.Skipper = config.Skipper
	if config.Handler != nil {
		cfg.Handler = config.Handler
	}
	if config.StackSize != 0 {
		cfg.StackSize = config.StackSize
	}
	cfg.DisableStack = config.DisableStack
	if cfg.Handler == nil {
		panic("zincrecover: Handler is required")
	}
	if cfg.StackSize < 0 {
		panic("zincrecover: StackSize must be greater than or equal to zero")
	}
	return cfg
}

func captureRecoverStack(cfg RecoverConfig) []byte {
	if cfg.DisableStack || cfg.StackSize == 0 {
		return nil
	}
	stack := make([]byte, cfg.StackSize)
	n := runtime.Stack(stack, false)
	return stack[:n]
}

func defaultRecoverHandler(_ *zinc.Context, err *RecoverError) error {
	return errors.Join(zinc.ErrInternalServerError, err)
}
