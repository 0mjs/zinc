# Zinc Differentiation Plan

- Status: Proposed
- Created: 2026-07-30
- Baseline: Zinc v0.1.3 development branch
- Target: Define and deliver Zinc's identity before a stable v1 API

## Start Here

When work begins:

1. Confirm the decisions in [Open Decisions](#open-decisions).
2. Create the Phase 0 issue from [Implementation Checklist](#implementation-checklist).
3. Capture a fresh benchmark baseline before changing router or context behavior.
4. Implement one phase at a time; do not begin typed endpoints before native `net/http` interoperability is complete.
5. Update the status table below whenever a pull request merges.

## Progress

| Phase | Outcome | Estimate | Status |
|---|---|---:|---|
| 0 | Decisions and baseline locked | 1-2 days | Not started |
| 1 | Zinc has a distinct product identity | 3-5 days | Not started |
| 2 | Routing speaks standard Go | 5-8 days | Not started |
| 3 | `net/http` interoperability is first-class | 5-8 days | Not started |
| 4 | Registration and convenience APIs are consistent | 3-5 days | Not started |
| 5 | Typed endpoints prove a distinct DX advantage | 1-2 weeks | Not started |
| 6 | Performance is protected by repeatable budgets | 2-4 days | Not started |
| 7 | Release, migration, and documentation are complete | 3-5 days | Not started |

Estimated total: 5-7 focused engineering weeks, including API review, tests, documentation, and one iteration on the typed endpoint experiment.

## Executive Decision

Zinc should become:

> A high-performance application layer for `net/http`.

The product should not present itself as an Express-inspired framework or as Fiber implemented on top of the standard library.

Fiber's identity is an Express-style developer experience built on `fasthttp`. Zinc's identity should be productive, idiomatic Go that remains native `net/http` from the server boundary through middleware, routing, request context, and handlers.

The differentiating promise is:

> Standard Go contracts, framework-level convenience, and no alternate HTTP ecosystem.

## Product Principles

Every API decision should be tested against these principles, in order.

### 1. Native Go Is the Default

- `App` remains an `http.Handler`.
- Standard `http.Handler` values can be registered without translation.
- Standard `func(http.Handler) http.Handler` middleware can be composed without translation.
- Route parameters are visible through `http.Request.PathValue`.
- Request cancellation and scoped data use `request.Context()`.
- Files use `fs.FS`.
- Cookies, headers, response streaming, server lifecycle, and protocol behavior preserve standard-library contracts.

Convenience APIs may sit above these contracts, but must not replace or obscure them.

### 2. The Golden Path Is Minimal

The ordinary API handler should require little ceremony:

```go
app.Get("/users/{id}", func(c *zinc.Context) error {
    user, err := users.Find(c.Context(), c.Param("id"))
    if err != nil {
        return err
    }
    return c.JSON(user)
})
```

The native handler path should be equally obvious:

```go
app.HandleHTTP("GET /metrics", promhttp.Handler())
```

Users should not need adapters, assertions, reflection-heavy handler dispatch, or framework-specific server infrastructure.

### 3. Compile-Time Clarity Beats Dynamic Cleverness

- Prefer narrow function signatures over `any`.
- Prefer explicit helpers such as `zinc.Text("ok")` over magical string handlers.
- Do not accept a large collection of unrelated handler signatures through runtime adaptation.
- Programmer configuration errors should fail at startup, not silently omit routes.
- Data-driven registration should retain an explicit error-returning path.

### 4. Performance Is a Constraint, Not the Personality

Zinc should be able to claim:

- Zero-allocation hot routing where practical.
- Competitive or leading realistic API latency against Gin, Chi, and Echo.
- Predictable allocation budgets.
- Standard `net/http` production throughput without a transport conversion layer.

Benchmark wins support the product. They are not the primary product identity.

### 5. Escape Hatches Are Designed In

Users must always be able to:

- Access `http.ResponseWriter`.
- Access and replace `*http.Request`.
- Use `http.Handler` and standard middleware.
- Own and configure `http.Server`.
- Use standard request contexts and typed context keys.
- Replace binding, validation, rendering, JSON encoding, and error handling.

The test for lock-in is simple: a Zinc handler or middleware should be able to delegate to standard Go without reconstructing the request or response.

## Current-State Assessment

### Existing Strengths to Preserve

Zinc already has valuable technical foundations:

- `App` implements `http.Handler`.
- The server uses `http.Server`.
- `Context` exposes the original `http.ResponseWriter` and `*http.Request`.
- Standard handlers can be mounted through `Mount`, `Wrap`, and `WrapFunc`.
- The router has fast static and dynamic paths.
- Contexts are pooled.
- Middleware chains are precomposed on common paths.
- Static files use standard `fs.FS`.
- Binding, validation, rendering, and JSON behavior are replaceable.
- The route metadata and named-route APIs are useful application-level features.
- The full test suite, race detector, and `go vet` passed during the 2026-07-29 audit.

These are assets. The plan should evolve their public shape without discarding the optimized internals.

### Current Similarity Risk

Zinc presently overlaps Fiber in several visible ways:

- "Express-inspired" positioning.
- `:param` and `*wildcard` route syntax.
- `func(*Context) error` as the central handler shape.
- `Map`, `Next`, `Status`, `JSON`, `Bind`, `Query`, groups, and middleware chains.
- Configuration names such as `ServerHeader`, `CaseSensitive`, `StrictRouting`, `BodyLimit`, `ProxyHeader`, and `ErrorHandler`.
- A broad middleware catalogue with repeated `WithConfig` constructors.
- A context object containing a growing collection of request and response helpers.

None of these elements is individually disqualifying. Together, they create the impression that Zinc is a `net/http` implementation of Fiber or Echo semantics.

### Current DX Problems

The differentiation work should also correct several practical inconsistencies:

- `Get` accepts `...any` to support string handlers, while the other route methods accept typed handlers.
- Quick-start examples ignore route registration errors.
- Quick-start examples ignore `Listen` errors.
- A registration error can therefore be missed even though the route was not installed.
- Route parameters are not populated through `Request.SetPathValue`.
- Standard middleware cannot be applied as a first-class Zinc middleware type.
- Standard handlers are mountable but are not the obvious route-level alternative.
- Default case-insensitive routing differs from normal `net/http` expectations.

## Intended Public Identity

### Positioning

Primary:

> Zinc is a high-performance application layer for `net/http`.

Expanded:

> Zinc adds fast routing, structured errors, binding, validation, response helpers, and production middleware while keeping ordinary `net/http` handlers, middleware, contexts, servers, and tooling usable.

Short comparison:

> Fiber creates an Express-style HTTP environment on `fasthttp`. Zinc makes the standard Go HTTP environment productive.

### Audience

Primary users:

- Go developers who prefer `net/http` contracts but want less application boilerplate.
- Teams that require compatibility with the standard Go observability, authentication, RPC, and middleware ecosystem.
- API services that need framework convenience without accepting a separate transport abstraction.
- Teams that care about both latency and long-term replaceability.

Not the primary target:

- Developers specifically seeking an Express clone.
- Users choosing a framework solely for the largest middleware catalogue.
- Projects that want the framework to own databases, service provisioning, outbound clients, and application architecture.

## Target API Shape

The examples in this section are design sketches. Exact names must be confirmed during Phase 0.

### Convenience Handler

Keep a narrow Zinc-native handler:

```go
type HandlerFunc func(*Context) error
```

This shape is useful for central error handling and concise response helpers. It does not need to be removed merely because Fiber and Echo use similar forms.

The differentiation comes from the surrounding contracts and the deliberate limits placed on the context API.

### Standard Handler Registration

Add a direct, obvious route for standard handlers:

```go
app.HandleHTTP("GET /metrics", promhttp.Handler())

app.HandleHTTP("GET /legacy/{id}", http.HandlerFunc(
    func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        // ...
    },
))
```

Possible names:

- `HandleHTTP(pattern string, handler http.Handler)`
- `HTTP(pattern string, handler http.Handler)`
- `Handle(pattern string, handler http.Handler)` if the existing `Handle(RouteSpec)` API is renamed

Selection criteria:

- Reads naturally next to `http.ServeMux`.
- Does not rely on `any`.
- Does not conflict confusingly with route metadata registration.
- Can be mirrored on `Group`.

### Standard Middleware Registration

Add direct support for the standard middleware shape:

```go
type HTTPMiddleware func(http.Handler) http.Handler

app.UseHTTP(requestTracing)
app.UseHTTP(authentication)
```

Required semantics:

- Preserves the original request and response writer.
- Can wrap the whole application.
- Has a documented ordering relationship with Zinc-native middleware.
- Supports group or prefix scoping if that can be done without rebuilding or translating requests.
- Does not add allocations to requests that do not use standard middleware.

Ordering must be explicit. A candidate model:

```text
outer HTTP middleware
  -> Zinc global middleware
    -> Zinc group/route middleware
      -> route handler
```

If group-scoped HTTP middleware introduces excessive complexity, launch with application-wide `UseHTTP` and add scoped composition later.

### Standard Route Syntax

Canonical syntax:

```go
app.Get("/users/{id}", handler)
app.Get("/files/{path...}", handler)
```

Canonical parameter access:

```go
c.Param("id")
c.Request().PathValue("id")
```

Both must return the same decoded value.

Compatibility plan:

- Continue parsing `:id` and `*path` during a deprecation window.
- Change all first-party documentation and examples to braces.
- Optionally emit a development-time diagnostic when legacy syntax is registered.
- Remove legacy syntax only in a release that clearly permits breaking changes.

Constraint decision:

Current route constraints should not dictate the public identity. Options:

1. Retain them only for legacy syntax.
2. Design a brace-compatible extension.
3. Move constraints to binding and validation, keeping routing structural.

Option 3 is preferred unless benchmarks or real use cases demonstrate that routing constraints are necessary. Structural routing plus typed binding is simpler and more idiomatic.

### Typed Text Handler

Replace:

```go
app.Get("/", "Hello, world!")
```

With:

```go
app.Get("/", zinc.Text("Hello, world!"))
```

Potential helpers:

```go
zinc.Text("ok")
zinc.JSONValue(value)
zinc.NoContent(http.StatusNoContent)
zinc.RedirectTo(http.StatusTemporaryRedirect, "/new")
```

Only add helpers with a common, unambiguous use case. Do not create a second response DSL.

### Typed Endpoint Layer

The typed endpoint layer is the main future DX experiment:

```go
type CreateUser struct {
    TeamID string `path:"teamID"`
    Name   string `json:"name"`
}

app.Post("/teams/{teamID}/users", zinc.Endpoint(
    func(ctx context.Context, input CreateUser) (User, error) {
        return users.Create(ctx, input)
    },
))
```

Responsibilities:

- Bind path, query, header, and body values.
- Validate input through the configured validator.
- Invoke a typed application function with `context.Context`.
- Encode the successful response.
- Send failures through the existing central error handler.

Non-goals:

- Dependency injection container.
- Controller classes.
- Runtime handler signature discovery.
- OpenAPI generation in the first iteration.
- Replacing ordinary Zinc or `net/http` handlers.

Potential signatures to prototype:

```go
func Endpoint[Input, Output any](
    fn func(context.Context, Input) (Output, error),
    options ...EndpointOption,
) HandlerFunc
```

```go
func Endpoint0[Output any](
    fn func(context.Context) (Output, error),
    options ...EndpointOption,
) HandlerFunc
```

Prototype this outside the core routing API first. If the API is not clearly simpler in two realistic applications, do not ship it.

## Registration Error Strategy

Route definitions in source code are programmer configuration. Ignoring an error during startup is worse DX than failing immediately.

Preferred model:

```go
app.Get("/users/{id}", handler) // panics on invalid/conflicting pattern
```

This follows the behavior developers already encounter with `http.ServeMux`.

Provide an explicit error-returning API for dynamic input:

```go
err := app.TryHandle(RouteSpec{
    Method:  http.MethodGet,
    Path:    patternFromConfig,
    Handler: handler,
})
```

Alternative model:

- Store the first registration error on `App`.
- Return it from `Listen`, `Serve`, and `Err`.

This avoids panics but delays feedback and allows invalid applications to exist. Use it only if API review rejects startup panics.

Required tests:

- Invalid patterns fail deterministically.
- Duplicate routes fail deterministically.
- Conflicting patterns fail deterministically.
- Dynamic registration returns errors rather than panicking.
- The server cannot start with a stored registration error if the alternative model is selected.

## Context Boundary

The context should become smaller and more standard-library-oriented over time.

### Keep

- `Request() *http.Request`
- `Writer() http.ResponseWriter`
- `Context() context.Context`
- Route parameter access
- Binding and validation entry points
- Central response helpers
- `Next()` for Zinc-native middleware
- Route metadata

### Integrate More Deeply

- Set matched parameters through `Request.SetPathValue`.
- Make request-scoped values available through `request.Context()`.
- Prefer typed helper functions for values rather than adding more `GetString`, `GetInt`, and map variants.
- Document context lifetime and pooled-context safety prominently.
- Preserve `Copy` only if there are demonstrated safe asynchronous use cases.

### Avoid

- Implementing dozens of conversions on `Context`.
- Making `Context` imitate Express request and response objects.
- Adding methods that already have a clear standard-library equivalent.
- Encouraging users to pass the pooled Zinc context into background goroutines.
- Implementing `context.Context` directly unless its lifetime and cancellation behavior are fully correct and demonstrably better than `c.Request().Context()`.

## Middleware Direction

Middleware names such as CORS, recovery, compression, request IDs, and authentication are common across Go. The names are not the differentiation problem.

The intended distinction is:

- Middleware operates on native `net/http` requests and responses.
- Middleware interoperates with `slog`, OpenTelemetry, Prometheus, `httputil`, and standard context values.
- Each middleware documents which standard contracts it reads or changes.
- Middleware does not depend on Zinc-specific storage when a request context or standard header is sufficient.

Audit every first-party middleware for:

- Request cancellation behavior.
- Response interface preservation, including flushing, hijacking, and unwrapping.
- Context-value compatibility.
- Header and status semantics.
- Allocation and latency overhead.

Do not pursue a middleware-count race. Prioritize a smaller set of production-grade middleware with strong contract tests.

## Jobs and Other Add-ons

The jobs package can remain useful, but it should not lead the differentiation story yet.

Near-term treatment:

- Keep it optional and clearly separate from the HTTP core.
- Remove it from the first screen of the README.
- Describe it as a first-party add-on, not evidence that Zinc owns the whole application runtime.
- Avoid coupling jobs to `App` or HTTP lifecycle types.

Revisit after the core identity is established. A separate module may be appropriate if dependency or release coupling becomes costly.

## Performance Contract

### What to Claim

Prefer:

> Zinc targets zero-allocation hot routing and leading realistic API performance while remaining native `net/http`.

Avoid:

> Zinc wins most benchmark rows.

The former is a durable engineering promise. The latter changes with hardware, dependency versions, benchmark selection, and noise.

### Benchmark Categories

Maintain no more than five top-level categories:

1. Routing: static, parameter, wildcard, not found, and method mismatch.
2. Middleware: zero, one, five, and ten handlers.
3. API paths: query plus JSON, binding, validation, and error responses.
4. Scale: large route sets, cold paths, registration time, and memory.
5. Production throughput: real TCP, multiple concurrency levels, and representative payloads.

### Comparison Set

Primary:

- `net/http.ServeMux`
- Gin
- Chi
- Echo

Optional:

- `httprouter` as a router-only reference

Fiber should be reported separately because its `fasthttp` transport is not an equivalent `net/http` execution path. Do not run Fiber through a `net/http` adaptor and describe that as Fiber's native performance.

### Required Benchmark Method

- Pin all dependency versions.
- Record Go version, OS, architecture, CPU, and commit.
- Run at least five samples for release claims.
- Use `benchstat` rather than selecting the best run.
- Report `ns/op`, `B/op`, and `allocs/op`.
- Keep competitor semantics equivalent.
- Publish the complete benchmark source.
- Separate in-process handler measurements from TCP throughput.
- Avoid reusing mutable requests where a framework's documented contract does not permit it.
- Add regression budgets only after stable baselines are collected.

### Initial Budgets

These are proposed guardrails, not final thresholds:

| Path | Latency regression | Allocation regression |
|---|---:|---:|
| Static route | No more than 10% | Remain at 0 allocs |
| Parameter route | No more than 10% | Remain at 0 allocs where possible |
| Five middleware | No more than 10% | Remain at 0 allocs on native path |
| API happy path | No more than 10% | No more than +1 alloc |
| Standard-handler route | Establish baseline | Avoid adapter-only allocations |

Any intentional budget exception must be described in the pull request with the measured DX or correctness benefit.

## Documentation Plan

### README

The first screen should contain:

- New positioning.
- One Zinc-native handler.
- One standard `http.Handler`.
- One standard middleware integration.
- One small performance statement.
- Links to migration, interoperability, and benchmarks.

Remove from the first screen:

- "Express-inspired."
- String handler magic.
- The complete middleware catalogue.
- Jobs.
- A large configuration block.
- Benchmark win counts.

### Guides

Add or revise:

- Why Zinc.
- Zinc and `net/http`.
- Routing with standard Go patterns.
- Using standard handlers.
- Using standard middleware.
- Context lifetime and cancellation.
- Migrating from legacy route syntax.
- Performance methodology.

### Comparison Page

The comparison should be factual, not adversarial:

| Area | Zinc | Fiber |
|---|---|---|
| HTTP engine | `net/http` | `fasthttp` |
| Canonical route syntax | Go brace patterns | Express colon patterns |
| Standard handlers | Native execution | Adapted compatibility |
| Standard middleware | First-class composition | Adapted compatibility |
| Request context | `*http.Request.Context()` | Fiber context bridge/interface |
| Primary abstraction | Thin application layer | Full framework environment |
| Performance goal | Fastest practical native `net/http` DX | Maximum framework/transport speed |

Avoid claiming that Fiber cannot use `net/http`; Fiber v3 can adapt standard handlers. Zinc's claim is that no adaptation boundary exists.

## Migration Strategy

Zinc is pre-v1, so this is the best time to correct the public shape. Breaking changes should still be deliberate.

### Route Syntax

1. Add brace syntax and `SetPathValue`.
2. Convert first-party tests, examples, and documentation.
3. Continue accepting colon and star syntax for one documented transition release.
4. Decide whether legacy syntax remains permanently as compatibility or is removed before v1.

### Handler Registration

1. Add typed handler helpers.
2. Deprecate string handlers.
3. Replace `Get(...any)` with `Get(...HandlerFunc)`.
4. Keep a temporary compatibility helper only if real users need it.

### Registration Errors

1. Introduce the chosen fail-fast behavior.
2. Add the dynamic error-returning registration API.
3. Update every example so startup errors cannot be ignored.
4. Document duplicate and conflicting route behavior.

### Configuration

Do not rename common HTTP concepts solely to look different from Fiber. Rename only when the existing name conflicts with standard Go expectations.

Review:

- Whether routing should default to case-sensitive.
- Whether strict routing needs to be public configuration after adopting Go patterns.
- Whether automatic `HEAD` and `OPTIONS` behavior matches documented HTTP semantics.
- Whether server timeouts belong in Zinc config or in user-owned `http.Server` examples.
- Whether `NewWithConfig` should remain or be simplified before v1.

## Implementation Checklist

### Phase 0: Decisions and Baseline

- [ ] Approve the positioning statement.
- [ ] Choose the canonical standard-handler registration name.
- [ ] Choose the standard-middleware registration name and ordering.
- [ ] Choose fail-fast or stored registration errors.
- [ ] Decide the legacy route-syntax deprecation window.
- [ ] Decide whether constraints remain in the router.
- [ ] Capture benchmark results with five samples.
- [ ] Record API surface with `go doc`.
- [ ] Tag or record the baseline commit.

Exit criterion: all blocking decisions are recorded in this document and the benchmark baseline is committed or attached to a tracking issue.

### Phase 1: Product Identity

- [ ] Replace "Express-inspired" positioning.
- [ ] Rewrite the README first screen.
- [ ] Move jobs and the middleware catalogue below the core story.
- [ ] Add a concise "Why Zinc" document.
- [ ] Add a factual Zinc-versus-Fiber explanation.
- [ ] Remove string-handler usage from examples.

Exit criterion: a new visitor can explain Zinc's difference from Fiber without mentioning implementation internals or benchmark row counts.

### Phase 2: Standard Go Routing

- [ ] Parse `{name}` segment parameters.
- [ ] Parse `{name...}` tail wildcards.
- [ ] Define route precedence and conflict behavior.
- [ ] Populate `Request.SetPathValue`.
- [ ] Make `Context.Param` read the same matched values.
- [ ] Preserve named-route URL generation.
- [ ] Add route parser unit tests.
- [ ] Add conflict and precedence tests.
- [ ] Add interoperability tests using `Request.PathValue`.
- [ ] Benchmark brace patterns against the existing router.
- [ ] Convert first-party examples and docs.
- [ ] Add legacy syntax tests for the transition period.

Exit criterion: all documented routes use Go brace patterns, and an unmodified `http.Handler` can read Zinc route parameters through `r.PathValue`.

### Phase 3: Native HTTP Interoperability

- [ ] Add direct route-level `http.Handler` registration.
- [ ] Mirror it on groups.
- [ ] Add application-wide standard middleware composition.
- [ ] Define and test middleware ordering.
- [ ] Preserve `Flusher`, `Hijacker`, `ReaderFrom`, `Pusher`, and `Unwrap` behavior.
- [ ] Add OpenTelemetry-style middleware example.
- [ ] Add Prometheus handler example.
- [ ] Add mixed native/Zinc middleware tests.
- [ ] Benchmark native handler and middleware paths.

Exit criterion: common `net/http` handlers and middleware work without adapters, request cloning, or transport conversion.

### Phase 4: API Consistency

- [ ] Add `zinc.Text`.
- [ ] Remove or deprecate `Get(...any)`.
- [ ] Make every method shortcut consistently typed.
- [ ] Implement the registration error decision.
- [ ] Add explicit dynamic registration.
- [ ] Audit `Map` usage and decide whether it remains a convenience alias.
- [ ] Audit context getter proliferation.
- [ ] Update all examples to handle server errors.
- [ ] Add API compile tests for intended usage.

Exit criterion: route registration is compile-time clear, invalid routes cannot be silently ignored, and all method helpers follow the same rules.

### Phase 5: Typed Endpoint Experiment

- [ ] Select two realistic APIs as design fixtures.
- [ ] Prototype typed input binding.
- [ ] Prototype typed output encoding.
- [ ] Integrate validation and HTTP errors.
- [ ] Define empty-body and no-content behavior.
- [ ] Define custom status and header behavior.
- [ ] Measure code reduction against ordinary handlers.
- [ ] Benchmark latency and allocations.
- [ ] Conduct API review before exporting from core.
- [ ] Ship in an optional package or reject the experiment.

Exit criterion: the typed API removes meaningful boilerplate, remains understandable without framework knowledge, and has an acceptable measured overhead.

### Phase 6: Performance Contract

- [ ] Add standard `ServeMux` baselines.
- [ ] Separate router, handler, and TCP suites.
- [ ] Add `benchstat` workflow documentation.
- [ ] Add allocation assertions for hot paths.
- [ ] Establish stable regression thresholds.
- [ ] Run benchmarks in repeatable CI or dedicated release hardware.
- [ ] Publish dependency versions and benchmark commands.
- [ ] Rewrite benchmark documentation around budgets and scenarios.

Exit criterion: maintainers can detect a meaningful regression before release and reproduce every public performance claim.

### Phase 7: Release

- [ ] Publish migration guide.
- [ ] Publish compatibility table.
- [ ] Add deprecation notices.
- [ ] Update all cookbook examples.
- [ ] Run full tests, race detector, vet, and benchmarks.
- [ ] Validate package documentation.
- [ ] Create release notes organized by user impact.
- [ ] Tag the differentiation release.

Exit criterion: existing users have a documented migration path and new users encounter only the new identity and canonical APIs.

## Test Strategy

### Correctness

- Router syntax, conflicts, precedence, and wildcard extraction.
- `PathValue` equivalence with `Context.Param`.
- `HEAD`, `OPTIONS`, 404, and 405 behavior.
- Request context cancellation.
- Response interface preservation.
- Middleware ordering and error propagation.
- Mixed `http.Handler` and Zinc-native routes.
- Pooled context lifetime and cleanup.

### Compatibility

Use real integrations where possible:

- `promhttp.Handler`.
- An OpenTelemetry `net/http` middleware.
- Standard `http.TimeoutHandler`.
- `http.FileServerFS`.
- `httputil.ReverseProxy`.
- A handler using `http.NewResponseController`.

### Performance

Every router or context change should run:

```bash
cd benchmarks
go test -run '^$' -bench 'Benchmark(HelloWorld|RouterParam|MiddlewareChain|APIHappyPath)$' \
  -benchmem -count=5
```

Release candidates should run the complete suite and compare through `benchstat`.

### Verification

Before merging a phase:

```bash
go test -count=1 ./...
go test -race -count=1 ./...
go vet ./...
```

Documentation examples that represent the golden path should compile in tests.

## Success Metrics

### Product Clarity

- Users describe Zinc as a productive `net/http` layer, not an Express or Fiber clone.
- The README demonstrates standard interoperability before listing framework breadth.
- Search results and package descriptions use the same positioning.

### Developer Experience

- A first route takes no more code than the current quick start.
- Standard handlers require no wrapper.
- Standard middleware requires one explicit registration call.
- Invalid source-defined routes fail immediately.
- Route parameters work through both Zinc and standard Go APIs.
- Typed endpoints, if shipped, materially reduce realistic handler boilerplate.

### Performance

- Static and common parameter routes retain zero-allocation hot paths where practical.
- No phase introduces an unexplained regression beyond the agreed budget.
- Zinc remains competitive with or faster than Gin, Chi, and Echo on realistic API paths.
- Production throughput is measured independently from in-process routing.

### Maintainability

- The core API grows more slowly than the middleware and add-on ecosystem.
- Standard-library contracts eliminate duplicate Zinc abstractions.
- Every compatibility promise has a test.
- Benchmark claims remain reproducible after dependency upgrades.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Brace syntax is perceived as copying Chi | Weak differentiation | Position it as the Go 1.22+ standard and implement `PathValue` interoperability |
| Supporting legacy and new syntax complicates the router | Performance or maintenance cost | Parse to one internal representation at registration time |
| Standard middleware ordering becomes confusing | DX regressions | Launch with one documented ordering model and application-wide scope first |
| Panic-on-registration surprises users | Migration friction | Restrict panics to source-defined registration and provide `TryHandle` |
| Typed endpoints become a large abstraction | Loss of simplicity | Prototype separately, define non-goals, and reject if two fixtures do not improve |
| Performance work dominates product work | Identity remains unclear | Finish positioning and interoperability before further micro-optimization |
| Context API keeps growing | Fiber/Echo similarity returns | Require a standard-library-gap justification for each new method |
| Jobs dilute the HTTP story | Confused positioning | Keep jobs optional and outside the README's primary narrative |

## Open Decisions

These decisions block implementation and should be resolved in Phase 0.

### Decision 1: Route Registration Name for `http.Handler`

Candidates:

- `HandleHTTP(pattern, handler)`
- `HTTP(pattern, handler)`
- Rename existing `Handle(RouteSpec)` and use `Handle(pattern, handler)`

Recommendation: `HandleHTTP` for the first release. It is explicit and avoids a rushed rename of route metadata APIs.

### Decision 2: Registration Errors

Candidates:

- Panic for source-defined routes plus `TryHandle` for dynamic routes.
- Store errors on `App` and reject `Listen` or `Serve`.

Recommendation: fail immediately for ordinary registration. It is simple, visible, and consistent with `http.ServeMux`.

### Decision 3: Legacy Syntax

Candidates:

- Permanent alias.
- One-release deprecation.
- Remove before the next release.

Recommendation: support for one transition release, then decide based on actual adoption and router complexity.

### Decision 4: Route Constraints

Candidates:

- Preserve the existing syntax.
- Create brace-compatible syntax.
- Move type constraints into binding and validation.

Recommendation: move application validation out of routing unless a route genuinely requires disambiguation.

### Decision 5: Default Route Semantics

Review:

- Case sensitivity.
- Trailing-slash behavior.
- Automatic `HEAD`.
- Automatic `OPTIONS`.
- Method-not-allowed handling.

Recommendation: defaults should match standard `net/http` expectations wherever that does not create a correctness or security problem.

### Decision 6: Typed Endpoint Package

Candidates:

- Core `zinc.Endpoint`.
- Separate `github.com/0mjs/zinc/endpoint` package.
- Separate experimental module.

Recommendation: prototype in a separate package, then promote only after API and performance validation.

## Decision Log

Record decisions here so future work does not reopen settled questions without new evidence.

| Date | Decision | Reason | Owner |
|---|---|---|---|
| 2026-07-30 | Proposed positioning: "A high-performance application layer for `net/http`" | Differentiates Zinc from Fiber while preserving current strengths | Pending approval |

## Pull Request Template for This Plan

Each implementation pull request should state:

```text
Plan phase:
Checklist items completed:
Public API changes:
Compatibility impact:
Benchmark before/after:
Allocation before/after:
Tests added:
Documentation updated:
Open follow-up:
```

Keep each pull request bounded to one behavior change where possible. Router syntax, `PathValue`, standard handlers, standard middleware, registration errors, and typed endpoints should not land as one large change.

## Definition of Done

This plan is complete when:

- Zinc no longer markets itself as Express-inspired.
- Go brace patterns are canonical.
- `Request.PathValue` works for Zinc routes.
- Standard handlers and middleware are first-class.
- Route registration is typed and cannot fail silently.
- Context growth is governed by standard-library compatibility.
- Typed endpoints have either shipped after validation or been explicitly rejected.
- Performance budgets are automated and public claims are reproducible.
- Migration documentation exists.
- A new user can understand Zinc's difference from Fiber from the first README screen.

## Immediate Next Action

Approve or edit the six items in [Open Decisions](#open-decisions), then create the Phase 0 tracking issue.
