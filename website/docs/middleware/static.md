---
id: static
title: Static
description: Serve static files from middleware.
sidebar_position: 24
---

Zinc already has app-level static helpers. `Static` is the middleware-shaped version for stacks that prefer `app.Use(...)`.

```go
app.Use(middleware.Static("./public"))
```

Serve from a prefix:

```go
app.Use(middleware.StaticFrom("/assets", "./public"))
```

Serve from an `fs.FS`:

```go
app.Use(middleware.StaticFS(embeddedFiles))
```

By default, `Static` falls through to the next handler when a file is not found.

```go
app.Use(middleware.StaticWithConfig(middleware.StaticConfig{
	Root:           "./public",
	Prefix:         "/assets",
	NextOnNotFound: true,
}))
```
