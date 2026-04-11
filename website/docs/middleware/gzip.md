---
id: gzip
title: Gzip
description: Compress responses for clients that accept gzip.
sidebar_position: 12
---

`Gzip` compresses response bodies when the request includes `Accept-Encoding: gzip`.

```go
app.Use(middleware.Gzip())
```

Zinc skips compression when:

- the client does not accept gzip
- the response status cannot include a body
- the response already has a `Content-Encoding`
- the request method is `HEAD`

Use `GzipWithConfig` to set the compression level.

```go
app.Use(middleware.GzipWithConfig(middleware.GzipConfig{
	Level: gzip.BestSpeed,
}))
```

`Gzip` also adds `Vary: Accept-Encoding` when it compresses a response.
