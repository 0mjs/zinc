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

Use it for auth pages, dashboards, and any route where cached browser history is not wanted.

## Heartbeat

```go
app.Use(middleware.Heartbeat("/healthz"))
```

Requests to the heartbeat path return `204 No Content` before route handlers run.

Use this for load balancers and uptime checks when the check does not need application logic.

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

Without trusted proxy configuration, do not trust arbitrary forwarded IP headers from the public internet.

## Throttle

```go
app.Use(middleware.Throttle(64))
```

`Throttle` limits concurrent in-flight requests. It is different from `RateLimiter`, which limits request rate over time.

Use `Throttle` when the protected resource is concurrency-sensitive: database pools, expensive exports, slow upstream calls, and similar work.

## Maybe

```go
app.Use(middleware.Maybe(func(c *zinc.Context) bool {
	return c.GetHeader("X-Debug") == "1"
}, debugMiddleware))
```

`Maybe` applies a middleware only when the predicate returns true.

That keeps conditional middleware out of the handler.

## SetHeader

```go
app.Use(middleware.SetHeader("X-App", "zinc"))
```

`SetHeader` writes a response header and continues the chain.

Use it for small static headers. Use `Secure` for common browser security headers.
