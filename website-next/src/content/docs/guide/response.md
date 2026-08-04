---
title: Response
description: Send strings, structured data, streams, files, redirects, cookies, and standard net/http responses.
---

Zinc response helpers write through the standard `http.ResponseWriter` exposed by `c.Writer()`.

## Status and headers

`Status`, `SetHeader`, `AppendHeader`, `Type`, `Location`, and `Vary` return the context for chaining.

```go
return c.
    Status(zinc.StatusCreated).
    SetHeader("X-Resource-ID", "42").
    JSON(zinc.Map{"id": 42})
```

## Common response helpers

```go
return c.String("hello")
return c.JSON(zinc.Map{"ok": true})
return c.JSONPretty(value, "  ")
return c.XML(value)
return c.YAML(value)
return c.TOML(value)
return c.HTML("<h1>Hello</h1>")
return c.NoContent()
```

`Send` selects an appropriate representation for common Go values. Prefer explicit helpers in public APIs when the response contract matters.

## Raw data and streams

```go
return c.Data("application/pdf", pdfBytes)
```

Stream an `io.Reader` without buffering the entire response:

```go
return c.Stream("text/plain; charset=utf-8", reader)
```

## Files

```go
return c.File("./public/manual.pdf")
return c.Download("./exports/report.csv", "report.csv")
return c.Inline("./public/manual.pdf")
```

`FileFS` serves from an `fs.FS`, which works well with `embed.FS`.

## Redirects

```go
return c.Redirect(zinc.StatusTemporaryRedirect, "/new-location")
```

## Standard library access

Use the writer directly whenever a helper is not appropriate:

```go
app.Get("/native", func(c *zinc.Context) error {
    http.Error(c.Writer(), "temporarily unavailable", http.StatusServiceUnavailable)
    return nil
})
```

Do not call another Zinc response helper after the response has been committed.
