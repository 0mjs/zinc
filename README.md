![Version](https://img.shields.io/badge/version-0.0.84-blue)
![Go Version](https://img.shields.io/badge/Go-1.25+-blue)
[![Docs](https://pkg.go.dev/badge/github.com/0mjs/zinc.svg)](https://pkg.go.dev/github.com/0mjs/zinc)
[![Coverage](https://img.shields.io/badge/coverage-90.6%25-brightgreen)](#quality-snapshot)
[![Go%20Report%20Card](https://goreportcard.com/badge/github.com/0mjs/zinc)](https://goreportcard.com/report/github.com/0mjs/zinc)
![License](https://img.shields.io/badge/license-MIT-green)

# Zinc

Zinc is an fast, minimal API framework for Go built on top of `net/http`.

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
- Zinc wins `31/43` non-throughput rows in the current framework-only suite.
- Zinc now leads static dispatch, API-path work, and route/build-time benchmarks.
- Gin's remaining latency edges are concentrated in a smaller set of `404` / `405` paths plus the heavier `ParseAPI26` / `GPlusAPI13` param-routing cases.

Framework-only scorecard:

| Framework | Wins | 2nd Place |
|---|---:|---:|
| `Zinc` | `31` | `10` |
| `Gin` | `11` | `26` |
| `Echo` | `1` | `3` |
| `Chi` | `0` | `4` |

Request-path highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `66.5` | `89.2` | `128.9` | `193.7` | 🥇 Zinc |
| `StaticRoute` | `66.9` | `90.4` | `129.6` | `174.5` | 🥇 Zinc |
| `RouterParam` | `83.9` | `120.5` | `150.4` | `345.2` | 🥇 Zinc |
| `JSONResponse` | `331.7` | `422.3` | `398.4` | `510.1` | 🥇 Zinc |
| `MiddlewareChain` | `379.2` | `443.9` | `543.5` | `871.4` | 🥇 Zinc |
| `LargeRouteSetStaticMixed` | `70.2` | `122.8` | `168.5` | `258.4` | 🥇 Zinc |
| `LargeRouteSetMethodMismatch` | `319.6` | `134.0` | `1316.2` | `438.2` | 🥇 Gin |
| `APIParamQueryJSON` | `933.9` | `2708.3` | `1831.3` | `953.1` | 🥇 Zinc |
| `APIHappyPath` | `1336.7` | `3123.7` | `2348.0` | `1684.0` | 🥇 Zinc |
| `APIBindJSONHappyPath` | `2631.7` | `4525.3` | `2592.0` | `2812.7` | 🥇 Echo |

Scenario highlights (`ns/op`, lower is better):

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/GitHubAPI203` | `80.2` | `105.0` | `149.6` | `272.6` | 🥇 Zinc |
| `ScenarioStatic/ParseAPI26` | `67.9` | `101.2` | `146.1` | `229.9` | 🥇 Zinc |
| `ScenarioAll/Static157` | `74.9` | `137.5` | `179.9` | `302.3` | 🥇 Zinc |
| `ScenarioAll/GPlusAPI13` | `147.4` | `192.5` | `208.7` | `304.7` | 🥇 Zinc |
| `ScenarioParam/GitHubAPI203` | `158.9` | `162.5` | `251.0` | `412.9` | 🥇 Zinc |
| `ScenarioParam/ParseAPI26` | `217.3` | `122.0` | `166.0` | `291.3` | 🥇 Gin |
| `ScenarioNotFound/GitHubAPI203` | `118.2` | `127.2` | `656.6` | `379.9` | 🥇 Zinc |
| `Scenario405/GPlusAPI13` | `174.6` | `206.7` | `1032.2` | `446.2` | 🥇 Zinc |

Build-time highlights:

| Registration Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `68.8 µs / 87.5 KB` | `79.4 µs / 53.0 KB` | `350.7 µs / 181.6 KB` | `87.7 µs / 120.7 KB` | 🥇 Zinc |
| `RouteRegistrationParam` | `46.0 µs / 70.2 KB` | `54.3 µs / 46.8 KB` | `162.4 µs / 152.1 KB` | `76.6 µs / 89.9 KB` | 🥇 Zinc |


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
