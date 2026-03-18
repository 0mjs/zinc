# Zinc Benchmarks

Latest full peer-only benchmark snapshot from the March 18, 2026 rerun. The suite now compares Zinc directly against `Gin`, `Echo`, and `Chi` only.

## Run Info

- Date: `2026-03-18`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command: `cd benchmarks && env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench . -benchmem -count=1`
- Units: lower is better for latency rows; higher is better for `reqs/s`.

## Headline

- Peer-only score: Zinc wins `65/85` comparable rows overall.
- Excluding throughput, Zinc wins `65/77` rows.
- Throughput only, Zinc wins `0/8` rows.

## Peer Scorecard

| Framework | Wins | 2nd Place | Top-3 Finishes |
|---|---:|---:|---:|
| `Zinc` | `65` | `12` | `83` |
| `Gin` | `13` | `52` | `73` |
| `Echo` | `2` | `12` | `59` |
| `Chi` | `5` | `9` | `40` |

## Remaining Peer Gaps

| Benchmark | Winner | Zinc Gap |
|---|---|---:|
| `StaticFileNotFound` | `Gin` | `63.5% slower` |
| `APIBindInvalidJSON` | `Gin` | `32.2% slower` |
| `RouterParamCold` | `Gin` | `23.0% slower` |
| `RouteRegistrationParam` | `Gin` | `21.0% slower` |
| `RequestsPerSecondFocused/Concurrency128` | `Echo` | `12.6% lower req/s` |
| `LargeRouteSetMethodMismatch` | `Gin` | `10.8% slower` |
| `ScenarioBuild/ParamsAny24` | `Gin` | `7.4% slower` |
| `ScenarioAll/ParamsAny24` | `Gin` | `5.4% slower` |
| `RequestsPerSecondFocused/Concurrency1` | `Chi` | `5.1% lower req/s` |
| `RequestsPerSecond/Concurrency1` | `Chi` | `3.9% lower req/s` |
| `LargeRouteSetNotFound` | `Gin` | `3.5% slower` |
| `NotFound` | `Gin` | `3.4% slower` |
| `RequestsPerSecondFocused/Concurrency8` | `Gin` | `2.9% lower req/s` |
| `RequestsPerSecondFocused/Concurrency32` | `Echo` | `2.3% lower req/s` |
| `ScenarioNotFound/ParamsAny24` | `Gin` | `2.2% slower` |
| `RequestsPerSecond/Concurrency32` | `Chi` | `2.1% lower req/s` |
| `RequestsPerSecond/Concurrency8` | `Chi` | `1.4% lower req/s` |
| `RequestsPerSecond/Concurrency128` | `Chi` | `1.2% lower req/s` |
| `APIBindMultipartHappyPath` | `Gin` | `0.5% slower` |
| `Scenario405/Static157` | `Gin` | `0.4% slower` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `61.82 ns` | `88.21 ns` | `127.2 ns` | `178.8 ns` | 🥇 `Zinc` |
| `StaticRoute` | `61.15 ns` | `89.03 ns` | `129.1 ns` | `177.0 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `72.05 ns` | `113.9 ns` | `154.6 ns` | `232.3 ns` | 🥇 `Zinc` |
| `RouterParam` | `77.54 ns` | `94.03 ns` | `134.9 ns` | `336.8 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `122.9 ns` | `99.93 ns` | `135.6 ns` | `360.0 ns` | 🥇 `Gin` |
| `JSONResponse` | `331.2 ns` | `372.9 ns` | `377.2 ns` | `489.2 ns` | 🥇 `Zinc` |
| `QueryParams` | `396.6 ns` | `422.0 ns` | `466.8 ns` | `507.0 ns` | 🥇 `Zinc` |
| `MiddlewareChain` | `358.3 ns` | `407.9 ns` | `518.5 ns` | `863.4 ns` | 🥇 `Zinc` |
| `NotFound` | `60.30 ns` | `58.30 ns` | `598.6 ns` | `323.6 ns` | 🥇 `Gin` |
| `LargeRouteSetStatic` | `68.48 ns` | `101.0 ns` | `145.3 ns` | `210.9 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `71.18 ns` | `120.1 ns` | `156.5 ns` | `248.2 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `65.68 ns` | `63.43 ns` | `605.8 ns` | `336.5 ns` | 🥇 `Gin` |
| `LargeRouteSetMethodMismatch` | `110.6 ns` | `99.85 ns` | `830.4 ns` | `301.2 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `78.32 ns` | `107.7 ns` | `150.5 ns` | `376.3 ns` | 🥇 `Zinc` |
| `LargeRouteSetParamMixed` | `124.4 ns` | `129.2 ns` | `168.1 ns` | `422.9 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `853.7 ns` | `2.69 µs` | `1.79 µs` | `938.7 ns` | 🥇 `Zinc` |
| `APIHappyPath` | `1.25 µs` | `3.02 µs` | `2.19 µs` | `1.66 µs` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2.41 µs` | `4.41 µs` | `2.52 µs` | `2.81 µs` | 🥇 `Zinc` |

## Routing Realism

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `Param5` | `135.4 ns` | `153.0 ns` | `216.2 ns` | `555.4 ns` | 🥇 `Zinc` |
| `Param10` | `243.7 ns` | `321.2 ns` | `383.5 ns` | `1.24 µs` | 🥇 `Zinc` |
| `NestedGroupStatic` | `70.14 ns` | `190.1 ns` | `148.0 ns` | `885.7 ns` | 🥇 `Zinc` |
| `NestedGroupParam` | `140.1 ns` | `145.0 ns` | `193.8 ns` | `1.15 µs` | 🥇 `Zinc` |
| `NestedGroupNotFound` | `68.69 ns` | `117.1 ns` | `179.2 ns` | `1.03 µs` | 🥇 `Zinc` |
| `NestedGroupMethodMismatch` | `117.1 ns` | `157.9 ns` | `435.4 ns` | `1.12 µs` | 🥇 `Zinc` |
| `WildcardTail` | `81.39 ns` | `105.5 ns` | `135.3 ns` | `402.3 ns` | 🥇 `Zinc` |
| `WildcardTailNotFound` | `71.99 ns` | `102.8 ns` | `135.4 ns` | `198.1 ns` | 🥇 `Zinc` |

## Feature Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `APIBindHeaderQueryJSON` | `2.73 µs` | `5.96 µs` | `3.12 µs` | `2.83 µs` | 🥇 `Zinc` |
| `APIBindInvalidJSON` | `967.7 ns` | `732.1 ns` | `857.1 ns` | `906.1 ns` | 🥇 `Gin` |
| `APIBindValidationFailure` | `883.8 ns` | `970.4 ns` | `921.9 ns` | `1.06 µs` | 🥇 `Zinc` |
| `APIBindMultipartHappyPath` | `11.71 µs` | `11.65 µs` | `11.80 µs` | `12.48 µs` | 🥇 `Gin` |
| `LargeJSONResponse` | `28.97 µs` | `33.54 µs` | `30.12 µs` | `29.63 µs` | 🥇 `Zinc` |
| `LargeJSONBind` | `255.4 µs` | `269.5 µs` | `314.6 µs` | `273.9 µs` | 🥇 `Zinc` |
| `StaticFileHit` | `15.64 µs` | `32.85 µs` | `26.95 µs` | `15.97 µs` | 🥇 `Zinc` |
| `StaticFileNotFound` | `1.75 µs` | `1.07 µs` | `1.83 µs` | `2.08 µs` | 🥇 `Gin` |
| `NestedGroupMiddlewareAPI` | `1.26 µs` | `3.44 µs` | `2.18 µs` | `2.98 µs` | 🥇 `Zinc` |
| `APIUnauthorizedReject` | `42.07 ns` | `48.59 ns` | `75.17 ns` | `212.9 ns` | 🥇 `Zinc` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `57.88 µs` | `76.33 µs` | `337.4 µs` | `85.04 µs` | 🥇 `Zinc` |
| `RouteRegistrationParam` | `61.97 µs` | `51.20 µs` | `154.9 µs` | `67.29 µs` | 🥇 `Gin` |

## Scenario Build

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioBuild/Static157` | `52.78 µs` | `85.59 µs` | `179.8 µs` | `75.80 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `127.2 µs` | `134.0 µs` | `329.1 µs` | `172.4 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `8.31 µs` | `8.87 µs` | `18.07 µs` | `10.75 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `14.80 µs` | `14.91 µs` | `23.56 µs` | `15.60 µs` | 🥇 `Zinc` |
| `ScenarioBuild/NestedAPI36` | `21.97 µs` | `22.96 µs` | `36.24 µs` | `25.36 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParamsAny24` | `25.14 µs` | `23.41 µs` | `39.94 µs` | `26.84 µs` | 🥇 `Gin` |

## Scenario Static

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `69.10 ns` | `119.8 ns` | `164.3 ns` | `259.0 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `72.15 ns` | `104.5 ns` | `144.7 ns` | `232.4 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `64.90 ns` | `185.2 ns` | `140.8 ns` | `227.1 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `70.04 ns` | `100.3 ns` | `143.9 ns` | `231.4 ns` | 🥇 `Zinc` |
| `ScenarioStatic/NestedAPI36` | `63.71 ns` | `99.58 ns` | `146.4 ns` | `236.7 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParamsAny24` | `64.76 ns` | `101.3 ns` | `135.4 ns` | `228.7 ns` | 🥇 `Zinc` |

## Scenario Param

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioParam/GitHubAPI203` | `134.2 ns` | `161.3 ns` | `221.5 ns` | `560.3 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `109.1 ns` | `121.7 ns` | `168.3 ns` | `446.6 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParseAPI26` | `105.4 ns` | `122.0 ns` | `164.0 ns` | `459.4 ns` | 🥇 `Zinc` |
| `ScenarioParam/NestedAPI36` | `120.6 ns` | `139.9 ns` | `190.1 ns` | `506.2 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParamsAny24` | `140.1 ns` | `166.1 ns` | `217.5 ns` | `555.1 ns` | 🥇 `Zinc` |

## Scenario Not Found

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `63.43 ns` | `75.32 ns` | `711.3 ns` | `376.4 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GitHubAPI203` | `63.09 ns` | `137.0 ns` | `673.0 ns` | `431.1 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `80.57 ns` | `80.73 ns` | `703.8 ns` | `440.2 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParseAPI26` | `74.77 ns` | `123.6 ns` | `716.2 ns` | `396.0 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/NestedAPI36` | `86.53 ns` | `130.1 ns` | `669.7 ns` | `413.0 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParamsAny24` | `78.81 ns` | `77.10 ns` | `667.2 ns` | `395.6 ns` | 🥇 `Gin` |

## Scenario 405

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `Scenario405/Static157` | `113.6 ns` | `113.1 ns` | `1.01 µs` | `379.6 ns` | 🥇 `Gin` |
| `Scenario405/GitHubAPI203` | `102.9 ns` | `180.5 ns` | `976.8 ns` | `340.2 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `111.9 ns` | `196.5 ns` | `931.8 ns` | `478.0 ns` | 🥇 `Zinc` |
| `Scenario405/ParseAPI26` | `113.7 ns` | `197.4 ns` | `1.01 µs` | `357.8 ns` | 🥇 `Zinc` |
| `Scenario405/NestedAPI36` | `118.6 ns` | `186.7 ns` | `919.9 ns` | `342.6 ns` | 🥇 `Zinc` |
| `Scenario405/ParamsAny24` | `109.8 ns` | `116.1 ns` | `863.6 ns` | `315.1 ns` | 🥇 `Zinc` |

## Scenario Mixed

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioAll/Static157` | `74.32 ns` | `134.1 ns` | `176.2 ns` | `291.1 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `150.5 ns` | `160.1 ns` | `207.6 ns` | `499.3 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `125.3 ns` | `182.7 ns` | `168.3 ns` | `418.1 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `118.2 ns` | `120.9 ns` | `155.4 ns` | `370.9 ns` | 🥇 `Zinc` |
| `ScenarioAll/NestedAPI36` | `126.6 ns` | `131.0 ns` | `168.9 ns` | `422.3 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParamsAny24` | `155.0 ns` | `147.0 ns` | `186.2 ns` | `484.3 ns` | 🥇 `Gin` |

## Parallel In-Process

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ParallelStaticRoute` | `10.98 ns` | `45.09 ns` | `56.55 ns` | `151.3 ns` | 🥇 `Zinc` |
| `ParallelRouterParam` | `15.92 ns` | `71.04 ns` | `64.60 ns` | `319.4 ns` | 🥇 `Zinc` |
| `ParallelMiddlewareChain` | `74.24 ns` | `209.2 ns` | `265.0 ns` | `885.2 ns` | 🥇 `Zinc` |
| `ParallelAPIHappyPath` | `344.7 ns` | `863.6 ns` | `623.0 ns` | `1.25 µs` | 🥇 `Zinc` |

## Throughput

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `19,134` | `19,253` | `19,565` | `19,918` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency8` | `58,651` | `57,048` | `58,662` | `59,501` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency32` | `88,112` | `88,183` | `87,854` | `89,971` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency128` | `100,526` | `85,485` | `99,570` | `101,749` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency1` | `18,479` | `18,333` | `19,100` | `19,469` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency8` | `59,579` | `61,337` | `59,540` | `59,814` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency32` | `88,535` | `79,992` | `90,587` | `90,547` | 🥇 `Echo` |
| `RequestsPerSecondFocused/Concurrency128` | `86,669` | `78,637` | `99,163` | `96,368` | 🥇 `Echo` |
