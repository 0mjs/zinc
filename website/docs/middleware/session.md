---
id: session
title: Session
description: Store signed cookie-backed string session values.
sidebar_position: 23
---

`SessionCookie` stores small string session values in a signed cookie.

```go
app.Use(middleware.SessionCookie("sid", os.Getenv("SESSION_SECRET")))
```

Inside handlers:

```go
app.Get("/profile", func(c *zinc.Context) error {
	session := middleware.MustSession(c)
	session.Set("last_path", c.Path())
	return c.String(session.Get("user_id"))
})
```

Use `SessionWithConfig` to customize cookie settings.

```go
app.Use(middleware.SessionWithConfig(middleware.SessionConfig{
	Name:     "sid",
	Secret:   []byte(os.Getenv("SESSION_SECRET")),
	MaxAge:   86400,
	Secure:   true,
	SameSite: http.SameSiteLaxMode,
}))
```

This middleware is for small cookie-backed values. Store large or sensitive session state server-side and keep only an opaque ID in the cookie.
