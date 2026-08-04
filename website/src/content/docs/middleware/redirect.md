---
title: Redirect
description: Redirect exact or wildcard paths before route dispatch.
---

`Redirect` maps a path to another path.

```go
app.Use(middleware.Redirect("/old", "/new"))
```

Pass a status code when the default `301 Moved Permanently` is not right.

```go
app.Use(middleware.Redirect("/login", "/signin", zinc.StatusTemporaryRedirect))
```

Wildcard redirects are supported through `RedirectWithRules`.

```go
app.Use(middleware.RedirectWithRules(map[string]string{
	"/v1/*": "/api/v1/*",
}))
```

Query strings are preserved.
