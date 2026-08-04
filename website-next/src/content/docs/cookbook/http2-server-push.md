---
title: HTTP/2 Server Push
description: Understand optional HTTP/2 push support and prefer modern preload alternatives.
---

Zinc's response writer preserves `http.Pusher` when the server supports it:

```go
app.Get("/", func(c *zinc.Context) error {
    if pusher, ok := c.Writer().(http.Pusher); ok {
        _ = pusher.Push("/app.css", nil)
    }
    return c.File("./public/index.html")
})
```

HTTP/2 server push is no longer supported by major browsers and should not be a new application's primary optimisation. Prefer cache headers, `preload` links, early hints, and a CDN. The interface remains useful when integrating with specialised clients that still support push.
