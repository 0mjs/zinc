---
title: Body Dump
description: Capture request and response bodies for logging, auditing, or test instrumentation.
sidebar_position: 8
---

`BodyDump` captures request and response payloads and sends them to an observer.

## Quick start

```go
app.Use(middleware.BodyDump(func(c *zinc.Context, snapshot middleware.BodyDumpSnapshot) {
	slog.Info("body_dump",
		"route", snapshot.RoutePath,
		"status", snapshot.Status,
		"request_bytes", snapshot.RequestBytes,
		"response_bytes", snapshot.ResponseBytes,
	)
}))
```

## Config fields

| Field | Meaning |
|---|---|
| `Skipper` | Skip capture for selected requests |
| `Observe` | Required observer callback |
| `Redact` | Mutate the snapshot before observation |
| `MaxRequestBytes` | Capture limit for the request body |
| `MaxResponseBytes` | Capture limit for the response body |

## Snapshot fields

`BodyDumpSnapshot` includes:

- `Method`
- `Path`
- `RoutePath`
- `Status`
- `RequestBody`
- `ResponseBody`
- `RequestBytes`
- `ResponseBytes`
- `RequestTruncated`
- `ResponseTruncated`
- `Error`

## Redaction example

```go
app.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
	Observe: func(_ *zinc.Context, snapshot middleware.BodyDumpSnapshot) {
		// send to logs, audit pipeline, or tests
	},
	Redact: func(_ *zinc.Context, snapshot *middleware.BodyDumpSnapshot) {
		snapshot.RequestBody = []byte("[redacted]")
	},
	MaxRequestBytes:  64 << 10,
	MaxResponseBytes: 64 << 10,
}))
```

## Good fit

Use this for debugging, auditing, and integration tests. Avoid it on very high-volume routes unless you are deliberate about truncation and redaction.
