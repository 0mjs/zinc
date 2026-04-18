---
slug: /
id: intro
title: Welcome
description: Build fast, explicit Go APIs on net/http with Zinc.
sidebar_position: 1
---

Zinc is a Go API framework for developers who want expressive routing and practical helpers without leaving `net/http`.

It gives you route groups, middleware chains, request binding, response helpers, static files, rendering, route metadata, and first-party middleware while keeping normal Go deployment and testing habits.

```bash
go get github.com/0mjs/zinc
```

```go
package main

import (
	"log"

	"github.com/0mjs/zinc"
)

func main() {
	app := zinc.New()

	app.Get("/", func(c *zinc.Context) error {
		return c.String("Hello, Zinc!")
	})

	log.Fatal(app.Listen(":8080"))
}
```

<div className="zinc-doc-cards">
  <a className="zinc-doc-card" href="/getting-started/quick-start">
    <strong>Start building</strong>
    Install Zinc, run a tiny server, and add your first JSON endpoint.
  </a>
  <a className="zinc-doc-card" href="/guide/routing">
    <strong>Learn the guide</strong>
    Work through routing, middleware, context, binding, responses, and errors.
  </a>
  <a className="zinc-doc-card" href="/middleware/overview">
    <strong>Pick middleware</strong>
    Add CORS, auth, logging, recovery, metrics, rate limiting, static files, and more.
  </a>
</div>

## What Zinc is good at

- **API ergonomics:** handlers and middleware use one `func(*zinc.Context) error` shape.
- **Standard library compatibility:** mount `http.Handler` values and keep normal `net/http` server behavior.
- **Fast request paths:** Zinc is built around a compact router and a pooled request context.
- **Practical defaults:** common response formats, request binding, error handling, static files, and middleware are first-party.

## Mental model

A Zinc app is a request pipeline:

1. app-level middleware runs
2. prefix and group middleware run when their route matches
3. the route handler writes a response or returns an error
4. Zinc's error handler turns returned errors into responses

```go
app.Post("/posts", requireUser, requireRole("editor"), createPost)
```

Each item in that chain can call `c.Next()` to continue, return a response to stop, or return an error for the app error handler.

## Where to go next

- [Installation](./getting-started/installation) covers module setup and Go version requirements.
- [Quick Start](./getting-started/quick-start) builds a small API in one file.
- [First Route](./getting-started/first-route) explains params, query values, JSON, and errors in one handler.
- [Routing](./guide/routing) is the first full guide page once you are ready for groups and named routes.
- [API Reference](./api/app) keeps the method-level details separate from the learning path.
