# Zinc Benchmarks

Local benchmark snapshot for Zinc against common `net/http`-based Go routers and frameworks.

- Machine: `Apple M1 Pro`
- OS: `darwin`
- Arch: `arm64`
- Metric: `ns/op`
- Interpretation: lower is better

## Legend

- `🪙` Zinc
- `⚡` HttpRouter
- `🍸` Gin
- `📣` Echo
- `🌐` ServeMux
- `🌿` Chi
- `🥇` fastest in benchmark
- `🥈` second
- `🥉` third

## Ultimate Leaderboard

Ranked by average placement across all 8 benchmarks.

| Rank | Framework | Wins | Top 3 Finishes | Avg Place | Performance Profile |
|---|---|---:|---:|---:|---|
| 1 | `⚡ HttpRouter` | 7 | 7 | 1.38 | Raw routing speed ceiling |
| 2 | `🍸 Gin` | 1 | 6 | 2.50 | Best full-framework param routing and middleware |
| 3 | `🪙 Zinc` | 0 | 8 | 2.75 | Most consistent top-3 finisher across all rows |
| 4 | `🌐 ServeMux` | 0 | 2 | 3.88 | Strong simple-route baseline, weaker on richer paths |
| 5 | `📣 Echo` | 0 | 1 | 4.50 | Competitive framework baseline, behind Gin/Zinc in this set |
| 6 | `🌿 Chi` | 0 | 0 | 6.00 | Flexible router, not a performance target |

## Ultimate Matrix

| Framework | `HelloWorld` | `StaticRoute` | `RouterParam` | `RouterParamCold` | `LargeRouteSetParam` | `LargeRouteSetParamMixed` | `JSONResponse` | `MiddlewareChain` |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `⚡ HttpRouter` | 🥇 17.6 | 🥇 18.1 | 🥇 42.3 | 🥇 46.4 | 🥇 57.6 | 🥇 74.1 | 🥇 329.7 | 690.5 |
| `🍸 Gin` | 83.9 | 85.6 | 🥈 91.7 | 🥈 97.7 | 🥈 103.4 | 🥈 124.2 | 🥉 377.4 | 🥇 373.9 |
| `🪙 Zinc` | 🥉 80.6 | 🥉 82.4 | 🥉 115.6 | 🥉 123.8 | 🥉 124.2 | 🥉 127.7 | 🥈 336.9 | 🥈 385.6 |
| `🌐 ServeMux` | 🥈 44.7 | 🥈 65.5 | 117.5 | 125.6 | 167.7 | 181.9 | 379.1 | 763.6 |
| `📣 Echo` | 119.1 | 121.6 | 134.0 | 129.5 | 149.1 | 160.1 | 414.2 | 🥉 512.8 |
| `🌿 Chi` | 177.2 | 174.7 | 315.8 | 217.0 | 380.3 | 289.8 | 492.7 | 848.4 |

## Benchmark Winners

| Benchmark | Winner | Runner-up | Third |
|---|---|---|---|
| `HelloWorld` | `⚡ HttpRouter` `17.6` | `🌐 ServeMux` `44.7` | `🪙 Zinc` `80.6` |
| `StaticRoute` | `⚡ HttpRouter` `18.1` | `🌐 ServeMux` `65.5` | `🪙 Zinc` `82.4` |
| `RouterParam` | `⚡ HttpRouter` `42.3` | `🍸 Gin` `91.7` | `🪙 Zinc` `115.6` |
| `RouterParamCold` | `⚡ HttpRouter` `46.4` | `🍸 Gin` `97.7` | `🪙 Zinc` `123.8` |
| `LargeRouteSetParam` | `⚡ HttpRouter` `57.6` | `🍸 Gin` `103.4` | `🪙 Zinc` `124.2` |
| `LargeRouteSetParamMixed` | `⚡ HttpRouter` `74.1` | `🍸 Gin` `124.2` | `🪙 Zinc` `127.7` |
| `JSONResponse` | `⚡ HttpRouter` `329.7` | `🪙 Zinc` `336.9` | `🍸 Gin` `377.4` |
| `MiddlewareChain` | `🍸 Gin` `373.9` | `🪙 Zinc` `385.6` | `📣 Echo` `512.8` |

## Takeaways

- `⚡ HttpRouter` still dominates 7 of 8 benchmarks and remains the routing speed ceiling.
- `🍸 Gin` wins `MiddlewareChain` and leads the full-framework race on average placement.
- `🪙 Zinc` is top-3 in all 8 rows and takes runner-up in `JSONResponse` and `MiddlewareChain`.
- `🌐 ServeMux` remains very strong for simple/static routes and a solid baseline.
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

- These are local measurements, not cross-machine absolutes.
- Throughput-style benchmarks are noisier than these in-process `ns/op` comparisons.
- `HelloWorld` includes Zinc's root fast path, so `StaticRoute` is the better general static-route comparison.
- This snapshot came from the optimized routing pass. Rerun it after major framework changes before publishing it as canonical.
