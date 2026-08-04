---
title: Request
description: Read route parameters, query strings, headers, forms, files, bodies, and the underlying net/http request.
---

Every Zinc handler receives a `*zinc.Context`. The original request is always available through `c.Request()`.

```go
app.Get("/inspect", func(c *zinc.Context) error {
    req := c.Request()
    return c.JSON(zinc.Map{
        "method": req.Method,
        "host":   req.Host,
        "path":   req.URL.Path,
    })
})
```

## Route and query values

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
    id := c.Param("id")
    page := c.QueryOr("page", "1")
    tags := c.QueryArray("tag")

    return c.JSON(zinc.Map{"id": id, "page": page, "tags": tags})
})
```

Zinc-native handlers read parameters with `c.Param` and avoid populating the
standard request's path-value map. When a route crosses into a standard handler
through `HandleHTTP` or `zinc.Wrap`, Zinc populates `Request.PathValue` first:

```go
app.HandleHTTP("GET /users/{id}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, r.PathValue("id"))
}))
```

## Headers and content type

```go
authorization := c.GetHeader(zinc.HeaderAuthorization)
contentType := c.ContentType()
```

## Forms and uploaded files

```go
name := c.FormValue("name")

file, err := c.FormFile("document")
if err != nil {
    return zinc.NewError(zinc.StatusBadRequest).WithMessage("document is required")
}
name := filepath.Base(file.Filename)
if err := c.SaveFile(file, filepath.Join("uploads", name)); err != nil {
	return err
}
```

Use `FormFiles` for multiple files and `MultipartForm` when you need direct access
to every value and file. Treat client filenames as untrusted input; keep only a
safe base name or replace it with an application-generated name.

## Request body

For typed input, prefer [binding](/guide/binding/). For raw access:

```go
body, err := c.BodyBytes()
if err != nil {
    return err
}
```

`BodyBytes` and `BodyString` cache the body so later binding or middleware can read it again.

## Cancellation and deadlines

The standard request context carries client cancellation and deadlines:

```go
select {
case result := <-work:
    return c.JSON(result)
case <-c.Context().Done():
    return c.Context().Err()
}
```

Use `c.SetContext(ctx)` when middleware needs to attach values or a derived deadline.
