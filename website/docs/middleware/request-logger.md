---
title: 🪵 Request Logger
description: Structured request logging with configurable fields, hooks, and error handling.
sidebar_position: 6
---

`RequestLogger` logs request and response details after the handler chain completes.

## Quick start

```go
app.Use(middleware.RequestLogger())
```

Zinc also exposes `Logger()` and `LoggerWithConfig(...)` as aliases.

## Default fields

With the default config, Zinc logs:

- latency
- method
- URI
- status
- error
- remote IP
- host
- user agent
- request ID
- content length
- response size

## Config fields

| Field | Meaning |
|---|---|
| `Skipper` | Skip logging for selected requests |
| `BeforeNextFunc` | Hook that runs before the rest of the chain |
| `LogValuesFunc` | Receive extracted values and emit logs yourself |
| `HandleError` | Route returned errors through Zinc before logging |
| `Logger` | Custom `slog.Logger` for the default logger |
| `LogLatency` | Include request duration |
| `LogMethod` | Include request method |
| `LogURI` | Include request URI |
| `LogRoutePath` | Include matched route path |
| `LogStatus` | Include response status |
| `LogError` | Include request error |
| `LogRemoteIP` | Include remote IP |
| `LogHost` | Include host |
| `LogUserAgent` | Include user agent |
| `LogRequestID` | Include request ID |
| `LogContentLength` | Include request body size from headers |
| `LogResponseSize` | Include response bytes written |
| `LogHeaders` | Capture selected request headers |
| `LogQueryParams` | Capture selected query parameters |

## Custom logger example

```go
app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
	LogRoutePath: true,
	LogHeaders:   []string{"X-Request-ID", "X-Forwarded-For"},
	LogValuesFunc: func(c *zinc.Context, v middleware.RequestLoggerValues) error {
		slog.Info("http",
			"method", v.Method,
			"route", v.RoutePath,
			"status", v.Status,
			"latency", v.Latency,
		)
		return nil
	},
}))
```

## When to use `BeforeNextFunc`

Use it when you need to stamp context, start tracing spans, or derive correlation values before the rest of the middleware stack runs.
