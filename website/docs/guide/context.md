---
id: context
title: Context
description: Work with request data, response state, route metadata, and request-scoped values.
sidebar_position: 3
---

`*zinc.Context` is Zinc’s request-scoped object.

It gives you access to:

- the request and response writer
- route params and query values
- request-scoped values
- binding and response helpers
- route metadata and client/network helpers

## At a glance

```go
app.Get("/users/:id", func(c *zinc.Context) error {
	return c.JSON(zinc.Map{
		"id":      c.Param("id"),
		"verbose": c.QueryOr("verbose", "false"),
		"route":   c.FullPath(),
	})
})
```

Treat the context as short-lived. Read what you need during the request, and use `c.Copy()` before sending request data to another goroutine.

## Basic request data

```go
app.Get("/users/:id", func(c *zinc.Context) error {
	method := c.Method()
	path := c.Path()
	id := c.Param("id")
	search := c.Query("search")

	_ = method
	_ = path
	_ = id
	_ = search

	return c.NoContent()
})
```

Useful helpers include:

- `Method()`
- `Path()`
- `Request()`
- `Writer()`
- `Param(name)`
- `ParamOr(name, fallback)`
- `Query(name)`
- `QueryOr(name, fallback)`
- `QueryArray(name)`
- `QueryMap(name)`
- `QueryValues()`
- `PostForm(name)`
- `PostFormOr(name, fallback)`
- `PostFormArray(name)`
- `PostFormMap(name)`
- `ContentType()`
- `IsWebSocket()`

For repeated query values:

```go
app.Get("/search", func(c *zinc.Context) error {
	tags := c.QueryArray("tag")
	return c.JSON(zinc.Map{"tags": tags})
})
```

For form posts:

```go
app.Post("/profile", func(c *zinc.Context) error {
	name := c.PostForm("name")
	roles := c.PostFormArray("roles")
	return c.JSON(zinc.Map{"name": name, "roles": roles})
})
```

For bracket-style query or form maps:

```go
app.Get("/search", func(c *zinc.Context) error {
	filters := c.QueryMap("filter")
	return c.JSON(filters)
})
```

For request shape checks:

```go
app.Post("/events", func(c *zinc.Context) error {
	if c.ContentType() != "application/json" {
		return zinc.ErrUnsupportedMediaType
	}
	return c.NoContent()
})
```

## Request-scoped values

Use `Set` and `Get` for request-local state.

```go
func loadUser(c *zinc.Context) error {
	c.Set("userID", "42")
	return c.Next()
}

func handler(c *zinc.Context) error {
	if userID := c.GetString("userID"); userID != "" {
		return c.String(userID)
	}
	return zinc.ErrUnauthorized
}
```

Use `Get` when you want the raw value and `MustGet` when missing state should be treated as a programmer error.

## Files and multipart form data

Zinc exposes direct multipart helpers on the context:

- `FormFile(name)`
- `FormFiles(name)`
- `MultipartForm()`
- `SaveFile(file, dst)`

These are useful when you want lower-level control instead of binding multipart fields into a struct.

## Route metadata

Use `Route()` when you need the matched route’s metadata inside a handler.

```go
app.Get("/users/:id", func(c *zinc.Context) error {
	route := c.Route()
	return c.JSON(route)
})
```

`RouteInfo` includes:

- `Name`
- `Method`
- `Path`
- `Params`
- `Mounted`
- `Handler`

## IP and request identity helpers

Zinc exposes client IP helpers and request identity access:

- `IP()`
- `IPs()`
- `RequestID()`

`IP()` respects the configured proxy header and trusted proxy settings.

## Copying context safely

Treat `*zinc.Context` as request-scoped and short-lived.

If you need to hand request state to a goroutine or persist it outside the handler, copy it first:

```go
app.Get("/jobs/:id", func(c *zinc.Context) error {
	cc := c.Copy()
	go func() {
		_ = cc.Route()
	}()
	return c.NoContent()
})
```

:::caution
Do not keep the original request context beyond the active handler lifetime. Use `Copy()` or extract the exact data you need.
:::

## Advanced: manual acquire and release

Most applications do not need manual context management. Use `AcquireContext` and `ReleaseContext` only when integrating Zinc with code that already owns an `http.ResponseWriter` and `*http.Request`.

## See also

- [Binding](./binding) for decoding requests into structs.
- [Responses and Rendering](./responses-and-rendering) for writing responses.
- [Context API](../api/context) for the full helper list.

Most applications should let Zinc manage context lifecycle automatically.

Advanced integrations can use:

- `app.AcquireContext(w, r)`
- `app.ReleaseContext(c)`

These are primarily useful for framework internals, adapters, and low-level testing.
