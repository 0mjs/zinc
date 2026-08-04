---
title: Subdomain
description: Restrict a Zinc route or group by the incoming request host.
---

Host matching is ordinary middleware because Zinc keeps the standard request available.

```go
func requireHost(host string) zinc.Middleware {
    return func(c *zinc.Context) error {
        requestHost := c.Request().Host
        if name, _, err := net.SplitHostPort(requestHost); err == nil {
            requestHost = name
        }
        if !strings.EqualFold(requestHost, host) {
            return zinc.NewError(zinc.StatusNotFound)
        }
        return c.Next()
    }
}

app.Get("/", requireHost("api.example.com"), func(c *zinc.Context) error {
    return c.String("API subdomain")
})
```

Behind a reverse proxy, make sure the proxy forwards the original host and that untrusted clients cannot bypass the proxy.
