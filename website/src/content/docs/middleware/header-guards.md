---
title: Header Guards
description: Reject unsupported content headers and route middleware by request headers.
---

Header guards stop unsupported requests before handlers run.

## Content type

```go
app.Use(middleware.AllowContentType("application/json"))
```

Requests with another `Content-Type` return `415 Unsupported Media Type`.

Zinc compares media types without parameters, so `application/json; charset=utf-8` matches `application/json`.

## Content encoding

```go
app.Use(middleware.AllowContentEncoding("identity", "gzip"))
```

Missing `Content-Encoding` is treated as `identity`.

Use this before binders or body readers when an endpoint only accepts a known set of encodings.

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

If `Value` is empty, any non-empty value for that header matches. Rules without a header or middleware are ignored.
