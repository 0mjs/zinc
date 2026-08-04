---
title: Rate Limiter
description: Token-bucket rate limiting for global, per-IP, or custom-key request control.
---

`RateLimiter` is Zinc's first-party token-bucket limiter.

## Quick start

```go
app.Use(middleware.RateLimiter())
```

Default behavior:

- refill rate: `10` tokens per second
- bucket capacity: `10`
- response code: `429 Too Many Requests`

## Per-IP limiting

```go
app.Use(middleware.IPRateLimiter(20, 40))
```

That uses `c.IP()` as the limiting key.

## Config fields

| Field | Meaning |
|---|---|
| `Rate` | Tokens added per second |
| `Capacity` | Maximum number of tokens stored |
| `IPLookup` | Extract an IP or client identity |
| `KeyGenerator` | Generate a custom limiting key |
| `StatusCode` | Status to use when the limit is exceeded |
| `LimitReachedHandler` | Override the exceeded-limit response |

## Custom key example

```go
app.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
	Rate:     30,
	Capacity: 60,
	KeyGenerator: func(c *zinc.Context) string {
		return c.GetHeader("X-API-Key")
	},
	LimitReachedHandler: func(c *zinc.Context) error {
		return c.Status(429).JSON(zinc.Map{
			"error": "rate limit exceeded",
		})
	},
}))
```

## Notes

- If you set `KeyGenerator`, Zinc uses that first.
- If `KeyGenerator` is nil and `IPLookup` is set, Zinc uses the returned IP.
- If both are nil, Zinc uses one global limiter for the whole app.
