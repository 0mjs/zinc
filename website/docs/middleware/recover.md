---
id: recover
title: Recover
description: Recover panics and route them through Zinc's error flow.
sidebar_position: 11
---

`Recover` catches panics from downstream middleware and handlers.

```go
app.Use(middleware.Recover())
```

By default, recovered panics become `500 Internal Server Error` responses through Zinc's error handling flow.

Use `RecoverWithConfig` to customize the response or log the stack.

```go
app.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
	Handler: func(c *zinc.Context, err *middleware.RecoverError) error {
		log.Printf("panic: %v\n%s", err.Value, err.Stack)
		return c.Status(zinc.StatusInternalServerError).JSON(zinc.Map{
			"error": "internal server error",
		})
	},
}))
```

Set `DisableStack` when stack capture is not needed.
