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

Latest local snapshot (`Apple M1 Pro`, `darwin/arm64`, averaged over `count=3`):

- This view intentionally compares full frameworks only: `Zinc`, `Gin`, `Echo`, and `Chi`.
- Loopback `req/s` is left out here on purpose; the cleaner signal for docs is in-process request-path cost.
- Zinc wins `25/43` non-throughput rows in the current framework-only suite.
- Zinc still leads static dispatch and most API-path work.
- Gin's current latency edges are broader than the previous snapshot, covering more cold `404` / `405` paths plus the heavier `ParseAPI26` / `GPlusAPI13` param-routing cases.
- Build-time leadership is now mixed: Zinc still wins param registration, but Gin is slightly ahead on static registration.

Framework-only scorecard:

| Framework | Wins | 2nd Place |
|---|---:|---:|
| `Zinc` | `25` | `15` |
| `Gin` | `17` | `20` |
| `Echo` | `1` | `4` |
| `Chi` | `0` | `4` |

Request-path highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `65.7` | `84.3` | `123.6` | `170.9` | 🥇 Zinc |
| `StaticRoute` | `63.8` | `83.9` | `124.2` | `166.3` | 🥇 Zinc |
| `RouterParam` | `92.8` | `93.0` | `134.6` | `315.9` | 🥇 Zinc |
| `JSONResponse` | `326.3` | `388.8` | `389.7` | `491.9` | 🥇 Zinc |
| `MiddlewareChain` | `354.1` | `385.6` | `485.8` | `810.6` | 🥇 Zinc |
| `LargeRouteSetStaticMixed` | `66.9` | `114.7` | `151.7` | `234.3` | 🥇 Zinc |
| `LargeRouteSetMethodMismatch` | `153.9` | `100.8` | `869.2` | `304.7` | 🥇 Gin |
| `APIParamQueryJSON` | `873.7` | `2742.0` | `1814.0` | `1004.6` | 🥇 Zinc |
| `APIHappyPath` | `1292.3` | `3073.3` | `2221.7` | `1696.7` | 🥇 Zinc |
| `APIBindJSONHappyPath` | `2689.7` | `4463.7` | `2569.7` | `2906.3` | 🥇 Echo |

Scenario highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/GitHubAPI203` | `70.5` | `97.6` | `147.4` | `205.7` | 🥇 Zinc |
| `ScenarioStatic/ParseAPI26` | `67.0` | `93.9` | `145.3` | `196.3` | 🥇 Zinc |
| `ScenarioAll/Static157` | `70.7` | `124.7` | `169.0` | `257.0` | 🥇 Zinc |
| `ScenarioAll/GPlusAPI13` | `153.6` | `173.9` | `169.8` | `275.0` | 🥇 Zinc |
| `ScenarioParam/GitHubAPI203` | `152.8` | `155.2` | `220.9` | `373.5` | 🥇 Zinc |
| `ScenarioParam/ParseAPI26` | `219.1` | `116.4` | `164.7` | `255.3` | 🥇 Gin |
| `ScenarioNotFound/GitHubAPI203` | `113.6` | `120.8` | `589.0` | `366.7` | 🥇 Zinc |
| `Scenario405/GPlusAPI13` | `166.1` | `164.3` | `774.5` | `382.4` | 🥇 Gin |

Build-time highlights:

| Registration Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `80.7 µs / 113.7 KB` | `77.1 µs / 53.0 KB` | `338.2 µs / 181.6 KB` | `92.5 µs / 120.7 KB` | 🥇 Gin |
| `RouteRegistrationParam` | `51.1 µs / 86.8 KB` | `51.7 µs / 46.8 KB` | `160.1 µs / 152.1 KB` | `76.8 µs / 89.9 KB` | 🥇 Zinc |


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
