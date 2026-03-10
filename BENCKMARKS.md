# Zinc Benchmarks

Framework-only snapshot from the latest full benchmark suite run. The raw suite also includes `HttpRouter` and `ServeMux`; this document focuses on Zinc's direct framework peers: `Gin`, `Echo`, and `Chi`.

## Run Info

- Date: `2026-03-10`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command: `GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench . -benchmem -count=3 -benchtime=200ms`
- Reporting: averages across `count=3`
- Units: lower is better for `ns/op`, higher is better for `req/s`

## Headline

- Zinc wins `32/51` framework-only rows overall, with `44/51` top-2 finishes.
- Excluding throughput, Zinc wins `31/43` rows.
- Zinc now leads static dispatch, API-path work, and route/build-time benchmarks.
- The remaining latency gaps are concentrated in some `404/405` cases plus the heavier `ParseAPI26` / `GPlusAPI13` param-routing rows.
- Loopback throughput remains noisier than the in-process suite; on this rerun `Chi` won most throughput rows.

## Full-Suite Scorecard

| Framework | Wins | 2nd Place | Top-2 Finishes |
|---|---:|---:|---:|
| `Zinc` | `32` | `12` | `44` |
| `Gin` | `11` | `29` | `40` |
| `Chi` | `7` | `4` | `11` |
| `Echo` | `1` | `6` | `7` |

## Category Scorecard

| Category | Zinc | Gin | Echo | Chi |
|---|---:|---:|---:|---:|
| `Routing + Miss Paths` | `25W / 9x 2nd` | `11W / 23x 2nd` | `0W / 2x 2nd` | `0W / 2x 2nd` |
| `API / Handler Path` | `6W / 1x 2nd` | `0W / 3x 2nd` | `1W / 1x 2nd` | `0W / 2x 2nd` |
| `Throughput` | `1W / 2x 2nd` | `0W / 3x 2nd` | `0W / 3x 2nd` | `7W / 0x 2nd` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `66.5 ns` | `89.2 ns` | `128.9 ns` | `193.7 ns` | 🥇 `Zinc` |
| `StaticRoute` | `66.9 ns` | `90.4 ns` | `129.6 ns` | `174.5 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `72.2 ns` | `114.6 ns` | `164.6 ns` | `251.2 ns` | 🥇 `Zinc` |
| `RouterParam` | `83.9 ns` | `120.5 ns` | `150.4 ns` | `345.2 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `88.5 ns` | `99.9 ns` | `145.9 ns` | `227.0 ns` | 🥇 `Zinc` |
| `JSONResponse` | `331.7 ns` | `422.3 ns` | `398.4 ns` | `510.1 ns` | 🥇 `Zinc` |
| `QueryParams` | `410.6 ns` | `433.4 ns` | `485.5 ns` | `536.2 ns` | 🥇 `Zinc` |
| `MiddlewareChain` | `379.2 ns` | `443.9 ns` | `543.5 ns` | `871.4 ns` | 🥇 `Zinc` |
| `NotFound` | `126.5 ns` | `67.6 ns` | `640.0 ns` | `396.7 ns` | 🥇 `Gin` |
| `LargeRouteSetStatic` | `70.5 ns` | `103.0 ns` | `148.4 ns` | `222.9 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `70.2 ns` | `122.8 ns` | `168.5 ns` | `258.4 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `107.9 ns` | `190.5 ns` | `1541.8 ns` | `683.3 ns` | 🥇 `Zinc` |
| `LargeRouteSetMethodMismatch` | `319.6 ns` | `134.0 ns` | `1316.2 ns` | `438.2 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `156.1 ns` | `119.8 ns` | `259.7 ns` | `375.0 ns` | 🥇 `Gin` |
| `LargeRouteSetParamMixed` | `116.6 ns` | `135.5 ns` | `173.1 ns` | `292.8 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `933.9 ns` | `2708.3 ns` | `1831.3 ns` | `953.1 ns` | 🥇 `Zinc` |
| `APIHappyPath` | `1336.7 ns` | `3123.7 ns` | `2348.0 ns` | `1684.0 ns` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2631.7 ns` | `4525.3 ns` | `2592.0 ns` | `2812.7 ns` | 🥇 `Echo` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `68.8 µs / 87.5 KB` | `79.4 µs / 53.0 KB` | `350.7 µs / 181.6 KB` | `87.7 µs / 120.7 KB` | 🥇 `Zinc` |
| `RouteRegistrationParam` | `46.0 µs / 70.2 KB` | `54.3 µs / 46.8 KB` | `162.4 µs / 152.1 KB` | `76.6 µs / 89.9 KB` | 🥇 `Zinc` |

## Scenario Routing

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `68.4 ns` | `117.6 ns` | `172.5 ns` | `283.3 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `80.2 ns` | `105.0 ns` | `149.6 ns` | `272.6 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `67.0 ns` | `191.3 ns` | `156.5 ns` | `230.5 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `67.9 ns` | `101.2 ns` | `146.1 ns` | `229.9 ns` | 🥇 `Zinc` |
| `ScenarioParam/GitHubAPI203` | `158.9 ns` | `162.5 ns` | `251.0 ns` | `412.9 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `166.9 ns` | `125.5 ns` | `180.5 ns` | `288.9 ns` | 🥇 `Gin` |
| `ScenarioParam/ParseAPI26` | `217.3 ns` | `122.0 ns` | `166.0 ns` | `291.3 ns` | 🥇 `Gin` |
| `ScenarioAll/Static157` | `74.9 ns` | `137.5 ns` | `179.9 ns` | `302.3 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `141.1 ns` | `163.0 ns` | `209.9 ns` | `388.8 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `147.4 ns` | `192.5 ns` | `208.7 ns` | `304.7 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `142.8 ns` | `126.1 ns` | `167.3 ns` | `307.4 ns` | 🥇 `Gin` |
| `ScenarioBuild/Static157` | `54.7 µs` | `91.0 µs` | `194.5 µs` | `75.6 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `84.9 µs` | `137.5 µs` | `374.7 µs` | `183.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `6.6 µs` | `8.9 µs` | `17.7 µs` | `13.0 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `9.1 µs` | `14.1 µs` | `23.4 µs` | `15.1 µs` | 🥇 `Zinc` |

## Scenario Miss Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `111.8 ns` | `72.2 ns` | `654.2 ns` | `386.9 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GitHubAPI203` | `118.2 ns` | `127.2 ns` | `656.6 ns` | `379.9 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `155.7 ns` | `73.1 ns` | `726.0 ns` | `392.1 ns` | 🥇 `Gin` |
| `ScenarioNotFound/ParseAPI26` | `161.6 ns` | `122.6 ns` | `686.7 ns` | `448.2 ns` | 🥇 `Gin` |
| `Scenario405/Static157` | `158.5 ns` | `121.7 ns` | `944.8 ns` | `368.7 ns` | 🥇 `Gin` |
| `Scenario405/GitHubAPI203` | `169.6 ns` | `186.6 ns` | `938.7 ns` | `436.6 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `174.6 ns` | `206.7 ns` | `1032.2 ns` | `446.2 ns` | 🥇 `Zinc` |
| `Scenario405/ParseAPI26` | `350.4 ns` | `181.0 ns` | `1061.0 ns` | `346.8 ns` | 🥇 `Gin` |

## Throughput

Loopback `req/s` is noisier than the in-process `ns/op` benchmarks above, but included here for completeness.

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `14508` | `18509` | `17460` | `18977` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency8` | `53886` | `54215` | `52601` | `59133` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency32` | `81989` | `78822` | `82257` | `85514` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency128` | `87721` | `89423` | `87965` | `90867` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency1` | `15718` | `15149` | `15176` | `15904` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency8` | `58229` | `53830` | `51417` | `59160` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency32` | `67115` | `72847` | `80993` | `82212` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency128` | `89346` | `81707` | `84019` | `78970` | 🥇 `Zinc` |
