# Feature Parity: Gin vs Chi vs Echo vs Zinc

A practical, source-backed parity matrix for API framework capabilities.

## Scope

This compares what is **built into each core repo** (out of the box), using local checkouts:

- Gin: `v1.12.0-5-g3e44fdc`
- Chi: `v5.2.5-7-ga54874f`
- Echo: `v5.0.4-1-g1753170`
- Zinc: `b4cd92c` (current local 0.0.73 work)

Legend:

- ✅ = built-in / first-party in that core repository
- ❌ = not built-in in that core repository (possible via stdlib, custom code, or external package)

## Parity Matrix

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| Named path params | ✅ | ✅ | ✅ | ✅ | Core routing primitive in all four |
| Wildcard / catch-all routes | ✅ | ✅ | ✅ | ✅ | Supported by each router syntax |
| Route groups | ✅ | ✅ | ✅ | ✅ | Prefix-based grouping supported everywhere |
| Global middleware | ✅ | ✅ | ✅ | ✅ | App/router-level middleware in all |
| Group-level middleware | ✅ | ✅ | ✅ | ✅ | Group-scoped middleware supported |
| Route-level middleware | ✅ | ✅ | ✅ | ✅ | Chi uses `With(...)` for inline route middleware |
| Multi-method helpers (`Any` / `Match`) | ✅ | ❌ | ✅ | ✅ | Chi has `Handle`/`Method`, but no direct `Any`/`Match` helpers |
| Prefix middleware without creating a group | ❌ | ❌ | ❌ | ✅ | Zinc `UsePrefix(...)` is a notable differentiator |
| Custom 404 handler | ✅ | ✅ | ✅ | ✅ | All support custom not-found behavior |
| Custom 405 handler | ✅ | ✅ | ✅ | ✅ | All support method-not-allowed customization |
| Mount handler tree at prefix | ❌ | ✅ | ❌ | ✅ | Chi/Zinc have explicit `Mount(...)` |
| Static file helper API | ✅ | ❌ | ✅ | ✅ | Chi uses stdlib patterns, no `Static(...)` helper on router |
| Shorthand string route registration | ❌ | ❌ | ❌ | ✅ | Zinc `Get("/", "Hello")` |
| Route metadata helper in handler | ✅ | ✅ | ✅ | ✅ | e.g. full/matched route path access |
| Route registration returns errors (not panic-first) | ❌ | ❌ | ❌ | ✅ | Zinc favors error returns; others often panic on invalid route registration |
| Unified bind (path + query + body) | ❌ | ❌ | ✅ | ✅ | Echo/Zinc provide single-call unified binding |
| Dedicated bind helpers (query/header/path/body) | ✅ | ❌ | ✅ | ✅ | Chi leaves binding to ecosystem/userland |
| Header binding helper | ✅ | ❌ | ✅ | ✅ | |
| Multipart file helper API | ✅ | ❌ | ✅ | ✅ | |
| Validation hook integrated with binding flow | ✅ | ❌ | ⚠️ | ✅ | Echo exposes validation in core, but typically as explicit `c.Validate(...)` after binding |
| JSON response helper | ✅ | ❌ | ✅ | ✅ | |
| XML response helper | ✅ | ❌ | ✅ | ✅ | |
| String/text response helper | ✅ | ❌ | ✅ | ✅ | |
| Stream response helper | ✅ | ❌ | ✅ | ✅ | |
| File / attachment response helpers | ✅ | ❌ | ✅ | ✅ | |
| SSE helper in core | ✅ | ❌ | ❌ | ✅ | Echo can stream manually; no dedicated SSE helper API |
| WebSocket upgrade/helper in core | ❌ | ❌ | ❌ | ✅ | Zinc has explicit WS helpers in core API |
| Request logger middleware (configurable) | ✅ | ✅ | ✅ | ✅ | All have configurable request logging middleware |
| Panic recovery middleware in core | ✅ | ✅ | ✅ | ❌ | Zinc gap today |
| CORS middleware in core repo | ❌ | ❌ | ✅ | ✅ | Gin/Chi usually use external first-party repos |
| Rate limiter middleware in core repo | ❌ | ❌ | ✅ | ✅ | Chi has throttle middleware, but not token-bucket rate limiter in core |
| Request ID middleware in core repo | ❌ | ✅ | ✅ | ❌ | Zinc has `RequestID()` accessor, but no request-id middleware yet |
| Trusted proxy / real IP configuration | ✅ | ❌ | ✅ | ✅ | Chi provides RealIP middleware but no built-in trusted-proxy list config |
| net/http interop wrappers/adapters | ✅ | ✅ | ✅ | ✅ | Gin/Echo/Zinc expose wrappers; Chi is natively `net/http` |

## What This Says About Zinc

Zinc is already strong in API ergonomics and core capability coverage, especially:

- `UsePrefix(...)` middleware targeting
- route shorthand (`Get("/", "...")`)
- explicit WS + SSE helpers
- non-panic route registration style

## Highest-Value Next Parity Targets For Zinc

1. Add a built-in `middleware.Recover()` (panic recovery) to close the most visible ops gap.
2. Add a built-in `middleware.RequestID()` to align with Echo/Chi operational defaults.
3. Keep CORS/rate-logger performance tight (already in good shape) while adding the above.

## Source Notes (Local Repos)

- Gin: `gin.go`, `routergroup.go`, `context.go`, `logger.go`, `recovery.go`, `binding/*`
- Chi: `chi.go`, `mux.go`, `context.go`, `middleware/logger.go`, `middleware/recoverer.go`, `middleware/request_id.go`
- Echo: `echo.go`, `group.go`, `context.go`, `bind.go`, `middleware/request_logger.go`, `middleware/cors.go`, `middleware/rate_limiter.go`, `middleware/request_id.go`
- Zinc: `app.go`, `group.go`, `router.go`, `ctx.go`, `bind.go`, `response.go`, `realtime.go`, `cfg.go`, `middleware/request_logger.go`, `middleware/cors/cors.go`, `middleware/ratelimiter.go`
