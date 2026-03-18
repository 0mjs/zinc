---
id: binding
title: Binding API
description: The binder interfaces, builder methods, formats, and validation flow.
sidebar_position: 4
---

Zinc’s binding story is centered around three types:

- `Binder`
- `Validator`
- `Binding`

## Binder interface

```go
type Binder interface {
	Bind(*Context, any) error
	BindBody(*Context, any) error
	BindQuery(*Context, any) error
	BindForm(*Context, any) error
	BindHeader(*Context, any) error
	BindPath(*Context, any) error
}
```

Provide a custom binder in `Config` if you need different decoding behavior.

## Binding builder

`c.Binding()` returns a `*Binding` helper with methods for explicit source decoding.

| Method | Purpose |
|---|---|
| `All` | Use the configured binder across supported request sources |
| `Body` | Decode request body only |
| `JSON`, `XML`, `YAML`, `TOML`, `Text` | Decode a specific body format |
| `Form` | Bind form or multipart form data |
| `Query` | Bind query values |
| `Header` | Bind request headers |
| `Path` | Bind route params |

## Validation

If `Config.Validator` is set, Zinc validates after binding.

This lets you centralize struct validation without changing individual handlers.

## BindError

`BindError` includes:

- `Source`
- `Field`
- `Err`

This is especially useful for APIs that want structured bad-request responses.
