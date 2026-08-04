---
title: Secure
description: Set common security response headers.
---

`Secure` sets a practical default set of security headers.

```go
app.Use(middleware.Secure())
```

Defaults include:

- `X-XSS-Protection: 0`
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: SAMEORIGIN`
- `Referrer-Policy: no-referrer`
- `Cross-Origin-Resource-Policy: same-origin`

Use `SecureWithConfig` to add CSP or HSTS.

```go
app.Use(middleware.SecureWithConfig(middleware.SecureConfig{
	ContentSecurityPolicy: "default-src 'self'",
	HSTSMaxAge:            31536000,
}))
```

HSTS is only written for secure requests.
