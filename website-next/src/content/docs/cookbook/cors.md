---
title: CORS
description: Allow selected browser origins, methods, headers, and credentials.
---

Allow a known frontend origin:

```go
cfg := middleware.DefaultCORSConfig()
cfg.AllowOrigins = []string{"https://app.example.com"}
cfg.AllowMethods = []string{
    http.MethodGet,
    http.MethodPost,
    http.MethodPut,
    http.MethodDelete,
}
cfg.AllowHeaders = []string{
    zinc.HeaderContentType,
    zinc.HeaderAuthorization,
}
cfg.AllowCredentials = true

app.Use(middleware.CORSWithConfig(cfg))
```

For a public unauthenticated API:

```go
app.Use(middleware.CORS("*"))
```

Do not combine the wildcard origin with credentialed browser requests. See the full [CORS middleware reference](/middleware/cors/).
