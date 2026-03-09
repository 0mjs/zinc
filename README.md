![Version](https://img.shields.io/badge/version-0.0.83-blue)
![Go Version](https://img.shields.io/badge/Go-1.26+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-90.6%25-brightgreen)](#quality-snapshot)
[![Go%20Report%20Card](https://goreportcard.com/badge/github.com/0mjs/zinc)](https://goreportcard.com/report/github.com/0mjs/zinc)
![License](https://img.shields.io/badge/license-MIT-green)

# Zinc

Zinc is an fast, minimal API framework for Go built on top of `net/http`.

- [Documentation](https://zinc.carbonsoft.com)
- [Quickstart](https://zinc.carbonsoft.sh/guide/quick-start)
- [Middleware](https://zinc.carbonsoft.sh/middleware)
- [pkg.go.dev]([pkg.go.dev](https://pkg.go.dev/github.com/0mjs/zinc))

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
- Zinc wins `23/43` non-throughput rows in the current framework-only suite.
- Zinc is strongest on static and mixed dispatch, middleware, JSON response, and scenario static sweeps.
- Gin still leads most `404` / `405` paths and the nastiest param-heavy scenario corpus.

Framework-only scorecard:

| Framework | Wins | 2nd Place |
|---|---:|---:|
| `Zinc` | `23` | `16` |
| `Gin` | `18` | `19` |
| `Echo` | `1` | `4` |
| `Chi` | `1` | `4` |

Request-path highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `73.1` | `91.4` | `133.7` | `199.7` | 🥇 Zinc |
| `StaticRoute` | `69.4` | `98.5` | `134.0` | `191.1` | 🥇 Zinc |
| `RouterParam` | `87.6` | `95.7` | `137.4` | `368.7` | 🥇 Zinc |
| `JSONResponse` | `338.0` | `400.4` | `447.0` | `561.5` | 🥇 Zinc |
| `MiddlewareChain` | `367.4` | `420.8` | `537.1` | `853.3` | 🥇 Zinc |
| `LargeRouteSetStaticMixed` | `71.3` | `123.0` | `162.9` | `260.6` | 🥇 Zinc |
| `LargeRouteSetMethodMismatch` | `153.8` | `101.3` | `856.3` | `328.2` | 🥇 Gin |
| `APIParamQueryJSON` | `1195.3` | `2741.0` | `1876.0` | `1095.3` | 🥇 Chi |
| `APIHappyPath` | `1487.7` | `3069.3` | `2239.7` | `1727.0` | 🥇 Zinc |
| `APIBindJSONHappyPath` | `2774.0` | `4503.0` | `2570.7` | `2926.7` | 🥇 Echo |

Scenario highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/GitHubAPI203` | `68.9` | `103.3` | `148.3` | `242.3` | 🥇 Zinc |
| `ScenarioStatic/ParseAPI26` | `70.1` | `98.5` | `145.4` | `238.1` | 🥇 Zinc |
| `ScenarioAll/Static157` | `72.9` | `130.2` | `171.6` | `294.0` | 🥇 Zinc |
| `ScenarioAll/GPlusAPI13` | `143.6` | `180.6` | `173.7` | `280.5` | 🥇 Zinc |
| `ScenarioParam/GitHubAPI203` | `171.4` | `159.0` | `225.4` | `436.6` | 🥇 Gin |
| `ScenarioParam/ParseAPI26` | `217.9` | `118.1` | `158.6` | `279.5` | 🥇 Gin |
| `ScenarioNotFound/GitHubAPI203` | `161.4` | `118.8` | `611.0` | `358.8` | 🥇 Gin |
| `Scenario405/GPlusAPI13` | `170.3` | `182.7` | `866.8` | `424.2` | 🥇 Zinc |

Build-time highlights:

| Registration Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `84.3 µs / 62.5 KB` | `74.8 µs / 54.3 KB` | `338.3 µs / 185.9 KB` | `88.5 µs / 123.6 KB` | 🥇 Gin |
| `RouteRegistrationParam` | `64.1 µs / 67.8 KB` | `55.8 µs / 47.9 KB` | `161.8 µs / 155.7 KB` | `73.8 µs / 92.1 KB` | 🥇 Gin |


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
