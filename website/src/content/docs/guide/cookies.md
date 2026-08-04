---
title: Cookies
description: Read, write, configure, and clear HTTP cookies with Zinc and net/http.
---

Zinc uses the standard `http.Cookie` type.

## Set a cookie

```go
app.Post("/login", func(c *zinc.Context) error {
    c.SetCookie(&http.Cookie{
        Name:     "session",
        Value:    token,
        Path:     "/",
        MaxAge:   3600,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteLaxMode,
    })
    return c.NoContent()
})
```

## Read cookies

```go
cookie, err := c.Cookie("session")
if errors.Is(err, http.ErrNoCookie) {
    return zinc.NewError(zinc.StatusUnauthorized).WithMessage("sign in required")
}
if err != nil {
    return err
}
```

Use `c.Cookies()` to read all cookies.

## Clear cookies

```go
c.ClearCookie("session", "preferences")
return c.NoContent()
```

Zinc's session, CSRF, JWT, and key-auth middleware also provide cookie-specific configuration and extractors.
