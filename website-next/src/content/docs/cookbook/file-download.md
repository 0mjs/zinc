---
title: File Download
description: Serve files inline or as named browser downloads.
---

Force a download with a safe public filename:

```go
app.Get("/reports/{id}", func(c *zinc.Context) error {
    id := c.Param("id")
    path := filepath.Join("exports", id+".csv")
    return c.Download(path, "report-"+id+".csv")
})
```

Display a file in the browser when supported:

```go
app.Get("/manual", func(c *zinc.Context) error {
    return c.Inline("./public/manual.pdf", "zinc-manual.pdf")
})
```

Use `c.File` when no content-disposition header is required. Never build filesystem paths directly from an unvalidated catch-all parameter.
