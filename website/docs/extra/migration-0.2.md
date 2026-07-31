---
id: migration-0.2
title: Migrating to Zinc 0.2
description: Update route patterns, registration, native HTTP integration, and background work for Zinc 0.2.
sidebar_position: 1
---

Zinc 0.2 is an intentional pre-1.0 cleanup. It removes compatibility aliases instead of carrying two ways to express the same route.

## Migration table

| Zinc 0.1 | Zinc 0.2 |
|---|---|
| `/users/:id` | `/users/{id}` |
| `/files/*path` | `/files/{path...}` |
| `/users/:id<\d+>` | `/users/{id}` plus validation in the handler or binder |
| `app.Get("/health", "ok")` | A typed handler returning `c.String("ok")` |
| `if err := app.Get(...); err != nil` | `app.Get(...)`; invalid source declarations panic during startup |
| `if err := app.Handle(spec); err != nil` | `app.Handle(spec)` for source declarations or `app.TryHandle(spec)` for runtime input |
| `c.Copy()` | Extract exact values or use the standard request context deliberately |
| `zinc.ParamIdentifier`, `zinc.WildcardIdentifier` | Removed; public route patterns use brace syntax directly |

## Route patterns

Replace colon parameters and star wildcards mechanically:

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
	return c.String(c.Param("id"))
})

app.Get("/files/{path...}", func(c *zinc.Context) error {
	return c.String(c.Param("path"))
})
```

Zinc also sets standard-library path values, so native handlers can read the same match:

```go
app.HandleHTTP("GET /users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.PathValue("id"))
}))
```

Regex-constrained route patterns are gone. Validate IDs and other formats in the handler, binder, or application layer.

## Typed handlers

All route methods accept `zinc.HandlerFunc`. Literal response shorthand is removed:

```go
app.Get("/health", func(c *zinc.Context) error {
	return c.String("ok")
})
```

Handler errors still flow through Zinc's central error handler. Only route-registration errors changed.

## Registration errors

Routes written in source fail immediately when their declaration is invalid or conflicts with another route:

```go
app.Get("/users/{id}", showUser)

app.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/named-users/{id}",
	Handler: showUser,
})
```

Use `TryHandle` when a route comes from configuration, a plugin, or another runtime source:

```go
if err := app.TryHandle(zinc.RouteSpec{
	Method:  methodFromConfig,
	Path:    patternFromConfig,
	Handler: handler,
}); err != nil {
	return fmt.Errorf("register configured route: %w", err)
}
```

## Standard HTTP integration

Zinc 0.2 can register a standard handler directly and wrap the whole application in standard middleware:

```go
app.HandleHTTP("GET /metrics", promhttp.Handler())
app.UseHTTP(requestTracing, authenticateRequest)
```

`App` continues to implement `http.Handler`, and Zinc handlers retain direct access to `c.Request()` and `c.Writer()`.

## Background work

`Context.Copy` is removed. A Zinc context is pooled and valid only during its handler.

Extract exact values before dispatching work:

```go
app.Post("/jobs/{id}", func(c *zinc.Context) error {
	jobs <- Job{ID: c.Param("id"), RequestID: c.RequestID()}
	return c.NoContent()
})
```

Use `c.Request().Context()` when work should share request cancellation. Use `context.WithoutCancel` only when deliberately detaching standard context values from that cancellation.

## Unchanged behavior

- Handlers remain `func(*zinc.Context) error`.
- `Get`, `Post`, `c.JSON`, `c.String`, `Map`, and the context store remain.
- `CaseSensitive` and `StrictRouting` remain configurable and default to `false`.
- Automatic `HEAD`, automatic `OPTIONS`, and `405 Method Not Allowed` behavior remain enabled.
- `app.Listen` remains the shortest server startup path; `App` still works with a user-owned `http.Server`.

## Upgrade checklist

1. Replace legacy route patterns with brace patterns.
2. Replace literal string routes with typed handlers.
3. Remove route-registration error checks; use `TryHandle` only for dynamic declarations.
4. Replace `Context.Copy` with explicit value extraction.
5. Run `go test ./...` and verify intentional case and trailing-slash configuration.
