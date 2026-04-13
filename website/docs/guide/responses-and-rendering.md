---
id: responses-and-rendering
title: 📤 Responses and Rendering
description: Return JSON, XML, YAML, TOML, HTML, templates, files, downloads, streams, and redirects.
sidebar_position: 5
---

Zinc provides response helpers for common API and web application flows.

## Plain responses

```go
return c.String("ok")
return c.HTML("<strong>ok</strong>")
return c.NoContent()
```

## Structured responses

```go
return c.JSON(zinc.Map{"ok": true})
return c.XML(payload)
return c.YAML(payload)
return c.TOML(payload)
```

For pretty JSON:

```go
return c.JSONPretty(payload, "  ")
```

If you already have encoded bytes:

```go
return c.JSONBlob(zinc.StatusOK, rawJSON)
return c.XMLBlob(zinc.StatusOK, rawXML)
return c.HTMLBlob(zinc.StatusOK, rawHTML)
return c.Blob(zinc.StatusOK, "application/custom", payload)
```

Blob helpers write the bytes you pass in. They do not re-encode the payload.

## Status and headers

Use the response builder style when you want to set headers or status before writing the body.

```go
return c.
	Status(zinc.StatusCreated).
	SetHeader(zinc.HeaderLocation, "/users/42").
	JSON(zinc.Map{"id": 42})
```

Useful helpers include:

- `Status(code)`
- `SetHeader(key, value)`
- `AppendHeader(key, values...)`
- `Type(ext)`
- `Location(url)`
- `Vary(fields...)`
- `SetSameSite(mode)`

For cookies:

```go
c.SetSameSite(http.SameSiteLaxMode)
c.SetCookie(&http.Cookie{Name: "session", Value: token, Path: "/"})
```

## Redirects

```go
return c.Redirect(zinc.StatusTemporaryRedirect, "/login")
```

## Templates

Configure a renderer in `Config` and then call `Render`.

```go
views := template.Must(template.ParseGlob("templates/*.html"))

app := zinc.NewWithConfig(zinc.Config{
	Renderer: zinc.NewHTMLTemplateRenderer(views),
})

app.Get("/dashboard", func(c *zinc.Context) error {
	return c.Render("dashboard", zinc.Map{"Title": "Overview"})
})
```

## Files and downloads

Serve files directly from disk or an `fs.FS`.

```go
return c.File("./public/report.pdf")
return c.FileFS("report.pdf", embeddedFiles)
```

For downloads and attachments:

```go
return c.Download("./exports/users.csv", "users-latest.csv")
return c.Inline("./public/report.pdf")
```

## Streams

```go
reader := strings.NewReader("streamed content")
return c.Stream("text/plain; charset=utf-8", reader)
```

For server-sent events, write one event at a time:

```go
app.Get("/events", func(c *zinc.Context) error {
	for _, msg := range messages {
		if err := c.SSE(zinc.SSEvent{
			Event: "message",
			Data:  zinc.Map{"text": msg},
		}); err != nil {
			return err
		}
		if flusher, ok := c.Writer().(http.Flusher); ok {
			flusher.Flush()
		}
	}
	return nil
})
```

`SSE` writes one event and leaves streaming control with the handler.

## Content negotiation

Use `Accepts` when the handler needs to branch before writing.

```go
switch c.Accepts("application/json", "text/html") {
case "application/json":
	return c.JSON(payload)
case "text/html":
	return c.Render("users/show", payload)
default:
	return zinc.ErrNotAcceptable
}
```

Use `Negotiate` when the response body can be offered in multiple content types.

```go
return c.Negotiate(zinc.StatusOK, map[string]any{
	"application/json": zinc.Map{"ok": true},
	"text/plain":       "ok",
})
```

`Accepts` follows the request `Accept` header, including quality values and `type/*` or `*/*` wildcards. When no `Accept` header is present, Zinc picks the first offered type.

## Response writer wrapping

Middleware that needs response status or byte counts can wrap the underlying writer.

```go
base := c.Writer()
rw := zinc.WrapResponseWriter(base)
c.SetWriter(rw)
defer c.SetWriter(base)
```

`WrapResponseWriter` returns `zinc.ResponseWriter`.

```go
type ResponseWriter interface {
	http.ResponseWriter
	Status() int
	BytesWritten() int
	Written() bool
}
```

It preserves common optional writer behavior such as flushing, hijacking, `io.ReaderFrom`, server push, and unwrapping when the underlying writer supports it.

## Good API pattern

For most API handlers, a clean pattern is:

```go
return c.Status(zinc.StatusCreated).JSON(payload)
```

That keeps status, headers, and body shape close together and easy to read.
