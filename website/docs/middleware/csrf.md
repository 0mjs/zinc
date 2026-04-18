---
title: CSRF
description: Cookie-backed CSRF protection with configurable token readers and fetch metadata checks.
sidebar_position: 3
---

`CSRF` issues a token for safe requests and verifies it on unsafe requests.

## Default behavior

```go
app.Use(middleware.CSRF())
```

By default Zinc:

- stores the token in a cookie named `_csrf`
- reads request tokens from `X-CSRF-Token`
- issues tokens on safe methods like `GET` and `HEAD`
- verifies tokens on unsafe methods like `POST`, `PUT`, `PATCH`, and `DELETE`

## Custom readers

You can combine readers when your app accepts tokens from more than one place.

```go
app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
	Readers: []middleware.CSRFReader{
		middleware.CSRFFromFirst(
			middleware.CSRFFromHeader(zinc.HeaderXCSRFToken),
			middleware.CSRFFromForm("_csrf"),
			middleware.CSRFFromQuery("csrf"),
		),
	},
	ExposeHeader: zinc.HeaderXCSRFToken,
}))
```

## Reader helpers

- `CSRFFromHeader(name)`
- `CSRFFromQuery(name)`
- `CSRFFromForm(name)`
- `CSRFFromFirst(readers...)`

## Config fields

| Field | Meaning |
|---|---|
| `Skipper` | Skip protection for selected requests |
| `Readers` | Token readers for unsafe requests |
| `Generate` | Custom token generator |
| `TokenBytes` | Random token size when using the default generator |
| `Cookie` | Cookie name and attributes |
| `ExposeHeader` | Optional response header that publishes the token |
| `TrustedOrigins` | Additional origins accepted for fetch metadata checks |
| `AllowFetchSite` | Custom fetch-site decision hook |
| `ErrorHandler` | Override CSRF failure behavior |

## Cookie configuration

`CSRFCookie` supports:

- `Name`
- `Domain`
- `Path`
- `MaxAge`
- `Secure`
- `HTTPOnly`
- `SameSite`

If you choose `SameSite=None`, Zinc automatically forces `Secure=true`.

## Accessing CSRF state

```go
state, ok := middleware.CSRFCurrent(c)
token, ok := middleware.CSRFToken(c)
```

The state tells you:

- the token value
- whether it was newly issued
- whether the request was verified
- the cookie name
- the normalized fetch-site value

## Failure model

CSRF failures are returned as `*middleware.CSRFViolation`, with reason values such as:

- `token_missing`
- `cookie_missing`
- `token_invalid`
- `fetch_site_rejected`

That makes it easy to customize API error responses without losing the reason for rejection.
