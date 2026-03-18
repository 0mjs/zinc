---
id: binding
title: Binding
description: Decode path, query, headers, body formats, forms, and multipart files into Go types.
sidebar_position: 4
---

Zinc supports both a simple binding path and an explicit builder-style path.

## The simple path

Use `c.Bind(&dst)` when you want Zinc to bind the request using its configured binder.

```go
type CreateUserInput struct {
	TeamID int      `path:"teamID"`
	Page   int      `query:"page"`
	Name   string   `json:"name"`
	Auth   string   `header:"x-auth"`
	Roles  []string `json:"roles"`
}

app.Post("/teams/:teamID/users", func(c *zinc.Context) error {
	var input CreateUserInput
	if err := c.Bind(&input); err != nil {
		return err
	}
	return c.Status(zinc.StatusCreated).JSON(input)
})
```

`Bind` is ideal for conventional API handlers.

## The explicit path

Use `c.Binding()` when you want to be precise about the source you are decoding.

```go
var pathInput PathInput
var queryInput QueryInput
var bodyInput BodyInput

if err := c.Binding().Path(&pathInput); err != nil {
	return err
}
if err := c.Binding().Query(&queryInput); err != nil {
	return err
}
if err := c.Binding().JSON(&bodyInput); err != nil {
	return err
}
```

Available builder methods:

- `All`
- `Body`
- `JSON`
- `XML`
- `YAML`
- `TOML`
- `Text`
- `Form`
- `Query`
- `Header`
- `Path`

## Supported request formats

Out of the box, Zinc supports:

- JSON
- XML
- YAML
- TOML
- plain text
- URL-encoded forms
- multipart forms

## Multipart file binding

Zinc can bind file fields directly into structs.

Supported multipart targets include:

- `multipart.FileHeader`
- `*multipart.FileHeader`
- `[]multipart.FileHeader`
- `[]*multipart.FileHeader`

Example:

```go
type UploadInput struct {
	Name   string                  `form:"name"`
	Avatar *multipart.FileHeader   `form:"avatar"`
	Files  []*multipart.FileHeader `form:"files"`
}
```

## Validation

If you provide a custom `Validator` in `Config`, Zinc validates after binding.

```go
type Validator interface {
	Validate(any) error
}
```

That means both `c.Bind(&dst)` and `c.Binding().JSON(&dst)` participate in the same validation story.

## Bind errors

Zinc exposes a typed `BindError` to make source-aware error handling easier.

It includes:

- `Source`
- `Field`
- `Err`

This is useful when you want structured error responses rather than a single generic bad-request message.

## Body-only helpers

Zinc also exposes source-specific helpers directly on `Context`:

- `BindJSON`
- `BindXML`
- `BindYAML`
- `BindTOML`
- `BindText`
- `BindForm`
- `BindQuery`
- `BindHeader`
- `BindPath`

Use these when you want the explicit source path without going through `Binding()`.
