---
title: Groups and Middleware
description: Compose app-wide, prefix-scoped, and group-scoped behavior cleanly.
---

Zinc keeps middleware composition simple:

- `app.UseHTTP(...)` wraps the application in standard `net/http` middleware
- `app.Use(...)` applies globally
- `app.UsePrefix(...)` applies to matching prefixes
- `group.Use(...)` applies inside a specific group tree
- route registration accepts middleware before the final handler

Middleware uses the same handler signature as routes:

```go
type Middleware = func(*zinc.Context) error
```

Every item in the chain runs in order. Call `c.Next()` to continue.

## At a glance

```go
app.Use(middleware.RequestID())

api := app.Group("/api", requireAPIKey)
api.Get("/users/{id}", loadUser, showUser)
```

Middleware belongs at the narrowest level that still matches the behavior: app-wide for every request, group-level for route families, and route-level for endpoint-specific guards.

Use `UseHTTP` when middleware already follows the standard
`func(http.Handler) http.Handler` contract. It runs outside every Zinc
middleware layer. See [net/http Interoperability](/guide/http-interoperability/).

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

## Route middleware

Pass middleware directly into a route when the behavior belongs to one endpoint.

```go
app.Post("/posts", authUser, requireRole("editor", "admin"), createPost)
```

That reads as:

1. authenticate the request
2. check the allowed roles
3. run the handler

The guard is just middleware.

```go
func requireRole(allowed ...string) zinc.Middleware {
	return func(c *zinc.Context) error {
		user, ok := c.Get("user")
		if !ok || user == nil {
			return c.AbortWithStatus(zinc.StatusUnauthorized)
		}
		if !hasAnyRole(user, allowed...) {
			return c.AbortWithStatus(zinc.StatusForbidden)
		}
		return c.Next()
	}
}
```

The role lookup can live in your auth package. The middleware shape stays the same.

Use groups when many routes share a guard.

```go
admin := app.Group("/admin", authUser, requireRole("admin"))
admin.Get("/users", listUsers)
admin.Delete("/users/{id}", deleteUser)
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

Zinc ships common middleware in one package:

```go
import "github.com/0mjs/zinc/middleware"
```

Start with [Request ID](/middleware/request-id/), [Request Logger](/middleware/request-logger/), [Recover](/middleware/recover/), [CORS](/middleware/cors/), and [Secure](/middleware/secure/) for a typical API stack.

## See also

- [Middleware Overview](/middleware/overview/) for the full first-party middleware map.
- [Routing](/guide/routing/) for groups and route registration.
- [Errors](/guide/errors/) for returned middleware errors.
