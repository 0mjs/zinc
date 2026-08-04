---
title: Quick Start
description: Install Go, create a module, and run your first Zinc API.
---

This guide takes you from an empty folder to a running Zinc API in a few minutes.

## Before you start

Zinc requires **Go 1.25 or newer**.

If Go is not installed, download it from [go.dev/dl](https://go.dev/dl/) and follow the installer for your operating system. Then open a new terminal and confirm the installed version:

```bash
go version
```

The output should report Go `1.25` or newer. Upgrade Go before continuing if it reports an older version.

## Create a project

Create a folder for the application and initialize a Go module inside it:

```bash
mkdir hello-zinc
cd hello-zinc
go mod init example.com/hello-zinc
```

The module path identifies your project. Replace `example.com/hello-zinc` with your repository path when you have one.

## Add Zinc

Install Zinc through the normal Go module tooling:

```bash
go get github.com/0mjs/zinc
```

This adds Zinc to `go.mod` and records the selected version in `go.sum`. The first-party middleware package is included in the same module.

## Create `main.go`

Create a file named `main.go` with the following application:

```go
package main

import (
	"log"

	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/middleware"
)

func main() {
	app := zinc.New()

	app.Use(
		middleware.RequestLogger(),
		middleware.Recover(),
	)

	app.Get("/", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{
			"message": "Hello from Zinc!",
		})
	})

	log.Fatal(app.Listen(":8080"))
}
```

## Run the application

Start the server from the project folder:

```bash
go run .
```

Zinc is now listening at `http://localhost:8080`.

## Make a request

Open another terminal and call the route:

```bash
curl -i http://localhost:8080/
```

The response includes `HTTP/1.1 200 OK` and a JSON body:

```json
{"message":"Hello from Zinc!"}
```

Press `Ctrl+C` in the server terminal when you are finished.

## What you just built

- `zinc.New()` created an application that also satisfies `http.Handler`.
- `app.Use(...)` installed request logging and panic recovery.
- `app.Get(...)` registered a standard `GET /` route.
- Returning an error keeps response and middleware failures in one error flow.
- `c.JSON(...)` encoded the response and set its content type.
- `app.Listen(":8080")` started the standard-library HTTP server.

Zinc's primary static routing path dispatches this route with zero request-time heap allocations.

## Next steps

- [Routing](/guide/routing/) — parameters, wildcards, route groups, and matching rules.
- [Middleware](/middleware/overview/) — the complete first-party middleware catalogue.
- [HTTP interoperability](/guide/http-interoperability/) — use Zinc with standard `net/http` handlers and servers.
