---
id: overview
title: Middleware Overview
description: Zinc’s first-party middleware package and the common use cases it covers.
sidebar_position: 1
---

Zinc ships first-party middleware under:

```go
github.com/0mjs/zinc/middleware
```

## Available middleware

- `RequestLogger`
- `CORS`
- `CSRF`
- `JWT`
- `BasicAuth`
- `RateLimiter`
- `BodyLimit`
- `BodyDump`
- `ContextTimeout`

## Example

```go
app.Use(middleware.RequestLogger())
app.Use(middleware.CORS("https://app.example.com"))

admin := app.Group("/admin")
admin.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
	Validator: middleware.BasicAuthStatic("admin", "secret"),
}))
```

## Design goal

The middleware package is intended to cover the common API and security cases while keeping Zinc’s core small and focused.

For full API details, the source package docs remain the best low-level reference:

- [pkg.go.dev middleware package](https://pkg.go.dev/github.com/0mjs/zinc/middleware)
