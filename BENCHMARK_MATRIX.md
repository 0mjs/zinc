# Zinc GitHub Routing Cache Matrix

This benchmark separates Zinc's public Gin-suite-style result from the workloads needed to judge whether an optimization generalizes.

The matrix uses Zinc's local recreation of the 203-route GitHub API corpus.

## Run It

From the repository root:

```sh
cd benchmarks
go test -run=^$ -bench '^BenchmarkZincGitHubCacheMatrix$' -benchmem -count=10
```

For a quick iteration:

```sh
cd benchmarks
go test -run=^$ -bench '^BenchmarkZincGitHubCacheMatrix$' -benchmem -benchtime=200ms -count=3
```

Run the correctness corpus separately:

```sh
cd benchmarks
go test -run '^TestZincGitHubCacheMatrixCorpus$'
```

## The Four Workloads

| Workload | Cache | Request population | Purpose |
| :------- | :---- | :----------------- | :------ |
| `DefaultCache` | Default 1,000 entries | Fixed 203 requests | Closest local representation of the public warmed benchmark |
| `CacheDisabled` | Disabled | Same fixed 203 requests | Measures the underlying route tree without memoized concrete paths |
| `HighCardinality` | Default 1,000 entries | 3,248 requests with changing parameter values | Exceeds cache capacity and exposes churn from real ID-like traffic |
| `ParallelHighCardinality` | Default 1,000 entries | Same 3,248 requests | Exposes cache contention and concurrent replacement behavior |

Every request is proven to return `200 OK` with an empty body before timing begins. The high-cardinality corpus is also tested with the cache both enabled and disabled.

## Initial Baseline

System:

- Apple M1 Pro
- Go 1.26.1
- 10 samples
- Zero application changes beyond adding this benchmark

The table reports the median of ten samples. Ranges show the minimum and maximum sample.

| Workload | Median | Range | B/op | allocs/op |
| :------- | -----: | :---- | ---: | --------: |
| `DefaultCache` | 110.10 ns/op | 105.8–120.3 | 0 | 0 |
| `CacheDisabled` | 123.15 ns/op | 121.0–139.1 | 0 | 0 |
| `HighCardinality` | 430.25 ns/op | 417.8–457.9 | 216 | 0* |
| `ParallelHighCardinality` | 388.45 ns/op | 384.4–419.8 | 27–29 | 0* |

`*` Go's benchmark output rounds allocations per operation to an integer. The non-zero bytes show that cache replacement still creates amortized heap traffic.

## First Cache Optimization

The first implementation replaces the growing FIFO key slice with a fixed eviction ring. Once the cache reaches capacity, it also admits only one of every four misses, using sixteen path-sharded counters so parallel requests do not contend on one global sampling counter.

This protects established entries from one-hit URLs while still allowing the cache to adapt.

Ten-sample medians before and after:

| Workload | Baseline | Optimized | Change | Baseline B/op | Optimized B/op |
| :------- | -------: | --------: | -----: | ------------: | -------------: |
| `DefaultCache` | 110.10 ns | 105.15 ns | **4.5% faster** | 0 | 0 |
| `CacheDisabled` | 123.15 ns | 121.80 ns | 1.1% faster | 0 | 0 |
| `HighCardinality` | 430.25 ns | 236.90 ns | **44.9% faster** | 216 | **0** |
| `ParallelHighCardinality` | 388.45 ns | 337.05 ns | **13.2% faster** | 27–29 | **0** |

The cache-disabled change is within expected run noise; the optimization does not alter radix matching.

The exact external Gin suite was then rerun in full:

| External workload | Before | Optimized | Change | Optimized allocs |
| :---------------- | -----: | --------: | -----: | ---------------: |
| GitHub API, 203 routes | 22,517 ns | 20,702 ns | **8.1% faster** | 0 |
| Parse API, 26 routes | 1,963 ns | 1,910 ns | 2.7% faster | 0 |
| Google+ API, 13 routes | 1,169 ns | 1,160 ns | 0.8% faster | 0 |
| Static, 157 routes | 7,410 ns | 7,352 ns | 0.8% faster | 0 |
| Twenty parameters | 123.0 ns | 121.8 ns | 1.0% faster | 0 |

Zinc remains fifth on the fresh GitHub workload: HttpRouter scored 19,626 ns/op in the same run, 5.5% ahead of Zinc, while allocating 13,792 B and 167 objects. Among native `net/http` routers, Zinc remains the fastest zero-allocation option outside Gin, BunRouter, and Echo.

The sixteen admission shards add 80 bytes to each constructed Zinc app after alignment. That is a 0.06% increase for the GitHub route set and remains below the 2% memory guardrail.

## Static Fast-Path Optimization

Profiling the GitHub static route showed that a successful static request checked the negative route cache before consulting the exact static map. A dedicated cache-enabled/cache-disabled A/B benchmark measured that tax directly.

The dispatch order now checks the exact static map first. Successful static routes return without constructing a cache key or loading the atomic hot-cache entry. Dynamic routes, misses, invalidation, and cache admission behavior are unchanged.

Ten-sample static A/B medians:

| Workload | Before | Optimized | Change | B/op |
| :------- | -----: | --------: | -----: | ---: |
| Static hit, default cache | 45.13 ns | 39.81 ns | **11.8% faster** | 0 |
| Static hit, cache disabled | 39.43 ns | 39.58 ns | Within noise | 0 |

The four-way matrix remained inside its 3% noise guardrail and retained zero B/op in every row.

The complete external Gin suite was rerun again:

| External workload | Before | Optimized | Change | Optimized allocs |
| :---------------- | -----: | --------: | -----: | ---------------: |
| GitHub static route | 42.60 ns | 38.79 ns | **8.9% faster** | 0 |
| Static, 157 routes | 7,352 ns | 6,721 ns | **8.6% faster** | 0 |
| GitHub API, 203 routes | 20,702 ns | 19,912 ns | **3.8% faster** | 0 |

Fresh same-run GitHub ranking:

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | Gin | 13,976 | 0 | 0 |
| 2 | BunRouter | 14,927 | 0 | 0 |
| 3 | Echo | 16,154 | 0 | 0 |
| **4** | **Zinc** | **19,912** | **0** | **0** |
| 5 | HttpRouter | 25,199 | 13,792 | 167 |

Zinc now completes the GitHub workload with 21.0% lower latency than HttpRouter while avoiding all of its request allocations. Run-to-run machine conditions varied materially, so the ranking is meaningful within this shared run; the 3.8% Zinc before/after improvement should be confirmed again when making a release claim.

## Context Lifecycle Optimization

A CPU profile attributed a meaningful share of request time to context acquisition, reset, release, and the surrounding `sync.Pool` operations. Inspection found that eight request-owned fields were cleared during `release`, then cleared a second time immediately after the same context was acquired.

Zinc must clear request references before returning a context to the pool so it does not retain bodies, requests, handlers, or parsed values. The optimization preserves that release-time cleanup and removes only the duplicate reset writes. A dedicated reuse test dirties request, response, route, parameter, body, cookie, handler, error, and store state before proving that none survives the release/acquire boundary.

Ten-sample lifecycle medians:

| Workload | Before | Optimized | Change | B/op | allocs/op |
| :------- | -----: | --------: | -----: | ---: | --------: |
| Context acquire + release | 16.35 ns | 14.60 ns | **10.7% faster** | 0 | 0 |

The stable post-change routing matrix retained zero request allocations:

| Workload | Median | B/op | allocs/op |
| :------- | -----: | ---: | --------: |
| `DefaultCache` | 108.70 ns | 0 | 0 |
| `CacheDisabled` | 119.05 ns | 0 | 0 |
| `HighCardinality` | 231.00 ns | 0 | 0 |
| `ParallelHighCardinality` | 313.70 ns | 0 | 0 |
| Static hit, default cache | 37.43 ns | 0 | 0 |
| Static hit, cache disabled | 37.52 ns | 0 | 0 |

One later paired matrix attempt was excluded because the machine entered visible thermal throttling midway through the post-change run: unrelated cache-disabled and static rows slowed abruptly by 30–50%. It is not used as evidence for or against the change.

The complete external Gin suite passed again:

| External workload | Previous run | Optimized | Change | Optimized allocs |
| :---------------- | -----------: | --------: | -----: | ---------------: |
| GitHub static route | 38.79 ns | 35.76 ns | **7.8% faster** | 0 |
| Static, 157 routes | 6,721 ns | 6,266 ns | **6.8% faster** | 0 |
| Parse API, 26 routes | 1,930 ns | 1,816 ns | **5.9% faster** | 0 |
| Google+ API, 13 routes | 1,181 ns | 1,126 ns | **4.7% faster** | 0 |
| GitHub API, 203 routes | 19,912 ns | 19,800 ns | Within noise | 0 |

Zinc remained fourth on the same-run GitHub workload, behind Gin at 13,624 ns, BunRouter at 14,368 ns, and Echo at 16,397 ns. HttpRouter measured 22,071 ns with 13,792 B and 167 allocations. The context change clearly helps per-request overhead, but the 203-route result confirms that large mixed-tree lookup—not context reuse—is now the dominant optimization target.

## What We Learned

### Fixed-Path Cache Benefit

The default cache reduces median latency from 123.15 ns/op to 110.10 ns/op:

- 13.05 ns/op saved
- 10.6% lower latency
- Zero request allocations in both modes

The cache helps, but the underlying radix tree is already reasonably close. The public result is not entirely produced by memoization.

### High-Cardinality Cost

Changing concrete parameter values raises the median to 430.25 ns/op:

- 3.91× the fixed-path default-cache result
- 3.49× the fixed-path cache-disabled result
- 216 B/op of amortized cache churn

This is the most important baseline. Optimizing only `DefaultCache` can improve the public score while leaving a major production weakness untouched.

### Parallel Cost

Parallel high-cardinality routing reaches 388.45 ns/op with 27–29 B/op:

- Slightly better aggregate time per operation than sequential high cardinality
- Still 3.53× the fixed-path default-cache median
- Continues to create amortized heap traffic

The current mutex-protected cache does not collapse under this eight-thread run, but replacement behavior remains much more expensive than either a cache hit or a direct radix lookup.

## Optimization Acceptance Rule

Every routing optimization must be evaluated against all four rows.

An optimization is acceptable when it:

1. Improves `DefaultCache`, or provides a documented architectural benefit.
2. Does not add request allocations to fixed-path cache or radix-tree routing.
3. Does not materially regress `CacheDisabled`.
4. Improves or preserves both high-cardinality workloads.
5. Passes the complete Zinc test suite and the external Gin correctness suite.

Treat changes below 3% as noise until repeated. Prefer changes that improve at least one workload by 5% without regressing another by more than 2%.

## First Engineering Target — Achieved

The cache replacement target was:

- Keep the roughly 10% fixed-path cache benefit.
- Remove the 216 B/op high-cardinality churn.
- Bring sequential high cardinality below 250 ns/op.
- Preserve zero allocations in fixed and cache-disabled routing.

All four conditions now pass. Zinc has also moved ahead of HttpRouter in two consecutive complete-suite runs while retaining zero allocations.

The next target is dynamic lookup in large mixed route trees. Profile and optimize `GithubAll` first, using `ParseAll`, `GPlusAll`, cache-disabled routing, and high-cardinality routing as guardrails. The context round improved static and smaller suites but moved `GithubAll` by less than 1%, so further context micro-tuning is not the highest-value work.
