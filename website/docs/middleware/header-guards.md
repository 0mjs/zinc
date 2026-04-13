---
id: header-guards
title: Header Guards
description: Reject unsupported content headers and route middleware by request headers.
sidebar_position: 27
---

Header guards stop unsupported requests before handlers run.

## Content type

```go
app.Use(middleware.AllowContentType("application/json"))
```

Requests with another `Content-Type` return `415 Unsupported Media Type`.

## Content encoding

```go
app.Use(middleware.AllowContentEncoding("identity", "gzip"))
```

Missing `Content-Encoding` is treated as `identity`.

## Header routes

```go
app.Use(middleware.RouteHeaders(middleware.HeaderRoute{
	Header: "X-Admin",
	Value:  "1",
	Middleware: func(c *zinc.Context) error {
		c.Set("admin", true)
		return c.Next()
	},
}))
```

`RouteHeaders` runs the first matching header middleware. If nothing matches, the request continues normally.
