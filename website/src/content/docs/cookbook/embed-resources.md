---
title: Embed Resources
description: Compile static assets into a Zinc binary with embed.FS.
---

```go
package main

import (
    "embed"
    "io/fs"
    "log"

    "github.com/0mjs/zinc"
)

//go:embed public
var assets embed.FS

func main() {
    public, err := fs.Sub(assets, "public")
    if err != nil {
        log.Fatal(err)
    }

    app := zinc.New()
    if err := app.StaticFS("/", public); err != nil {
        log.Fatal(err)
    }

    log.Fatal(app.Listen())
}
```

Files under `public/` are now part of the executable. Configure `zinc.WithStaticIndex("index.html")` for a different index filename or `zinc.WithStaticBrowse(true)` to allow directory listings.
