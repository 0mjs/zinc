---
slug: /
id: intro
title: Welcome
description: Build fast, explicit Go APIs on net/http with Zinc.
sidebar_position: 1
hide_title: true
hide_table_of_contents: true
---

import DocCards from "@site/src/components/DocCards";
import DocsHomeHero from "@site/src/components/DocsHomeHero";

<div className="zinc-home">

<DocsHomeHero />

<div className="zinc-section-heading">
  <span>Start here</span>
  <h2>Find the shortest path to what you need.</h2>
</div>

<DocCards
  cards={[
    {
      to: "/getting-started/quick-start",
      title: "Start building",
      description:
        "Install Zinc, run a tiny server, and add your first JSON endpoint.",
    },
    {
      to: "/guide/routing",
      title: "Learn the guide",
      description:
        "Work through routing, middleware, context, binding, responses, and errors.",
    },
    {
      to: "/middleware/overview",
      title: "Pick middleware",
      description:
        "Add CORS, auth, logging, recovery, metrics, rate limiting, static files, and more.",
    },
    {
      to: "/api/app",
      title: "Look something up",
      description:
        "Jump into focused reference pages for the app, context, config, and errors.",
    },
  ]}
/>

<div className="zinc-home-prose">

## What Zinc gives you

Zinc adds route groups, middleware chains, request binding, response helpers, static files, rendering, route metadata, and first-party middleware while keeping normal Go deployment and testing habits.

- **API ergonomics:** handlers and middleware use one `func(*zinc.Context) error` shape.
- **Standard library compatibility:** mount `http.Handler` values and keep normal `net/http` server behavior.
- **Fast request paths:** Zinc is built around a compact router and a pooled request context.
- **Practical defaults:** common response formats, request binding, error handling, static files, and middleware are first-party.

## Mental model

A Zinc app is a request pipeline:

1. app-level middleware runs
2. prefix and group middleware run when their route matches
3. the route handler writes a response or returns an error
4. Zinc's error handler turns returned errors into responses

```go
app.Post("/posts", requireUser, requireRole("editor"), createPost)
```

Each item in that chain can call `c.Next()` to continue, return a response to stop, or return an error for the app error handler.

## Where to go next

- [Installation](./getting-started/installation) covers module setup and Go version requirements.
- [Quick Start](./getting-started/quick-start) builds a small API in one file.
- [First Route](./getting-started/first-route) explains params, query values, JSON, and errors in one handler.
- [Routing](./guide/routing) is the first full guide page once you are ready for groups and named routes.
- [API Reference](./api/app) keeps the method-level details separate from the learning path.

</div>

</div>
