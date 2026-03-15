![Version](https://img.shields.io/badge/version-0.0.84-blue)
![Go Version](https://img.shields.io/badge/Go-1.25+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-90.6%25-brightgreen)](#quality-snapshot)
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
Overall                                      [##################..] 90.6%
Core (github.com/0mjs/zinc)                  [###################.] 95.1%
Middleware (github.com/0mjs/zinc/middleware) [#################...] 83.9%
```

## Benchmark Snapshot

Latest local snapshot (`Apple M1 Pro`, `darwin/arm64`, sequential `count=1` reruns on `2026-03-13`):

- This view intentionally compares full frameworks only: `Zinc`, `Gin`, `Echo`, and `Chi`.
- Zinc wins `45/75` framework-only rows overall and `42/67` non-throughput rows in the current rerun.
- Full framework-only tables live in [BENCKMARKS.md](/Users/matt/dev/oss/zinc/BENCKMARKS.md).

Framework-only scorecard:

| Framework | Wins | 2nd Place | Top-3 |
|---|---:|---:|---:|
| `Zinc` | `45` | `25` | `73` |
| `Gin` | `27` | `35` | `69` |
| `Echo` | `1` | `10` | `51` |
| `Chi` | `2` | `5` | `32` |

Request-path highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `72.7` | `94.3` | `132.2` | `193.0` | 🥇 Zinc |
| `StaticRoute` | `71.6` | `108.9` | `153.1` | `224.2` | 🥇 Zinc |
| `RouterParamCold` | `102.5` | `105.7` | `145.5` | `278.5` | 🥇 Zinc |
| `LargeRouteSetParam` | `118.3` | `134.7` | `159.1` | `428.8` | 🥇 Zinc |
| `LargeRouteSetParamMixed` | `128.7` | `138.4` | `184.1` | `317.9` | 🥇 Zinc |
| `NotFound` | `150.2` | `103.1` | `883.6` | `695.5` | 🥇 Gin |
| `APIParamQueryJSON` | `914.9` | `2836.0` | `1865.0` | `1236.0` | 🥇 Zinc |
| `APIHappyPath` | `1281.0` | `3146.0` | `2230.0` | `1826.0` | 🥇 Zinc |
| `APIBindJSONHappyPath` | `2666.0` | `4367.0` | `2527.0` | `2993.0` | 🥇 Echo |

Scenario highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/GitHubAPI203` | `68.4` | `104.0` | `146.6` | `244.9` | 🥇 Zinc |
| `ScenarioParam/GitHubAPI203` | `173.2` | `270.4` | `325.8` | `453.9` | 🥇 Zinc |
| `ScenarioAll/ParseAPI26` | `121.1` | `158.4` | `164.5` | `333.6` | 🥇 Zinc |
| `ScenarioNotFound/GitHubAPI203` | `119.7` | `132.2` | `730.7` | `400.2` | 🥇 Zinc |
| `ScenarioNotFound/NestedAPI36` | `119.7` | `133.7` | `703.5` | `386.9` | 🥇 Zinc |
| `Scenario405/NestedAPI36` | `160.8` | `223.8` | `1293.0` | `1507.0` | 🥇 Zinc |
| `Scenario405/GPlusAPI13` | `201.4` | `200.2` | `980.3` | `460.9` | 🥇 Gin |

Build-time highlights:

| Registration Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `81.1 µs / 113.8 KB` | `78.4 µs / 53.0 KB` | `363.0 µs / 181.6 KB` | `113.8 µs / 120.7 KB` | 🥇 Gin |
| `RouteRegistrationParam` | `53.2 µs / 86.9 KB` | `55.1 µs / 46.8 KB` | `179.4 µs / 152.1 KB` | `74.1 µs / 89.9 KB` | 🥇 Zinc |


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
