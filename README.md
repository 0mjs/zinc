# zinc

![Zinc](https://img.shields.io/badge/Zinc-%20A%20web%20framework%20for%20Go-silver)
![Version](https://img.shields.io/badge/version-0.0.7-red)
![Go Version](https://img.shields.io/badge/Go-1.24+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

Zinc is a focused web framework for Go built on top of `net/http`. It keeps the transport and server model familiar,
while adding a fast router, a compact request context, binding helpers, response helpers, and explicit lifecycle APIs.

## Features

- `App` is an `http.Handler`
- Express-style routes with `:param` and `*wildcard`
- Route groups, prefix middleware, and route metadata
- Binding helpers for path, query, headers, JSON, XML, and forms
- Response helpers for JSON, XML, HTML, streams, redirects, files, and rendering
- Built-in real-time helpers for Server-Sent Events (SSE) and WebSockets
- Static/file serving and stdlib handler interop through `Mount`, `Wrap`, and `WrapFunc`
- Explicit startup and shutdown with `Listen`, `Serve`, and `Shutdown`

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
func requestLogger(c *zinc.Context) error {
	log.Printf("%s %s", c.Method(), c.Path())
	return c.Next()
}

app.Use(requestLogger)
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

## Real-Time (SSE + WebSocket)

```go
package main

import (
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"

	"github.com/0mjs/zinc"
)

func main() {
	app := zinc.New()

	app.SSE("/events", func(c *zinc.Context, stream *zinc.SSEStream) error {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		count := 0
		for {
			select {
			case <-stream.Done():
				return nil
			case tick := <-ticker.C:
				count++
				if err := stream.Send(zinc.SSEEvent{
					ID:    strconv.Itoa(count),
					Event: "tick",
					Data: zinc.Map{
						"at": tick.UTC().Format(time.RFC3339),
					},
				}); err != nil {
					return err
				}
			}
		}
	})

	app.WS("/ws", func(c *zinc.Context, conn *websocket.Conn) error {
		for {
			messageType, payload, err := conn.ReadMessage()
			if err != nil {
				return nil
			}
			if err := conn.WriteMessage(messageType, payload); err != nil {
				return err
			}
		}
	}, zinc.WebSocketConfig{
		CheckOrigin: func(c *zinc.Context) bool {
			return c.GetHeader(zinc.HeaderOrigin) == "http://localhost:3000"
		},
	})

	log.Fatal(app.Listen(":8080"))
}
```

Browser client examples (plain HTML, React, Next.js, and Vue can all use these browser APIs):

```html
<script>
  const events = new EventSource("/events");
  events.addEventListener("tick", (event) => {
    const data = JSON.parse(event.data);
    console.log("sse tick:", data);
  });

  const ws = new WebSocket("ws://localhost:8080/ws");
  ws.onopen = () => ws.send("hello");
  ws.onmessage = (event) => console.log("ws message:", event.data);
</script>
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
