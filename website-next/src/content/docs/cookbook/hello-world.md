---
title: Hello World
description: Run the smallest useful Zinc application and call it with curl.
---

Create `main.go`:

```go
package main

import (
    "log"

    "github.com/0mjs/zinc"
)

func main() {
    app := zinc.New()

    app.Get("/", func(c *zinc.Context) error {
        return c.String("Hello, World!")
    })

    log.Fatal(app.Listen())
}
```

Run it:

```sh
go mod init example.com/hello
go get github.com/0mjs/zinc
go run .
```

```sh
curl http://localhost:8080/
```

Zinc listens on `:8080` when `Listen` is called without an address.
