---
title: net/http Interoperability
description: Use standard Go handlers and middleware directly inside Zinc.
---

Zinc is an `http.Handler`. Standard handlers and middleware can remain standard
handlers and middleware, including when they need route parameters.

## Register a standard handler

Use `HandleHTTP` with a `METHOD /pattern` string:

```go
app.HandleHTTP("GET /native/{id}", http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		fmt.Fprintf(w, "user %s", id)
	},
))
```

The handler receives the request and response writer used for the Zinc request.
Zinc does not clone the request or translate it to another HTTP representation.
Matched parameters are copied into `PathValue` immediately before the standard
handler runs. Zinc-native handlers use `c.Param`, so they do not pay this
interoperability cost when it is not needed.

Groups support the same API:

```go
api := app.Group("/api", requireAPIKey)
api.HandleHTTP("GET /metrics", metricsHandler)
```

Group and application Zinc middleware still run around the standard handler.

## Register standard middleware

Use `UseHTTP` for middleware with the usual Go signature:

```go
func requestTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add tracing data to r.Context() here.
		next.ServeHTTP(w, r)
	})
}

app.UseHTTP(requestTracing)
```

Standard middleware wraps the complete application. The first registered HTTP
middleware is outermost:

```text
HTTP middleware
  -> Zinc application middleware
    -> Zinc group and route middleware
      -> route handler
```

`UseHTTP` is deliberately application-wide. Use Zinc middleware when behavior
belongs to a particular group or route.

## Prometheus

Register a Prometheus exporter without an adapter:

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

app.HandleHTTP("GET /metrics", promhttp.Handler())
```

## OpenTelemetry

OpenTelemetry's standard HTTP instrumentation can wrap the whole app:

```go
import "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

app.UseHTTP(func(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "zinc")
})
```

The instrumented request context is the same context available through
`c.Request().Context()` and `c.Context()` inside Zinc handlers.

## Response writer capabilities

Standard handlers receive the original response writer. Capabilities such as
`http.Flusher`, `http.Hijacker`, `io.ReaderFrom`, `http.Pusher`, and `Unwrap` are
therefore retained when the server's writer supports them.

Use `Mount` when an existing handler owns a whole path subtree. Use `HandleHTTP`
when it owns one routed endpoint.
