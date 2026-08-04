---
title: Errors
description: HTTPError, predefined status errors, and custom error handling hooks.
---

Zinc uses `HTTPError` for structured HTTP failures.

## HTTPError

```go
type HTTPError struct {
	Code    int
	Message string
	Cause   error
	Meta    zinc.Map
	Headers http.Header
}
```

Useful methods:

- `WithMessage`
- `WithCause`
- `WithMeta`
- `WithHeader`

## Create errors

```go
return zinc.NewError(zinc.StatusBadRequest).WithMessage("invalid payload")
```

## Predefined errors

Zinc includes ready-made errors for common HTTP statuses, including:

- `ErrBadRequest`
- `ErrUnauthorized`
- `ErrForbidden`
- `ErrNotFound`
- `ErrMethodNotAllowed`
- `ErrTooManyRequests`
- `ErrInternalServerError`

## Error handling hook

Set `Config.ErrorHandler` to control how returned errors are rendered.

That is the right place to implement a project-wide JSON error envelope.
