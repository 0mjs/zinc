---
title: Templates
description: Render html/template, text/template, or a custom template engine from Zinc handlers.
---

Configure a renderer once, then call `c.Render` from handlers.

```go
package main

import (
    "html/template"
    "log"

    "github.com/0mjs/zinc"
)

func main() {
    views := template.Must(template.ParseGlob("views/*.html"))

    cfg := zinc.DefaultConfig
    cfg.Renderer = zinc.NewHTMLTemplateRenderer(views)
    app := zinc.NewWithConfig(cfg)

    app.Get("/", func(c *zinc.Context) error {
        return c.Render("home.html", zinc.Map{"Title": "Zinc"})
    })

    log.Fatal(app.Listen())
}
```

`NewTextTemplateRenderer` supports `text/template`. `NewTemplateRenderer` accepts any type implementing Zinc's `TemplateExecutor` interface.

## Template suffixes

Some engines register names without file suffixes. Configure fallback suffixes when required:

```go
renderer := zinc.NewHTMLTemplateRenderer(
    views,
    zinc.WithTemplateSuffixes(".html", ".tmpl"),
)
```

For Templ components, call the generated component's `Render` method with `c.Context()` and `c.Writer()`. See the [Templ UI recipe](/cookbook/templ-ui/).
