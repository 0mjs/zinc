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
| 1 | `🍸 Gin` | 4 | 8 | 1.50 | Best overall average placement in this framework set |
| 2 | `🪙 Zinc` | 4 | 8 | 1.63 | Tied for most wins; strongest middleware and JSON results |
| 3 | `📣 Echo` | 0 | 8 | 2.88 | Consistent third-place baseline with one cold-param runner-up |
| 4 | `🌿 Chi` | 0 | 0 | 4.00 | Flexible router, not a performance target |

## Ultimate Matrix

| Framework | `HelloWorld` | `StaticRoute` | `RouterParam` | `RouterParamCold` | `LargeRouteSetParam` | `LargeRouteSetParamMixed` | `JSONResponse` | `MiddlewareChain` |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `🍸 Gin` | 🥈 91.5 | 🥈 93.5 | 🥇 103.1 | 🥇 107.0 | 🥇 108.4 | 🥇 130.6 | 🥈 385.4 | 🥈 403.6 |
| `🪙 Zinc` | 🥇 80.0 | 🥇 93.3 | 🥈 133.9 | 🥉 141.7 | 🥈 140.4 | 🥈 150.2 | 🥇 356.9 | 🥇 224.5 |
| `📣 Echo` | 🥉 124.7 | 🥉 122.1 | 🥉 137.9 | 🥈 140.9 | 🥉 153.7 | 🥉 167.7 | 🥉 414.2 | 🥉 606.2 |
| `🌿 Chi` | 189.6 | 188.1 | 361.7 | 249.9 | 451.2 | 303.4 | 510.1 | 950.1 |

## Benchmark Winners

| Benchmark | Winner | Runner-up | Third |
|---|---|---|---|
| `HelloWorld` | `🪙 Zinc` `80.0` | `🍸 Gin` `91.5` | `📣 Echo` `124.7` |
| `StaticRoute` | `🪙 Zinc` `93.3` | `🍸 Gin` `93.5` | `📣 Echo` `122.1` |
| `RouterParam` | `🍸 Gin` `103.1` | `🪙 Zinc` `133.9` | `📣 Echo` `137.9` |
| `RouterParamCold` | `🍸 Gin` `107.0` | `📣 Echo` `140.9` | `🪙 Zinc` `141.7` |
| `LargeRouteSetParam` | `🍸 Gin` `108.4` | `🪙 Zinc` `140.4` | `📣 Echo` `153.7` |
| `LargeRouteSetParamMixed` | `🍸 Gin` `130.6` | `🪙 Zinc` `150.2` | `📣 Echo` `167.7` |
| `JSONResponse` | `🪙 Zinc` `356.9` | `🍸 Gin` `385.4` | `📣 Echo` `414.2` |
| `MiddlewareChain` | `🪙 Zinc` `224.5` | `🍸 Gin` `403.6` | `📣 Echo` `606.2` |

## Takeaways

- `🍸 Gin` and `🪙 Zinc` split wins `4-4` across this benchmark set.
- `🍸 Gin` leads the param-routing benchmarks on average.
- `🪙 Zinc` leads `HelloWorld`, `StaticRoute`, `JSONResponse`, and `MiddlewareChain`.
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
