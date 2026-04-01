---
id: group
title: 👥 Group
description: Group routes by prefix, middleware chain, and route subtree.
sidebar_position: 2
---

`Group` is Zinc’s route subtree type.

Create one with:

```go
api := app.Group("/api", requireAPIKey)
```

## What groups do

Groups let you:

- apply a common path prefix
- apply a shared middleware chain
- build nested APIs without repeating paths

## Common methods

| Method | Purpose |
|---|---|
| `Use` | Add group-local middleware |
| `Group`, `Route` | Create nested route trees |
| `Get`, `Post`, `Put`, `Patch`, `Delete`, `Head`, `Options` | Register routes under the group prefix |
| `Add`, `Match`, `All`, `Any` | Register more general route sets |
| `Handle` | Register a named route with `RouteSpec` |
| `RouteNotFound` | Register a prefix-scoped not-found route |
| `Mount`, `Static`, `StaticFS`, `File`, `FileFS` | Mount handlers or serve files below the group prefix |

## Example

```go
api := app.Group("/api", requireAPIKey)
v1 := api.Group("/v1")

v1.Get("/users/:id", showUser)
v1.Post("/users", createUser)

if err := v1.Handle(zinc.RouteSpec{
	Name:    "users.show",
	Method:  zinc.MethodGet,
	Path:    "/users/:id",
	Handler: showUser,
}); err != nil {
	log.Fatal(err)
}
```

Group route registration inherits the group prefix and middleware stack automatically.
