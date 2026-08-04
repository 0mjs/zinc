---
title: Customization
description: Replace Zinc's binder, validator, renderer, JSON codec, error handler, routing policy, and server defaults.
---

Start from `zinc.DefaultConfig`, change only what your application owns, then construct the app with `NewWithConfig`.

```go
cfg := zinc.DefaultConfig
cfg.CaseSensitive = true
cfg.StrictRouting = true
cfg.BodyLimit = 8 << 20
cfg.ReadTimeout = 10 * time.Second
cfg.WriteTimeout = 20 * time.Second

app := zinc.NewWithConfig(cfg)
```

## Replace extension points

```go
cfg := zinc.DefaultConfig
cfg.RequestBinder = myBinder
cfg.Validator = myValidator
cfg.Renderer = myRenderer
cfg.JSONCodec = myJSONCodec
cfg.ErrorHandler = myErrorHandler

app := zinc.NewWithConfig(cfg)
```

Each extension point is a small interface or function type. Implement only the behavior you need.

## Standard HTTP middleware

`UseHTTP` accepts the normal `func(http.Handler) http.Handler` shape:

```go
app.UseHTTP(requestTracing, compression)
```

You can also wrap the entire application because `*zinc.App` implements `http.Handler`.

## Server ownership

For complete control over listeners and lifecycle, use `app.Handler()` with your own `http.Server`, or pass an existing listener to `app.Serve(listener)`.
