# zinc

![Zinc](https://img.shields.io/badge/Zinc-%20Go%20API%20Framework-silver)
![Version](https://img.shields.io/badge/version-0.0.7-red)
![Go Version](https://img.shields.io/badge/Go-1.24+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Zinc is an API framework for Go built on top of `net/http`. It keeps the server and transport model familiar,
while adding practical routing, middleware, binding, response helpers, and explicit lifecycle APIs.

## Features

- `App` is an `http.Handler`
- Express-style routes with `:param` and `*wildcard`
- Route groups, prefix middleware, and route metadata
- Binding helpers for path, query, headers, JSON, XML, and forms
- Response helpers for JSON, XML, HTML, streams, redirects, files, and rendering
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

	app.Get("/", func(c *zinc.Context) error {
		return c.String("Hello, Zinc!")
	})

	app.Get("/hello", "hello from zinc")

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

`WS` uses `github.com/gorilla/websocket`.

```go
app.SSE("/events", func(c *zinc.Context, stream *zinc.SSEStream) error {
	return stream.Send(zinc.SSEEvent{
		Event: "ready",
		Data:  zinc.Map{"ok": true},
	})
})

app.WS("/ws", func(c *zinc.Context, conn *websocket.Conn) error {
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

## Documentation

- [pkg.go.dev](https://pkg.go.dev/github.com/0mjs/zinc)
- [Docs app source](./docs)
- [Benchmarks](./BENCHMARKS.md)

## Optional Middleware

Zinc ships focused middleware packages outside the core, such as:

- `github.com/0mjs/zinc/middleware/cors`

## License

MIT
