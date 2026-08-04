---
title: File Upload
description: Accept multipart uploads and save them safely with Zinc.
---

```go
package main

import (
    "log"
    "path/filepath"

    "github.com/0mjs/zinc"
)

func main() {
    app := zinc.New()

    app.Post("/upload", func(c *zinc.Context) error {
        file, err := c.FormFile("document")
        if err != nil {
            return zinc.NewError(zinc.StatusBadRequest).WithCause(err)
        }

        name := filepath.Base(file.Filename)
        if err := c.SaveFile(file, filepath.Join("uploads", name)); err != nil {
            return err
        }

        return c.Status(zinc.StatusCreated).JSON(zinc.Map{
            "name": name,
            "size": file.Size,
        })
    })

    log.Fatal(app.Listen())
}
```

Create the `uploads` directory before starting the app. Validate size, extension, content type, and authorization before keeping untrusted files.
