# Zinc Benchmarks

Latest full benchmark snapshot from the March 18, 2026 rerun. Row tables focus on Zinc's direct framework peers: `Gin`, `Echo`, and `Chi`. The raw harness also runs `HttpRouter` and `ServeMux`; those are summarized in a separate all-framework scorecard where directly comparable.

## Run Info

- Date: `2026-03-18`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Commands:
  - Non-throughput cross-framework suite: `env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'Benchmark(HelloWorld|StaticRoute|StaticRouteCold|RouterParam|RouterParamCold|JSONResponse|QueryParams|MiddlewareChain|NotFound|LargeRouteSetStatic|LargeRouteSetStaticMixed|LargeRouteSetNotFound|LargeRouteSetMethodMismatch|LargeRouteSetParam|LargeRouteSetParamMixed|RouteRegistrationStatic|RouteRegistrationParam|APIParamQueryJSON|APIHappyPath|APIBindJSONHappyPath|Param5|Param10|NestedGroupStatic|NestedGroupParam|NestedGroupNotFound|NestedGroupMethodMismatch|WildcardTail|WildcardTailNotFound|ScenarioRouteSet(Build|Static|Param|NotFound|MethodMismatch|All)|ParallelStaticRoute|ParallelRouterParam|ParallelMiddlewareChain|ParallelAPIHappyPath|APIBindHeaderQueryJSON|APIBindInvalidJSON|APIBindValidationFailure|APIBindMultipartHappyPath|LargeJSONResponse|LargeJSONBind|StaticFileHit|StaticFileNotFound|NestedGroupMiddlewareAPI|APIUnauthorizedReject)$' -benchmem -count=1`
  - Throughput suite: `env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'BenchmarkRequestsPerSecond(Focused)?$' -benchmem -count=1`
- Units: lower is better for latency rows; higher is better for `reqs/s`.

## Headline

- Peer-only score: Zinc wins `70/85` comparable rows overall.
- Excluding throughput, Zinc wins `69/77` rows.
- Throughput only, Zinc wins `1/8` rows.
- All-framework directly comparable rows: Zinc wins `11/41`, `HttpRouter` wins `29/41`, and `Gin` wins `1/41`.

## Peer Scorecard

| Framework | Wins | 2nd Place | Top-3 Finishes |
|---|---:|---:|---:|
| `Zinc` | `70` | `8` | `79` |
| `Gin` | `6` | `58` | `79` |
| `Echo` | `3` | `13` | `58` |
| `Chi` | `6` | `6` | `39` |

## All-Framework Scorecard

| Framework | Wins | 2nd Place | Top-3 Finishes |
|---|---:|---:|---:|
| `Zinc` | `11` | `22` | `38` |
| `ServeMux` | `0` | `5` | `12` |
| `HttpRouter` | `29` | `2` | `36` |
| `Chi` | `0` | `2` | `6` |
| `Echo` | `0` | `0` | `5` |
| `Gin` | `1` | `10` | `26` |

## Remaining Peer Gaps

| Benchmark | Winner | Zinc Gap |
|---|---|---:|
| `StaticFileNotFound` | `Gin` | `48.3% slower` |
| `APIBindInvalidJSON` | `Gin` | `34.4% slower` |
| `RouterParamCold` | `Gin` | `24.1% slower` |
| `LargeRouteSetMethodMismatch` | `Gin` | `8.6% slower` |
| `APIBindValidationFailure` | `Echo` | `1.3% slower` |
| `ScenarioAll/ParamsAny24` | `Gin` | `1.2% slower` |
| `APIBindMultipartHappyPath` | `Gin` | `1.0% slower` |
| `LargeJSONResponse` | `Echo` | `0.1% slower` |
| `RequestsPerSecond/Concurrency8` | `Chi` | `15.0% lower req/s` |
| `RequestsPerSecondFocused/Concurrency1` | `Chi` | `13.7% lower req/s` |
| `RequestsPerSecond/Concurrency32` | `Chi` | `6.8% lower req/s` |
| `RequestsPerSecondFocused/Concurrency32` | `Chi` | `1.9% lower req/s` |
| `RequestsPerSecondFocused/Concurrency8` | `Chi` | `1.9% lower req/s` |
| `RequestsPerSecondFocused/Concurrency128` | `Chi` | `1.0% lower req/s` |
| `RequestsPerSecond/Concurrency128` | `Echo` | `0.7% lower req/s` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `62.7 ns` | `94.7 ns` | `125.7 ns` | `183.4 ns` | 🥇 `Zinc` |
| `StaticRoute` | `61.5 ns` | `89.5 ns` | `126.3 ns` | `182.6 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `67.5 ns` | `115.3 ns` | `154.4 ns` | `241.9 ns` | 🥇 `Zinc` |
| `RouterParam` | `78.3 ns` | `95.0 ns` | `132.8 ns` | `369.3 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `123.0 ns` | `99.1 ns` | `135.6 ns` | `228.2 ns` | 🥇 `Gin` |
| `JSONResponse` | `335.6 ns` | `389.3 ns` | `384.8 ns` | `510.6 ns` | 🥇 `Zinc` |
| `QueryParams` | `407.7 ns` | `439.1 ns` | `472.1 ns` | `562.8 ns` | 🥇 `Zinc` |
| `MiddlewareChain` | `363.9 ns` | `425.0 ns` | `544.6 ns` | `890.8 ns` | 🥇 `Zinc` |
| `NotFound` | `60.7 ns` | `69.9 ns` | `615.4 ns` | `340.6 ns` | 🥇 `Zinc` |
| `LargeRouteSetStatic` | `66.4 ns` | `104.1 ns` | `145.5 ns` | `220.4 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `71.7 ns` | `122.1 ns` | `157.3 ns` | `263.6 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `60.8 ns` | `72.5 ns` | `631.3 ns` | `357.8 ns` | 🥇 `Zinc` |
| `LargeRouteSetMethodMismatch` | `110.0 ns` | `101.3 ns` | `864.9 ns` | `316.4 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `79.6 ns` | `113.5 ns` | `152.6 ns` | `391.9 ns` | 🥇 `Zinc` |
| `LargeRouteSetParamMixed` | `125.6 ns` | `130.0 ns` | `170.1 ns` | `297.7 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `869.7 ns` | `2.70 µs` | `1.80 µs` | `1.02 µs` | 🥇 `Zinc` |
| `APIHappyPath` | `1.23 µs` | `3.05 µs` | `2.21 µs` | `1.71 µs` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2.45 µs` | `4.45 µs` | `2.55 µs` | `2.93 µs` | 🥇 `Zinc` |

## Routing Realism

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `Param5` | `135.1 ns` | `154.5 ns` | `215.0 ns` | `574.9 ns` | 🥇 `Zinc` |
| `Param10` | `241.6 ns` | `325.0 ns` | `378.3 ns` | `1.26 µs` | 🥇 `Zinc` |
| `NestedGroupStatic` | `66.9 ns` | `187.4 ns` | `145.8 ns` | `881.9 ns` | 🥇 `Zinc` |
| `NestedGroupParam` | `142.2 ns` | `148.9 ns` | `192.8 ns` | `1.17 µs` | 🥇 `Zinc` |
| `NestedGroupNotFound` | `68.9 ns` | `117.2 ns` | `177.8 ns` | `1.07 µs` | 🥇 `Zinc` |
| `NestedGroupMethodMismatch` | `116.2 ns` | `164.4 ns` | `439.4 ns` | `1.12 µs` | 🥇 `Zinc` |
| `WildcardTail` | `80.5 ns` | `105.0 ns` | `135.3 ns` | `406.1 ns` | 🥇 `Zinc` |
| `WildcardTailNotFound` | `74.2 ns` | `101.2 ns` | `135.6 ns` | `208.3 ns` | 🥇 `Zinc` |

## Feature Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `APIBindHeaderQueryJSON` | `2.60 µs` | `5.79 µs` | `3.02 µs` | `2.78 µs` | 🥇 `Zinc` |
| `APIBindInvalidJSON` | `926.9 ns` | `689.9 ns` | `834.6 ns` | `843.8 ns` | 🥇 `Gin` |
| `APIBindValidationFailure` | `908.2 ns` | `963.9 ns` | `896.2 ns` | `1.00 µs` | 🥇 `Echo` |
| `APIBindMultipartHappyPath` | `11.6 µs` | `11.5 µs` | `11.6 µs` | `12.1 µs` | 🥇 `Gin` |
| `LargeJSONResponse` | `29.7 µs` | `32.9 µs` | `29.7 µs` | `29.8 µs` | 🥇 `Echo` |
| `LargeJSONBind` | `244.1 µs` | `249.5 µs` | `251.5 µs` | `248.6 µs` | 🥇 `Zinc` |
| `StaticFileHit` | `14.5 µs` | `25.9 µs` | `26.6 µs` | `15.0 µs` | 🥇 `Zinc` |
| `StaticFileNotFound` | `1.55 µs` | `1.04 µs` | `1.80 µs` | `2.04 µs` | 🥇 `Gin` |
| `NestedGroupMiddlewareAPI` | `1.24 µs` | `3.22 µs` | `2.29 µs` | `2.76 µs` | 🥇 `Zinc` |
| `APIUnauthorizedReject` | `41.1 ns` | `49.5 ns` | `76.0 ns` | `198.9 ns` | 🥇 `Zinc` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `50.1 µs` | `75.5 µs` | `344.4 µs` | `88.5 µs` | 🥇 `Zinc` |
| `RouteRegistrationParam` | `49.0 µs` | `52.0 µs` | `166.0 µs` | `70.4 µs` | 🥇 `Zinc` |

## Scenario Build

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioBuild/Static157` | `39.5 µs` | `84.8 µs` | `179.7 µs` | `75.4 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `97.1 µs` | `139.5 µs` | `340.6 µs` | `176.8 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `6.81 µs` | `8.60 µs` | `16.2 µs` | `11.2 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `11.1 µs` | `14.8 µs` | `24.1 µs` | `15.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/NestedAPI36` | `17.0 µs` | `23.1 µs` | `37.3 µs` | `24.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParamsAny24` | `20.5 µs` | `23.9 µs` | `42.2 µs` | `27.0 µs` | 🥇 `Zinc` |

## Scenario Static

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `70.1 ns` | `118.5 ns` | `170.3 ns` | `272.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `67.1 ns` | `106.8 ns` | `146.6 ns` | `245.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `63.2 ns` | `195.4 ns` | `144.6 ns` | `243.2 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `63.1 ns` | `103.3 ns` | `144.2 ns` | `240.8 ns` | 🥇 `Zinc` |
| `ScenarioStatic/NestedAPI36` | `62.7 ns` | `103.1 ns` | `150.1 ns` | `251.0 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParamsAny24` | `64.4 ns` | `106.4 ns` | `137.4 ns` | `242.0 ns` | 🥇 `Zinc` |

## Scenario Param

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioParam/GitHubAPI203` | `137.6 ns` | `166.6 ns` | `230.4 ns` | `417.9 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `110.8 ns` | `125.3 ns` | `171.6 ns` | `307.4 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParseAPI26` | `104.7 ns` | `125.9 ns` | `163.8 ns` | `307.8 ns` | 🥇 `Zinc` |
| `ScenarioParam/NestedAPI36` | `121.5 ns` | `142.9 ns` | `192.1 ns` | `369.7 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParamsAny24` | `141.9 ns` | `168.5 ns` | `214.0 ns` | `580.7 ns` | 🥇 `Zinc` |

## Scenario Not Found

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `60.3 ns` | `69.3 ns` | `667.2 ns` | `396.5 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GitHubAPI203` | `60.8 ns` | `129.3 ns` | `657.6 ns` | `390.2 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `60.1 ns` | `73.2 ns` | `662.8 ns` | `391.5 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParseAPI26` | `60.1 ns` | `118.0 ns` | `667.7 ns` | `390.6 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/NestedAPI36` | `60.3 ns` | `129.6 ns` | `663.9 ns` | `390.1 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParamsAny24` | `60.3 ns` | `83.0 ns` | `667.6 ns` | `407.2 ns` | 🥇 `Zinc` |

## Scenario 405

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `Scenario405/Static157` | `111.1 ns` | `113.4 ns` | `956.7 ns` | `377.9 ns` | 🥇 `Zinc` |
| `Scenario405/GitHubAPI203` | `103.2 ns` | `175.9 ns` | `932.3 ns` | `346.1 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `111.2 ns` | `200.6 ns` | `952.3 ns` | `469.1 ns` | 🥇 `Zinc` |
| `Scenario405/ParseAPI26` | `103.3 ns` | `169.7 ns` | `935.0 ns` | `340.6 ns` | 🥇 `Zinc` |
| `Scenario405/NestedAPI36` | `112.5 ns` | `191.7 ns` | `948.6 ns` | `353.5 ns` | 🥇 `Zinc` |
| `Scenario405/ParamsAny24` | `111.3 ns` | `119.1 ns` | `931.2 ns` | `354.1 ns` | 🥇 `Zinc` |

## Scenario Mixed

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioAll/Static157` | `72.7 ns` | `137.8 ns` | `177.5 ns` | `307.3 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `152.6 ns` | `162.5 ns` | `205.7 ns` | `392.1 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `124.9 ns` | `191.7 ns` | `173.5 ns` | `300.9 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `117.0 ns` | `125.3 ns` | `159.3 ns` | `300.7 ns` | 🥇 `Zinc` |
| `ScenarioAll/NestedAPI36` | `125.8 ns` | `135.1 ns` | `171.9 ns` | `315.2 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParamsAny24` | `152.6 ns` | `150.8 ns` | `189.3 ns` | `521.3 ns` | 🥇 `Gin` |

## Parallel In-Process

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ParallelStaticRoute` | `11.8 ns` | `47.6 ns` | `58.3 ns` | `160.4 ns` | 🥇 `Zinc` |
| `ParallelRouterParam` | `14.9 ns` | `54.0 ns` | `61.3 ns` | `307.2 ns` | 🥇 `Zinc` |
| `ParallelMiddlewareChain` | `64.0 ns` | `193.6 ns` | `264.5 ns` | `908.5 ns` | 🥇 `Zinc` |
| `ParallelAPIHappyPath` | `340.2 ns` | `863.6 ns` | `641.0 ns` | `1.33 µs` | 🥇 `Zinc` |

## Throughput

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `15,665` | `15,319` | `13,113` | `15,158` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency8` | `53,148` | `59,637` | `59,921` | `61,121` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency32` | `82,895` | `86,221` | `86,337` | `88,545` | 🥇 `Chi` |
| `RequestsPerSecond/Concurrency128` | `98,355` | `97,575` | `98,999` | `87,675` | 🥇 `Echo` |
| `RequestsPerSecondFocused/Concurrency1` | `16,692` | `18,131` | `18,670` | `18,986` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency8` | `59,894` | `60,256` | `59,928` | `61,030` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency32` | `87,672` | `87,997` | `89,145` | `89,340` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency128` | `99,100` | `98,425` | `94,459` | `100,072` | 🥇 `Chi` |
