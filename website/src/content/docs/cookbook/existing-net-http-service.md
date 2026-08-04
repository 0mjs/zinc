---
title: Add Zinc to an Existing net/http Service
description: Introduce Zinc routes without rewriting working standard-library handlers or middleware.
---

Zinc is an `http.Handler`, so adopting it does not require a flag day. Keep an
existing `http.ServeMux`, mount it under Zinc, then add Zinc routes where they
make the application clearer.

## Setup

```bash
mkdir zinc-adoption
cd zinc-adoption
go mod init example.com/zinc-adoption
go get github.com/0mjs/zinc
```

## Application

```go
package main

import (
	"log"
	"net/http"
	"time"

	"github.com/0mjs/zinc"
)

func accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(started))
	})
}

func main() {
	legacy := http.NewServeMux()
	legacy.HandleFunc("GET /reports", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("legacy report\n"))
	})

	app := zinc.New()

	// Standard middleware can continue to wrap the whole service.
	app.UseHTTP(accessLog)

	// Mount strips /legacy before the request reaches legacy.
	app.Mount("/legacy", legacy)

	// New endpoints can use Zinc's concise handler API.
	app.Get("/api/health", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{"status": "ok"})
	})

	// Standard handlers can also own individual routed endpoints.
	app.HandleHTTP("GET /metrics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("requests_total 1\n"))
	}))

	log.Fatal(app.Listen(":8080"))
}
```

## Try it

```bash
curl http://localhost:8080/legacy/reports
curl http://localhost:8080/api/health
curl http://localhost:8080/metrics
```

## Choose the smallest integration point

- Use `app.Mount("/prefix", handler)` when an existing handler owns a subtree.
- Use `app.HandleHTTP("METHOD /path", handler)` for one standard handler.
- Use `app.UseHTTP(middleware)` for standard middleware around the whole app.
- Use Zinc handlers for new routes that benefit from context helpers, binding,
  groups, and framework middleware.

Mounted handlers receive the path with the mount prefix removed. `HandleHTTP`
parameters are available through `r.PathValue`, just as they are with Go's
standard `ServeMux`.
