---
title: Context Timeout
description: Attach per-request deadlines to the handler chain.
sidebar_position: 9
---

`ContextTimeout` wraps the request context with a deadline before the rest of the chain runs.

## Quick start

```go
app.Use(middleware.ContextTimeout(2 * time.Second))
```

## Config fields

| Field | Meaning |
|---|---|
| `Skipper` | Skip timeout handling for selected requests |
| `Timeout` | Required timeout duration |
| `ErrorHandler` | Customize timeout failures |

## Accessing timeout state

```go
info, ok := middleware.ContextTimeoutCurrent(c)
remaining, ok := middleware.ContextTimeoutRemaining(c)
```

`ContextTimeoutInfo` contains:

- `Timeout`
- `Deadline`

## Error behavior

If the handler chain returns `context.DeadlineExceeded`, Zinc wraps it in `*middleware.ContextTimeoutError`.

The default error handler returns `503 Service Unavailable` joined with the timeout error.

## Example

```go
app.Use(middleware.ContextTimeoutWithConfig(middleware.ContextTimeoutConfig{
	Timeout: 2 * time.Second,
	ErrorHandler: func(c *zinc.Context, err *middleware.ContextTimeoutError) error {
		return zinc.ErrServiceUnavailable.WithMessage("request timed out")
	},
}))
```
