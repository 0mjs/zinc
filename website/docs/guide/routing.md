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
app.Get("/users/:id", showUser)
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
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
```

For custom methods or more dynamic registration:

```go
app.Add("PURGE", "/cache/:key", purgeCache)
app.Match([]string{zinc.MethodGet, zinc.MethodHead}, "/health", healthHandler)
```

## Route params

Zinc supports named params and wildcard captures.

```go
app.Get("/users/:id", func(c *zinc.Context) error {
	return c.String(c.Param("id"))
})

app.Get("/assets/*tail", func(c *zinc.Context) error {
	return c.String(c.Param("tail"))
})
```

### Regex-constrained params

Use `:name<expr>` when a route should only match values that satisfy a regular expression.

```go
app.Get("/users/:id<\\d+>", showUser)
app.Get("/posts/:slug<[a-z0-9-]+>", showPost)
```

If the constraint does not match, Zinc treats the request as not found for that route.

:::info
Constrained params are useful for disambiguating routes without moving the logic into handlers.
:::

## Named routes and reverse URLs

For application tooling, URL generation, and link building, use `RouteSpec`.

```go
if err := app.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/users/:id<\\d+>",
	Handler: showUser,
}); err != nil {
	log.Fatal(err)
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

v1.Get("/users/:id", showUser)
v1.Post("/users", createUser)
```

The same API exists on nested groups, including `Handle`, `Static`, `Mount`, and `RouteNotFound`.

## Route blocks

Use `Route` when you want a clear nested declaration block.

```go
app.Route("/api", func(api *zinc.Group) {
	api.Route("/v1", func(v1 *zinc.Group) {
		v1.Get("/users/:id", showUser)
		v1.Post("/users", createUser)
	})
}, requireAPIKey)
```

## Route-scoped 404 handling

Zinc supports app-wide and prefix-scoped not-found flows.

```go
app.RouteNotFound("/api/*tail", func(c *zinc.Context) error {
	return c.Status(zinc.StatusNotFound).JSON(zinc.Map{
		"error": "unknown api route",
	})
})
```

This is useful for APIs that should return structured JSON in one subtree while leaving the rest of the app with different not-found behavior.

## Mounting stdlib handlers

Mount standard `net/http` handlers without leaving Zinc.

```go
app.Mount("/metrics", promhttp.Handler())
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
