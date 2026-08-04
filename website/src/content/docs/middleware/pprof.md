---
title: Pprof
description: Mount standard library pprof handlers behind Zinc middleware.
---

`Pprof` mounts the standard library pprof handlers at `/debug/pprof`.

```go
app.Use(middleware.Pprof())
```

Use a custom prefix when profiling should live somewhere else.

```go
app.Use(middleware.PprofWithPrefix("/internal/pprof"))
```

:::danger[Keep profiling private]
Do not expose pprof publicly. Protect it with authentication, network restrictions, or both.
:::
