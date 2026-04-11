---
id: method-override
title: Method Override
description: Override POST methods from headers or custom getters.
sidebar_position: 16
---

`MethodOverride` lets clients tunnel `PUT`, `PATCH`, or `DELETE` through `POST`.

```go
app.Use(middleware.MethodOverride())
```

The default source is:

```text
X-HTTP-Method-Override
```

The original method is preserved in:

```text
X-Original-Method
```

Use query-based or custom lookup when needed.

```go
app.Use(middleware.MethodOverrideWithConfig(middleware.MethodOverrideConfig{
	Getter: middleware.MethodOverrideFromFirst(
		middleware.MethodOverrideFromHeader("X-HTTP-Method-Override"),
		middleware.MethodOverrideFromQuery("_method"),
	),
}))
```
