# Go HTTP Router Benchmark — Zinc Included

This is a local run of [gin-gonic/go-http-routing-benchmark](https://github.com/gin-gonic/go-http-routing-benchmark), with the current Zinc checkout added to the complete upstream suite.

The suite compares HTTP request routers by implementing real-world API routing structures. Some APIs are slightly adapted because they cannot be implemented identically in every router.

## Tested routers & frameworks

- [Beego](http://beego.me/)
- [BunRouter](https://github.com/uptrace/bunrouter)
- [Chi](https://github.com/go-chi/chi)
- [Echo](https://github.com/labstack/echo)
- [Fiber](https://github.com/gofiber/fiber) _(fasthttp-based, benchmarked separately)_
- [Gin](https://github.com/gin-gonic/gin)
- [Goji v2](https://github.com/goji/goji)
- [Go-Restful](https://github.com/emicklei/go-restful)
- [Gorilla Mux](https://github.com/gorilla/mux)
- [http.ServeMux](https://pkg.go.dev/net/http#ServeMux) _(static routes only)_
- [HttpRouter](https://github.com/julienschmidt/httprouter)
- [HttpTreeMux](https://github.com/dimfeld/httptreemux)
- [Macaron](https://github.com/go-macaron/macaron)
- [Zinc](https://github.com/0mjs/zinc)

> **Fiber note:** Fiber v2 uses fasthttp and has a separate benchmark harness. Its absolute results are listed for completeness but are not directly comparable with the native `net/http` routers.

## Reproduction

Benchmark source:

- Upstream repository: [gin-gonic/go-http-routing-benchmark](https://github.com/gin-gonic/go-http-routing-benchmark)
- Upstream commit: `ff3cdf55eccd0aa6a272991db9611a86c734dc51`
- Zinc source: `/Users/matt/dev/oss/zinc`
- Zinc commit: `05ad9894dd54c6b29e11cd1b0007ec031143747d`
- Zinc module replacement: `replace github.com/0mjs/zinc => /Users/matt/dev/oss/zinc`

Command:

```sh
go test -bench=. -benchmem -timeout=20m -count=1
```

Benchmark system:

- Apple M1 Pro
- Go 1.26.1 darwin/arm64
- macOS 26.3 (Build 25D125)
- Run date: 2026-07-30
- Total suite time: 301.224 seconds

## Results

### Summary

The table ranks every router by **GitHub API throughput**: 203 routes across all supported methods. Lower `ns/op` is better.

| Rank | Router | ns/op | B/op | allocs/op | Zero-alloc |
| :--: | :----- | ----: | ---: | --------: | :--------: |
| 1 | **Gin** | 13,624 | 0 | 0 | :white_check_mark: |
| 2 | **BunRouter** | 14,368 | 0 | 0 | :white_check_mark: |
| 3 | **Echo** | 16,397 | 0 | 0 | :white_check_mark: |
| 4 | **Zinc** | 19,800 | 0 | 0 | :white_check_mark: |
| 5 | **HttpRouter** | 22,071 | 13,792 | 167 |  |
| 6 | **HttpTreeMux** | 66,930 | 65,856 | 671 |  |
| 7 | **Chi** | 123,835 | 130,817 | 740 |  |
| 8 | **Beego** | 127,494 | 71,456 | 609 |  |
| 9 | **Fiber** | 156,388 | 0 | 0 | :white_check_mark: |
| 10 | **Macaron** | 166,764 | 147,784 | 1,624 |  |
| 11 | **Goji v2** | 312,370 | 313,744 | 3,712 |  |
| 12 | **GoRestful** | 1,155,200 | 1,006,744 | 3,009 |  |
| 13 | **GorillaMux** | 1,894,683 | 225,667 | 1,588 |  |

**Zinc takeaways:**

- Zinc ranks **4 of 13** on the 203-route GitHub workload at **19,800 ns/op**, with **0 allocations**.
- Zinc is the **fourth zero-allocation native `net/http` framework** in the GitHub test, behind Gin, BunRouter, and Echo.
- Zinc completes the 157-route static workload in **6,266 ns/op**, second only to HttpRouter in this run.
- Zinc is fastest in the 20-parameter microbenchmark at **128 ns/op**.
- Zinc uses **35,472 B** for the static route set and **138,680 B** for the GitHub route set.

---

### Memory Consumption

Heap memory retained after loading each routing structure, measured by the upstream `calcMem` helper:

| Router | Static | GitHub | Google+ | Parse |
| :----- | -----: | -----: | ------: | ----: |
| HttpServeMux | 71,728 B | — | — | — |
| Beego | 98,824 B | 150,840 B | 10,256 B | 19,256 B |
| BunRouter | 51,232 B | 93,776 B | 7,360 B | 9,336 B |
| Chi | 83,160 B | 94,888 B | 8,008 B | 9,656 B |
| Echo | 91,976 B | 117,784 B | 10,968 B | 13,816 B |
| Fiber | 59,248 B | 163,832 B | 10,840 B | 15,352 B |
| Gin | 34,408 B | 58,840 B | 4,576 B | 7,896 B |
| Goji v2 | 118,000 B | 118,640 B | 8,096 B | 16,064 B |
| GoRestful | 819,688 B | 1,270,704 B | 72,520 B | 121,184 B |
| GorillaMux | 599,496 B | 1,319,696 B | 68,000 B | 105,384 B |
| HttpRouter | 21,680 B | 37,072 B | 2,776 B | 5,024 B |
| HttpTreeMux | 73,448 B | 78,800 B | 7,440 B | 7,848 B |
| Macaron | 36,976 B | 90,632 B | 8,672 B | 13,704 B |
| Zinc | 35,472 B | 138,680 B | 10,832 B | 17,424 B |

### Static Routes

```text
BenchmarkHttpServeMux_StaticAll             58444      20,440 ns/op           0 B/op         0 allocs/op
BenchmarkBeego_StaticAll                    12817      89,920 ns/op      55,264 B/op       471 allocs/op
BenchmarkBunRouter_StaticAll               142521       8,646 ns/op           0 B/op         0 allocs/op
BenchmarkChi_StaticAll                      18949      62,252 ns/op      57,776 B/op       314 allocs/op
BenchmarkEcho_StaticAll                    120732      10,114 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_StaticAll                    28750      41,628 ns/op           0 B/op         0 allocs/op
BenchmarkGin_StaticAll                     115159      10,265 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_StaticAll                    9128     115,169 ns/op     175,840 B/op     1,099 allocs/op
BenchmarkGoRestful_StaticAll                 1728     674,227 ns/op     677,824 B/op     2,193 allocs/op
BenchmarkGorillaMux_StaticAll                3014     446,279 ns/op     133,138 B/op     1,099 allocs/op
BenchmarkHttpRouter_StaticAll              202986       5,712 ns/op           0 B/op         0 allocs/op
BenchmarkHttpTreeMux_StaticAll             143336       8,271 ns/op           0 B/op         0 allocs/op
BenchmarkMacaron_StaticAll                  10000     114,199 ns/op     114,296 B/op     1,256 allocs/op
BenchmarkZinc_StaticAll                    189236       6,266 ns/op           0 B/op         0 allocs/op
```

### Micro Benchmarks

#### Single Route with Param

```text
BenchmarkBeego_Param                      2405223         506 ns/op         352 B/op         3 allocs/op
BenchmarkBunRouter_Param                 67064552       18.58 ns/op           0 B/op         0 allocs/op
BenchmarkChi_Param                        1933922       580.1 ns/op         704 B/op         4 allocs/op
BenchmarkEcho_Param                      46266247       27.41 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_Param                      7734930       165.3 ns/op           0 B/op         0 allocs/op
BenchmarkGin_Param                       38398924       31.39 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_Param                     1636260       730.5 ns/op       1,136 B/op         8 allocs/op
BenchmarkGoRestful_Param                   469903       2,284 ns/op       4,600 B/op        15 allocs/op
BenchmarkGorillaMux_Param                 1000000       1,027 ns/op       1,152 B/op         8 allocs/op
BenchmarkHttpRouter_Param                22836553       48.75 ns/op          32 B/op         1 allocs/op
BenchmarkHttpTreeMux_Param                4763328       246.4 ns/op         352 B/op         3 allocs/op
BenchmarkMacaron_Param                    1000000       1,024 ns/op       1,064 B/op        10 allocs/op
BenchmarkZinc_Param                      27210986       44.55 ns/op           0 B/op         0 allocs/op
```

#### Route with 5 Params

```text
BenchmarkBeego_Param5                     1866901         625 ns/op         352 B/op         3 allocs/op
BenchmarkBunRouter_Param5                14963704       79.82 ns/op           0 B/op         0 allocs/op
BenchmarkChi_Param5                       1601617       752.8 ns/op         704 B/op         4 allocs/op
BenchmarkEcho_Param5                     17372526       69.02 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_Param5                     3376689       353.3 ns/op           0 B/op         0 allocs/op
BenchmarkGin_Param5                      19682953       58.73 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_Param5                    1397863       872.5 ns/op       1,200 B/op         8 allocs/op
BenchmarkGoRestful_Param5                  451645       2,423 ns/op       4,712 B/op        15 allocs/op
BenchmarkGorillaMux_Param5                 838705       1,565 ns/op       1,216 B/op         8 allocs/op
BenchmarkHttpRouter_Param5                7860688       139.1 ns/op         160 B/op         1 allocs/op
BenchmarkHttpTreeMux_Param5               2356395       511.6 ns/op         576 B/op         6 allocs/op
BenchmarkMacaron_Param5                   1000000       1,140 ns/op       1,064 B/op        10 allocs/op
BenchmarkZinc_Param5                     19723284       61.94 ns/op           0 B/op         0 allocs/op
```

#### Route with 20 Params

```text
BenchmarkBeego_Param20                     909722       1,476 ns/op         352 B/op         3 allocs/op
BenchmarkBunRouter_Param20                3118816       382.6 ns/op           0 B/op         0 allocs/op
BenchmarkChi_Param20                       405828       3,046 ns/op       2,504 B/op         9 allocs/op
BenchmarkEcho_Param20                     6002026       199.7 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_Param20                    1823250       665.3 ns/op           0 B/op         0 allocs/op
BenchmarkGin_Param20                      7097480       172.6 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_Param20                   1000000       1,165 ns/op       1,440 B/op         8 allocs/op
BenchmarkGoRestful_Param20                 247063       4,797 ns/op       7,008 B/op        20 allocs/op
BenchmarkGorillaMux_Param20                384590       3,244 ns/op       3,272 B/op        13 allocs/op
BenchmarkHttpRouter_Param20               2598848       478.7 ns/op         704 B/op         1 allocs/op
BenchmarkHttpTreeMux_Param20               400652       2,938 ns/op       3,144 B/op        13 allocs/op
BenchmarkMacaron_Param20                   407355       3,032 ns/op       2,864 B/op        15 allocs/op
BenchmarkZinc_Param20                     9703689         128 ns/op           0 B/op         0 allocs/op
```

#### Param Write

Reads the named parameter and writes it to the response.

```text
BenchmarkBeego_ParamWrite                 2202913       545.4 ns/op         360 B/op         4 allocs/op
BenchmarkBunRouter_ParamWrite            31752121        38.1 ns/op           0 B/op         0 allocs/op
BenchmarkChi_ParamWrite                   2192654       643.7 ns/op         704 B/op         4 allocs/op
BenchmarkEcho_ParamWrite                 17133963       80.99 ns/op           8 B/op         1 allocs/op
BenchmarkFiber_ParamWrite                 6842467       177.3 ns/op           0 B/op         0 allocs/op
BenchmarkGin_ParamWrite                  27260662       39.52 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_ParamWrite                1265094       816.6 ns/op       1,168 B/op        10 allocs/op
BenchmarkGoRestful_ParamWrite              470337       2,416 ns/op       4,608 B/op        16 allocs/op
BenchmarkGorillaMux_ParamWrite            1000000       1,045 ns/op       1,152 B/op         8 allocs/op
BenchmarkHttpRouter_ParamWrite           21889071       55.32 ns/op          32 B/op         1 allocs/op
BenchmarkHttpTreeMux_ParamWrite           4696696       268.5 ns/op         352 B/op         3 allocs/op
BenchmarkMacaron_ParamWrite               1000000       1,200 ns/op       1,112 B/op        13 allocs/op
BenchmarkZinc_ParamWrite                 21832554       57.69 ns/op           0 B/op         0 allocs/op
```

### Parse API — 26 Routes

```text
BenchmarkBeego_ParseAll                     85455      14,449 ns/op       9,152 B/op        78 allocs/op
BenchmarkBunRouter_ParseAll               1439668       813.2 ns/op           0 B/op         0 allocs/op
BenchmarkChi_ParseAll                       92289      13,377 ns/op      14,944 B/op        84 allocs/op
BenchmarkEcho_ParseAll                    1000000       1,060 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_ParseAll                    199063       6,078 ns/op           0 B/op         0 allocs/op
BenchmarkGin_ParseAll                     1000000       1,038 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_ParseAll                    56637      20,889 ns/op      29,456 B/op       199 allocs/op
BenchmarkGoRestful_ParseAll                 14458      79,566 ns/op     131,728 B/op       380 allocs/op
BenchmarkGorillaMux_ParseAll                29163      41,165 ns/op      26,960 B/op       198 allocs/op
BenchmarkHttpRouter_ParseAll               947623       1,516 ns/op         640 B/op        16 allocs/op
BenchmarkHttpTreeMux_ParseAll              217609       5,662 ns/op       5,728 B/op        51 allocs/op
BenchmarkMacaron_ParseAll                   55873      21,179 ns/op      18,928 B/op       208 allocs/op
BenchmarkZinc_ParseAll                     670556       1,816 ns/op           0 B/op         0 allocs/op
```

### Google+ API — 13 Routes

```text
BenchmarkBeego_GPlusAll                    168876       7,383 ns/op       4,576 B/op        39 allocs/op
BenchmarkBunRouter_GPlusAll               2471661       498.7 ns/op           0 B/op         0 allocs/op
BenchmarkChi_GPlusAll                      153986       6,919 ns/op       8,480 B/op        48 allocs/op
BenchmarkEcho_GPlusAll                    1886781       630.6 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_GPlusAll                    346766       3,483 ns/op           0 B/op         0 allocs/op
BenchmarkGin_GPlusAll                     1934193       630.5 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_GPlusAll                   120990      11,217 ns/op      15,120 B/op       115 allocs/op
BenchmarkGoRestful_GPlusAll                 37622      32,098 ns/op      60,720 B/op       193 allocs/op
BenchmarkGorillaMux_GPlusAll                54672      22,817 ns/op      14,448 B/op       102 allocs/op
BenchmarkHttpRouter_GPlusAll              1000000       1,064 ns/op         640 B/op        11 allocs/op
BenchmarkHttpTreeMux_GPlusAll              317504       3,848 ns/op       4,032 B/op        38 allocs/op
BenchmarkMacaron_GPlusAll                   99793      11,879 ns/op       9,464 B/op       104 allocs/op
BenchmarkZinc_GPlusAll                    1000000       1,126 ns/op           0 B/op         0 allocs/op
```

### GitHub API — 203 Routes

```text
BenchmarkBeego_GithubAll                     9264     127,494 ns/op      71,456 B/op       609 allocs/op
BenchmarkBunRouter_GithubAll                81684      14,368 ns/op           0 B/op         0 allocs/op
BenchmarkChi_GithubAll                      10000     123,835 ns/op     130,817 B/op       740 allocs/op
BenchmarkEcho_GithubAll                     75001      16,397 ns/op           0 B/op         0 allocs/op
BenchmarkFiber_GithubAll                     7580     156,388 ns/op           0 B/op         0 allocs/op
BenchmarkGin_GithubAll                      89120      13,624 ns/op           0 B/op         0 allocs/op
BenchmarkGojiv2_GithubAll                    3739     312,370 ns/op     313,744 B/op     3,712 allocs/op
BenchmarkGoRestful_GithubAll                 1112   1,155,200 ns/op   1,006,744 B/op     3,009 allocs/op
BenchmarkGorillaMux_GithubAll                 670   1,894,683 ns/op     225,667 B/op     1,588 allocs/op
BenchmarkHttpRouter_GithubAll               54688      22,071 ns/op      13,792 B/op       167 allocs/op
BenchmarkHttpTreeMux_GithubAll              17330      66,930 ns/op      65,856 B/op       671 allocs/op
BenchmarkMacaron_GithubAll                   7748     166,764 ns/op     147,784 B/op     1,624 allocs/op
BenchmarkZinc_GithubAll                     57319      19,800 ns/op           0 B/op         0 allocs/op
```

## Interpretation Notes

- This is one complete upstream-style run, matching the public repository's reporting method. Repeated samples with `benchstat` are preferable for optimization decisions.
- Framework handlers are configured to do the minimum work required by the upstream suite.
- The API-wide benchmarks execute every route once per benchmark iteration, so their allocation counts are totals for the complete route set.
- Results should only be compared with runs from the same machine, Go toolchain, dependency versions, and power/thermal conditions.
- The Zinc integration uses its native `http.Handler` implementation and the same `benchRequest` and `benchRoutes` helpers as Gin, Echo, Chi, and the other `net/http` routers.
