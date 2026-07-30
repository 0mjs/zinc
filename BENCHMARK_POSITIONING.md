# Zinc Benchmark Positioning

This document compares:

1. The [current upstream Gin routing benchmark report](./GIN_GO_HTTP_ROUTING_BENCHMARK_CURRENT.md).
2. The [local full-suite run with Zinc included](./GO_HTTP_ROUTING_BENCHMARK.md).

## Bottom Line

Zinc is a **zero-allocation, high-performance `net/http` router** that sits immediately behind the established top tier on full API workloads.

Its clearest competitive position is:

- **Near-HttpRouter throughput without HttpRouter's request-time allocations.**
- **Excellent scaling as route parameter count increases.**
- **Faster static routing than Gin, BunRouter, and Echo in this run.**
- **Not yet as fast as Gin, BunRouter, or Echo across complete API route sets.**
- **Dynamic route-table memory is the most obvious weakness.**

The short positioning statement supported by this benchmark is:

> Zinc combines idiomatic `net/http` compatibility and zero-allocation request routing with throughput close to HttpRouter. It is especially strong on static and parameter-heavy routes, while aggregate API throughput and route-table memory remain the primary optimization opportunities.

## Important Comparison Rule

The upstream and Zinc-inclusive reports were produced on different systems:

| Report | CPU | Go | OS |
| :----- | :-- | :-- | :-- |
| Current upstream | Apple M4 Pro | Go 1.25.8 | Darwin 25.3.0 |
| Zinc-inclusive local run | Apple M1 Pro | Go 1.26.1 | macOS 26.3 |

Absolute `ns/op` values must **not** be compared directly across these two reports. Zinc's ranking must come from the local run, where every framework ran on the same machine under the same conditions.

The upstream report is useful for checking whether the established frameworks retain a similar order and relative shape. They do.

## Zinc's Same-Machine Results

Lower `ns/op` is better. A negative value in the “vs Gin” column means Zinc was faster.

| Workload | Zinc rank | Zinc | Leader | vs leader | vs Gin | Zinc allocs |
| :------- | :-------: | ---: | :----- | --------: | -----: | -----------: |
| Single parameter | 4 / 13 | 47.45 ns | BunRouter, 19.17 ns | 147.5% slower | 41.3% slower | 0 |
| Five parameters | 2 / 13 | 63.03 ns | Gin, 58.26 ns | 8.2% slower | 8.2% slower | 0 |
| Twenty parameters | **1 / 13** | **123 ns** | **Zinc** | — | **28.1% faster** | 0 |
| Read and write parameter | 4 / 13 | 55.67 ns | BunRouter, 40.45 ns | 37.6% slower | 18.9% slower | 0 |
| 157 static routes | **2 / 14** | **7,410 ns** | HttpRouter, 5,554 ns | 33.4% slower | **26.2% faster** | 0 |
| Parse API, 26 routes | 5 / 13 | 1,963 ns | BunRouter, 811.1 ns | 142.0% slower | 90.0% slower | 0 |
| Google+ API, 13 routes | 5 / 13 | 1,169 ns | BunRouter, 499.1 ns | 134.2% slower | 81.7% slower | 0 |
| GitHub API, 203 routes | 5 / 13 | 22,517 ns | Gin, 13,751 ns | 63.7% slower | 63.7% slower | 0 |

### What This Pattern Means

Zinc becomes more competitive as the number of parameters in one route increases:

- It moves from fourth place with one parameter to second with five.
- It takes first place with twenty parameters.
- Its request path remains allocation-free in every measured case.

That suggests Zinc's parameter extraction scales well. Its weaker aggregate API results are therefore more likely to involve route selection across larger trees, method dispatch, or route-tree layout than parameter decoding itself.

## The Most Representative Workload

The suite treats the 203-route GitHub API benchmark as its best approximation of real-world routing.

| Rank | Router | ns/op | Relative to Gin | B/op | allocs/op |
| :--: | :----- | ----: | --------------: | ---: | --------: |
| 1 | Gin | 13,751 | 1.000× | 0 | 0 |
| 2 | BunRouter | 14,663 | 1.066× | 0 | 0 |
| 3 | Echo | 16,872 | 1.227× | 0 | 0 |
| 4 | HttpRouter | 22,308 | 1.622× | 13,792 | 167 |
| **5** | **Zinc** | **22,517** | **1.637×** | **0** | **0** |
| 6 | HttpTreeMux | 70,577 | 5.132× | 65,856 | 671 |

This places Zinc in a useful gap:

- Zinc is only **0.94% slower than HttpRouter**.
- Zinc avoids HttpRouter's **167 allocations and 13,792 bytes** per complete GitHub workload.
- Zinc is **3.13× faster than HttpTreeMux**, the next router in the ranking.
- Zinc is **33.5% slower than Echo**, the nearest zero-allocation framework above it.

The practical tiering is:

| Tier | Routers | Interpretation |
| :--- | :------ | :------------- |
| Top | Gin, BunRouter, Echo | Best aggregate API throughput and zero allocations |
| **Near-top** | **HttpRouter, Zinc** | Similar throughput; Zinc has the allocation advantage |
| Middle | HttpTreeMux, Beego, Chi, Fiber, Macaron | Substantially slower aggregate routing |
| Long tail | Goji v2, GoRestful, GorillaMux | Highest routing overhead in this suite |

Fiber uses a separate `fasthttp` harness, so its absolute result should not be treated as directly comparable with the native `net/http` routers.

## Does the Upstream Report Support This Position?

Yes. The established router order is stable between the current upstream M4 report and the local M1 report.

The following table normalizes each result to Gin on the same machine. This is more informative than comparing raw times across machines.

| Router | Upstream rank | Upstream vs Gin | Local rank | Local vs Gin |
| :----- | ------------: | --------------: | ---------: | -----------: |
| Gin | 1 | 1.000× | 1 | 1.000× |
| BunRouter | 2 | 1.034× | 2 | 1.066× |
| Echo | 3 | 1.113× | 3 | 1.227× |
| HttpRouter | 4 | 1.514× | 4 | 1.622× |
| **Zinc** | — | — | **5** | **1.637×** |
| HttpTreeMux | 5 | 4.958× | 6 | 5.132× |

The upstream top four remain the local top four in the same order. Zinc lands almost exactly beside HttpRouter and well before HttpTreeMux. That makes Zinc's fifth-place local position structurally credible rather than an artifact of comparing unrelated machines.

It does **not** prove what Zinc would score on the upstream M4 Pro. Only a same-machine upstream run with Zinc can establish that number.

## Memory Position

Zinc allocates nothing while serving the measured requests, but building its route structures uses more heap than its strongest throughput competitors.

| Route set | Zinc memory | Rank | Gin | BunRouter | Echo | HttpRouter |
| :-------- | ----------: | :--: | --: | --------: | ---: | ---------: |
| Static | 35,392 B | **3 / 14** | 34,408 B | 51,232 B | 91,976 B | 21,680 B |
| GitHub | 138,600 B | 9 / 13 | 58,840 B | 93,776 B | 117,784 B | 37,072 B |
| Google+ | 10,752 B | 9 / 13 | 4,576 B | 7,360 B | 10,968 B | 2,776 B |
| Parse | 17,344 B | 10 / 13 | 7,896 B | 9,336 B | 13,816 B | 5,024 B |

The memory story has two parts:

- Static route storage is strong: Zinc is third overall and uses only **2.9% more memory than Gin**.
- Mixed and parameterized route sets are weaker: Zinc's GitHub structure uses **2.36× Gin's memory**, **47.8% more than BunRouter**, and **17.7% more than Echo**.

This memory is paid primarily when constructing the router, not on every request. It matters most for applications with many router instances, very large route tables, memory-constrained deployments, or frequent router reconstruction.

## Competitive Strengths

### 1. Zero-Allocation Request Path

Zinc records zero bytes and zero allocations across every reported workload. That puts its request-time memory behavior in the same class as Gin, BunRouter, and most Echo benchmarks.

### 2. Deep Parameter Scaling

Zinc wins the twenty-parameter benchmark and comes within 8.2% of Gin on five parameters. This is its strongest raw performance differentiator.

### 3. Static Routing

Zinc is second on the full static workload. It beats Gin by 26.2%, BunRouter by 14.5%, and Echo by 25.7% in the local run.

### 4. Clear Separation From the Middle

On the GitHub workload, Zinc is over three times faster than the next-ranked router after HttpRouter. Zinc is not buried in a crowded middle; it sits directly below the leading group.

## Competitive Weaknesses

### 1. Aggregate Route-Tree Throughput

Gin, BunRouter, and Echo remain materially faster across the complete API workloads. Zinc needs roughly:

- A **25.1% latency reduction** to tie Echo on GitHub.
- A **34.9% latency reduction** to tie BunRouter.
- A **38.9% latency reduction** to tie Gin.

### 2. Dynamic Route-Table Memory

Zinc's static storage is efficient, but its memory use grows less favorably when parameterized routes are added. The GitHub, Google+, and Parse structures place Zinc in the lower half of the memory ranking.

### 3. Evidence Depth

These conclusions come from one complete local run. The run is valid for positioning, but repeated samples analyzed with `benchstat` are needed before attributing small differences to implementation changes.

The 0.94% gap between Zinc and HttpRouter is too small to treat as definitive without repeated runs. The zero-allocation difference is definitive in this benchmark.

## Recommended Performance Priorities

1. **Optimize large mixed route trees.** Profile `GithubAll`, then `ParseAll` and `GPlusAll`; do not optimize the already-leading twenty-parameter path first.
2. **Reduce route construction memory.** Focus on parameterized node representation and per-route metadata while preserving static-route efficiency.
3. **Protect zero allocations.** Add benchmark or test gates so throughput work cannot silently add request-time heap allocation.
4. **Measure with repeated samples.** Run at least 10 samples and use `benchstat` before accepting changes based on single-digit percentage improvements.
5. **Preserve the product position.** Zinc's benchmark story is strongest when speed is paired with idiomatic `net/http` compatibility and minimal API design—not speed in isolation.

## Positioning Verdict

| Dimension | Zinc position |
| :-------- | :------------ |
| Full API throughput | **Fifth overall; near HttpRouter** |
| Request allocations | **Top tier; zero allocation** |
| Static routing | **Second overall** |
| Deep parameter routing | **First at twenty parameters** |
| Route-table memory | **Strong for static routes; below average for mixed routes** |
| Credible claim today | **Fast, allocation-free, idiomatic `net/http` routing with excellent parameter scaling** |
| Claim not yet supported | **Faster than Gin, BunRouter, or Echo overall** |

The benchmark supports presenting Zinc as a **minimal, idiomatic router with near-top-tier performance**, not as the outright fastest framework. Its most defensible technical differentiator is the combination of `net/http` compatibility, zero request allocations, strong static routing, and unusually good scaling for parameter-heavy routes.
