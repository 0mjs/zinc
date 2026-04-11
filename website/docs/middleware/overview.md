---
id: overview
title: 🛡️ Middleware Overview
description: Zinc’s first-party middleware package and the common use cases it covers.
sidebar_position: 1
---

Zinc ships first-party middleware under:

```go
github.com/0mjs/zinc/middleware
```

## What ships today

| Middleware | Purpose |
|---|---|
| [`CORS`](./cors) | Cross-origin request policy and preflight handling |
| [`CSRF`](./csrf) | Cookie-backed CSRF protection for browser-facing apps |
| [`JWT`](./jwt) | Bearer token parsing, validation, and claims access |
| [`BasicAuth`](./basic-auth) | Basic auth extraction and validation helpers |
| [`RequestLogger`](./request-logger) | Structured request logging with status, timing, and selected fields |
| [`RequestID`](./request-id) | Request ID generation and response header publishing |
| [`Recover`](./recover) | Panic recovery through Zinc's normal error flow |
| [`Gzip`](./gzip) | Gzip response compression for clients that opt in with `Accept-Encoding` |
| [`Decompress`](./decompress) | Gzip request body decompression from `Content-Encoding` |
| [`KeyAuth`](./key-auth) | API key extraction and validation from headers, query values, or cookies |
| [`CasbinAuth`](./casbin-auth) | Casbin-compatible authorization through a small `Enforce(...any)` adapter |
| [`MethodOverride`](./method-override) | POST method override from headers or custom getters |
| [`Prometheus`](./prometheus) | Dependency-free Prometheus text metrics for request counts and duration |
| [`Jaeger`](./jaeger) | Jaeger `uber-trace-id` propagation with an observer hook |
| [`Proxy`](./proxy) | `net/http/httputil` reverse proxy middleware |
| [`Redirect`](./redirect) | Exact and wildcard path redirects |
| [`Rewrite`](./rewrite) | Exact and wildcard path rewrites before route dispatch |
| [`Secure`](./secure) | Common security response headers |
| [`SessionCookie`](./session) | Signed cookie-backed string session values |
| [`Static`](./static) | Static file serving as middleware |
| [`TrailingSlash`](./trailing-slash) | Add, remove, or redirect trailing slash variants |
| [`BodyLimit`](./body-limit) | Request-size enforcement before handlers consume the body |
| [`BodyDump`](./body-dump) | Request/response body capture with truncation and redaction hooks |
| [`ContextTimeout`](./context-timeout) | Per-request context deadlines for handler chains |
| [`RateLimiter`](./rate-limiter) | Global, per-IP, or custom-key token-bucket rate limiting |

## Typical stack

```go
app.Use(middleware.RequestLogger())
app.Use(middleware.RequestID())
app.Use(middleware.Recover())
app.Use(middleware.CORS("https://app.example.com"))
app.Use(middleware.Decompress())
app.Use(middleware.Gzip())
app.Use(middleware.MethodOverride())
app.Use(middleware.Secure())
app.Use(middleware.TrailingSlash())
app.Use(middleware.BodyLimit(10 * middleware.MB))
app.Use(middleware.Prometheus())

app.Get("/metrics", middleware.PrometheusHandler())

admin := app.Group("/admin")
admin.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
	Validator: middleware.BasicAuthStatic("admin", "secret"),
}))

api := app.Group("/api")
api.Use(middleware.JWT(func(_ *zinc.Context, token *jwt.Token) (any, error) {
	return signingKey, nil
}))
```

## How to read this section

- Start with the page for the middleware you want to install.
- Use the constructor examples first.
- Drop into the config tables only when you need custom behavior.
- Reach for `pkg.go.dev` when you want every exported helper in one place.

## Design goal

The middleware package is intended to cover the common API and security cases while keeping Zinc’s core small and focused. The docs here stay practical and operational; the source package docs remain the best low-level reference:

- [pkg.go.dev middleware package](https://pkg.go.dev/github.com/0mjs/zinc/middleware)
