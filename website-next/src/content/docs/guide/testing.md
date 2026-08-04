---
title: Testing
description: Test Zinc handlers, routes, middleware, errors, and native HTTP integration with httptest.
---

Zinc implements `http.Handler`, so normal `net/http/httptest` tools are enough.

```go
func TestGreeting(t *testing.T) {
    app := zinc.New()
    app.Get("/hello/{name}", func(c *zinc.Context) error {
        return c.JSON(zinc.Map{"hello": c.Param("name")})
    })

    req := httptest.NewRequest(http.MethodGet, "/hello/gopher", nil)
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
    }
    if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
        t.Fatalf("content-type=%q", got)
    }
}
```

## JSON requests

```go
body := strings.NewReader(`{"name":"Zinc"}`)
req := httptest.NewRequest(http.MethodPost, "/widgets", body)
req.Header.Set("Content-Type", "application/json")
```

## Test middleware

Register middleware exactly as production does and assert observable headers, status, body, or side effects. Avoid testing private middleware implementation details.

## Test a live server

Use `httptest.NewServer(app)` when the test needs redirects, cookies, streaming, or a real HTTP client:

```go
server := httptest.NewServer(app)
defer server.Close()

res, err := server.Client().Get(server.URL + "/health")
```
