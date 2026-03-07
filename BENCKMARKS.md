# Zinc Benchmarks

Local benchmark snapshot for Zinc against `Gin`, `Echo`, and `Chi` (filtered framework-only view).

- Machine: `Apple M1 Pro`
- OS: `darwin`
- Arch: `arm64`
- Metric: `ns/op`
- Interpretation: lower is better

## Legend

- `🪙` Zinc
- `🍸` Gin
- `📣` Echo
- `🌿` Chi
- `🥇` fastest in benchmark
- `🥈` second
- `🥉` third

## Ultimate Leaderboard

Ranked by average placement across all 8 benchmarks.

| Rank | Framework | Wins | Top 3 Finishes | Avg Place | Performance Profile |
|---|---|---:|---:|---:|---|
| 1 | `🍸 Gin` | 5 | 8 | 1.38 | Best overall average placement, especially param and middleware paths |
| 2 | `🪙 Zinc` | 3 | 8 | 1.63 | Wins static/hello/JSON and remains consistently close to Gin |
| 3 | `📣 Echo` | 0 | 8 | 3.00 | Consistent third-place baseline |
| 4 | `🌿 Chi` | 0 | 0 | 4.00 | Flexible router, not a performance target |

## Ultimate Matrix

| Framework | `HelloWorld` | `StaticRoute` | `RouterParam` | `RouterParamCold` | `LargeRouteSetParam` | `LargeRouteSetParamMixed` | `JSONResponse` | `MiddlewareChain` |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `🍸 Gin` | 🥈 83.9 | 🥈 85.6 | 🥇 91.7 | 🥇 97.7 | 🥇 103.4 | 🥇 124.2 | 🥈 377.4 | 🥇 373.9 |
| `🪙 Zinc` | 🥇 80.6 | 🥇 82.4 | 🥈 115.6 | 🥈 123.8 | 🥈 124.2 | 🥈 127.7 | 🥇 336.9 | 🥈 385.6 |
| `📣 Echo` | 🥉 119.1 | 🥉 121.6 | 🥉 134.0 | 🥉 129.5 | 🥉 149.1 | 🥉 160.1 | 🥉 414.2 | 🥉 512.8 |
| `🌿 Chi` | 177.2 | 174.7 | 315.8 | 217.0 | 380.3 | 289.8 | 492.7 | 848.4 |

## Benchmark Winners

| Benchmark | Winner | Runner-up | Third |
|---|---|---|---|
| `HelloWorld` | `🪙 Zinc` `80.6` | `🍸 Gin` `83.9` | `📣 Echo` `119.1` |
| `StaticRoute` | `🪙 Zinc` `82.4` | `🍸 Gin` `85.6` | `📣 Echo` `121.6` |
| `RouterParam` | `🍸 Gin` `91.7` | `🪙 Zinc` `115.6` | `📣 Echo` `134.0` |
| `RouterParamCold` | `🍸 Gin` `97.7` | `🪙 Zinc` `123.8` | `📣 Echo` `129.5` |
| `LargeRouteSetParam` | `🍸 Gin` `103.4` | `🪙 Zinc` `124.2` | `📣 Echo` `149.1` |
| `LargeRouteSetParamMixed` | `🍸 Gin` `124.2` | `🪙 Zinc` `127.7` | `📣 Echo` `160.1` |
| `JSONResponse` | `🪙 Zinc` `336.9` | `🍸 Gin` `377.4` | `📣 Echo` `414.2` |
| `MiddlewareChain` | `🍸 Gin` `373.9` | `🪙 Zinc` `385.6` | `📣 Echo` `512.8` |

## Takeaways

- `🍸 Gin` and `🪙 Zinc` remain close, with wins now split `5-3` for Gin.
- `🍸 Gin` leads param-routing and middleware in this run.
- `🪙 Zinc` leads `HelloWorld`, `StaticRoute`, and `JSONResponse`.
- `📣 Echo` stays competitive but trails `🍸 Gin` and `🪙 Zinc` on most paths.
- `🌿 Chi` remains valuable for flexibility, but not for raw speed.

## Reproduce

Focused suite:

```bash
GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'BenchmarkHelloWorld$|BenchmarkStaticRoute$|BenchmarkRouterParam$|BenchmarkRouterParamCold$|BenchmarkJSONResponse$|BenchmarkMiddlewareChain$|BenchmarkLargeRouteSetParam$|BenchmarkLargeRouteSetParamMixed$' -benchmem
```

Param-routing deep dive:

```bash
GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'BenchmarkRouterParam$|BenchmarkRouterParamCold$|BenchmarkLargeRouteSetParam$|BenchmarkLargeRouteSetParamMixed$|BenchmarkRouterFindIntoWarm$|BenchmarkRouterFindIntoCold$|BenchmarkRouterLookupRawParam$' -benchmem
```

## Notes

- This is a filtered presentation from the same snapshot, excluding `HttpRouter` and `ServeMux`.
- These are local measurements, not cross-machine absolutes.
- Throughput-style benchmarks are noisier than these in-process `ns/op` comparisons.
- `HelloWorld` includes Zinc's root fast path, so `StaticRoute` is the better general static-route comparison.
- This snapshot came from the optimized routing pass. Rerun it after major framework changes before publishing it as canonical.
