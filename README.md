# zinc

![Version](https://img.shields.io/badge/version-0.0.78-red)
![Go Version](https://img.shields.io/badge/Go-1.26+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-90.6%25-brightgreen)](#quality-snapshot)
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
Overall                                      [##################..] 90.6%
Core (github.com/0mjs/zinc)                  [###################.] 95.1%
Middleware (github.com/0mjs/zinc/middleware) [#################...] 83.9%
```

## Benchmark Snapshot

Benchmark suites now live in [`./benchmarks`](./benchmarks), so Zinc's main module does not need Gin, Echo, Chi, or HttpRouter as direct dependencies.

Latest local snapshot (`Apple M1 Pro`, `darwin/arm64`, lower is better for `ns/op`):

- Zinc wins `9/11` request-path rows in the broader in-process suite against `Gin`, `Echo`, and `Chi`.
- Zinc is still `0 allocs/op` on the main routing hot paths.
- Static dispatch is a real strength: Zinc leads the scenario static sweeps and large static route-set benchmarks.
- Route registration memory is no longer a major outlier: Zinc is now in the same class as Gin instead of multiple times larger.
- Focused loopback RPS is competitive, but Zinc is not consistently first there.

High-signal examples:

| Benchmark | Zinc | Gin | Echo | Chi | Read |
|---|---:|---:|---:|---:|---|
| `Static157 All` | `71.8 ns` | `133.7 ns` | `173.9 ns` | `274.1 ns` | Zinc win |
| `GitHubAPI203 All` | `162.1 ns` | `156.6 ns` | `202.8 ns` | `365.1 ns` | Near Gin, ahead of Echo/Chi |
| `HelloWorld` | `67.1 ns` | `93.4 ns` | `131.0 ns` | `172.9 ns` | Zinc win |
| `RouterParam` | `87.5 ns` | `97.9 ns` | `131.4 ns` | `331.5 ns` | Zinc win |
| `LargeRouteSetStatic` | `66.1 ns` | `100.8 ns` | `146.8 ns` | `208.9 ns` | Zinc win |
| `LargeRouteSetParam` | `108.1 ns` | `107.2 ns` | `154.1 ns` | `381.4 ns` | Effectively tied with Gin |

Build-time highlights:

| Registration Benchmark | Zinc | Gin | Echo | Chi |
|---|---:|---:|---:|---:|
| `RouteRegistrationStatic` | `68.4 µs / 62.5 KB` | `76.7 µs / 54.3 KB` | `401.4 µs / 185.9 KB` | `113.3 µs / 123.6 KB` |
| `RouteRegistrationParam` | `49.0 µs / 67.8 KB` | `134.6 µs / 47.9 KB` | `157.5 µs / 155.7 KB` | `73.0 µs / 92.1 KB` |

Focused loopback RPS average (`count=3`):

| Framework | Req/s |
|---|---:|
| `Chi` | `88.9k` |
| `Gin` | `88.5k` |
| `Zinc` | `82.6k` |
| `Echo` | `80.5k` |

The short version: Zinc already has a strong benchmark story on real request dispatch, especially for static and mixed route sets. The remaining obvious runtime gap is dedicated param-heavy paths against Gin, not broad request-path throughput.

## Documentation

- [pkg.go.dev](https://pkg.go.dev/github.com/0mjs/zinc)
- [Docs app source](./docs)
- [Benchmark module](./benchmarks)
- [Benchmark summary](_docs/BENCHMARK_SUMMARY.md)

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
