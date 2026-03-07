# Zinc Specification

Status: draft  
Scope: Zinc core framework direction after studying local copies of:

- `/Users/matt/dev/sandbox/chi`
- `/Users/matt/dev/sandbox/echo`
- `/Users/matt/dev/sandbox/gin`
- `/Users/matt/dev/sandbox/fiber`
- `/Users/matt/dev/sandbox/httprouter`

This is a product and API specification for what Zinc should become. It is not a line-by-line copy of any existing framework.

## One Sentence

Zinc should be the best Express/Fiber-style developer experience available on top of `net/http`, with `HttpRouter`-class routing discipline, strong Go ergonomics, and no non-idiomatic framework baggage.

## Product Definition

Zinc is:

- a `net/http` web framework
- primarily optimized for REST APIs
- Express/Fiber-inspired in readability and route ergonomics
- explicitly Go-idiomatic in lifecycle, error handling, and stdlib interoperability
- performance-sensitive by default

Zinc is not:

- a `fasthttp` framework
- a service container
- an ORM wrapper
- a cron runner
- a websocket room manager
- a template engine product
- a framework that prints to stdout or manages app process concerns by default

## Philosophy

### Fiber inspiration we want

- expressive route API
- clean grouping
- route-first readability
- convenient request/response helpers
- low-friction middleware authoring
- “nice to write” handler code

### Fiber choices we do not want

- `fasthttp`
- custom transport contract instead of `net/http`
- request context semantics that diverge from stdlib expectations
- variadic `...any` route registration
- framework magic that works against Go readability

### Go principles Zinc must respect

- `App` is an `http.Handler`
- raw `*http.Request` and `http.ResponseWriter` stay available
- startup and shutdown are explicit and return errors
- route registration should return errors, not panic
- no hidden service locator in core
- no framework-managed database connection abstraction in core
- no stdout logging by default
- configuration should be explicit and mostly immutable after startup

## Research Synthesis

## Chi

What to take:

- uncompromising `net/http` compatibility
- `Mount`, `Route`, `Group`, `With`
- route introspection
- custom `NotFound` and `MethodNotAllowed`
- minimalism

What to reject:

- stdlib middleware shape as Zinc’s primary UX
- lack of rich context/response helpers
- regex-heavy routing philosophy in the hot path

## Echo

What to take:

- centralized error handler
- binder, validator, renderer, serializer interfaces
- `Any`, `Match`, static/file helpers
- `NoContent`, `Redirect`, `Blob`, `Stream`, `File`, cookie helpers
- route metadata on context
- strong stdlib alignment

What to reject:

- panic-oriented route registration ergonomics
- too much mutable framework state on the app object

## Gin

What to take:

- `NoRoute` / `NoMethod`
- strong trusted-proxy handling
- broad request/response helper coverage
- `FullPath`
- listener-based startup APIs

What to reject:

- abort-centric control model as the primary Zinc mental model
- enormous typed getter surface on context
- overly broad serialization surface in core

## Fiber

What to take:

- route ergonomics
- `All`, `Group`, `Route`
- strong response convenience API
- request/response helper breadth
- `FullPath`, `RequestID`, `Redirect`, `Bind`

What to reject:

- `fasthttp`
- custom request/response transport types as the main contract
- `...any` registration APIs
- service subsystem in core
- context lifetime caveats caused by non-stdlib transport semantics

## HttpRouter

What to take:

- per-method compressed radix trees
- hot-path discipline
- automatic `OPTIONS`
- proper `405` + `Allow`
- small focused API

What to reject:

- framework minimalism to the point of missing standard REST ergonomics

## Zinc Identity

Zinc should position as:

- “Fiber DX on `net/http`”
- “Go-idiomatic Express-style API”
- “Fast enough to compete with Gin/Echo, with less baggage”

The core wedge is:

- readable route declarations
- no `fasthttp`
- strong stdlib interoperability
- better ergonomics than raw routers
- better Go taste than cargo-cult JS abstractions

## Core Design Rules

1. `net/http` is non-negotiable.
2. Route syntax stays Express-style: `:id`, `*`.
3. Route registration APIs stay typed, not `...any`.
4. Handlers return `error`.
5. Middleware and handlers use one function shape.
6. Optional features must have near-zero cost when unused.
7. Core remains focused on HTTP, routing, middleware, binding, responses, and lifecycle.
8. Everything else moves to optional packages.

## What Zinc Keeps

These are core identity features and should remain:

- Express-style path syntax
- `App`, `Group`, `Context`
- route methods like `Get`, `Post`, `Put`, `Delete`
- `Match` and `All`
- middleware via `Next()`
- handler signature returning `error`
- `Map` alias
- `http.Handler` compatibility
- high-performance internal router

## What Zinc Removes From Core

These features should be removed from Zinc core and moved to optional packages or dropped entirely.

| Current Feature | Decision | Reason |
|---|---|---|
| `Injectable`, `GetService`, `ServiceOf`, `Context.Service` | Remove from core | service locator is not idiomatic Go |
| `ConnectDB`, `GetDB`, `Context.DB`, `WithDB` | Remove from core | database wiring belongs in user code |
| `Cron`, scheduler methods | Remove from core | not an HTTP framework concern |
| built-in websocket room manager | Remove from core | too opinionated for framework core |
| built-in template engine implementation | Move behind interface / optional package | renderer interface belongs in core, engine does not |
| upload manager object | Remove from core | multipart helpers on context are enough |
| built-in concrete validator implementation | Move behind interface / optional package | apps should choose validator |
| startup logging / route printing flags | Remove | library should not print by default |
| `SetConfig` after construction | Remove | mutating config after creation is poor API design |
| `App.Version()` / `Context.Version()` | Remove | package constant is enough |
| `BodyParser` | Deprecate in favor of `Bind*` | duplicate concept |

## Core API Direction

## Handler Types

```go
type HandlerFunc func(*Context) error

// Backwards compatibility alias only.
type Middleware = HandlerFunc
```

Zinc should not maintain separate conceptual handler types for route handlers vs middleware.

## App Lifecycle

Current `Serve(port ...string)` should not be the primary API.

Target lifecycle:

```go
func New() *App
func NewWithConfig(cfg Config) *App

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (a *App) Handler() http.Handler

func (a *App) Listen(addr string) error
func (a *App) ListenTLS(addr, certFile, keyFile string) error
func (a *App) Serve(ln net.Listener) error
func (a *App) Shutdown(ctx context.Context) error
```

Rules:

- no implicit default port in the primary API
- no startup banners
- no shutdown printing
- `App` remains directly usable as `http.Handler`

Recommended usage:

```go
app := zinc.New()

if err := app.Listen(":8080"); err != nil {
	log.Fatal(err)
}
```

Advanced stdlib usage:

```go
srv := &http.Server{
	Addr:    ":8080",
	Handler: app,
}
if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	log.Fatal(err)
}
```

## Routing API

Core app and group APIs should converge on this surface:

```go
func (a *App) Use(handlers ...HandlerFunc)
func (a *App) UsePrefix(prefix string, handlers ...HandlerFunc)

func (a *App) Group(prefix string, handlers ...HandlerFunc) *Group
func (a *App) Route(prefix string, fn func(*Group), handlers ...HandlerFunc) *Group
func (a *App) Mount(prefix string, h http.Handler)

func (a *App) Add(method, path string, handlers ...HandlerFunc) error
func (a *App) Get(path string, handlers ...any) error // accepts handler funcs or string shorthand
func (a *App) Post(path string, handlers ...HandlerFunc) error
func (a *App) Put(path string, handlers ...HandlerFunc) error
func (a *App) Delete(path string, handlers ...HandlerFunc) error
func (a *App) Patch(path string, handlers ...HandlerFunc) error
func (a *App) Head(path string, handlers ...HandlerFunc) error
func (a *App) Options(path string, handlers ...HandlerFunc) error
func (a *App) Connect(path string, handlers ...HandlerFunc) error
func (a *App) Trace(path string, handlers ...HandlerFunc) error
func (a *App) Match(methods []string, path string, handlers ...HandlerFunc) error
func (a *App) All(path string, handlers ...HandlerFunc) error
func (a *App) Any(path string, handlers ...HandlerFunc) error // alias to All
```

Group gets the same methods.

### Routing semantics

- `Use(...)` is global pre-routing middleware
- `UsePrefix(...)` is path-scoped pre-routing middleware
- group middleware is post-match middleware
- route-level handlers execute in declared order
- route registration returns `error` on conflicts or invalid patterns
- no route-registration panics in the public API

### Router capabilities Zinc must support

- static routes
- named params: `:id`
- catch-all params: `*`
- method-specific trees
- `405 Method Not Allowed` with `Allow`
- automatic `OPTIONS` responses
- automatic `HEAD` for `GET` unless explicitly overridden
- route metadata for introspection

### Router capabilities Zinc should not put in the hot path

- general regex path segments
- reflection-based route matching
- per-request path splitting allocations

If parameter constraints are added later, they should use fast named constraints, not general regex as the default mechanism.

## Mounted stdlib support

Zinc should have first-class support for stdlib interop:

```go
func (a *App) Mount(prefix string, h http.Handler)
func Wrap(h http.Handler) HandlerFunc
func WrapFunc(fn http.HandlerFunc) HandlerFunc
```

This is mandatory for the “Fiber DX, stdlib contract” positioning.

## Not Found / Method Not Allowed

Zinc must add:

```go
func (a *App) NotFound(handler HandlerFunc)
func (a *App) MethodNotAllowed(handler HandlerFunc)
func (a *App) Routes() []RouteInfo
```

`RouteInfo` should include at minimum:

```go
type RouteInfo struct {
	Method  string
	Path    string
	Handler string
}
```

`Context` should expose:

```go
func (c *Context) FullPath() string
func (c *Context) Route() RouteInfo
```

## Static / File APIs

Zinc currently lacks a full framework-level static/file surface.

Add:

```go
func (a *App) Static(prefix, root string, opts ...StaticOption) error
func (a *App) StaticFS(prefix string, filesystem fs.FS, opts ...StaticOption) error
func (a *App) File(path, file string) error
func (a *App) FileFS(path, file string, filesystem fs.FS) error
```

Group equivalents should exist too.

Rules:

- use `fs.FS` where possible
- default to stdlib file serving behavior
- no custom file server stack in the hot path unless explicitly configured

## Context API

Zinc’s context should be a thin ergonomic layer over `net/http`, not a replacement for it.

### Raw stdlib escape hatches

```go
func (c *Context) Request() *http.Request
func (c *Context) SetRequest(r *http.Request)
func (c *Context) Writer() http.ResponseWriter
func (c *Context) SetWriter(w http.ResponseWriter)

func (c *Context) Context() context.Context
func (c *Context) SetContext(ctx context.Context)
```

### Flow control

```go
func (c *Context) Next() error
```

No primary `Abort()` API is required. Zinc’s middleware model should stay closer to Express/Fiber: if a middleware does not call `Next()`, the chain stops.

### Request helpers

```go
func (c *Context) Method() string
func (c *Context) Path() string
func (c *Context) SetPath(path string)
func (c *Context) OriginalURL() string

func (c *Context) Param(name string) string
func (c *Context) ParamOr(name, fallback string) string

func (c *Context) Query(name string) string
func (c *Context) QueryOr(name, fallback string) string
func (c *Context) QueryValues() url.Values

func (c *Context) FormValue(name string) string
func (c *Context) FormFile(name string) (*multipart.FileHeader, error)
func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error)
func (c *Context) MultipartForm() (*multipart.Form, error)
func (c *Context) SaveFile(file *multipart.FileHeader, dst string) error

func (c *Context) GetHeader(key string) string
func (c *Context) Cookie(name string) (*http.Cookie, error)
func (c *Context) Cookies() []*http.Cookie

func (c *Context) BodyBytes() ([]byte, error)
func (c *Context) BodyString() (string, error)
```

`BodyParser` should be removed. Binding APIs cover that responsibility.

### Request metadata helpers

```go
func (c *Context) Scheme() string
func (c *Context) IP() string
func (c *Context) IPs() []string
func (c *Context) RemoteIP() string
func (c *Context) Secure() bool
func (c *Context) IsWebSocket() bool
func (c *Context) IsPreflight() bool
func (c *Context) RequestID() string
```

### Request-local storage

Current `Set/Get` should evolve to accept typed keys:

```go
func (c *Context) Set(key any, value any)
func (c *Context) Get(key any) (any, bool)
func (c *Context) MustGet(key any) any
```

Zinc should not copy Gin’s huge typed getter matrix.

## Binding and Validation

Zinc should support a complete REST binder surface, but the implementation should sit behind interfaces.

Core interfaces:

```go
type Binder interface {
	Bind(*Context, any) error
	BindBody(*Context, any) error
	BindQuery(*Context, any) error
	BindForm(*Context, any) error
	BindHeader(*Context, any) error
	BindPath(*Context, any) error
}

type Validator interface {
	Validate(any) error
}
```

Context helpers:

```go
func (c *Context) Bind(v any) error
func (c *Context) BindJSON(v any) error
func (c *Context) BindXML(v any) error
func (c *Context) BindForm(v any) error
func (c *Context) BindQuery(v any) error
func (c *Context) BindHeader(v any) error
func (c *Context) BindPath(v any) error
func (c *Context) Validate(v any) error
```

Rules:

- default binder uses stdlib decoding
- validator is optional
- no concrete validation engine in Zinc core
- no reflection on request hot path beyond binder use

## Rendering / Serialization

The core should expose interfaces, not concrete engines.

```go
type Renderer interface {
	Render(w io.Writer, name string, data any, c *Context) error
}

type JSONCodec interface {
	Encode(w io.Writer, v any, indent string) error
	Decode(r io.Reader, v any) error
}
```

App config holds:

- `Renderer`
- `Binder`
- `Validator`
- `JSONCodec`
- `ErrorHandler`

Current built-in template engine should move to an optional package.

## Response API

Zinc needs a complete response surface that matches what developers expect from a real Go web framework.

### Required response helpers

```go
func (c *Context) Status(code int) *Context

func (c *Context) SetHeader(key, value string) *Context
func (c *Context) AppendHeader(key string, values ...string) *Context
func (c *Context) Type(ext string) *Context
func (c *Context) Location(url string) *Context
func (c *Context) Vary(fields ...string) *Context

func (c *Context) String(s string) error
func (c *Context) Send(v any) error
func (c *Context) Data(contentType string, b []byte) error
func (c *Context) JSON(v any) error
func (c *Context) JSONPretty(v any, indent string) error
func (c *Context) XML(v any) error
func (c *Context) HTML(s string) error
func (c *Context) Stream(contentType string, r io.Reader) error
func (c *Context) NoContent() error
func (c *Context) Redirect(code int, location string) error

func (c *Context) File(path string) error
func (c *Context) FileFS(path string, filesystem fs.FS) error
func (c *Context) Attachment(path string, name ...string) error
func (c *Context) Download(path string, name ...string) error
func (c *Context) Render(name string, data any) error

func (c *Context) SetCookie(cookie *http.Cookie)
func (c *Context) ClearCookie(names ...string)
```

### API style decisions

- explicit methods like `JSON`, `XML`, `String`, `Data` are the primary path
- `Send(any)` remains convenience sugar, not the primary example API
- response methods must obey `HEAD`, `204`, and `304` semantics correctly

## Error Model

Zinc should adopt a centralized error model similar to Echo/Fiber, while keeping handler-returned errors as the main flow.

Core:

```go
type ErrorHandler func(*Context, error)

type HTTPError struct {
	Code    int
	Message string
}
```

Behavior:

- handler returns error
- app-level `ErrorHandler` renders it
- `HTTPError` maps cleanly to status code + message
- default error handler is predictable and minimal

This is a strength Zinc already has and should keep.

## Middleware Strategy

Zinc core should stay small, but the official middleware package should grow into a standard set.

### Tier 1 official middleware

- `recover`
- `requestid`
- `logger` / `requestlogger`
- `cors`
- `compress`
- `timeout`
- `limiter`
- `etag`
- `static`
- `methodoverride`
- `realip`
- `slash` / `cleanpath`
- `secure` / `helmet`

### Tier 2 official middleware

- `basicauth`
- `keyauth`
- `csrf`
- `proxy`
- `redirect`
- `idempotency`
- `session`

JWT should not be hard-baked into core. Keep auth middleware modular.

## Config Specification

The core config should be reduced and focused.

Target `Config` direction:

```go
type Config struct {
	ServerHeader string // default ""

	CaseSensitive bool
	StrictRouting bool
	AutoHead bool
	AutoOptions bool
	HandleMethodNotAllowed bool

	BodyLimit int64
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	IdleTimeout time.Duration

	ProxyHeader string
	TrustedProxies []string

	Binder Binder
	Validator Validator
	Renderer Renderer
	JSONCodec JSONCodec
	ErrorHandler ErrorHandler

	RouteCacheSize int
}
```

Rules:

- `ServerHeader` default must be empty
- no default port/address in config
- no app name/version/startup message fields
- no route-printing fields
- no mutable `SetConfig`

## Stdlib Compatibility Contract

This is a non-negotiable part of Zinc’s identity.

Zinc must:

- implement `http.Handler`
- mount arbitrary `http.Handler`
- expose raw `*http.Request`
- expose raw `http.ResponseWriter`
- work naturally with `http.Server`
- support `fs.FS`
- not require a custom client or transport

Zinc should not:

- replace the stdlib request model
- require wrapper servers
- invent parallel HTTP types where stdlib types are good enough

## Performance Rules

Every feature in this spec must respect the following constraints:

1. Static and param route lookup remain zero-allocation on the hot path.
2. Router features must be per-method and radix-tree based.
3. Optional subsystems must not add overhead when unused.
4. Adapters and rich helpers may allocate, but only on the paths that use them.
5. Core request serving must not depend on reflection.
6. Binding, validation, rendering, and serialization hooks stay off the routing hot path.
7. Route registration remains explicit and error-returning.

## Package Structure Direction

Core package:

- `zinc`

Official packages:

- `zinc/middleware`
- `zinc/render/...`
- `zinc/bind/...`
- `zinc/validate/...`
- `zinc/static/...`
- `zinc/ws/...` optional

Features to move out of core:

- current template engine
- current validator implementation
- websocket helpers
- upload helper object
- cron scheduler

## Migration Direction

### Deprecate

- `Serve(port ...string)`
- `Injectable`
- `GetService`
- `ServiceOf`
- `Context.Service`
- `ConnectDB`
- `GetDB`
- `Context.DB`
- `WithDB`
- `Cron` and related scheduler methods
- `SetTemplateEngine`
- `SetWebSocketHandler`
- `SetFileUpload`
- `SetConfig`
- `BodyParser`
- `App.Version()`
- `Context.Version()`

### Add

- `NewWithConfig`
- `Listen`
- `ListenTLS`
- `Serve(net.Listener)`
- `Shutdown(context.Context)`
- `UsePrefix`
- `Route`
- `Mount`
- `Any`
- `Static`, `StaticFS`, `File`, `FileFS`
- `NotFound`, `MethodNotAllowed`, `Routes`
- `FullPath`, `Route`
- `ParamOr`, `QueryOr`
- `BindHeader`, `BindPath`
- `BodyBytes`, `BodyString`
- `SetHeader`, `AppendHeader`, `Type`, `Location`, `Vary`
- `NoContent`, `Redirect`, `Data`, `Stream`, `Attachment`, `Download`, `Render`
- `Cookie`, `SetCookie`, `ClearCookie`

## Recommended Example Style

```go
app := zinc.NewWithConfig(zinc.Config{
	BodyLimit:              4 << 20,
	HandleMethodNotAllowed: true,
	AutoOptions:            true,
	AutoHead:               true,
})

app.Use(middleware.RequestID())
app.Use(middleware.Recover())
app.Use(middleware.Logger())

app.Get("/health", func(c *zinc.Context) error {
	return c.JSON(zinc.Map{"ok": true})
})

app.Route("/api", func(api *zinc.Group) {
	api.Get("/users/:id", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{"id": c.Param("id")})
	})
})

app.Mount("/metrics", promhttp.Handler())

if err := app.Listen(":8080"); err != nil {
	log.Fatal(err)
}
```

## Implementation Phases

### Phase 1: Core cleanup

- remove non-core subsystems from `App`
- add idiomatic lifecycle methods
- remove stdout behavior
- add centralized error handler config

### Phase 2: Surface completion

- add `NotFound`, `MethodNotAllowed`
- add static/file APIs
- add cookie/header/file/stream helpers
- add `FullPath`, `Route`, `Routes`
- add missing bind methods

### Phase 3: Stdlib-strength integrations

- add `Mount`
- add stdlib handler adapters
- add renderer/binder/validator/json interfaces
- move template and validation implementations out of core

### Phase 4: Middleware and polish

- expand official middleware surface
- document recommended app patterns
- provide migration notes and deprecation shims

### Phase 5: Performance enforcement

- benchmark gates for static, param, 405, middleware, and file serving
- ensure added features are cold-path or opt-in

## Final Direction

Zinc should become:

- smaller in core scope
- broader in real HTTP framework coverage
- faster in the router
- more complete in request/response ergonomics
- more idiomatic in lifecycle and configuration
- more useful in stdlib-heavy Go codebases

The target is not “Echo clone”, “Gin clone”, or “Fiber on `net/http`”.

The target is:

Go’s best Express-style REST framework on `net/http`.
