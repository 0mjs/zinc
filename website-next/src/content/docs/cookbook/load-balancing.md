---
title: Load Balancing
description: Distribute proxied Zinc requests across multiple upstream services.
---

```go
apiA, err := url.Parse("http://10.0.0.11:9000")
if err != nil {
    log.Fatal(err)
}
apiB, err := url.Parse("http://10.0.0.12:9000")
if err != nil {
    log.Fatal(err)
}

targets := []*middleware.ProxyTarget{
    {Name: "api-a", URL: apiA},
    {Name: "api-b", URL: apiB},
}

app.UsePrefix("/api", middleware.ProxyWithConfig(middleware.ProxyConfig{
    Balancer: middleware.NewRoundRobinBalancer(targets),
    Retries:  1,
    Rewrite: map[string]string{
        "/api/*": "/*",
    },
}))
```

`NewRoundRobinBalancer` rotates deterministically. `NewRandomBalancer` selects a target randomly.

Application-level balancing does not replace health checking. Remove unhealthy targets through deployment configuration or provide a custom `ProxyBalancer` that understands service health.
