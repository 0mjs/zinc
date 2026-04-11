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

The target must be a valid absolute URL.
