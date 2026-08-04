---
title: IP Address
description: Read remote and forwarded client addresses safely behind trusted proxies.
---

`c.RemoteIP()` reads the peer address from `Request.RemoteAddr`. It does not trust forwarding headers.

```go
peer := c.RemoteIP()
```

`c.IP()` can resolve a forwarded client address when proxy trust is configured:

```go
cfg := zinc.DefaultConfig
cfg.ProxyHeader = "X-Forwarded-For"
cfg.TrustedProxies = []string{
    "10.0.0.0/8",
    "192.168.0.0/16",
}

app := zinc.NewWithConfig(cfg)
```

```go
app.Get("/whoami", func(c *zinc.Context) error {
    return c.JSON(zinc.Map{
        "client": c.IP(),
        "chain":  c.IPs(),
        "peer":   c.RemoteIP(),
    })
})
```

Only list proxies you operate or explicitly trust. Trusting forwarded headers from arbitrary clients allows IP spoofing and can break rate limits, auditing, and access control.

When no trusted proxy matches, Zinc falls back to the direct peer address.
