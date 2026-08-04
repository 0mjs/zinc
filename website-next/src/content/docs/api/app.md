---
title: App
description: The primary Zinc application type and its lifecycle, routing, and introspection APIs.
---

`App` is Zinc’s main application type.

## Create an app

```go
app := zinc.New()
app := zinc.NewWithConfig(zinc.Config{StrictRouting: true})
```

## Core responsibilities

`App` is responsible for:

- route registration
- middleware registration
- mounting sub-handlers
- lifecycle methods such as `Listen`, `Serve`, and `Shutdown`
- route introspection and named route lookup

## Common methods

| Method | Purpose |
|---|---|
| `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options`, `Connect`, `Trace` | Register routes |
| `Add`, `Match`, `All`, `Any` | Register source-defined routes more generically |
| `Handle`, `TryHandle` | Register named source or runtime-defined routes |
| `Use`, `UsePrefix`, `UseHTTP` | Register Zinc or standard `net/http` middleware |
| `Group`, `Route` | Create grouped route trees |
| `HandleHTTP`, `Mount`, `Wrap`, `WrapFunc` | Integrate stdlib or external handlers |
| `Static`, `StaticFS`, `File`, `FileFS` | Serve files and directories |
| `NotFound`, `RouteNotFound`, `MethodNotAllowed` | Customize error routing |
| `Routes`, `FindRoute`, `RouteByName`, `URL` | Inspect routes and generate URLs |
| `Listen`, `ListenTLS`, `Serve`, `Shutdown` | Run and stop the app |

Route methods accept a handler chain. Middleware goes before the final handler.

```go
app.Post("/posts", authUser, requireRole("editor"), createPost)
```

Every route method accepts typed `HandlerFunc` values. Responses stay explicit in the handler:

```go
app.Get("/health", func(c *zinc.Context) error {
	return c.String("ok")
})
```

Register an ordinary `http.Handler` directly with a method and route pattern:

```go
app.HandleHTTP("GET /metrics", promhttp.Handler())
```

Wrap the complete application in standard Go middleware with `UseHTTP`:

```go
app.UseHTTP(requestTracing, authenticateRequest)
```

Standard middleware runs outside Zinc application, group, and route middleware.

## Named route registration

Use `Handle(RouteSpec)` when you want route naming and reverse URL generation. Source-defined routes fail immediately if the declaration is invalid or conflicts with an existing route.

```go
app.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/users/{id}",
	Handler: showUser,
})
```

When a route pattern comes from configuration, a plugin, or another runtime source, use `TryHandle` and handle the error explicitly:

```go
if err := app.TryHandle(zinc.RouteSpec{
	Method:  zinc.MethodGet,
	Path:    patternFromConfig,
	Handler: showUser,
}); err != nil {
	return err
}
```

## Introspection

```go
routes := app.Routes()
route, ok := app.FindRoute("GET", "/users/42")
named, ok := app.RouteByName("users.show")
url, err := app.URL("users.show", "42")
```

`Routes()` returns `RouteInfo` values containing:

- `Name`
- `Method`
- `Path`
- `Params`
- `Mounted`
- `Handler`

## Advanced lifecycle helpers

`Listen` is the shortest way to start an app. With no address, it listens on `:8080`.

```go
app.Listen()
app.Listen(":3000")
```

Most apps do not need these directly, but Zinc exposes:

- `AcquireContext`
- `ReleaseContext`

They are useful for adapters, framework integration, and low-level tests.

For normal applications, register routes and let Zinc manage context lifecycle.
