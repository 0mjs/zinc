# zinc

![Version](https://img.shields.io/badge/version-0.0.78-red)
![Go Version](https://img.shields.io/badge/Go-1.26+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-94.39%25-brightgreen)](#quality-snapshot)
[![Go%20Report%20Card](https://goreportcard.com/badge/github.com/0mjs/zinc)](https://goreportcard.com/report/github.com/0mjs/zinc)
![License](https://img.shields.io/badge/license-MIT-green)

Zinc is an API framework for Go built on top of `net/http`. It keeps the server and transport model familiar,
while adding practical routing, middleware, binding, response helpers, and explicit lifecycle APIs.

## Features

- `App` is an `http.Handler`
- Express-style routes with `:param` and `*wildcard`
- Route groups, prefix middleware, and route metadata
- Binding helpers for path, query, headers, JSON, XML, and forms
- Response helpers for JSON, XML, HTML, streams, redirects, files, and rendering
- First-party template renderer helper for `html/template` and `text/template`
- Static/file serving and stdlib handler interop through `Mount`, `Wrap`, and `WrapFunc`
- Explicit startup and shutdown with `Listen`, `Serve`, and `Shutdown`
- Optional real-time endpoints with SSE and Gorilla WebSocket integration

## Typical Use Cases

- JSON APIs and backend services
- Internal tools and admin endpoints
- Public REST APIs with middleware and validation
- Real-time features (chat, activity feeds, live dashboards) when needed

## Installation

```bash
go get github.com/0mjs/zinc
```

## Quick Start

```go
package main

import (
	"log"

	"github.com/0mjs/zinc"
)

func main() {
	app := zinc.New()

	app.Get("/", "Hello, world!") // Shorthand

	app.Get("/greet", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"greeting": "Hello, world!",
		})
	})

	app.Get("/users/:id", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"id":        c.Param("id"),
			"full_path": c.FullPath(),
		})
	})

	api := app.Group("/api")
	api.Get("/health", func(c *zinc.Context) error {
		return c.String("ok")
	})

	log.Fatal(app.Listen(":8080"))
}
```

## Routing And Middleware

```go
app.Use(middleware.RequestLogger())
app.UsePrefix("/api", authMiddleware)

app.Route("/api", func(api *zinc.Group) {
	api.Get("/users/:id", showUser)
	api.Post("/users", createUser)
})
```

## Binding And Responses

```go
type CreateUserInput struct {
	TeamID int    `path:"teamID"`
	Page   int    `query:"page"`
	Name   string `json:"name"`
	Auth   string `header:"x-auth"`
}

app.Post("/teams/:teamID/users", func(c *zinc.Context) error {
	var input CreateUserInput
	if err := c.Bind(&input); err != nil {
		return err
	}

	return c.Status(zinc.StatusCreated).JSON(input)
})
```

## Optional Real-Time Endpoints (SSE + WebSocket)

`WS` exposes Zinc's `WebSocketConn` for common server-side handlers. If you need client dialers or advanced websocket helpers, you can still import `github.com/gorilla/websocket` directly.

```go
app.SSE("/events", func(c *zinc.Context, stream *zinc.SSEStream) error {
	return stream.Send(zinc.SSEEvent{
		Event: "ready",
		Data:  zinc.Map{"ok": true},
	})
})

app.WS("/ws", func(c *zinc.Context, conn *zinc.WebSocketConn) error {
	for {
		msgType, payload, err := conn.ReadMessage()
		if err != nil {
			return nil
		}
		if err := conn.WriteMessage(msgType, payload); err != nil {
			return err
		}
	}
}, zinc.WebSocketConfig{})
```

## Configuration

```go
app := zinc.NewWithConfig(zinc.Config{
	ServerHeader:           "zinc/example",
	CaseSensitive:          true,
	StrictRouting:          true,
	AutoHead:               true,
	AutoOptions:            true,
	HandleMethodNotAllowed: true,
	BodyLimit:              8 << 20,
	ProxyHeader:            zinc.HeaderXForwardedFor,
	TrustedProxies:         []string{"10.0.0.1"},
})
```

`Config` also lets you plug in a custom `Binder`, `Validator`, `Renderer`, `JSONCodec`, and `ErrorHandler`.

```go
views := template.Must(template.ParseGlob("templates/*.html"))

app := zinc.NewWithConfig(zinc.Config{
	Renderer: zinc.NewHTMLTemplateRenderer(
		views,
		zinc.WithTemplateSuffixes(".html", ".tmpl"),
	),
})

app.Get("/dashboard", func(c *zinc.Context) error {
	return c.Render("dashboard", zinc.Map{"Title": "Overview"})
})
```

## Quality Snapshot

Latest local coverage run (`go test -count=1 ./... -coverprofile=coverage.out`):

```text
Overall                                      [###################.] 94.39%
Core (github.com/0mjs/zinc)                  [###################.] 95.3%
Middleware (github.com/0mjs/zinc/middleware) [####################] 98.9%
```

## Benchmark Snapshot (vs Gin, Echo, Chi)

Source: [`BENCKMARKS.md`](./BENCKMARKS.md) (`ns/op`, lower is better, Apple M1 Pro snapshot).

Headline results:

- Zinc wins `3/8` benchmarks (`HelloWorld`, `StaticRoute`, `JSONResponse`).
- Gin wins `5/8` benchmarks; Zinc is runner-up in all five.
- Zinc is faster than Echo and Chi in all 8 listed benchmarks.

Wins chart:

```text
Gin   [#####...] 5
Zinc  [###.....] 3
Echo  [........] 0
Chi   [........] 0
```

Selected benchmark table:

| Benchmark | Zinc | Gin | Echo | Chi | Zinc vs Gin |
|---|---:|---:|---:|---:|---:|
| `HelloWorld` | `80.6` | `83.9` | `119.1` | `177.2` | `1.04x faster` |
| `StaticRoute` | `82.4` | `85.6` | `121.6` | `174.7` | `1.04x faster` |
| `RouterParam` | `115.6` | `91.7` | `134.0` | `315.8` | `1.26x slower` |
| `JSONResponse` | `336.9` | `377.4` | `414.2` | `492.7` | `1.12x faster` |
| `MiddlewareChain` | `385.6` | `373.9` | `512.8` | `848.4` | `1.03x slower` |

## Documentation

- [pkg.go.dev](https://pkg.go.dev/github.com/0mjs/zinc)
- [Docs app source](./docs)
- [Benchmarks](https://github.com/0mjs/zinc/blob/dev/BENCKMARKS.md)

## Optional Middleware

Zinc ships middleware behind one unified import:

- `github.com/0mjs/zinc/middleware`

```go
import (
	"log"
	"os"
	"time"

	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/middleware"
	jwt "github.com/golang-jwt/jwt/v5"
)

app.Use(middleware.BodyDump(func(c *zinc.Context, snapshot middleware.BodyDumpSnapshot) {
	log.Printf("%s %s -> %d", snapshot.Method, snapshot.Path, snapshot.Status)
}))

app.Use(middleware.CORS("https://app.example.com"))

admin := app.Group("/admin")
admin.Use(middleware.BodyLimit(256 * middleware.KB))
admin.Use(middleware.ContextTimeout(250 * time.Millisecond))

app.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
	ExposeHeader: zinc.HeaderXCSRFToken,
}))

app.Use(middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
	Validator: middleware.BasicAuthStatic("admin", os.Getenv("ADMIN_PASSWORD")),
}))

app.Use(middleware.JWTWithConfig(middleware.JWTConfig{
	KeyFunc: func(*zinc.Context, *jwt.Token) (any, error) {
		return []byte("secret"), nil
	},
}))
```

Zinc now exposes middleware through a single public package: `github.com/0mjs/zinc/middleware`.

## License

MIT
