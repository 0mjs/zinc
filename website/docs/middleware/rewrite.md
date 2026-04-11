---
id: rewrite
title: Rewrite
description: Rewrite exact or wildcard paths before route dispatch.
sidebar_position: 21
---

`Rewrite` changes the request path before Zinc dispatches the route.

```go
app.Use(middleware.Rewrite("/old", "/new"))
```

The client does not receive a redirect. Handlers see the rewritten path.

Wildcard rewrites use `*` in the source and destination.

```go
app.Use(middleware.RewriteWithRules(map[string]string{
	"/v1/*": "/api/v1/*",
}))
```

Use `RewriteWithConfig` when a `Skipper` is needed.
