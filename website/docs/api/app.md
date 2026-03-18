---
id: app
title: App
description: The primary Zinc application type and its lifecycle, routing, and introspection APIs.
sidebar_position: 1
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
| `Add`, `Match`, `All`, `Any` | Register routes more generically |
| `Use`, `UsePrefix` | Register middleware |
| `Group`, `Route` | Create grouped route trees |
| `Mount`, `Wrap`, `WrapFunc` | Integrate stdlib or external handlers |
| `Static`, `StaticFS`, `File`, `FileFS` | Serve files and directories |
| `NotFound`, `RouteNotFound`, `MethodNotAllowed` | Customize error routing |
| `Routes`, `FindRoute`, `RouteByName`, `URL` | Inspect routes and generate URLs |
| `Listen`, `ListenTLS`, `Serve`, `Shutdown` | Run and stop the app |

## Named route registration

Use `Handle(RouteSpec)` when you want route naming and reverse URL generation.

```go
if err := app.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/users/:id",
	Handler: showUser,
}); err != nil {
	log.Fatal(err)
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

Most apps do not need these directly, but Zinc exposes:

- `AcquireContext`
- `ReleaseContext`

They are useful for adapters, framework integration, and low-level tests.
