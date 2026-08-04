---
title: JWT
description: Protect Zinc routes with signed bearer tokens and read typed claims.
---

```go
package main

import (
    "log"
    "os"

    jwt "github.com/golang-jwt/jwt/v5"
    "github.com/0mjs/zinc"
    "github.com/0mjs/zinc/middleware"
)

func main() {
    app := zinc.New()
    secret := []byte(os.Getenv("JWT_SECRET"))

    api := app.Group("/api", middleware.JWT(
        func(_ *zinc.Context, _ *jwt.Token) (any, error) {
            return secret, nil
        },
    ))

    api.Get("/me", func(c *zinc.Context) error {
        claims, ok := middleware.JWTClaims[jwt.MapClaims](c)
        if !ok {
            return zinc.NewError(zinc.StatusUnauthorized)
        }
        return c.JSON(claims)
    })

    log.Fatal(app.Listen())
}
```

The default extractor expects `Authorization: Bearer <token>`. Zinc also provides extractors for custom headers, cookies, query parameters, and fallback chains. See [JWT middleware](/middleware/jwt/).
