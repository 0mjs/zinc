---
id: casbin-auth
title: Casbin Auth
description: Authorize requests with a Casbin-compatible enforcer.
sidebar_position: 15
---

`CasbinAuth` adapts any value with this method:

```go
Enforce(args ...any) (bool, error)
```

That keeps Zinc free of a hard Casbin dependency while still working with Casbin enforcers.

```go
app.Use(middleware.CasbinAuth(enforcer, middleware.CasbinSubjectFromBasicAuth()))
```

By default Zinc enforces:

```go
enforcer.Enforce(subject, c.Path(), c.Method())
```

Use `CasbinAuthWithConfig` to choose a different subject, object, or action.

```go
app.Use(middleware.CasbinAuthWithConfig(middleware.CasbinAuthConfig{
	Enforcer: enforcer,
	Subject:  middleware.CasbinSubjectFromContext(userKey),
	Object:   middleware.CasbinObjectPath(),
	Action:   middleware.CasbinActionMethod(),
}))
```
