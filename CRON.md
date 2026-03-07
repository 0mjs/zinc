# Cron API Notes

This is a design note for adding scheduled jobs to Zinc with minimal overhead.

## Goals

- Keep core framework lean.
- Provide very simple DX: `app.Cron(...)`.
- Allow external cron engines under the hood.
- Avoid per-request overhead.

## Proposed API

```go
app := zinc.New(
    zinc.WithCron(zinccron.New()), // optional
)

app.Cron("0 */5 * * *", func(ctx context.Context) error {
    // job code
    return nil
})
```

Advanced/custom engine path:

```go
app := zinc.New(
    zinc.WithCron(zinccron.FromEngine(myEngine)),
)
```

## Handler Type

Cron handlers should use:

```go
func(context.Context) error
```

They should **not** use `zinc.Context`, because cron jobs are not request/response scoped.

## Architecture

- Core Zinc:
  - Small scheduler interface.
  - `WithCron(...)` option.
  - `app.Cron(spec, handler)` helper method.
- Extension package (`zinccron`):
  - Default `New()` constructor.
  - Internal adapter to external cron library.
  - Optional `FromEngine(...)` for custom engines.

## Backwards Compatibility

If an implementation-specific constructor like `NewRobfig()` exists, keep it as a deprecated alias to `New()`.

## Performance Constraints

- No cron dependency required in core at runtime unless enabled.
- No cron goroutines started unless cron is configured.
- No request-path cost from cron support.

## Operational Caveat

In multi-instance deployments, each instance will run the same schedules unless distributed locking/leader election is added.
