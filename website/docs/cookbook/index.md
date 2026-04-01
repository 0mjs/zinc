---
id: cookbook
slug: /cookbook
title: 📚 Cookbook
description: Runnable Zinc examples that combine routing, templates, browser code, storage, and real request flows.
---

Zinc's cookbook is for complete examples, not isolated API snippets.

These recipes are the right place to see how Zinc fits together when you need:

- routing plus real handlers
- templates plus static assets
- browser behavior plus server rendering
- storage plus JSON APIs
- WebSockets plus normal HTTP endpoints

## Recipes

### [🧱 Templated HTML + JS Page](/cookbook/templated-html-js-page)

Render a small server-side HTML page, serve browser assets from `/static`, and progressively enhance the page with a lightweight script.

### [💬 Realtime Chat App](/cookbook/realtime-chat-app)

Build a live chat UI with Zinc templates, a WebSocket endpoint, and a small hub for fan-out broadcasting.

### [🗃️ SQLite CRUD API](/cookbook/sqlite-crud-api)

Wire Zinc into a compact SQLite-backed API with request binding, route params, and JSON responses.

## Why this section exists

The guide explains concepts. The API reference explains surface area. The cookbook shows how real pieces fit together.

That is usually the fastest way to answer questions like:

- "How should I structure a small Zinc app?"
- "What does a template + static asset setup look like?"
- "How do I mix normal routes and WebSockets?"
- "What does CRUD with binding and JSON responses look like?"
