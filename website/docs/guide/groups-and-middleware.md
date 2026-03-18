---
id: groups-and-middleware
title: Groups and Middleware
description: Compose app-wide, prefix-scoped, and group-scoped behavior cleanly.
sidebar_position: 2
---

Zinc keeps middleware composition simple:

- `app.Use(...)` applies globally
- `app.UsePrefix(...)` applies to matching prefixes
- `group.Use(...)` applies inside a specific group tree

Middleware uses the same handler signature as routes:

```go
type Middleware = func(*zinc.Context) error
```

## Global middleware

```go
app.Use(
	requestLogger,
	recoverPanic,
)
```

Global middleware wraps every request before Zinc dispatches the matched route.

## Prefix middleware

Use `UsePrefix` when behavior should apply to one part of the app but you do not want to build a full group up front.

```go
app.UsePrefix("/api", requireAPIKey)
app.UsePrefix("/admin", requireSession)
```

## Group middleware

Group middleware composes naturally with route prefixes.

```go
admin := app.Group("/admin")
admin.Use(requireSession, requireAdmin)

admin.Get("/dashboard", dashboard)
admin.Get("/users", listAdminUsers)
```

Nested groups inherit parent middleware.

```go
api := app.Group("/api", requireAPIKey)
v1 := api.Group("/v1", withVersionHeader)
```

## Returning from middleware

Middleware can:

- continue with `return c.Next()`
- short-circuit with a response
- short-circuit by returning an error

```go
func requireAPIKey(c *zinc.Context) error {
	if c.GetHeader("X-API-Key") == "" {
		return c.AbortWithStatus(zinc.StatusUnauthorized)
	}
	return c.Next()
}
```

## First-party middleware package

Zinc ships first-party middleware under:

```go
github.com/0mjs/zinc/middleware
```

Current built-in middleware includes:

- `RequestLogger`
- `CORS`
- `CSRF`
- `JWT`
- `BasicAuth`
- `RateLimiter`
- `BodyLimit`
- `BodyDump`
- `ContextTimeout`

See the [middleware overview](../middleware/overview) for a quick map.
