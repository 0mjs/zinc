---
id: quick-start
title: Quick Start
description: Build and run a tiny Zinc API in one file.
sidebar_position: 2
---

This page builds a small API with a health route, a JSON route, and a grouped endpoint.

## Create `main.go`

```go
package main

import (
	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/middleware"
)

func main() {
	app := zinc.New()

	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger())
	app.Use(middleware.Recover())

	app.Get("/", func(c *zinc.Context) error {
		return c.String("Hello, Zinc!")
	})

	app.Get("/health", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{"status": "ok"})
	})

	api := app.Group("/api")
	api.Get("/users/:id", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"id": c.Param("id"),
		})
	})

	app.Listen()
}
```

## Run it

```bash
go run .
```

Try the routes:

```bash
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/users/42
```

## What happened

- `zinc.New()` created the app.
- `app.Use(...)` added middleware that runs before matched routes.
- `app.Get(...)` registered handlers for `GET` requests.
- `app.Group("/api")` created a route group with a shared prefix.
- `c.Param("id")` read the `:id` route parameter.
- `c.JSON(...)` encoded and wrote a JSON response.
- `app.Listen()` started the server on `:8080`.

Continue with [First Route](./first-route) for a closer look at handlers.
