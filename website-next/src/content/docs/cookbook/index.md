---
slug: /cookbook
title: Cookbook
description: Runnable Zinc examples that combine routing, templates, browser code, storage, and real request flows.
---

Zinc's cookbook is for complete, runnable examples — not isolated API snippets.

Each recipe is a small service you could deploy: routing, storage, background work, templates, and browser code in one file.

## Recipes

### [Templated HTML + JS Page](/cookbook/templated-html-js-page)

Render a server-side HTML page, serve browser assets from `/static`, and progressively enhance the page with a lightweight script.

### [SQLite CRUD API](/cookbook/sqlite-crud-api)

A compact SQLite-backed API with request binding, route params, and JSON responses.

### [Scheduled Jobs](/cookbook/scheduled-jobs)

Run cron schedules and an HTTP API from one process, enqueue ad-hoc jobs from a handler, and shut both down cleanly.

### [Server-Rendered UI with Templ](/cookbook/templ-ui)

Render type-safe Templ components from a Zinc handler, use a real Templ UI Button, and serve Tailwind output with `app.Static`.

## Why this section exists

The guide explains concepts. The API reference explains surface area. The cookbook shows how the pieces fit together on a real request path.

That is usually the fastest way to answer questions like:

- "How should I structure a small Zinc service?"
- "What does a template + static asset setup look like?"
- "What does CRUD with binding and JSON responses look like?"
- "How do I run background work next to my HTTP handlers?"

More recipes — Redis-backed sessions, Kafka producers and consumers, Postgres with migrations, file uploads, and server-sent events — are on the way.
