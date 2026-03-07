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
| 2 | `🍸 Gin` | 0 | 5 | 2.75 | Best full-framework param routing |
| 3 | `🪙 Zinc` | 1 | 6 | 3.00 | Best middleware result, strong overall framework contender |
| 4 | `🌐 ServeMux` | 0 | 5 | 3.50 | Quietly strong stdlib baseline on simple routes |
| 5 | `📣 Echo` | 0 | 1 | 4.38 | Competitive, but behind Gin and Zinc in this set |
| 6 | `🌿 Chi` | 0 | 0 | 6.00 | Flexible router, not a performance target |

## Ultimate Matrix

| Framework | `HelloWorld` | `StaticRoute` | `RouterParam` | `RouterParamCold` | `LargeRouteSetParam` | `LargeRouteSetParamMixed` | `JSONResponse` | `MiddlewareChain` |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `⚡ HttpRouter` | 🥇 18.4 | 🥇 19.0 | 🥇 45.3 | 🥇 53.6 | 🥇 60.1 | 🥇 76.9 | 🥇 336.2 | 748.4 |
| `🍸 Gin` | 91.5 | 93.5 | 🥈 103.1 | 🥈 107.0 | 🥈 108.4 | 🥈 130.6 | 385.4 | 🥈 403.6 |
| `🪙 Zinc` | 🥉 80.0 | 🥉 93.3 | 133.9 | 141.7 | 🥉 140.4 | 🥉 150.2 | 🥈 356.9 | 🥇 224.5 |
| `🌐 ServeMux` | 🥈 43.1 | 🥈 63.7 | 🥉 114.6 | 🥉 129.1 | 181.6 | 185.9 | 🥉 380.4 | 790.0 |
| `📣 Echo` | 124.7 | 122.1 | 137.9 | 140.9 | 153.7 | 167.7 | 414.2 | 🥉 606.2 |
| `🌿 Chi` | 189.6 | 188.1 | 361.7 | 249.9 | 451.2 | 303.4 | 510.1 | 950.1 |

## Benchmark Winners

| Benchmark | Winner | Runner-up | Third |
|---|---|---|---|
| `HelloWorld` | `⚡ HttpRouter` `18.4` | `🌐 ServeMux` `43.1` | `🪙 Zinc` `80.0` |
| `StaticRoute` | `⚡ HttpRouter` `19.0` | `🌐 ServeMux` `63.7` | `🪙 Zinc` `93.3` |
| `RouterParam` | `⚡ HttpRouter` `45.3` | `🍸 Gin` `103.1` | `🌐 ServeMux` `114.6` |
| `RouterParamCold` | `⚡ HttpRouter` `53.6` | `🍸 Gin` `107.0` | `🌐 ServeMux` `129.1` |
| `LargeRouteSetParam` | `⚡ HttpRouter` `60.1` | `🍸 Gin` `108.4` | `🪙 Zinc` `140.4` |
| `LargeRouteSetParamMixed` | `⚡ HttpRouter` `76.9` | `🍸 Gin` `130.6` | `🪙 Zinc` `150.2` |
| `JSONResponse` | `⚡ HttpRouter` `336.2` | `🪙 Zinc` `356.9` | `🌐 ServeMux` `380.4` |
| `MiddlewareChain` | `🪙 Zinc` `224.5` | `🍸 Gin` `403.6` | `📣 Echo` `606.2` |

## Takeaways

- `⚡ HttpRouter` dominates 7 of 8 benchmarks and remains the routing speed ceiling.
- `🍸 Gin` slightly edges `🪙 Zinc` on average placement because of stronger param-route performance.
- `🪙 Zinc` wins `MiddlewareChain`, places second in `JSONResponse`, and beats `📣 Echo` across most high-value paths.
- `🌐 ServeMux` is a better speed baseline than many people expect on simple and static routes.
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
