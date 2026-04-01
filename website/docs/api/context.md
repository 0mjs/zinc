---
id: context
title: 🧠 Context
description: The request-scoped API surface for handlers and middleware.
sidebar_position: 3
---

`*zinc.Context` is the primary object you work with inside handlers.

## Request access

| Method | Purpose |
|---|---|
| `Request()` | Get the underlying `*http.Request` |
| `Writer()` | Get the response writer |
| `Method()` | Current request method |
| `Path()` | Current request path |
| `Route()` | Matched `RouteInfo` |

## Route and query helpers

| Method | Purpose |
|---|---|
| `Param(name)` | Read a route param |
| `ParamOr(name, fallback)` | Route param with fallback |
| `Query(name)` | Read a query value |
| `QueryOr(name, fallback)` | Query value with fallback |
| `QueryValues()` | Full query map |

## Request-local state

| Method | Purpose |
|---|---|
| `Set(key, value)` | Store request-local data |
| `Get(key)` | Retrieve request-local data |

## Binding

Primary binding entry point:

- `Bind().All(...)`
- `Bind().JSON(...)`, `Bind().Query(...)`, and related explicit helpers

## Responses

The response helpers live on `Context` too:

- `Status`
- `String`
- `JSON`
- `XML`
- `YAML`
- `TOML`
- `HTML`
- `Render`
- `Redirect`
- `File`
- `FileFS`
- `Stream`
- `NoContent`

## Multipart helpers

- `FormFile`
- `FormFiles`
- `MultipartForm`
- `SaveFile`

## Client helpers

- `IP()`
- `IPs()`
- `RequestID()`

## Safety

Use `Copy()` when context data needs to outlive the handler.
