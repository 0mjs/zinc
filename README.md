![Version](https://img.shields.io/badge/version-0.0.87-blue)
![Go Version](https://img.shields.io/badge/Go-1.25+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-83.4%25-brightgreen)](#quality-snapshot)
[![Go%20Report%20Card](https://goreportcard.com/badge/github.com/0mjs/zinc)](https://goreportcard.com/report/github.com/0mjs/zinc)
![License](https://img.shields.io/badge/license-MIT-green)

# Zinc

Zinc is an Express-inspired, idiomatic Go API framework built on `net/http`.

- [Documentation](https://zinc.carbonsoft.com)
- [Quickstart](https://zinc.carbonsoft.sh/guide/quick-start)
- [Middleware](https://zinc.carbonsoft.sh/middleware)
- [pkg.go.dev](https://pkg.go.dev/github.com/0mjs/zinc)

### Features

- Express-style routes with `:param` and `*wildcard`
- Route groups, prefix middleware, and route metadata
- Binding helpers for path, query, headers, JSON, XML, and forms
- Response helpers for JSON, XML, HTML, streams, redirects, files, and rendering
- First-party template renderer helper for `html/template` and `text/template`
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
	middleware "github.com/0mjs/zinc/middleware"
)

func main() {
	app := zinc.New()

	app.Use(middleware.RequestLogger())

	app.Get("/", "Hello, world!") // Shorthand

	app.Get("/greet", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"greeting": "Hello, world!",
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
Overall                                      [#################...] 83.4%
Core (github.com/0mjs/zinc)                  [#################...] 83.3%
Middleware (github.com/0mjs/zinc/middleware) [#################...] 83.9%
```

## Benchmark Snapshot

Latest local peer-only snapshot (`Apple M1 Pro`, `darwin/arm64`, `count=1` rerun on `2026-03-18`):

- Zinc wins `65/85` rows overall against `Gin`, `Echo`, and `Chi`.
- Excluding throughput, Zinc wins `65/77` rows.
- Full tables and remaining gaps live in [BENCKMARKS.md](/Users/matt/dev/oss/zinc/BENCKMARKS.md).

Selected highlights:

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `61.82 ns` | `88.21 ns` | `127.2 ns` | `178.8 ns` | 🥇 Zinc |
| `APIHappyPath` | `1.25 µs` | `3.02 µs` | `2.19 µs` | `1.66 µs` | 🥇 Zinc |
| `APIBindJSONHappyPath` | `2.41 µs` | `4.41 µs` | `2.52 µs` | `2.81 µs` | 🥇 Zinc |
| `StaticFileHit` | `15.64 µs` | `32.85 µs` | `26.95 µs` | `15.97 µs` | 🥇 Zinc |
| `RouteRegistrationStatic` | `57.88 µs` | `76.33 µs` | `337.4 µs` | `85.04 µs` | 🥇 Zinc |
| `ScenarioAll/ParseAPI26` | `118.2 ns` | `120.9 ns` | `155.4 ns` | `370.9 ns` | 🥇 Zinc |

Throughput is currently the weakest category in the peer-only suite; the detailed breakdown is in [BENCKMARKS.md](/Users/matt/dev/oss/zinc/BENCKMARKS.md).


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

[MIT](./LICENSE)
