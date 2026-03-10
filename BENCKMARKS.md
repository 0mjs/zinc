# Zinc Benchmarks

Framework-only snapshot from the latest full benchmark suite run. The raw suite also includes `HttpRouter` and `ServeMux`; this document focuses on Zinc's direct framework peers: `Gin`, `Echo`, and `Chi`.

## Run Info

- Date: `2026-03-10`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command: split run with `GOCACHE=/tmp/zinc-go-build`, `-benchmem`, `-count=3` for non-throughput and throughput slices
- Reporting: averages across `count=3`
- Units: lower is better for `ns/op`, higher is better for `req/s`

## Headline

- Zinc wins `28/51` framework-only rows overall, with `44/51` top-2 finishes.
- Excluding throughput, Zinc wins `25/43` rows.
- Zinc still leads static dispatch, most API-path work, and the scenario build rows.
- Gin now owns more cold `404/405` cases plus the heavier `ParseAPI26` / `GPlusAPI13` param-routing rows than the previous snapshot.
- Loopback throughput remains noisier than the in-process suite; on this rerun the `8` throughput rows split `Zinc 3`, `Gin 2`, `Chi 3`.

## Full-Suite Scorecard

| Framework | Wins | 2nd Place | Top-2 Finishes |
|---|---:|---:|---:|
| `Zinc` | `28` | `16` | `44` |
| `Gin` | `19` | `22` | `41` |
| `Chi` | `3` | `4` | `7` |
| `Echo` | `1` | `9` | `10` |

## Category Scorecard

| Category | Zinc | Gin | Echo | Chi |
|---|---:|---:|---:|---:|
| `Routing + Miss Paths` | `20W / 13x 2nd` | `16W / 17x 2nd` | `0W / 4x 2nd` | `0W / 2x 2nd` |
| `API / Handler Path` | `5W / 2x 2nd` | `1W / 3x 2nd` | `1W / 0x 2nd` | `0W / 2x 2nd` |
| `Throughput` | `3W / 1x 2nd` | `2W / 2x 2nd` | `0W / 5x 2nd` | `3W / 0x 2nd` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `65.7 ns` | `84.3 ns` | `123.6 ns` | `170.9 ns` | 🥇 `Zinc` |
| `StaticRoute` | `63.8 ns` | `83.9 ns` | `124.2 ns` | `166.3 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `68.1 ns` | `108.0 ns` | `150.5 ns` | `218.9 ns` | 🥇 `Zinc` |
| `RouterParam` | `92.8 ns` | `93.0 ns` | `134.6 ns` | `315.9 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `100.4 ns` | `95.8 ns` | `133.2 ns` | `209.6 ns` | 🥇 `Gin` |
| `JSONResponse` | `326.3 ns` | `388.8 ns` | `389.7 ns` | `491.9 ns` | 🥇 `Zinc` |
| `QueryParams` | `415.7 ns` | `397.6 ns` | `431.3 ns` | `467.5 ns` | 🥇 `Gin` |
| `MiddlewareChain` | `354.1 ns` | `385.6 ns` | `485.8 ns` | `810.6 ns` | 🥇 `Zinc` |
| `NotFound` | `116.9 ns` | `82.5 ns` | `536.8 ns` | `301.3 ns` | 🥇 `Gin` |
| `LargeRouteSetStatic` | `63.9 ns` | `97.0 ns` | `140.6 ns` | `194.6 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `66.9 ns` | `114.7 ns` | `151.7 ns` | `234.3 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `109.7 ns` | `66.7 ns` | `808.6 ns` | `587.2 ns` | 🥇 `Gin` |
| `LargeRouteSetMethodMismatch` | `153.9 ns` | `100.8 ns` | `869.2 ns` | `304.7 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `114.0 ns` | `108.0 ns` | `158.6 ns` | `389.0 ns` | 🥇 `Gin` |
| `LargeRouteSetParamMixed` | `133.6 ns` | `129.8 ns` | `172.6 ns` | `281.8 ns` | 🥇 `Gin` |
| `APIParamQueryJSON` | `873.7 ns` | `2742.0 ns` | `1814.0 ns` | `1004.6 ns` | 🥇 `Zinc` |
| `APIHappyPath` | `1292.3 ns` | `3073.3 ns` | `2221.7 ns` | `1696.7 ns` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2689.7 ns` | `4463.7 ns` | `2569.7 ns` | `2906.3 ns` | 🥇 `Echo` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `80.7 µs / 113.7 KB` | `77.1 µs / 53.0 KB` | `338.2 µs / 181.6 KB` | `92.5 µs / 120.7 KB` | 🥇 `Gin` |
| `RouteRegistrationParam` | `51.1 µs / 86.8 KB` | `51.7 µs / 46.8 KB` | `160.1 µs / 152.1 KB` | `76.8 µs / 89.9 KB` | 🥇 `Zinc` |

## Scenario Routing

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `68.5 ns` | `116.4 ns` | `164.6 ns` | `239.5 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `70.5 ns` | `97.6 ns` | `147.4 ns` | `205.7 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `67.5 ns` | `163.2 ns` | `146.2 ns` | `193.3 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `67.0 ns` | `93.9 ns` | `145.3 ns` | `196.3 ns` | 🥇 `Zinc` |
| `ScenarioParam/GitHubAPI203` | `152.8 ns` | `155.2 ns` | `220.9 ns` | `373.5 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `172.6 ns` | `116.6 ns` | `169.9 ns` | `250.3 ns` | 🥇 `Gin` |
| `ScenarioParam/ParseAPI26` | `219.1 ns` | `116.4 ns` | `164.7 ns` | `255.3 ns` | 🥇 `Gin` |
| `ScenarioAll/Static157` | `70.7 ns` | `124.7 ns` | `169.0 ns` | `257.0 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `147.1 ns` | `151.2 ns` | `196.5 ns` | `336.8 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `153.6 ns` | `173.9 ns` | `169.8 ns` | `275.0 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `152.2 ns` | `115.2 ns` | `157.2 ns` | `263.7 ns` | 🥇 `Gin` |
| `ScenarioBuild/Static157` | `49.5 µs` | `77.5 µs` | `172.8 µs` | `76.0 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `92.7 µs` | `125.4 µs` | `322.2 µs` | `169.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `5.8 µs` | `7.8 µs` | `13.5 µs` | `9.7 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `9.1 µs` | `13.0 µs` | `20.6 µs` | `13.3 µs` | 🥇 `Zinc` |

## Scenario Miss Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `112.3 ns` | `64.1 ns` | `605.9 ns` | `343.8 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GitHubAPI203` | `113.6 ns` | `120.8 ns` | `589.0 ns` | `366.7 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `159.5 ns` | `78.8 ns` | `552.5 ns` | `311.8 ns` | 🥇 `Gin` |
| `ScenarioNotFound/ParseAPI26` | `154.6 ns` | `100.7 ns` | `550.9 ns` | `317.8 ns` | 🥇 `Gin` |
| `Scenario405/Static157` | `150.7 ns` | `106.2 ns` | `824.2 ns` | `306.1 ns` | 🥇 `Gin` |
| `Scenario405/GitHubAPI203` | `160.2 ns` | `161.3 ns` | `798.7 ns` | `279.4 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `166.1 ns` | `164.3 ns` | `774.5 ns` | `382.4 ns` | 🥇 `Gin` |
| `Scenario405/ParseAPI26` | `324.2 ns` | `156.8 ns` | `793.8 ns` | `266.2 ns` | 🥇 `Gin` |

## Throughput

Loopback `req/s` is noisier than the in-process `ns/op` benchmarks above, but included here for completeness.

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `17844` | `17381` | `16365` | `15539` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency8` | `59442` | `57480` | `58724` | `60521` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency32` | `72305` | `82178` | `84582` | `86404` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency128` | `96457` | `88239` | `88903` | `66149` | 🥇 `Zinc` |
| `RequestsPerSecondFocused/Concurrency1` | `18370` | `18794` | `18297` | `18918` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency8` | `49509` | `57638` | `54415` | `40969` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency32` | `84956` | `69833` | `81602` | `79891` | 🥇 `Zinc` |
| `RequestsPerSecondFocused/Concurrency128` | `78708` | `94771` | `93132` | `87596` | 🥇 `Gin` |
