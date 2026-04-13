---
id: decompress
title: Decompress
description: Decompress gzip request bodies before handlers read them.
sidebar_position: 13
---

`Decompress` unwraps request bodies sent with `Content-Encoding: gzip`.

```go
app.Use(middleware.Decompress())
```

Handlers can then read the body normally.

```go
app.Post("/ingest", func(c *zinc.Context) error {
	var input Event
	if err := c.Bind().JSON(&input); err != nil {
		return err
	}
	return c.NoContent()
})
```

Unsupported content encodings return `415 Unsupported Media Type`. Invalid gzip bodies return `400 Bad Request`.

Use `MaxDecompressedSize` when compressed uploads should be capped after decompression.

```go
app.Use(middleware.DecompressWithConfig(middleware.DecompressConfig{
	MaxDecompressedSize: 4 << 20,
}))
```

Requests that expand past the limit return `413 Payload Too Large`.
