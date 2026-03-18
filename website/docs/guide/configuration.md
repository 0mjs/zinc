---
id: configuration
title: Configuration
description: Understand Zinc’s routing, server, proxy, and extensibility settings.
sidebar_position: 8
---

Use `NewWithConfig` when you need more than the defaults.

```go
app := zinc.NewWithConfig(zinc.Config{
	CaseSensitive:          true,
	StrictRouting:          true,
	AutoHead:               true,
	AutoOptions:            true,
	HandleMethodNotAllowed: true,
	BodyLimit:              8 << 20,
	ProxyHeader:            zinc.HeaderXForwardedFor,
	TrustedProxies:         []string{"10.0.0.1"},
	RouteCacheSize:         1000,
})
```

## Routing behavior

- `CaseSensitive`
- `StrictRouting`
- `AutoHead`
- `AutoOptions`
- `HandleMethodNotAllowed`
- `RouteCacheSize`

These control how Zinc matches and caches routes.

## Server behavior

- `ServerHeader`
- `BodyLimit`
- `ReadTimeout`
- `WriteTimeout`
- `IdleTimeout`

## Proxy awareness

- `ProxyHeader`
- `TrustedProxies`

These affect how client IP helpers behave.

## Extension points

Plug in application-specific behavior with:

- `Binder`
- `Validator`
- `Renderer`
- `JSONCodec`
- `ErrorHandler`

Use these when you need custom serialization, validation, rendering, or error-envelope policies without replacing the rest of Zinc.
