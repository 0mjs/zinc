---
id: binding
title: Bind
description: The request binding interfaces, `c.Bind()` helper, supported formats, and validation flow.
sidebar_position: 4
---

Zinc’s binding story is centered around three pieces:

- `RequestBinder`
- `Validator`
- the `Bind` helper returned by `c.Bind()`

## RequestBinder interface

```go
type RequestBinder interface {
	Bind(*Context, any) error
	BindBody(*Context, any) error
	BindQuery(*Context, any) error
	BindForm(*Context, any) error
	BindHeader(*Context, any) error
	BindPath(*Context, any) error
}
```

Provide a custom request binder in `Config` if you need different decoding behavior.

## `c.Bind()` helper

`c.Bind()` is Zinc's primary binding API.

```go
if err := c.Bind().All(&input); err != nil {
	return err
}
```

Use it when you want either:

- the inferred all-source bind path with `All(...)`
- an explicit source path like `JSON(...)` or `Query(...)`

| Method | Purpose |
|---|---|
| `All` | Use the configured request binder across supported request sources |
| `Body` | Decode request body only |
| `JSON`, `XML`, `YAML`, `TOML`, `Text` | Decode a specific body format |
| `Form` | Bind form or multipart form data |
| `Query` | Bind query values |
| `Header` | Bind request headers |
| `Path` | Bind route params |

Common struct tags:

- `path:"id"`
- `query:"page"`
- `header:"x-request-id"`
- `form:"name"`
- `json:"name"`
- `xml:"name"`
- `yaml:"name"`
- `toml:"name"`

## Validation

If `Config.Validator` is set, Zinc validates after binding.

This lets you centralize struct validation without changing individual handlers.

## BindError

`BindError` includes:

- `Source`
- `Field`
- `Err`

This is especially useful for APIs that want structured bad-request responses.
