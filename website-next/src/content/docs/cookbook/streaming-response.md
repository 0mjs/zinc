---
title: Streaming Response
description: Produce a Zinc response incrementally without buffering it in memory.
---

Use an `io.Pipe` when one goroutine produces the response over time:

```go
app.Get("/stream", func(c *zinc.Context) error {
    reader, writer := io.Pipe()

    go func() {
        defer writer.Close()
        for i := 1; i <= 5; i++ {
            if _, err := fmt.Fprintf(writer, "chunk %d\n", i); err != nil {
                return
            }
            time.Sleep(500 * time.Millisecond)
        }
    }()

    return c.Stream("text/plain; charset=utf-8", reader)
})
```

The request remains active until the reader reaches EOF or returns an error. Stop producing data when the request context is cancelled for long-running streams.
