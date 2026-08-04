---
title: Reverse Proxy
description: Forward a Zinc path prefix to an HTTP upstream.
---

```go
app := zinc.New()

app.UsePrefix(
    "/api",
    middleware.ProxyWithConfig(middleware.ProxyConfig{
        Target: "http://127.0.0.1:9000",
        Rewrite: map[string]string{
            "/api/*": "/*",
        },
    }),
)

log.Fatal(app.Listen(":8080"))
```

A request to `/api/users` is forwarded to `http://127.0.0.1:9000/users`.

Configure a custom `http.Transport` for connection pooling, TLS policy, proxy settings, and upstream timeouts. See [Proxy middleware](/middleware/proxy/) for response modification, retries, exact rewrites, and regular-expression rewrites.
