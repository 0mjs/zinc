---
id: trailing-slash
title: Trailing Slash
description: Add, remove, or redirect trailing slash variants.
sidebar_position: 25
---

`TrailingSlash` removes trailing slashes before route dispatch.

```go
app.Use(middleware.TrailingSlash())
```

This lets `/users/` match a route registered as `/users`.

Use `AddTrailingSlash` for the opposite behavior.

```go
app.Use(middleware.AddTrailingSlash())
```

Use redirects when the client should see the canonical URL.

```go
app.Use(middleware.TrailingSlashWithConfig(middleware.TrailingSlashConfig{
	Redirect:   true,
	StatusCode: zinc.StatusPermanentRedirect,
}))
```
