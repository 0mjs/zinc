---
title: CORS
description: Configure cross-origin policies, preflight responses, and exposed headers.
sidebar_position: 2
---

`CORS` handles standard cross-origin response headers and browser preflight requests.

## Quick start

```go
app.Use(middleware.CORS("https://app.example.com"))
```

That gives you a narrow allow-origin policy with Zinc's default methods and headers.

## Constructors

| API | Use when |
|---|---|
| `middleware.CORS(origins...)` | You only need to set allowed origins |
| `middleware.CORSWithConfig(cfg)` | You want full config control |
| `middleware.CORSWithOptions(...)` | You prefer additive option helpers |

## Option helpers

Zinc exposes small helpers for the common knobs:

- `CORSAllowOrigins(...)`
- `CORSAllowMethods(...)`
- `CORSAllowHeaders(...)`
- `CORSExposeHeaders(...)`
- `CORSAllowCredentials(bool)`
- `CORSMaxAge(time.Duration)`
- `CORSMaxAgeSeconds(int)`
- `CORSSkipper(func(*zinc.Context) bool)`

## Config fields

| Field | Meaning |
|---|---|
| `AllowOrigins` | Exact origins to allow. `*` is supported. |
| `AllowMethods` | Methods returned on preflight responses. |
| `AllowHeaders` | Allowed request headers for preflight checks. |
| `ExposeHeaders` | Headers the browser may expose to client code. |
| `AllowCredentials` | Enables `Access-Control-Allow-Credentials`. |
| `MaxAge` | Preflight cache duration in seconds. |
| `Skipper` | Skip middleware for selected requests. |

## Example with options

```go
app.Use(middleware.CORSWithOptions(
	middleware.CORSAllowOrigins("https://app.example.com", "https://admin.example.com"),
	middleware.CORSAllowMethods("GET", "POST", "PATCH", "DELETE"),
	middleware.CORSAllowHeaders("Authorization", "Content-Type", "X-Request-ID"),
	middleware.CORSExposeHeaders("X-Request-ID"),
	middleware.CORSAllowCredentials(true),
	middleware.CORSMaxAge(10*time.Minute),
))
```

## Notes

- If you enable `AllowCredentials`, browsers will reject `Access-Control-Allow-Origin: *`, so Zinc reflects the request origin instead.
- Zinc automatically handles real preflight requests and returns `204 No Content`.
- Requests without an `Origin` header pass straight through.
