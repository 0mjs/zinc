---
id: faq
title: ❓ FAQ
description: Common Zinc questions about performance, compatibility, and framework shape.
sidebar_position: 2
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

Yes. Use `Mount`, `Wrap`, and `WrapFunc` to integrate normal `http.Handler` and `http.HandlerFunc` values.

## Does Zinc support uploads and multipart forms?

Yes. Zinc supports multipart binding and direct multipart helpers such as `FormFile`, `FormFiles`, `MultipartForm`, and `SaveFile`.

## Can Zinc generate URLs from route names?

Yes. Register routes with `RouteSpec` and then use `RouteByName` and `URL`.

## Is Zinc optimized for performance?

Yes, especially on non-throughput request-path and API-path benchmarks. Zinc currently performs strongest on routing, grouping, scenario routing, and API response/binding latency. Throughput is the category where there is still the most room to improve.
