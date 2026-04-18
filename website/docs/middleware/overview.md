---
id: overview
title: Middleware Overview
description: Choose and compose Zinc's first-party middleware package.
sidebar_position: 1
---

Zinc middleware uses the same shape as a route handler, so it composes directly with apps, groups, and individual routes.

```go
func(*zinc.Context) error
```

Import first-party middleware from one package:

```go
import "github.com/0mjs/zinc/middleware"
```

## Start with a normal stack

Most APIs start with request identity, logging, panic recovery, CORS, body limits, and security headers.

```go
app.Use(middleware.RequestID())
app.Use(middleware.RequestLogger())
app.Use(middleware.Recover())
app.Use(middleware.CORS("https://app.example.com"))
app.Use(middleware.BodyLimit(10 * middleware.MB))
app.Use(middleware.Secure())
```

Add auth, metrics, rate limiting, compression, and static files only where the application needs them.

## Security

| Middleware | Use it for |
|---|---|
| [`CORS`](./cors) | Browser cross-origin policy and preflight responses |
| [`CSRF`](./csrf) | Cookie-backed CSRF protection for browser-facing forms and fetch requests |
| [`Secure`](./secure) | Common browser security response headers |
| [`Header Guards`](./header-guards) | Reject unsupported content types or route by request headers |

## Authentication and sessions

| Middleware | Use it for |
|---|---|
| [`JWT`](./jwt) | Bearer token parsing, validation, and typed claims |
| [`Basic Auth`](./basic-auth) | Simple admin or internal route protection |
| [`Key Auth`](./key-auth) | API keys from headers, query values, cookies, or custom extractors |
| [`Casbin Auth`](./casbin-auth) | Authorization through a Casbin-compatible enforcer |
| [`Session`](./session) | Small signed cookie-backed session values |

## Observability

| Middleware | Use it for |
|---|---|
| [`Request Logger`](./request-logger) | Structured request logs with status, timing, route, headers, and errors |
| [`Request ID`](./request-id) | Generate or propagate request IDs through `X-Request-ID` |
| [`Prometheus`](./prometheus) | Dependency-free request counters and duration metrics |
| [`Jaeger`](./jaeger) | `Uber-Trace-Id` propagation and span observation |
| [`Pprof`](./pprof) | Standard library profiling routes behind Zinc middleware |
| [`Body Dump`](./body-dump) | Request and response body snapshots for debugging or auditing |

## Traffic control

| Middleware | Use it for |
|---|---|
| [`Rate Limiter`](./rate-limiter) | Token-bucket limits by app, IP, or custom key |
| [`Body Limit`](./body-limit) | Reject oversized request bodies |
| [`Context Timeout`](./context-timeout) | Attach per-request deadlines to the handler chain |
| [`Recover`](./recover) | Convert panics into Zinc's normal error flow |
| [`Utility`](./utility) | `Throttle`, `Heartbeat`, `NoCache`, `Maybe`, `RealIP`, and `SetHeader` |

## Transport and routing helpers

| Middleware | Use it for |
|---|---|
| [`Gzip`](./gzip) | Compress responses for clients that accept gzip |
| [`Decompress`](./decompress) | Decode gzip request bodies before handlers read them |
| [`Method Override`](./method-override) | Tunnel `PUT`, `PATCH`, or `DELETE` through `POST` |
| [`Trailing Slash`](./trailing-slash) | Add, remove, or redirect trailing slash variants |
| [`Rewrite`](./rewrite) | Rewrite request paths before route dispatch |
| [`Redirect`](./redirect) | Redirect exact or wildcard paths |
| [`Proxy`](./proxy) | Reverse proxy requests with `net/http/httputil` |
| [`Static`](./static) | Serve static files from middleware |

## Where middleware belongs

Use app-level middleware for behavior that should wrap every request.

```go
app.Use(middleware.RequestID(), middleware.RequestLogger())
```

Use group middleware for a route family.

```go
api := app.Group("/api")
api.Use(middleware.JWT(keyFunc))
```

Use route middleware when the behavior belongs to one endpoint.

```go
app.Post("/exports", middleware.RateLimiter(), startExport)
```

## See also

- [Groups and Middleware](../guide/groups-and-middleware) explains middleware order and `c.Next()`.
- [Errors](../guide/errors) explains how returned middleware errors become responses.
- [Configuration](../guide/configuration) explains app-level extension points.
