# Zinc Benchmarks

Framework-only snapshot from the latest full benchmark suite run. The raw suite also includes `HttpRouter` and `ServeMux`; this document focuses on Zinc's direct framework peers: `Gin`, `Echo`, and `Chi`.

## Run Info

- Date: `2026-03-09`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command: `GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench . -benchmem -count=3 -benchtime=200ms`
- Reporting: averages across `count=3`
- Units: lower is better for `ns/op`, higher is better for `req/s`

## Headline

- Zinc wins `25/51` framework-only rows overall, with `44/51` top-2 finishes.
- Excluding throughput, Zinc wins `23/43` rows.
- Zinc is strongest on static and mixed dispatch, middleware, JSON response, and scenario static sweeps.
- Gin still leads the miss-path and heavier param-routing cases.

## Full-Suite Scorecard

| Framework | Wins | 2nd Place | Top-2 Finishes |
|---|---:|---:|---:|
| `Zinc` | `25` | `19` | `44` |
| `Gin` | `21` | `20` | `41` |
| `Chi` | `4` | `5` | `9` |
| `Echo` | `1` | `7` | `8` |

## Category Scorecard

| Category | Zinc | Gin | Echo | Chi |
|---|---:|---:|---:|---:|
| `Routing + Miss Paths` | `18W / 14x 2nd` | `18W / 15x 2nd` | `0W / 4x 2nd` | `0W / 3x 2nd` |
| `API / Handler Path` | `5W / 2x 2nd` | `0W / 4x 2nd` | `1W / 0x 2nd` | `1W / 1x 2nd` |
| `Throughput` | `2W / 3x 2nd` | `3W / 1x 2nd` | `0W / 3x 2nd` | `3W / 1x 2nd` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `73.1 ns` | `91.4 ns` | `133.7 ns` | `199.7 ns` | 🥇 `Zinc` |
| `StaticRoute` | `69.4 ns` | `98.5 ns` | `134.0 ns` | `191.1 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `76.4 ns` | `114.9 ns` | `158.8 ns` | `235.0 ns` | 🥇 `Zinc` |
| `RouterParam` | `87.6 ns` | `95.7 ns` | `137.4 ns` | `368.7 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `87.9 ns` | `100.8 ns` | `141.8 ns` | `231.3 ns` | 🥇 `Zinc` |
| `JSONResponse` | `338.0 ns` | `400.4 ns` | `447.0 ns` | `561.5 ns` | 🥇 `Zinc` |
| `QueryParams` | `424.6 ns` | `434.2 ns` | `473.1 ns` | `548.4 ns` | 🥇 `Zinc` |
| `MiddlewareChain` | `367.4 ns` | `420.8 ns` | `537.1 ns` | `853.3 ns` | 🥇 `Zinc` |
| `NotFound` | `122.9 ns` | `67.2 ns` | `608.9 ns` | `333.1 ns` | 🥇 `Gin` |
| `LargeRouteSetStatic` | `69.5 ns` | `102.4 ns` | `152.9 ns` | `214.8 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `71.3 ns` | `123.0 ns` | `162.9 ns` | `260.6 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `120.3 ns` | `74.8 ns` | `648.7 ns` | `364.8 ns` | 🥇 `Gin` |
| `LargeRouteSetMethodMismatch` | `153.8 ns` | `101.3 ns` | `856.3 ns` | `328.2 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `108.7 ns` | `109.1 ns` | `158.8 ns` | `383.1 ns` | 🥇 `Zinc` |
| `LargeRouteSetParamMixed` | `114.2 ns` | `128.6 ns` | `170.4 ns` | `289.5 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `1195.3 ns` | `2741.0 ns` | `1876.0 ns` | `1095.3 ns` | 🥇 `Chi` |
| `APIHappyPath` | `1487.7 ns` | `3069.3 ns` | `2239.7 ns` | `1727.0 ns` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2774.0 ns` | `4503.0 ns` | `2570.7 ns` | `2926.7 ns` | 🥇 `Echo` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `84.3 µs / 62.5 KB` | `74.8 µs / 54.3 KB` | `338.3 µs / 185.9 KB` | `88.5 µs / 123.6 KB` | 🥇 `Gin` |
| `RouteRegistrationParam` | `64.1 µs / 67.8 KB` | `55.8 µs / 47.9 KB` | `161.8 µs / 155.7 KB` | `73.8 µs / 92.1 KB` | 🥇 `Gin` |

## Scenario Routing

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `69.7 ns` | `114.3 ns` | `165.2 ns` | `264.1 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `68.9 ns` | `103.3 ns` | `148.3 ns` | `242.3 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `68.6 ns` | `186.8 ns` | `144.8 ns` | `228.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `70.1 ns` | `98.5 ns` | `145.4 ns` | `238.1 ns` | 🥇 `Zinc` |
| `ScenarioParam/GitHubAPI203` | `171.4 ns` | `159.0 ns` | `225.4 ns` | `436.6 ns` | 🥇 `Gin` |
| `ScenarioParam/GPlusAPI13` | `178.8 ns` | `119.0 ns` | `167.7 ns` | `279.8 ns` | 🥇 `Gin` |
| `ScenarioParam/ParseAPI26` | `217.9 ns` | `118.1 ns` | `158.6 ns` | `279.5 ns` | 🥇 `Gin` |
| `ScenarioAll/Static157` | `72.9 ns` | `130.2 ns` | `171.6 ns` | `294.0 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `163.0 ns` | `157.5 ns` | `207.4 ns` | `372.8 ns` | 🥇 `Gin` |
| `ScenarioAll/GPlusAPI13` | `143.6 ns` | `180.6 ns` | `173.7 ns` | `280.5 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `140.1 ns` | `118.4 ns` | `155.5 ns` | `271.3 ns` | 🥇 `Gin` |
| `ScenarioBuild/Static157` | `68.3 µs` | `83.3 µs` | `181.3 µs` | `72.8 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `125.5 µs` | `135.8 µs` | `343.5 µs` | `177.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `8.3 µs` | `8.3 µs` | `15.7 µs` | `10.9 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `15.6 µs` | `14.2 µs` | `24.8 µs` | `15.4 µs` | 🥇 `Gin` |

## Scenario Miss Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `141.9 ns` | `59.9 ns` | `609.7 ns` | `370.6 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GitHubAPI203` | `161.4 ns` | `118.8 ns` | `611.0 ns` | `358.8 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GPlusAPI13` | `162.7 ns` | `74.1 ns` | `608.0 ns` | `360.2 ns` | 🥇 `Gin` |
| `ScenarioNotFound/ParseAPI26` | `163.0 ns` | `114.2 ns` | `636.9 ns` | `387.2 ns` | 🥇 `Gin` |
| `Scenario405/Static157` | `169.8 ns` | `109.0 ns` | `898.7 ns` | `351.9 ns` | 🥇 `Gin` |
| `Scenario405/GitHubAPI203` | `200.2 ns` | `168.7 ns` | `872.1 ns` | `320.9 ns` | 🥇 `Gin` |
| `Scenario405/GPlusAPI13` | `170.3 ns` | `182.7 ns` | `866.8 ns` | `424.2 ns` | 🥇 `Zinc` |
| `Scenario405/ParseAPI26` | `341.6 ns` | `162.5 ns` | `861.0 ns` | `304.8 ns` | 🥇 `Gin` |

## Throughput

Loopback `req/s` is noisier than the in-process `ns/op` benchmarks above, but included here for completeness.

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `18252` | `18089` | `17816` | `17142` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency8` | `58460` | `57103` | `51935` | `58698` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency32` | `82746` | `76472` | `70594` | `79068` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency128` | `95965` | `80066` | `90752` | `96081` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency1` | `17615` | `18181` | `18258` | `18470` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency8` | `57063` | `59614` | `45444` | `56161` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency32` | `85473` | `88454` | `88223` | `85192` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency128` | `88778` | `100297` | `97923` | `88640` | 🥇 `Gin` |
