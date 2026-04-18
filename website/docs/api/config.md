---
id: config
title: Config
description: Routing behavior, server limits, proxy handling, and extension points.
sidebar_position: 5
---

`Config` controls Zinc’s runtime behavior.

## Default shape

Important defaults include:

- `CaseSensitive: false`
- `StrictRouting: false`
- `AutoHead: true`
- `AutoOptions: true`
- `HandleMethodNotAllowed: true`
- `BodyLimit: 4 << 20`
- `ReadTimeout: 5s`
- `WriteTimeout: 10s`
- `IdleTimeout: 120s`
- `ProxyHeader: "X-Forwarded-For"`
- `RouteCacheSize: 1000`

## Main fields

| Field | Purpose |
|---|---|
| `ServerHeader` | Override the server header |
| `CaseSensitive` | Make route matching case-sensitive |
| `StrictRouting` | Distinguish `/users` from `/users/` |
| `AutoHead` | Automatically support `HEAD` for `GET` routes |
| `AutoOptions` | Automatically respond to `OPTIONS` |
| `HandleMethodNotAllowed` | Return `405` when a path exists for another method |
| `BodyLimit` | Maximum request body size |
| `ReadTimeout`, `WriteTimeout`, `IdleTimeout` | Server timeouts |
| `ProxyHeader`, `TrustedProxies` | Proxy-aware client IP behavior |
| `RequestBinder`, `Validator`, `Renderer`, `JSONCodec`, `ErrorHandler` | Extension points |
| `RouteCacheSize` | Router cache size |

Use `NewWithConfig` whenever your application needs more than Zinc’s defaults.

Unset zero-value fields are normalized to safe defaults for body size, timeouts, proxy header, JSON codec, request binder, and error handler.

## Proxy trust

`ProxyHeader` controls which forwarded header Zinc reads for `c.IP()` and `c.IPs()`. `TrustedProxies` controls when that header is trusted.

```go
app := zinc.NewWithConfig(zinc.Config{
	ProxyHeader:    zinc.HeaderXForwardedFor,
	TrustedProxies: []string{"10.0.0.1"},
})
```

Leave `TrustedProxies` empty when the app is exposed directly to users and forwarded headers should not be trusted.

## Extension points

Use the extension fields to replace one part of Zinc without changing the handler API:

```go
app := zinc.NewWithConfig(zinc.Config{
	Validator:    validator,
	Renderer:     renderer,
	ErrorHandler: writeError,
})
```

`RequestBinder` controls request decoding, `JSONCodec` controls JSON encode/decode behavior, and `ErrorHandler` controls how returned errors are written.
