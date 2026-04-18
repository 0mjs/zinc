---
title: Basic Auth
description: Extract, validate, and expose HTTP Basic auth credentials.
sidebar_position: 5
---

`BasicAuth` is a small middleware for admin panels, internal tools, and simple protected routes.

## Quick start

```go
admin.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
	Validator: middleware.BasicAuthStatic("admin", "secret"),
}))
```

## Extractor helpers

- `BasicAuthFromAuthorizationHeader()`
- `BasicAuthFromHeader(name)`
- `BasicAuthFromHeaderPrefix(name, prefix)`
- `BasicAuthFromFirst(extractors...)`

## Validator helpers

- `BasicAuthStatic(username, password)`
- `BasicAuthStaticPairs(pairs...)`

`BasicAuthStaticPairs` hashes usernames and passwords before comparison and uses constant-time checks.

## Config fields

| Field | Meaning |
|---|---|
| `Skipper` | Skip auth for selected requests |
| `Extractor` | Read credentials from a custom source |
| `Validator` | Required validator |
| `SuccessHandler` | Override the success path |
| `ErrorHandler` | Override error/challenge behavior |
| `Realm` | Sets the `WWW-Authenticate` realm |

## Reading the authenticated identity

```go
identity, ok := middleware.BasicAuthCurrent(c)
username, ok := middleware.BasicAuthUsername(c)
```

Zinc stores:

- `Username`
- `Source`

That lets you distinguish whether credentials came from the normal authorization header or a custom header-based extractor.

## When to use this

Use Basic Auth when you genuinely want Basic Auth. For API access tokens, use [`JWT`](./jwt) instead.
