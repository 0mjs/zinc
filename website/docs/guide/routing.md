---
id: routing
title: Routing
description: Register routes, use params and wildcards, group APIs, and generate URLs from named routes.
sidebar_position: 1
toc_max_heading_level: 4
---

Routing in Zinc is intentionally straightforward:

- use `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`, `Connect`, and `Trace`
- fall back to `Add`, `Match`, `All`, or `Any` for broader patterns
- use groups to apply prefixes and middleware once

## At a glance

```go
app.Get("/users/{id}", showUser)
app.Post("/users", createUser)

api := app.Group("/api")
api.Get("/health", health)
```

Use method helpers for normal routes, params for path values, wildcards for trailing captures, and groups when routes share a prefix or middleware.

## Basic routes

```go
app.Get("/", func(c *zinc.Context) error {
	return c.String("ok")
})

app.Post("/users", createUser)
app.Put("/users/{id}", updateUser)
app.Delete("/users/{id}", deleteUser)
```

For custom methods or more dynamic registration:

```go
app.Add("PURGE", "/cache/{key}", purgeCache)
app.Match([]string{zinc.MethodGet, zinc.MethodHead}, "/health", healthHandler)
```

## Route params

Zinc supports named params and wildcard captures.

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
	return c.String(c.Param("id"))
})

app.Get("/assets/{tail...}", func(c *zinc.Context) error {
	return c.String(c.Param("tail"))
})
```

Route patterns are structural: static segments, `{name}` parameters, and a final
`{name...}` wildcard. Validate numeric IDs, slugs, and other formats in binding or
application code rather than embedding regular expressions in the route.

### Route pattern grammar

Zinc uses the path wildcard syntax introduced by Go 1.22's `net/http` router.
It deliberately supports a small path-only subset:

```text
pattern   = "/" [ segment { "/" segment } ]
segment   = literal | parameter | catch-all
parameter = "{" identifier "}"
catch-all = "{" identifier "...}"  // final segment only
```

An identifier may contain Unicode letters, digits, and underscores, but cannot
be empty or begin with a digit. Parameters must occupy a complete path segment,
and a name can appear only once in a pattern.

| Pattern | Meaning |
|---|---|
| `/users` | Static path |
| `/users/{id}` | One non-empty segment |
| `/files/{path...}` | The remaining path, including an empty value for `/files/` |

Static routes take priority over parameters, and parameters take priority over
catch-alls. Parameter names do not distinguish otherwise identical patterns, so
`/users/{id}` and `/users/{name}` conflict for the same method.

The default configuration is case-insensitive and non-strict. Literal route
segments are compared without case, while captured values keep their original
case. `/users` and `/users/` are treated as the same path unless
`StrictRouting` is enabled.

Zinc matches `Request.URL.Path`. It does not use `URL.RawPath`, so an encoded
slash decoded into `URL.Path` participates in path segmentation.

The following `net/http` pattern features are not supported:

- method or host prefixes inside ordinary route paths
- the `{$}` end marker
- regular-expression parameters
- `ServeMux` specificity and overlap resolution

Use `HandleHTTP("GET /users/{id}", handler)` when registering a native handler.
Its method and path are separated before the same Zinc path grammar is applied.

Invalid or legacy patterns fail during registration:

```text
/users/:id             use /users/{id}
/files/*path           use /files/{path...}
/users/prefix-{id}     parameters must occupy a complete segment
/files/{path...}/meta  catch-alls must be final
/users/{id}/{id}       names must be unique
```

### Migrating from Zinc 0.1

Zinc 0.2 uses one route syntax and does not keep compatibility aliases:

| Zinc 0.1 | Zinc 0.2 |
|---|---|
| `/users/:id` | `/users/{id}` |
| `/files/*path` | `/files/{path...}` |
| `/users/:id<\\d+>` | `/users/{id}` plus validation in the handler or binder |

Legacy patterns panic during source registration with a message that points to the new form.

## Named routes and reverse URLs

For application tooling, URL generation, and link building, use `RouteSpec`.

```go
app.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/users/{id}",
	Handler: showUser,
})
```

Route declarations written in source fail fast on invalid patterns and conflicts. For patterns loaded from configuration or plugins, use the error-returning form:

```go
if err := app.TryHandle(zinc.RouteSpec{
	Method:  zinc.MethodGet,
	Path:    patternFromConfig,
	Handler: showUser,
}); err != nil {
	return err
}
```

Generate URLs later with the route name:

```go
url, err := app.URL("users.show", "42")
// /users/42
```

You can also inspect route metadata:

```go
route, ok := app.RouteByName("users.show")
```

## Groups

Groups let you apply prefixes and middleware once.

```go
api := app.Group("/api", requireAPIKey)
v1 := api.Group("/v1")

v1.Get("/users/{id}", showUser)
v1.Post("/users", createUser)
```

The same API exists on nested groups, including `Handle`, `Static`, `Mount`, and `RouteNotFound`.

## Route blocks

Use `Route` when you want a clear nested declaration block.

```go
app.Route("/api", func(api *zinc.Group) {
	api.Route("/v1", func(v1 *zinc.Group) {
		v1.Get("/users/{id}", showUser)
		v1.Post("/users", createUser)
	})
}, requireAPIKey)
```

## Route-scoped 404 handling

Zinc supports app-wide and prefix-scoped not-found flows.

```go
app.RouteNotFound("/api/{tail...}", func(c *zinc.Context) error {
	return c.Status(zinc.StatusNotFound).JSON(zinc.Map{
		"error": "unknown api route",
	})
})
```

This is useful for APIs that should return structured JSON in one subtree while leaving the rest of the app with different not-found behavior.

## Standard library handlers

Register a standard handler at one endpoint with `HandleHTTP`:

```go
app.HandleHTTP("GET /metrics", promhttp.Handler())
```

The handler can read brace parameters through `r.PathValue`.

```go
app.HandleHTTP("GET /users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.PathValue("id"))
}))
```

Zinc handlers use `c.Param("id")`. `PathValue` is populated when control crosses
into a standard handler through `HandleHTTP` or `Wrap`, keeping the ordinary Zinc
route path minimal.

Use `Mount` when a standard handler owns a whole subtree.

```go
app.Mount("/debug", http.DefaultServeMux)
```

Mounted handlers also show up in route introspection with `Mounted: true`.

## Introspection helpers

Zinc keeps route metadata available for tooling and diagnostics.

```go
routes := app.Routes()
users := app.RoutesByPrefix("/users")
route, ok := app.FindRoute(zinc.MethodGet, "/users/42")
```

Use these helpers for tests, debug pages, generated route tables, and application tooling.

## See also

- [Groups and Middleware](./groups-and-middleware) for composing route trees.
- [Context](./context) for reading params and query values in handlers.
- [App API](../api/app) for method-level route registration details.
