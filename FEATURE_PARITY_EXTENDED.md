# Extended Feature Parity: Gin vs Chi vs Echo vs Zinc

A broader parity matrix for "normal API framework" expectations beyond routing/binding basics.

## Scope And Legend

Compared from local core repositories only (no external ecosystem plugins):

- Gin: `v1.12.0-5-g3e44fdc`
- Chi: `v5.2.5-7-ga54874f`
- Echo: `v5.0.4-1-g1753170`
- Zinc: `b4cd92c`

Legend:

- ✅ = built-in and first-class in core
- ⚠️ = partial/indirect support (possible, but not first-class or not equivalent)
- ❌ = not built into core

---

## 1) Core API Surface

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| Named path params | ✅ | ✅ | ✅ | ✅ | |
| Wildcard / catch-all routes | ✅ | ✅ | ✅ | ✅ | |
| Route groups | ✅ | ✅ | ✅ | ✅ | |
| Global middleware | ✅ | ✅ | ✅ | ✅ | |
| Group-level middleware | ✅ | ✅ | ✅ | ✅ | |
| Route-level middleware | ✅ | ✅ | ✅ | ✅ | Chi uses `With(...)` |
| Multi-method helper (`Any`/`Match`) | ✅ | ⚠️ | ✅ | ✅ | Chi has `Handle`/`Method`, but no direct `Any`/`Match` helpers |
| Prefix middleware without creating a group | ❌ | ❌ | ❌ | ✅ | Zinc `UsePrefix(...)` |
| Mount sub-router/handler tree | ❌ | ✅ | ❌ | ✅ | Chi/Zinc explicit `Mount(...)` |
| Custom 404 handler | ✅ | ✅ | ✅ | ✅ | |
| Custom 405 handler | ✅ | ✅ | ✅ | ✅ | |
| Route metadata helper in handler | ✅ | ✅ | ✅ | ✅ | Full/matched route info |
| Route introspection list API | ✅ | ✅ | ✅ | ✅ | Echo via `e.Router().Routes()` |
| Shorthand string route registration | ❌ | ❌ | ❌ | ✅ | `Get("/", "Hello")` |
| Error-first route registration (panic-avoidant style) | ❌ | ❌ | ❌ | ✅ | |

---

## 2) Binding, Validation, Response

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| Unified bind (path+query+body in one call) | ❌ | ❌ | ✅ | ✅ | |
| Dedicated bind helpers (path/query/header/body/form) | ✅ | ❌ | ✅ | ✅ | |
| Multipart file helper API | ✅ | ❌ | ✅ | ✅ | |
| Validator hook in core | ✅ | ❌ | ✅ | ✅ | Echo validation is explicit (`c.Validate(...)`) |
| JSON helper | ✅ | ❌ | ✅ | ✅ | |
| XML helper | ✅ | ❌ | ✅ | ✅ | |
| String/text helper | ✅ | ❌ | ✅ | ✅ | |
| Stream helper | ✅ | ❌ | ✅ | ✅ | |
| File/attachment helpers | ✅ | ❌ | ✅ | ✅ | |
| Template render helper | ✅ | ❌ | ✅ | ✅ | |
| SSE helper API | ✅ | ❌ | ❌ | ✅ | Echo streams manually; no dedicated SSE helper API |
| WebSocket helper API | ❌ | ❌ | ❌ | ✅ | Zinc has explicit WS helper |

---

## 3) Ops And Observability

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| Request logger middleware (configurable) | ✅ | ✅ | ✅ | ✅ | |
| slog-oriented built-in request logger | ❌ | ❌ | ✅ | ✅ | Echo and Zinc default to slog in logger middleware |
| Panic recovery middleware | ✅ | ✅ | ✅ | ❌ | Zinc gap |
| Request ID middleware | ❌ | ✅ | ✅ | ❌ | Zinc has `RequestID()` accessor, but no generator middleware |
| Timeout middleware | ❌ | ✅ | ✅ | ❌ | |
| Request size limiting middleware | ❌ | ✅ | ✅ | ❌ | Zinc has global body limit, not middleware-level |
| Global body limit setting | ⚠️ | ❌ | ❌ | ✅ | Zinc `Config.BodyLimit`; Gin only multipart memory knobs |
| Compression middleware | ❌ | ✅ | ✅ | ❌ | |
| Rate limiting (req/time) middleware | ❌ | ⚠️ | ✅ | ✅ | Chi has throttle concurrency middleware, not direct rate limiter |
| Concurrency throttling middleware | ❌ | ✅ | ❌ | ❌ | Chi `Throttle(...)` |
| Trusted proxy / real IP strategy | ✅ | ⚠️ | ✅ | ✅ | Chi uses middleware (`RealIP`), not trusted-proxy config model |
| Route tree walker utility | ❌ | ✅ | ❌ | ❌ | Chi `Walk(...)` |

---

## 4) Security And Auth Middleware

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| CORS middleware in core repo | ❌ | ❌ | ✅ | ✅ | |
| Basic Auth middleware/helper in core repo | ✅ | ❌ | ✅ | ❌ | Gin `BasicAuth(...)`, Echo middleware |
| API key auth middleware in core repo | ❌ | ❌ | ✅ | ❌ | Echo `key_auth` middleware |
| CSRF middleware in core repo | ❌ | ❌ | ✅ | ❌ | |
| Secure headers middleware in core repo | ❌ | ❌ | ✅ | ❌ | Echo `Secure()` |

---

## 5) Transport And Runtime

| Feature | Gin | Chi | Echo | Zinc | Notes |
|---|---|---|---|---|---|
| One-call server start helper | ✅ | ❌ | ✅ | ✅ | Chi is router-only by design |
| TLS start helper | ✅ | ❌ | ✅ | ✅ | Echo via `StartConfig.StartTLS(...)` |
| Graceful shutdown helper in framework API | ❌ | ❌ | ⚠️ | ✅ | Echo uses `StartConfig`/context pattern; Zinc has `Shutdown(ctx)` |
| HTTP/3 / QUIC helper | ✅ | ❌ | ❌ | ❌ | Gin `RunQUIC(...)` |
| h2c helper support | ✅ | ❌ | ❌ | ❌ | Gin `UseH2C` |
| net/http interop wrappers/adapters | ✅ | ✅ | ✅ | ✅ | Chi is native; others provide wrappers |

---

## Zinc Status (Pragmatic Read)

Zinc is **strong in core API ergonomics** and already competitive with mainstream frameworks for routing, binding, responses, SSE/WS, and general middleware composition.

Most important parity gaps are now in **ops/security middleware breadth**, not core API shape.

## Highest-Value Next Steps For Zinc

1. Add `middleware.Recover()` (panic recovery).
2. Add `middleware.RequestID()` (ID generation + propagation).
3. Add `middleware.Timeout(...)` and request-size middleware for operational controls.
4. Add `middleware.Compress(...)`.
5. Add auth/security baseline middleware (`BasicAuth` first, then secure headers/CSRF depending on API/browser focus).
6. Optional: add route walker/introspection utility if docs/tooling generation becomes a priority.

---

## Source Pointers (Local)

- Gin: `gin.go`, `routergroup.go`, `context.go`, `auth.go`, `logger.go`, `recovery.go`, `utils.go`
- Chi: `chi.go`, `mux.go`, `context.go`, `tree.go`, `middleware/*`
- Echo: `echo.go`, `router.go`, `context.go`, `bind.go`, `middleware/*`
- Zinc: `app.go`, `group.go`, `router.go`, `ctx.go`, `bind.go`, `response.go`, `realtime.go`, `cfg.go`, `middleware/*`
