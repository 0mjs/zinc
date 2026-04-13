---
id: utility
title: Utility
description: Small middleware helpers for caching, health checks, real IPs, throttling, and conditional middleware.
sidebar_position: 26
---

These helpers cover small request and response behaviors that show up often.

## NoCache

```go
app.Use(middleware.NoCache())
```

`NoCache` sets response headers that tell browsers and intermediaries not to cache the response.

## Heartbeat

```go
app.Use(middleware.Heartbeat("/healthz"))
```

Requests to the heartbeat path return `204 No Content` before route handlers run.

## RealIP

```go
app.Use(middleware.RealIP())
```

`RealIP` updates the request remote address from `c.IP()`.

Configure `TrustedProxies` before using forwarded IP headers.

```go
app := zinc.NewWithConfig(zinc.Config{
	ProxyHeader:    zinc.HeaderXForwardedFor,
	TrustedProxies: []string{"10.0.0.1"},
})
```

## Throttle

```go
app.Use(middleware.Throttle(64))
```

`Throttle` limits concurrent in-flight requests. It is different from `RateLimiter`, which limits request rate over time.

## Maybe

```go
app.Use(middleware.Maybe(func(c *zinc.Context) bool {
	return c.GetHeader("X-Debug") == "1"
}, debugMiddleware))
```

`Maybe` applies a middleware only when the predicate returns true.

## SetHeader

```go
app.Use(middleware.SetHeader("X-App", "zinc"))
```

`SetHeader` writes a response header and continues the chain.
