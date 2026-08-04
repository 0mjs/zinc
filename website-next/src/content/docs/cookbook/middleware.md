---
title: Custom Middleware
description: Write reusable Zinc middleware with before and after behavior.
---

Middleware has the same signature as a route handler. Call `c.Next()` to continue the chain.

```go
func timing(c *zinc.Context) error {
    started := time.Now()
    err := c.Next()

    status := 0
    if writer, ok := c.Writer().(zinc.ResponseWriter); ok {
        status = writer.Status()
    }

    log.Printf(
        "%s %s status=%d duration=%s",
        c.Method(),
        c.Path(),
        status,
        time.Since(started),
    )
    return err
}
```

Register it globally:

```go
app.Use(timing)
```

Or scope it to a group or route:

```go
admin := app.Group("/admin", requireAdmin)
admin.Get("/stats", stats)

app.Get("/expensive", rateLimit, expensiveHandler)
```

Standard `net/http` middleware works through `app.UseHTTP` without an adapter.
