---
id: errors
title: Errors
description: Return structured HTTP errors, short-circuit handlers cleanly, and customize error handling globally.
sidebar_position: 7
---

Zinc uses normal Go error returns for handler failure flow.

## Returning errors

```go
app.Get("/users/:id", func(c *zinc.Context) error {
	return zinc.ErrNotFound
})
```

If the returned error is an `HTTPError`, Zinc writes the associated status and message.

## Structured HTTP errors

Build more specific errors with `WithMessage`, `WithCause`, `WithMeta`, and `WithHeader`.

```go
return zinc.ErrBadRequest.
	WithMessage("invalid user payload").
	WithCause(err).
	WithMeta("field", "email")
```

## Predefined errors

Zinc ships a broad set of predefined HTTP errors, including:

- `ErrBadRequest`
- `ErrUnauthorized`
- `ErrForbidden`
- `ErrNotFound`
- `ErrMethodNotAllowed`
- `ErrUnprocessableEntity`
- `ErrInternalServerError`

You can also create your own with `NewError(code)`.

## Abort helpers

For explicit short-circuiting inside handlers or middleware:

- `AbortWithStatus(code)`
- `AbortWithJSON(code, value)`
- `Fail(err)`

Example:

```go
func requireAuth(c *zinc.Context) error {
	if c.GetHeader("Authorization") == "" {
		return c.AbortWithStatus(zinc.StatusUnauthorized)
	}
	return c.Next()
}
```

## Custom error handlers

Override Zinc’s default error writer with `Config.ErrorHandler`.

```go
app := zinc.NewWithConfig(zinc.Config{
	ErrorHandler: func(c *zinc.Context, err error) {
		_ = c.Status(zinc.StatusInternalServerError).JSON(zinc.Map{
			"error": err.Error(),
		})
	},
})
```

That gives you one place to standardize API error envelopes.
