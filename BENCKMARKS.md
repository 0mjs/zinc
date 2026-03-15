# Zinc Benchmarks

Framework-only snapshot from the latest benchmark rerun. The raw suite also includes `HttpRouter` and `ServeMux`; this document focuses on Zinc's direct framework peers: `Gin`, `Echo`, and `Chi`.

## Run Info

- Date: `2026-03-13`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command:
  - Core/runtime/build/parallel slice: `env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'Benchmark(HelloWorld|StaticRoute|StaticRouteCold|RouterParam|RouterParamCold|JSONResponse|QueryParams|MiddlewareChain|NotFound|LargeRouteSetStatic|LargeRouteSetStaticMixed|LargeRouteSetNotFound|LargeRouteSetMethodMismatch|LargeRouteSetParam|LargeRouteSetParamMixed|APIParamQueryJSON|APIHappyPath|APIBindJSONHappyPath|Param5|Param10|NestedGroupStatic|NestedGroupParam|NestedGroupNotFound|NestedGroupMethodMismatch|WildcardTail|WildcardTailNotFound|RouteRegistrationStatic|RouteRegistrationParam|ParallelStaticRoute|ParallelRouterParam|ParallelMiddlewareChain|ParallelAPIHappyPath)$' -benchmem -count=1`
  - Scenario + throughput slice: `env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench 'BenchmarkScenarioRouteSet(Build|Static|Param|NotFound|MethodMismatch|All)$|BenchmarkRequestsPerSecond(Focused)?$' -benchmem -count=1`
- Units: lower is better for `ns/op`, higher is better for `req/s`

## Headline

- Zinc wins `45/75` framework-only rows overall, with `70/75` top-2 finishes and `73/75` top-3 finishes.
- Excluding throughput, Zinc wins `42/67` rows.

## Full-Suite Scorecard

| Framework | Wins | 2nd Place | Top-3 Finishes |
|---|---:|---:|---:|
| `Zinc` | `45` | `25` | `73` |
| `Gin` | `27` | `35` | `69` |
| `Echo` | `1` | `10` | `51` |
| `Chi` | `2` | `5` | `32` |

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `72.7 ns` | `94.3 ns` | `132.2 ns` | `193.0 ns` | 🥇 `Zinc` |
| `StaticRoute` | `71.6 ns` | `108.9 ns` | `153.1 ns` | `224.2 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `95.2 ns` | `145.8 ns` | `195.2 ns` | `286.8 ns` | 🥇 `Zinc` |
| `RouterParam` | `102.7 ns` | `104.1 ns` | `169.6 ns` | `391.7 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `102.5 ns` | `105.7 ns` | `145.5 ns` | `278.5 ns` | 🥇 `Zinc` |
| `JSONResponse` | `362.2 ns` | `557.4 ns` | `438.0 ns` | `564.6 ns` | 🥇 `Zinc` |
| `QueryParams` | `519.2 ns` | `475.7 ns` | `556.0 ns` | `636.7 ns` | 🥇 `Gin` |
| `MiddlewareChain` | `392.8 ns` | `642.0 ns` | `603.2 ns` | `1161.0 ns` | 🥇 `Zinc` |
| `NotFound` | `150.2 ns` | `103.1 ns` | `883.6 ns` | `695.5 ns` | 🥇 `Gin` |
| `LargeRouteSetStatic` | `106.1 ns` | `133.3 ns` | `187.0 ns` | `332.2 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `100.0 ns` | `126.9 ns` | `180.4 ns` | `326.7 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `120.1 ns` | `86.3 ns` | `644.5 ns` | `368.8 ns` | 🥇 `Gin` |
| `LargeRouteSetMethodMismatch` | `160.2 ns` | `119.1 ns` | `888.2 ns` | `335.1 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `118.3 ns` | `134.7 ns` | `159.1 ns` | `428.8 ns` | 🥇 `Zinc` |
| `LargeRouteSetParamMixed` | `128.7 ns` | `138.4 ns` | `184.1 ns` | `317.9 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `914.9 ns` | `2836.0 ns` | `1865.0 ns` | `1236.0 ns` | 🥇 `Zinc` |
| `APIHappyPath` | `1281.0 ns` | `3146.0 ns` | `2230.0 ns` | `1826.0 ns` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2666.0 ns` | `4367.0 ns` | `2527.0 ns` | `2993.0 ns` | 🥇 `Echo` |

## Routing Realism

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `Param5` | `189.2 ns` | `152.6 ns` | `219.2 ns` | `531.2 ns` | 🥇 `Gin` |
| `Param10` | `394.6 ns` | `340.3 ns` | `381.0 ns` | `1172.0 ns` | 🥇 `Gin` |
| `NestedGroupStatic` | `100.6 ns` | `177.2 ns` | `146.8 ns` | `842.1 ns` | 🥇 `Zinc` |
| `NestedGroupParam` | `193.9 ns` | `144.8 ns` | `200.7 ns` | `1099.0 ns` | 🥇 `Gin` |
| `NestedGroupNotFound` | `199.6 ns` | `132.2 ns` | `179.8 ns` | `1024.0 ns` | 🥇 `Gin` |
| `NestedGroupMethodMismatch` | `217.1 ns` | `158.3 ns` | `471.2 ns` | `1044.0 ns` | 🥇 `Gin` |
| `WildcardTail` | `87.6 ns` | `106.1 ns` | `145.6 ns` | `494.7 ns` | 🥇 `Zinc` |
| `WildcardTailNotFound` | `89.7 ns` | `102.4 ns` | `132.9 ns` | `183.1 ns` | 🥇 `Zinc` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `81.1 µs / 113.8 KB` | `78.4 µs / 53.0 KB` | `363.0 µs / 181.6 KB` | `113.8 µs / 120.7 KB` | 🥇 `Gin` |
| `RouteRegistrationParam` | `53.2 µs / 86.9 KB` | `55.1 µs / 46.8 KB` | `179.4 µs / 152.1 KB` | `74.1 µs / 89.9 KB` | 🥇 `Zinc` |

## Scenario Routing

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `69.1 ns` | `118.9 ns` | `165.9 ns` | `267.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `68.4 ns` | `104.0 ns` | `146.6 ns` | `244.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `66.5 ns` | `184.6 ns` | `144.0 ns` | `233.5 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `75.7 ns` | `101.1 ns` | `158.2 ns` | `242.1 ns` | 🥇 `Zinc` |
| `ScenarioParam/GitHubAPI203` | `173.2 ns` | `270.4 ns` | `325.8 ns` | `453.9 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `176.1 ns` | `138.5 ns` | `177.2 ns` | `315.7 ns` | 🥇 `Gin` |
| `ScenarioParam/ParseAPI26` | `152.3 ns` | `134.0 ns` | `168.5 ns` | `315.5 ns` | 🥇 `Gin` |
| `ScenarioAll/Static157` | `84.1 ns` | `139.3 ns` | `184.0 ns` | `319.7 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `156.5 ns` | `187.2 ns` | `220.9 ns` | `458.3 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `160.9 ns` | `198.0 ns` | `176.0 ns` | `320.6 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `121.1 ns` | `158.4 ns` | `164.5 ns` | `333.6 ns` | 🥇 `Zinc` |
| `ScenarioBuild/Static157` | `57.4 µs` | `87.3 µs` | `184.6 µs` | `76.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `116.0 µs` | `141.6 µs` | `372.5 µs` | `179.5 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `7.5 µs` | `8.5 µs` | `16.2 µs` | `11.0 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParseAPI26` | `13.2 µs` | `14.8 µs` | `23.4 µs` | `15.2 µs` | 🥇 `Zinc` |

## Scenario Miss Paths

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `123.9 ns` | `69.5 ns` | `649.6 ns` | `390.9 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GitHubAPI203` | `119.7 ns` | `132.2 ns` | `730.7 ns` | `400.2 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `182.7 ns` | `75.6 ns` | `662.4 ns` | `382.3 ns` | 🥇 `Gin` |
| `ScenarioNotFound/ParseAPI26` | `157.1 ns` | `122.9 ns` | `671.4 ns` | `446.2 ns` | 🥇 `Gin` |
| `Scenario405/Static157` | `178.0 ns` | `116.0 ns` | `996.1 ns` | `381.6 ns` | 🥇 `Gin` |
| `Scenario405/GitHubAPI203` | `172.5 ns` | `178.6 ns` | `973.6 ns` | `344.4 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `201.4 ns` | `200.2 ns` | `980.3 ns` | `460.9 ns` | 🥇 `Gin` |
| `Scenario405/ParseAPI26` | `178.6 ns` | `169.5 ns` | `1122.0 ns` | `351.9 ns` | 🥇 `Gin` |

## Scenario Expansion

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ScenarioBuild/NestedAPI36` | `19.3 µs` | `23.7 µs` | `38.7 µs` | `24.0 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParamsAny24` | `23.2 µs` | `24.1 µs` | `43.4 µs` | `30.6 µs` | 🥇 `Zinc` |
| `ScenarioStatic/NestedAPI36` | `69.7 ns` | `120.5 ns` | `154.9 ns` | `281.9 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParamsAny24` | `72.5 ns` | `121.1 ns` | `147.6 ns` | `265.9 ns` | 🥇 `Zinc` |
| `ScenarioParam/NestedAPI36` | `167.8 ns` | `143.0 ns` | `194.2 ns` | `337.4 ns` | 🥇 `Gin` |
| `ScenarioParam/ParamsAny24` | `190.1 ns` | `168.8 ns` | `214.7 ns` | `620.7 ns` | 🥇 `Gin` |
| `ScenarioNotFound/NestedAPI36` | `119.7 ns` | `133.7 ns` | `703.5 ns` | `386.9 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParamsAny24` | `123.0 ns` | `83.9 ns` | `655.0 ns` | `387.7 ns` | 🥇 `Gin` |
| `Scenario405/NestedAPI36` | `160.8 ns` | `223.8 ns` | `1293.0 ns` | `1507.0 ns` | 🥇 `Zinc` |
| `Scenario405/ParamsAny24` | `189.7 ns` | `132.8 ns` | `973.1 ns` | `359.1 ns` | 🥇 `Gin` |
| `ScenarioAll/NestedAPI36` | `138.5 ns` | `135.2 ns` | `174.8 ns` | `369.4 ns` | 🥇 `Gin` |
| `ScenarioAll/ParamsAny24` | `156.2 ns` | `155.6 ns` | `200.7 ns` | `552.0 ns` | 🥇 `Gin` |

## Parallel In-Process

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `ParallelStaticRoute` | `14.7 ns` | `57.8 ns` | `72.9 ns` | `200.5 ns` | 🥇 `Zinc` |
| `ParallelRouterParam` | `20.0 ns` | `72.2 ns` | `72.2 ns` | `339.3 ns` | 🥇 `Zinc` |
| `ParallelMiddlewareChain` | `104.3 ns` | `300.6 ns` | `388.6 ns` | `1048.0 ns` | 🥇 `Zinc` |
| `ParallelAPIHappyPath` | `540.6 ns` | `1148.0 ns` | `837.8 ns` | `1451.0 ns` | 🥇 `Zinc` |

## Throughput

| Benchmark | Zinc | Gin | Echo | Chi | Winner |
|---|---:|---:|---:|---:|---|
| `RequestsPerSecond/Concurrency1` | `15892` | `17975` | `17765` | `17717` | 🥇 `Gin` |
| `RequestsPerSecond/Concurrency8` | `59669` | `58065` | `56077` | `57713` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency32` | `81277` | `71350` | `77299` | `80206` | 🥇 `Zinc` |
| `RequestsPerSecond/Concurrency128` | `80447` | `88910` | `88359` | `92344` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency1` | `17325` | `17717` | `17000` | `15520` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency8` | `55964` | `56564` | `55619` | `55929` | 🥇 `Gin` |
| `RequestsPerSecondFocused/Concurrency32` | `82463` | `80157` | `82649` | `84996` | 🥇 `Chi` |
| `RequestsPerSecondFocused/Concurrency128` | `91778` | `90014` | `91086` | `91321` | 🥇 `Zinc` |
