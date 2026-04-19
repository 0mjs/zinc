---
slug: /cookbook/templ-ui
title: Server-Rendered UI with Templ
description: Render type-safe components from a Zinc handler, with a Templ UI–style Button and a Tailwind pipeline.
---

Zinc renders [Templ](https://templ.guide) components with one line — no framework glue. [Templ UI](https://templui.io) is a component library built on Templ; this recipe stands up the toolchain both want (Templ + Tailwind v4) and shows the integration with Zinc.

The Zinc side stays thin. The weight is tooling: `templ generate`, a Tailwind build, and a tiny `Render` helper that pairs `c.Context()` with `c.Writer()`.

## What you get

- a typed `@components.Button` invoked from a page template
- `POST /subscribe` that reads a form, re-renders the page with a flash message, and keeps the URL stable
- a Tailwind-built stylesheet served by `app.Static`
- a `Render(c, t)` helper you can drop into any Zinc project using Templ

## Setup

```bash
go mod init zinc-templ
go get github.com/0mjs/zinc
go get github.com/a-h/templ
go install github.com/a-h/templ/cmd/templ@latest
go install github.com/templui/templui/cmd/templui@latest
```

Grab the Tailwind v4.1+ standalone binary (no Node). Pick the [asset for your OS/arch](https://github.com/tailwindlabs/tailwindcss/releases/latest); macOS arm64 shown here:

```bash
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-macos-arm64
chmod +x tailwindcss-macos-arm64 && mv tailwindcss-macos-arm64 tailwindcss
```

## Project layout

```text
.
├── main.go
├── components
│   ├── button.go
│   └── button.templ
├── views
│   └── home.templ
├── public            # tailwind output goes here
└── input.css
```

## Tailwind configuration

Tailwind v4 dropped the JS config file. Configure sources from CSS:

```css title="input.css"
@import "tailwindcss";

@source "./views/**/*.templ";
@source "./components/**/*.templ";
```

## Components

```go title="components/button.go"
package components

func buttonClasses(variant string) string {
	base := "inline-flex items-center justify-center rounded-md px-4 py-2 text-sm font-medium transition focus:outline-none focus:ring-2 focus:ring-offset-2"
	switch variant {
	case "primary":
		return base + " bg-slate-900 text-white hover:bg-slate-800 focus:ring-slate-500"
	case "secondary":
		return base + " border border-slate-300 bg-white text-slate-900 hover:bg-slate-100 focus:ring-slate-400"
	default:
		return base + " bg-slate-900 text-white"
	}
}
```

```templ title="components/button.templ"
package components

type ButtonProps struct {
	Text    string
	Type    string
	Variant string
}

templ Button(props ButtonProps) {
	<button type={ props.Type } class={ buttonClasses(props.Variant) }>
		{ props.Text }
	</button>
}
```

## Page template

```templ title="views/home.templ"
package views

import "zinc-templ/components"

templ Home(flash string) {
	<!DOCTYPE html>
	<html lang="en">
		<head>
			<meta charset="utf-8"/>
			<meta name="viewport" content="width=device-width, initial-scale=1"/>
			<title>Zinc + Templ</title>
			<link rel="stylesheet" href="/static/app.css"/>
		</head>
		<body class="min-h-screen bg-slate-50 text-slate-900">
			<main class="mx-auto max-w-lg p-10">
				<h1 class="text-2xl font-semibold">Subscribe</h1>
				<p class="mt-1 text-sm text-slate-600">Server-rendered with Templ. Styled with Tailwind. Served by Zinc.</p>
				if flash != "" {
					<p class="mt-4 rounded-md bg-emerald-50 px-3 py-2 text-sm text-emerald-800">{ flash }</p>
				}
				<form method="post" action="/subscribe" class="mt-6 space-y-3">
					<input
						name="email"
						type="email"
						required
						placeholder="you@example.com"
						class="w-full rounded-md border border-slate-300 px-3 py-2 focus:border-slate-500 focus:outline-none focus:ring-1 focus:ring-slate-500"/>
					@components.Button(components.ButtonProps{
						Text:    "Subscribe",
						Type:    "submit",
						Variant: "primary",
					})
				</form>
			</main>
		</body>
	</html>
}
```

## Application

```go title="main.go"
package main

import (
	"log"

	"github.com/0mjs/zinc"
	"github.com/a-h/templ"

	"zinc-templ/views"
)

func render(c *zinc.Context, t templ.Component) error {
	c.Type("html")
	return t.Render(c.Context(), c.Writer())
}

func main() {
	app := zinc.New()

	if err := app.Static("/static", "./public"); err != nil {
		log.Fatal(err)
	}

	app.Get("/", func(c *zinc.Context) error {
		return render(c, views.Home(""))
	})

	app.Post("/subscribe", func(c *zinc.Context) error {
		if err := c.Request().ParseForm(); err != nil {
			return err
		}
		email := c.Request().FormValue("email")
		if email == "" {
			return zinc.ErrBadRequest.WithMessage("email is required")
		}
		return render(c, views.Home("Subscribed "+email+"."))
	})

	app.Listen()
}
```

## Dev loop

Three watchers, one per concern. Run each in its own terminal (or wrap with `air`, `overmind`, `make -j3`, whatever you already use):

```bash
templ generate --watch
./tailwindcss -i ./input.css -o ./public/app.css --watch
go run .
```

Visit `http://localhost:8080`, submit the form, and the same page re-renders with a flash message — no JSON, no client-side framework, no reload-on-navigate.

## Using real Templ UI components

The `Button` above is structured like Templ UI's components — a `Props` struct, a named component, Tailwind classes from a helper — so swapping it for the real one is a file replacement, not a refactor.

Initialize your project for Templ UI and copy components into your tree:

```bash
templui init
templui add button input card
```

That drops real components into your project (Templ UI is copy-into-your-tree, like shadcn/ui). Import them from your views and call them the same way:

```templ
import "zinc-templ/components/button"

@button.Button(button.Props{ ... })
```

The Zinc handler doesn't change. `render(c, t)` renders any `templ.Component`.

## Notes

- `c.Context()` is the request context — pass it to `Render` so cancellations and per-request values propagate into components that read from `ctx`.
- `c.Writer()` is the raw `http.ResponseWriter`, so streaming, flushing, and chunked output work through Templ unchanged.
- Set `Content-Type` with `c.Type("html")` before calling `Render`; once Templ writes a byte, the header is committed.
- For HTMX-style partial updates, render a smaller component from a handler that matches the target — the same `render` helper serves full pages and fragments.
- For production, run `templ generate` and `./tailwindcss -i input.css -o public/app.css --minify` in your build step. The Go binary then embeds `./public` via `embed.FS` + `app.StaticFS` if you want a single-file deploy.
