---
id: proxy
title: Proxy
description: Reverse proxy requests with net/http/httputil.
sidebar_position: 19
---

`Proxy` forwards requests through `httputil.ReverseProxy`.

```go
app.UsePrefix("/upstream", middleware.Proxy("https://api.example.com"))
```

Use `ProxyWithConfig` when the request or response needs adjustment.

```go
app.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
	Target: "https://api.example.com",
	Director: func(req *http.Request) {
		req.Header.Set("X-Forwarded-App", "zinc")
	},
	Modify: func(resp *http.Response) error {
		resp.Header.Del("Server")
		return nil
	},
}))
```

Proxy can also balance across multiple targets.

```go
apiA, _ := url.Parse("https://api-a.example.com")
apiB, _ := url.Parse("https://api-b.example.com")

app.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
	Targets: []*middleware.ProxyTarget{
		{Name: "a", URL: apiA},
		{Name: "b", URL: apiB},
	},
}))
```

`Targets` uses round-robin balancing by default. You can also pass `Balancer`, `Transport`, `Rewrite`, `RegexRewrite`, `Retries`, `RetryFilter`, and `ModifyResponse`.

The target must be a valid absolute URL.
