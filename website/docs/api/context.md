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
| `SetWriter(writer)` | Replace the response writer for middleware |
| `Method()` | Current request method |
| `Path()` | Current request path |
| `SetPath(path)` | Rewrite the current path before continuing |
| `OriginalURL()` | Original request URI |
| `Route()` | Matched `RouteInfo` |
| `Context()` | Underlying request context |
| `SetContext(ctx)` | Replace the underlying request context |

## Route and query helpers

| Method | Purpose |
|---|---|
| `Param(name)` | Read a route param |
| `ParamOr(name, fallback)` | Route param with fallback |
| `Query(name)` | Read a query value |
| `QueryOr(name, fallback)` | Query value with fallback |
| `QueryArray(name)` | Repeated query values |
| `QueryMap(name)` | Bracket-style query map |
| `QueryValues()` | Full query map |
| `PostForm(name)` | Read a body form value |
| `PostFormOr(name, fallback)` | Body form value with fallback |
| `PostFormArray(name)` | Repeated body form values |
| `PostFormMap(name)` | Bracket-style body form map |
| `ContentType()` | Request content type without parameters |
| `IsWebSocket()` | Whether the request is a websocket upgrade |
| `GetHeader(key)` | Read a request header |
| `Cookie(name)` | Read a request cookie |
| `Cookies()` | Read all request cookies |
| `BodyBytes()` | Read and cache the request body |
| `BodyString()` | Read and cache the request body as a string |

## Request-local state

| Method | Purpose |
|---|---|
| `Set(key, value)` | Store request-local data |
| `Get(key)` | Retrieve request-local data |
| `MustGet(key)` | Retrieve request-local data or panic |
| `GetString(key)` | Retrieve a stored string |
| `GetBool(key)` | Retrieve a stored bool |
| `GetInt(key)` | Retrieve a stored int |
| `GetInt64(key)` | Retrieve a stored int64 |
| `GetFloat64(key)` | Retrieve a stored float64 |
| `GetStringSlice(key)` | Retrieve a stored string slice |
| `GetStringMap(key)` | Retrieve a stored string-to-any map |
| `GetStringMapString(key)` | Retrieve a stored string-to-string map |
| `GetStringMapStringSlice(key)` | Retrieve a stored string-to-string-slice map |

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
- `Attachment`
- `Download`
- `Inline`
- `Stream`
- `Blob`
- `JSONBlob`
- `XMLBlob`
- `HTMLBlob`
- `SSE`
- `Accepts`
- `Negotiate`
- `SetSameSite`
- `SetCookie`
- `ClearCookie`
- `NoContent`

## Multipart helpers

- `FormFile`
- `FormFiles`
- `MultipartForm`
- `SaveFile`

## Client helpers

- `IP()`
- `IPs()`
- `RemoteIP()`
- `Scheme()`
- `Secure()`
- `IsPreflight()`
- `RequestID()`

`IP()` and `IPs()` use `Config.ProxyHeader` only when `TrustedProxies` allows the direct peer.

## Safety

Use `Copy()` when context data needs to outlive the handler.
