---
title: Timeout
description: Apply request deadlines and stop downstream work when the client context expires.
---

```go
app.Use(middleware.ContextTimeout(3 * time.Second))

app.Get("/report", func(c *zinc.Context) error {
    report, err := reports.Generate(c.Context())
    if err != nil {
        return err
    }
    return c.JSON(report)
})
```

Database, HTTP, and other downstream calls must receive `c.Context()` for cancellation to propagate.

Use `ContextTimeoutWithConfig` to skip selected routes or replace the timeout error response:

```go
app.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
    Timeout: 3 * time.Second,
    Skipper: func(c *zinc.Context) bool {
        return c.Path() == "/events"
    },
}))
```

Do not apply a short request timeout to WebSockets or intentionally long-lived streams.
