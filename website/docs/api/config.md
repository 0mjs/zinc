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
| `Binder`, `Validator`, `Renderer`, `JSONCodec`, `ErrorHandler` | Extension points |
| `RouteCacheSize` | Router cache size |

Use `NewWithConfig` whenever your application needs more than Zinc’s defaults.
