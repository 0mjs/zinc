package middleware

import (
	"fmt"
	"strings"

	"github.com/0mjs/zinc"
)

type SecureConfig struct {
	Skipper                         func(*zinc.Context) bool
	XSSProtection                   string
	ContentTypeNosniff              string
	XFrameOptions                   string
	HSTSMaxAge                      int
	HSTSExcludeSubdomains           bool
	ContentSecurityPolicy           string
	ContentSecurityPolicyReportOnly string
	ReferrerPolicy                  string
	CrossOriginResourcePolicy       string
	PermissionsPolicy               string
}

func DefaultSecureConfig() SecureConfig {
	return SecureConfig{
		XSSProtection:             "0",
		ContentTypeNosniff:        "nosniff",
		XFrameOptions:             "SAMEORIGIN",
		ReferrerPolicy:            "no-referrer",
		CrossOriginResourcePolicy: "same-origin",
	}
}

func Secure() zinc.Middleware {
	return SecureWithConfig(DefaultSecureConfig())
}

func SecureWithConfig(config SecureConfig) zinc.Middleware {
	cfg := resolveSecureConfig(config)

	return func(c *zinc.Context) error {
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		if cfg.XSSProtection != "" {
			c.SetHeader(zinc.HeaderXXSSProtection, cfg.XSSProtection)
		}
		if cfg.ContentTypeNosniff != "" {
			c.SetHeader(zinc.HeaderXContentTypeOptions, cfg.ContentTypeNosniff)
		}
		if cfg.XFrameOptions != "" {
			c.SetHeader(zinc.HeaderXFrameOptions, cfg.XFrameOptions)
		}
		if cfg.ReferrerPolicy != "" {
			c.SetHeader(zinc.HeaderReferrerPolicy, cfg.ReferrerPolicy)
		}
		if cfg.CrossOriginResourcePolicy != "" {
			c.SetHeader(zinc.HeaderCrossOriginResourcePolicy, cfg.CrossOriginResourcePolicy)
		}
		if cfg.PermissionsPolicy != "" {
			c.SetHeader("Permissions-Policy", cfg.PermissionsPolicy)
		}
		if cfg.ContentSecurityPolicy != "" {
			c.SetHeader(zinc.HeaderContentSecurityPolicy, cfg.ContentSecurityPolicy)
		}
		if cfg.ContentSecurityPolicyReportOnly != "" {
			c.SetHeader(zinc.HeaderContentSecurityPolicyReportOnly, cfg.ContentSecurityPolicyReportOnly)
		}
		if cfg.HSTSMaxAge > 0 && c.Secure() {
			c.SetHeader(zinc.HeaderStrictTransportSecurity, hstsHeaderValue(cfg))
		}

		return c.Next()
	}
}

func resolveSecureConfig(config SecureConfig) SecureConfig {
	cfg := DefaultSecureConfig()
	cfg.Skipper = config.Skipper
	if config.XSSProtection != "" {
		cfg.XSSProtection = config.XSSProtection
	}
	if config.ContentTypeNosniff != "" {
		cfg.ContentTypeNosniff = config.ContentTypeNosniff
	}
	if config.XFrameOptions != "" {
		cfg.XFrameOptions = config.XFrameOptions
	}
	if config.HSTSMaxAge > 0 {
		cfg.HSTSMaxAge = config.HSTSMaxAge
	}
	cfg.HSTSExcludeSubdomains = config.HSTSExcludeSubdomains
	if config.ContentSecurityPolicy != "" {
		cfg.ContentSecurityPolicy = config.ContentSecurityPolicy
	}
	if config.ContentSecurityPolicyReportOnly != "" {
		cfg.ContentSecurityPolicyReportOnly = config.ContentSecurityPolicyReportOnly
	}
	if config.ReferrerPolicy != "" {
		cfg.ReferrerPolicy = config.ReferrerPolicy
	}
	if config.CrossOriginResourcePolicy != "" {
		cfg.CrossOriginResourcePolicy = config.CrossOriginResourcePolicy
	}
	if config.PermissionsPolicy != "" {
		cfg.PermissionsPolicy = config.PermissionsPolicy
	}
	return cfg
}

func hstsHeaderValue(cfg SecureConfig) string {
	value := fmt.Sprintf("max-age=%d", cfg.HSTSMaxAge)
	if !cfg.HSTSExcludeSubdomains {
		value += "; includeSubDomains"
	}
	return strings.TrimSpace(value)
}
