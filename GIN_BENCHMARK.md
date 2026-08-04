# Zinc in the Gin HTTP Routing Benchmark

**Machine:** Apple M1 Pro
**OS:** macOS 26.3 (Darwin 25.3.0), arm64
**Date:** August 4th, 2026
**Zinc Version:** 0.3.0 development (`66578db`)
**Go Version:** 1.26.5 darwin/arm64
**Source:** [Go HTTP Router Benchmark](https://github.com/gin-gonic/go-http-routing-benchmark)
**Suite Commit:** `ff3cdf55eccd0aa6a272991db9611a86c734dc51`

> Results were generated using the
> [gin-gonic/go-http-routing-benchmark](https://github.com/gin-gonic/go-http-routing-benchmark)
> suite, modified locally only to add a Zinc adapter. The benchmark suite is
> available under the [BSD 3-Clause License](https://github.com/gin-gonic/go-http-routing-benchmark/blob/master/LICENSE).
> Benchmark suite copyright © 2013 Julien Schmidt. All rights reserved.
> Zinc is not affiliated with or endorsed by Gin or the benchmark's authors.

---

## Table of Contents

- [Summary](#summary)
- [Zinc at a Glance](#zinc-at-a-glance)
- [Memory Consumption](#memory-consumption)
- [Benchmark Results](#benchmark-results)
  - [GitHub API (203 routes)](#github-api-203-routes)
  - [Google+ API (13 routes)](#google-api-13-routes)
  - [Parse API (26 routes)](#parse-api-26-routes)
  - [Static Routes (157 routes)](#static-routes-157-routes)
- [Micro Benchmarks](#micro-benchmarks)
  - [Single Param](#single-param)
  - [5 Params](#5-params)
  - [20 Params](#20-params)
  - [Param Write](#param-write)

---

## Summary

The table below ranks all routers by **GitHub API routing time** (203 routes, all methods), which best represents real-world routing workloads. _Lower ns/op is better._

| Rank | Router | ns/op | B/op | allocs/op | Zero-alloc |
| :--: | :----- | ----: | ---: | --------: | :--------: |
| 1 | **Gin** | 13,372 | 0 | 0 | :white_check_mark: |
| 2 | **BunRouter** | 14,612 | 0 | 0 | :white_check_mark: |
| 3 | **Echo** | 15,852 | 0 | 0 | :white_check_mark: |
| 4 | **Zinc** | 16,057 | 0 | 0 | :white_check_mark: |
| 5 | HttpRouter | 20,393 | 13,792 | 167 |  |
| 6 | HttpTreeMux | 62,934 | 65,856 | 671 |  |
| 7 | Beego | 117,328 | 71,456 | 609 |  |
| 8 | Chi | 124,766 | 130,817 | 740 |  |
| 9 | Macaron | 152,643 | 147,784 | 1,624 |  |
| 10 | Fiber | 153,679 | 0 | 0 | :white_check_mark: |
| 11 | Goji v2 | 292,715 | 313,744 | 3,712 |  |
| 12 | GoRestful | 1,216,620 | 1,006,744 | 3,009 |  |
| 13 | GorillaMux | 1,745,084 | 225,667 | 1,588 |  |

**Key takeaways:**

- **Gin**, **BunRouter**, **Echo**, and **Zinc** form the zero-allocation top tier, routing the complete GitHub API workload in 13.4–16.1 µs.
- **Zinc ranks fourth** on the GitHub workload, 1.3% behind Echo and 21.3% faster than HttpRouter.
- **Zinc ranks second** on the 157-route static workload and **first** on the 20-parameter microbenchmark.
- Zinc records zero heap allocations in every reported workload.

> **Fiber caveat:** Fiber benchmarks use `fasthttp.RequestCtx` with per-iteration Reset, which adds constant overhead not present in net/http benchmarks. Fiber-vs-Fiber comparisons are valid; cross-framework comparisons should be interpreted with care.

---

## Zinc at a Glance

| Workload | Zinc rank | Result | Allocations |
| :------- | :-------: | -----: | ----------: |
| GitHub API: 203 routes | **4th of 13** | 16,057 ns/op | 0 |
| Google+ API: 13 routes | **5th of 13** | 1,043 ns/op | 0 |
| Parse API: 26 routes | **5th of 13** | 1,665 ns/op | 0 |
| Static: 157 routes | **2nd of 13** | 7,022 ns/op | 0 |
| Single parameter | **5th of 13** | 47.84 ns/op | 0 |
| Five parameters | **2nd of 13** | 63.81 ns/op | 0 |
| Twenty parameters | **1st of 13** | 123.4 ns/op | 0 |
| Parameter read and write | **4th of 13** | 56.30 ns/op | 0 |

Zinc's clearest position in this suite is a **zero-allocation, top-tier router**: close to Echo on the representative GitHub workload, particularly strong on static and multi-parameter routes, and consistently ahead of the heavier `net/http` routers.

---

## Memory Consumption

Memory required for loading the routing structure (lower is better). Sorted by bytes ascending.

### Static Routes: 157

| Router | Bytes |
| :----- | ----: |
| **HttpRouter** | **21,680** |
| **Gin** | **34,408** |
| **Macaron** | **36,976** |
| **Zinc** | 39,296 |
| BunRouter | 51,232 |
| Fiber | 59,248 |
| HttpServeMux | 69,216 |
| HttpTreeMux | 73,448 |
| Chi | 83,160 |
| Echo | 91,976 |
| Beego | 98,824 |
| Goji v2 | 117,952 |
| GorillaMux | 599,496 |
| GoRestful | 819,688 |

### GitHub API Routes: 203

| Router | Bytes |
| :----- | ----: |
| **HttpRouter** | **37,072** |
| **Gin** | **58,840** |
| **HttpTreeMux** | **78,800** |
| Macaron | 90,632 |
| BunRouter | 93,776 |
| Chi | 94,888 |
| **Zinc** | 102,504 |
| Echo | 117,784 |
| Goji v2 | 118,640 |
| Beego | 150,840 |
| Fiber | 163,832 |
| GoRestful | 1,270,816 |
| GorillaMux | 1,319,696 |

### Google+ API Routes: 13

| Router | Bytes |
| :----- | ----: |
| **HttpRouter** | **2,776** |
| **Gin** | **4,576** |
| **BunRouter** | **7,360** |
| HttpTreeMux | 7,440 |
| Chi | 8,008 |
| Goji v2 | 8,096 |
| **Zinc** | 8,320 |
| Macaron | 8,672 |
| Beego | 10,256 |
| Fiber | 10,840 |
| Echo | 10,968 |
| GorillaMux | 68,000 |
| GoRestful | 72,520 |

### Parse API Routes: 26

| Router | Bytes |
| :----- | ----: |
| **HttpRouter** | **5,024** |
| **HttpTreeMux** | **7,848** |
| **Gin** | **7,896** |
| BunRouter | 9,336 |
| Chi | 9,656 |
| **Zinc** | 13,160 |
| Macaron | 13,704 |
| Echo | 13,816 |
| Fiber | 15,352 |
| Goji v2 | 16,064 |
| Beego | 19,256 |
| GorillaMux | 105,384 |
| GoRestful | 121,184 |

---

## Benchmark Results

### GitHub API (203 routes)

Routing all 203 GitHub API endpoints per operation.

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **Gin** | 13,372 | 0 | 0 |
| 2 | **BunRouter** | 14,612 | 0 | 0 |
| 3 | **Echo** | 15,852 | 0 | 0 |
| 4 | **Zinc** | 16,057 | 0 | 0 |
| 5 | HttpRouter | 20,393 | 13,792 | 167 |
| 6 | HttpTreeMux | 62,934 | 65,856 | 671 |
| 7 | Beego | 117,328 | 71,456 | 609 |
| 8 | Chi | 124,766 | 130,817 | 740 |
| 9 | Macaron | 152,643 | 147,784 | 1,624 |
| 10 | Fiber | 153,679 | 0 | 0 |
| 11 | Goji v2 | 292,715 | 313,744 | 3,712 |
| 12 | GoRestful | 1,216,620 | 1,006,744 | 3,009 |
| 13 | GorillaMux | 1,745,084 | 225,667 | 1,588 |

### Google+ API (13 routes)

Routing all 13 Google+ API endpoints per operation.

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **BunRouter** | 486.2 | 0 | 0 |
| 2 | **Gin** | 610.6 | 0 | 0 |
| 3 | **Echo** | 653.2 | 0 | 0 |
| 4 | HttpRouter | 1,003 | 640 | 11 |
| 5 | **Zinc** | 1,043 | 0 | 0 |
| 6 | HttpTreeMux | 3,357 | 4,032 | 38 |
| 7 | Fiber | 3,518 | 0 | 0 |
| 8 | Chi | 6,735 | 8,480 | 48 |
| 9 | Beego | 6,939 | 4,576 | 39 |
| 10 | Macaron | 9,861 | 9,464 | 104 |
| 11 | Goji v2 | 11,079 | 15,120 | 115 |
| 12 | GorillaMux | 21,358 | 14,448 | 102 |
| 13 | GoRestful | 31,147 | 60,720 | 193 |

### Parse API (26 routes)

Routing all 26 Parse API endpoints per operation.

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **BunRouter** | 803.1 | 0 | 0 |
| 2 | **Gin** | 1,108 | 0 | 0 |
| 3 | **Echo** | 1,363 | 0 | 0 |
| 4 | HttpRouter | 1,512 | 640 | 16 |
| 5 | **Zinc** | 1,665 | 0 | 0 |
| 6 | HttpTreeMux | 5,666 | 5,728 | 51 |
| 7 | Fiber | 6,338 | 0 | 0 |
| 8 | Beego | 12,524 | 9,152 | 78 |
| 9 | Chi | 13,338 | 14,944 | 84 |
| 10 | Goji v2 | 20,455 | 29,456 | 199 |
| 11 | Macaron | 21,216 | 18,928 | 208 |
| 12 | GorillaMux | 41,204 | 26,960 | 198 |
| 13 | GoRestful | 83,763 | 131,728 | 380 |

### Static Routes (157 routes)

Routing all 157 static routes per operation. Includes http.ServeMux as baseline.

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **HttpRouter** | 5,793 | 0 | 0 |
| 2 | **Zinc** | 7,022 | 0 | 0 |
| 3 | **HttpTreeMux** | 8,566 | 0 | 0 |
| 4 | BunRouter | 8,808 | 0 | 0 |
| 5 | Gin | 10,252 | 0 | 0 |
| 6 | Echo | 10,503 | 0 | 0 |
| — | HttpServeMux | 21,547 | 0 | 0 |
| 7 | Fiber | 42,243 | 0 | 0 |
| 8 | Chi | 62,158 | 57,776 | 314 |
| 9 | Beego | 91,662 | 55,264 | 471 |
| 10 | Macaron | 129,302 | 114,296 | 1,256 |
| 11 | Goji v2 | 133,092 | 175,840 | 1,099 |
| 12 | GorillaMux | 463,559 | 133,138 | 1,099 |
| 13 | GoRestful | 685,145 | 677,824 | 2,193 |

---

## Micro Benchmarks

### Single Param

Route: `/user/:name` — Request: `GET /user/gordon`

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **BunRouter** | 17.58 | 0 | 0 |
| 2 | **Echo** | 26.85 | 0 | 0 |
| 3 | **Gin** | 30.80 | 0 | 0 |
| 4 | HttpRouter | 46.41 | 32 | 1 |
| 5 | **Zinc** | 47.84 | 0 | 0 |
| 6 | Fiber | 161.0 | 0 | 0 |
| 7 | HttpTreeMux | 269.1 | 352 | 3 |
| 8 | Beego | 464.2 | 352 | 3 |
| 9 | Chi | 474.9 | 704 | 4 |
| 10 | Goji v2 | 656.4 | 1,136 | 8 |
| 11 | GorillaMux | 902.4 | 1,152 | 8 |
| 12 | Macaron | 940.5 | 1,064 | 10 |
| 13 | GoRestful | 2,087 | 4,600 | 15 |

### 5 Params

Route: `/:a/:b/:c/:d/:e` — Request: `GET /test/test/test/test/test`

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **Gin** | 56.63 | 0 | 0 |
| 2 | **Zinc** | 63.81 | 0 | 0 |
| 3 | **Echo** | 69.54 | 0 | 0 |
| 4 | BunRouter | 78.37 | 0 | 0 |
| 5 | HttpRouter | 144.6 | 160 | 1 |
| 6 | Fiber | 348.0 | 0 | 0 |
| 7 | HttpTreeMux | 491.2 | 576 | 6 |
| 8 | Beego | 646.8 | 352 | 3 |
| 9 | Chi | 710.9 | 704 | 4 |
| 10 | Goji v2 | 831.1 | 1,200 | 8 |
| 11 | Macaron | 1,027 | 1,064 | 10 |
| 12 | GorillaMux | 1,453 | 1,216 | 8 |
| 13 | GoRestful | 2,293 | 4,712 | 15 |

### 20 Params

Route: `/:a/:b/.../:t` (20 segments) — Request: `GET /a/b/.../t`

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **Zinc** | 123.4 | 0 | 0 |
| 2 | **Gin** | 172.8 | 0 | 0 |
| 3 | **Echo** | 203.1 | 0 | 0 |
| 4 | HttpRouter | 434.7 | 704 | 1 |
| 5 | BunRouter | 632.5 | 0 | 0 |
| 6 | Fiber | 667.8 | 0 | 0 |
| 7 | Goji v2 | 1,010 | 1,440 | 8 |
| 8 | Beego | 1,487 | 352 | 3 |
| 9 | Chi | 2,767 | 2,504 | 9 |
| 10 | Macaron | 2,780 | 2,864 | 15 |
| 11 | HttpTreeMux | 2,982 | 3,144 | 13 |
| 12 | GorillaMux | 3,079 | 3,272 | 13 |
| 13 | GoRestful | 4,963 | 7,008 | 20 |

### Param Write

Route: `/user/:name` with response write — Request: `GET /user/gordon`

| Rank | Router | ns/op | B/op | allocs/op |
| :--: | :----- | ----: | ---: | --------: |
| 1 | **BunRouter** | 38.54 | 0 | 0 |
| 2 | **Gin** | 40.19 | 0 | 0 |
| 3 | **HttpRouter** | 49.13 | 32 | 1 |
| 4 | **Zinc** | 56.30 | 0 | 0 |
| 5 | Echo | 67.11 | 8 | 1 |
| 6 | Fiber | 173.4 | 0 | 0 |
| 7 | HttpTreeMux | 290.2 | 352 | 3 |
| 8 | Chi | 484.4 | 704 | 4 |
| 9 | Beego | 495.7 | 360 | 4 |
| 10 | Goji v2 | 789.5 | 1,168 | 10 |
| 11 | GorillaMux | 1,048 | 1,152 | 8 |
| 12 | Macaron | 1,155 | 1,112 | 13 |
| 13 | GoRestful | 2,462 | 4,608 | 16 |
