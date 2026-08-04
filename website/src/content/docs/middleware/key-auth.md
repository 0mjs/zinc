---
title: Key Auth
description: Validate API keys from headers, query values, or cookies.
---

`KeyAuth` protects routes with an API key validator.

```go
app.Use(middleware.KeyAuth(middleware.KeyAuthStatic(os.Getenv("API_KEY"))))
```

By default, Zinc reads `Authorization: Bearer <key>`.

Use another extractor when keys live somewhere else.

```go
app.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
	Extractor: middleware.KeyAuthFromHeader("X-API-Key"),
	Validator: middleware.KeyAuthStatic("secret"),
}))
```

Available extractors:

- `KeyAuthFromAuthorizationHeader()`
- `KeyAuthFromHeader(header)`
- `KeyAuthFromHeaderPrefix(header, prefix)`
- `KeyAuthFromQuery(name)`
- `KeyAuthFromCookie(name)`
- `KeyAuthFromFirst(...)`

Inside handlers, use `KeyAuthCurrent(c)` or `MustKeyAuthCurrent(c)` to read the accepted key source.
