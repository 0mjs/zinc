---
title: Response Writer
description: Wrap response writers in middleware to track status, bytes written, and write state.
---

`WrapResponseWriter` is for middleware authors.

Use it when middleware needs to inspect the response after the next handler runs.

```go
func measure() zinc.Middleware {
	return func(c *zinc.Context) error {
		base := c.Writer()
		rw := zinc.WrapResponseWriter(base)
		c.SetWriter(rw)
		defer c.SetWriter(base)

		err := c.Next()

		status := rw.Status()
		bytes := rw.BytesWritten()
		written := rw.Written()

		_ = status
		_ = bytes
		_ = written
		return err
	}
}
```

## Interface

```go
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
	Written() bool
}
```

`Status()` returns `200 OK` until another status is written. `Written()` tells you whether headers or body bytes have been sent.

## Optional behavior

The wrapper preserves common optional writer behavior when the underlying writer supports it:

- `http.Flusher`
- `http.Hijacker`
- `io.ReaderFrom`
- `http.Pusher`
- `Unwrap() http.ResponseWriter`

Use a narrow assertion when integration code needs the original writer.

```go
if unwrapper, ok := rw.(interface{ Unwrap() http.ResponseWriter }); ok {
	base := unwrapper.Unwrap()
	_ = base
}
```
