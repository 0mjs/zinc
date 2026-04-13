---
slug: /
id: intro
title: 👋 Welcome
description: Fast, explicit, net/http-native API development for Go.
sidebar_position: 1
---

Zinc is a Fiber-style Go API framework built on top of `net/http`.

It focuses on:

- Fiber-like routing and handler ergonomics
- explicit middleware composition
- practical binding and response helpers
- compatibility with the standard library
- strong request-path performance without forcing a non-`net/http` runtime model

If you like Fiber's developer experience but want to stay on `net/http`, Zinc is built for that shape of application.

## Installation

```bash
go get github.com/0mjs/zinc
```

## Hello, World

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

## Mental Model

Zinc handlers and middleware share one signature:

```go
func(*zinc.Context) error
```

That means route-level guards, group middleware, and final handlers compose in the same chain.

```go
app.Post("/posts", authUser, requireRole("editor"), createPost)
```

Each middleware can call `return c.Next()` to continue, return a response to stop, or return an error for the app error handler.

## Why Zinc

- **`net/http` native**: mount stdlib handlers, wrap existing middleware, and keep normal Go deployment patterns.
- **Fiber-style routing**: named params, wildcards, groups, prefix middleware, and route-scoped `404` handling.
- **API-first binding**: bind path, query, headers, forms, multipart files, JSON, XML, YAML, TOML, and plain text.
- **Practical responses**: JSON, XML, YAML, TOML, HTML, streams, files, redirects, downloads, and renderer-backed templates.
- **Tooling-friendly routing**: named routes, reverse URL generation, route introspection, and regex-constrained params.

## Start Here

- [Routing](./guide/routing) for route shapes, params, groups, and route naming
- [Binding](./guide/binding) for request decoding and validation patterns
- [Responses and rendering](./guide/responses-and-rendering) for response helpers
- [Errors](./guide/errors) for `HTTPError`, abort helpers, and custom error handling
- [App API](./api/app) for the primary app lifecycle and route APIs

## Current Benchmark Snapshot

As of the latest peer-only suite, Zinc wins `65/85` rows overall and `65/77` non-throughput rows against `Gin`, `Echo`, and `Chi`.

The full benchmark tables live in the repository benchmark report:

- [BENCKMARKS.md](https://github.com/0mjs/zinc/blob/main/BENCKMARKS.md)

## Docs Philosophy

These docs are written to be explicit, operational, and copy-paste friendly:

- guide pages explain how to use Zinc in real applications
- API pages explain the most important public types and behaviors
- benchmark and FAQ pages keep the “why Zinc?” story honest and easy to inspect
