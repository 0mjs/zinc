---
title: Streaming Response
description: Write and flush a Zinc response incrementally without buffering the whole result.
---

Use Go's `http.ResponseController` when every chunk must reach the client as it
is produced. It works with Zinc because the context exposes the standard
response writer.

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/0mjs/zinc"
)

func main() {
	app := zinc.New()

	app.Get("/stream", func(c *zinc.Context) error {
		c.Type("txt")
		controller := http.NewResponseController(c.Writer())

		for i := 1; i <= 5; i++ {
			if _, err := fmt.Fprintf(c.Writer(), "chunk %d\n", i); err != nil {
				return err
			}
			if err := controller.Flush(); err != nil {
				return err
			}

			select {
			case <-time.After(500 * time.Millisecond):
			case <-c.Context().Done():
				return nil
			}
		}

		return nil
	})

	log.Fatal(app.Listen(":8080"))
}
```

Test without client-side buffering:

```bash
curl --no-buffer http://localhost:8080/stream
```

Use `c.Stream(contentType, reader)` when you already have an `io.Reader` and do
not need to control flush boundaries. Long-running streams should always observe
`c.Context().Done()`.
