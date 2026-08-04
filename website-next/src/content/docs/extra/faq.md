---
title: FAQ
description: Common Zinc questions about performance, compatibility, and framework shape.
---

## Is Zinc just a router?

No. Zinc is a full API framework with routing, binding, responses, rendering, static serving, lifecycle control, route introspection, and first-party middleware.

## Why Zinc instead of a plain `net/http` stack?

Zinc gives you:

- cleaner route registration
- explicit binding helpers
- richer response helpers
- route groups and prefix middleware
- named routes and reverse URL generation
- a more coherent API framework surface

while still staying close to `net/http`.

## Is Zinc compatible with stdlib handlers?

Yes. Use `HandleHTTP` for one route, `Mount` for a subtree, or `Wrap` and `WrapFunc` inside a Zinc handler chain.

## Does Zinc support uploads and multipart forms?

Yes. Zinc supports multipart binding and direct multipart helpers such as `FormFile`, `FormFiles`, `MultipartForm`, and `SaveFile`.

## Can Zinc generate URLs from route names?

Yes. Register routes with `RouteSpec` and then use `RouteByName` and `URL`.

## Is Zinc optimized for performance?

Yes. In the current in-process suite, Zinc records the lowest latency in 62 of 77 comparable rows against Gin, Echo, and Chi. See the [benchmark report](./benchmarks) for the environment, command, and complete results.
