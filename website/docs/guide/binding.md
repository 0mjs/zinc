---
id: binding
title: Binding
description: Use `c.Bind()` to decode path, query, headers, body formats, forms, and multipart files into Go types.
sidebar_position: 4
---

Zinc's primary binding API is `c.Bind()`.

## At a glance

```go
type CreateUserInput struct {
	TeamID int    `path:"teamID"`
	Page   int    `query:"page"`
	Name   string `json:"name"`
}

app.Post("/teams/{teamID}/users", func(c *zinc.Context) error {
	var input CreateUserInput
	if err := c.Bind().All(&input); err != nil {
		return err
	}
	return c.Status(zinc.StatusCreated).JSON(input)
})
```

Use `All` for normal API handlers and the source-specific methods when a handler needs stricter control.

## Bind the whole request

Use `c.Bind().All(&dst)` when you want Zinc to bind the request using its configured binder.

```go
type CreateUserInput struct {
	TeamID int      `path:"teamID"`
	Page   int      `query:"page"`
	Name   string   `json:"name"`
	Auth   string   `header:"x-auth"`
	Roles  []string `json:"roles"`
}

app.Post("/teams/{teamID}/users", func(c *zinc.Context) error {
	var input CreateUserInput
	if err := c.Bind().All(&input); err != nil {
		return err
	}
	return c.Status(zinc.StatusCreated).JSON(input)
})
```

This is the normal path for conventional API handlers.

`All(...)` binds in Zinc's normal order:

1. route params
2. query values
3. request body or form data, based on `Content-Type`
4. validation, if a validator is configured

## Bind a specific source

Use `c.Bind()` when you want to be precise about the source you are decoding.

```go
var pathInput PathInput
var queryInput QueryInput
var bodyInput BodyInput

if err := c.Bind().Path(&pathInput); err != nil {
	return err
}
if err := c.Bind().Query(&queryInput); err != nil {
	return err
}
if err := c.Bind().JSON(&bodyInput); err != nil {
	return err
}
```

Available `c.Bind()` methods:

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

That means both `c.Bind().All(&dst)` and `c.Bind().JSON(&dst)` participate in the same validation story.

## Bind errors

Binding errors keep source information so your error handler can explain whether a failure came from path params, query values, headers, forms, or the request body.

## See also

- [First Route](../getting-started/first-route) for a smaller binding example.
- [Configuration](./configuration) for custom binders and validators.
- [Bind API](../api/binding) for interfaces and error types.
