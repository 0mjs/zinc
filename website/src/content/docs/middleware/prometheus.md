---
title: Prometheus
description: Expose dependency-free request metrics in Prometheus text format.
---

`Prometheus` records request counts and request duration sums.

```go
metrics := middleware.NewPrometheusMetrics()

app.Use(middleware.Prometheus(metrics))
app.Get("/metrics", middleware.PrometheusHandler(metrics))
```

The handler writes Prometheus text format:

```text
zinc_http_requests_total
zinc_http_request_duration_seconds_sum
zinc_http_request_duration_seconds_count
```

Labels include:

- `method`
- `route`
- `status`

Use a dedicated metrics instance for tests or multi-app processes.
