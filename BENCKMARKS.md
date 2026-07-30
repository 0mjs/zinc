# Zinc Benchmarks

Zinc is fastest in `64/77` directly comparable benchmark rows against Gin, Echo, and Chi. In another four rows, Zinc is within 2% of the fastest result, giving it a leading or near-tied result in `68/77` rows.

These results measure the default in-process comparison suite. Each framework performs equivalent work through its idiomatic API.

## Run Info

- Date: `2026-07-31`
- Zinc commit: `160b107`
- Machine: `Apple M1 Pro`
- OS / Arch: `darwin / arm64`
- Command: `cd benchmarks && env GOCACHE=/tmp/zinc-go-build go test -run=^$ -bench . -benchmem -count=1`
- Units: lower is better.

The result column uses `≈` when the two fastest measurements are less than 2% apart. This policy is applied consistently, including when Zinc records the lower raw value. Small differences from a single run should be confirmed with repeated samples and `benchstat`.

## Headline

- Zinc records the lowest raw latency in `64/77` comparable rows.
- Zinc wins or finishes within 2% of the fastest result in `68/77` rows.
- Zinc finishes first or second in all `77/77` rows.
- Zinc retains zero request-time allocations across its primary static, parameter, and not-found dispatch benchmarks.

## Peer Scorecard

| Framework | Raw Wins | 2nd Place | Top-3 Finishes |
|---|---:|---:|---:|
| `Zinc` | `64` | `13` | `77` |
| `Gin` | `11` | `52` | `71` |
| `Echo` | `1` | `8` | `51` |
| `Chi` | `1` | `4` | `32` |

## Latest Performance Improvements

Round 8 targeted feature paths without changing the router. Repeated before-and-after samples showed:

| Benchmark | Latency | Memory | Allocations |
|---|---:|---:|---:|
| `StaticFileNotFound` | `8.8% faster` | `929 → 736 B/op` | `14 → 12` |
| `StaticFileHit` | `2.5% faster` | `1,235 → 1,073 B/op` | `20 → 18` |
| `APIBindInvalidJSON` | `20.9% faster` | `1,040 → 576 B/op` | `18 → 18` |
| `APIBindJSONHappyPath` | `13.1% faster` | `30.0% lower` | — |
| `APIBindHeaderQueryJSON` | `8.5% faster` | `27.5% lower` | — |
| `APIBindValidationFailure` | `18.6% faster` | `50.4% lower` | — |
| `LargeJSONBind` | `7.7% faster` | `19.7% lower` | `0.7% lower` |

The full-suite values below are a fresh single snapshot. The repeated results above are the stronger evidence for the paths changed in Round 8.

## Remaining Meaningful Gaps

| Benchmark | Winner | Zinc Gap |
|---|---|---:|
| `ScenarioNotFound/Static157` | `Gin` | `61.6% slower` |
| `StaticFileNotFound` | `Gin` | `48.5% slower` |
| `ScenarioNotFound/ParamsAny24` | `Gin` | `37.1% slower` |
| `ScenarioNotFound/GPlusAPI13` | `Gin` | `22.5% slower` |
| `RouterParamCold` | `Gin` | `18.1% slower` |
| `APIBindInvalidJSON` | `Gin` | `12.5% slower` |
| `RouteRegistrationParam` | `Gin` | `8.4% slower` |
| `LargeRouteSetMethodMismatch` | `Gin` | `7.1% slower` |
| `ScenarioBuild/ParseAPI26` | `Gin` | `3.7% slower` |

Four additional raw losses are below the 2% near-tie threshold: `LargeJSONResponse`, `StaticFileHit`, `ScenarioBuild/GPlusAPI13`, and `ScenarioBuild/ParamsAny24`.

## Core Runtime

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `HelloWorld` | `60.53 ns` | `92.07 ns` | `126.3 ns` | `187.8 ns` | 🥇 `Zinc` |
| `StaticRoute` | `60.66 ns` | `106.4 ns` | `132.0 ns` | `190.7 ns` | 🥇 `Zinc` |
| `StaticRouteCold` | `74.57 ns` | `146.2 ns` | `176.7 ns` | `255.1 ns` | 🥇 `Zinc` |
| `RouterParam` | `86.44 ns` | `109.8 ns` | `164.8 ns` | `713.7 ns` | 🥇 `Zinc` |
| `RouterParamCold` | `120.0 ns` | `101.6 ns` | `135.9 ns` | `388.4 ns` | 🥇 `Gin` |
| `JSONResponse` | `333.8 ns` | `388.3 ns` | `385.9 ns` | `521.7 ns` | 🥇 `Zinc` |
| `QueryParams` | `417.6 ns` | `456.3 ns` | `484.2 ns` | `547.8 ns` | 🥇 `Zinc` |
| `MiddlewareChain` | `362.7 ns` | `435.7 ns` | `551.9 ns` | `920.4 ns` | 🥇 `Zinc` |
| `NotFound` | `73.51 ns` | `78.88 ns` | `639.2 ns` | `351.8 ns` | 🥇 `Zinc` |
| `LargeRouteSetStatic` | `69.63 ns` | `104.1 ns` | `144.9 ns` | `226.2 ns` | 🥇 `Zinc` |
| `LargeRouteSetStaticMixed` | `70.18 ns` | `128.6 ns` | `158.5 ns` | `266.1 ns` | 🥇 `Zinc` |
| `LargeRouteSetNotFound` | `71.18 ns` | `73.43 ns` | `646.3 ns` | `363.1 ns` | 🥇 `Zinc` |
| `LargeRouteSetMethodMismatch` | `108.9 ns` | `101.7 ns` | `902.7 ns` | `362.3 ns` | 🥇 `Gin` |
| `LargeRouteSetParam` | `81.53 ns` | `113.3 ns` | `150.7 ns` | `401.8 ns` | 🥇 `Zinc` |
| `LargeRouteSetParamMixed` | `121.3 ns` | `130.6 ns` | `169.4 ns` | `460.1 ns` | 🥇 `Zinc` |
| `APIParamQueryJSON` | `845.3 ns` | `2.72 µs` | `1.83 µs` | `1.04 µs` | 🥇 `Zinc` |
| `APIHappyPath` | `1.19 µs` | `3.10 µs` | `2.24 µs` | `1.70 µs` | 🥇 `Zinc` |
| `APIBindJSONHappyPath` | `2.22 µs` | `4.58 µs` | `2.56 µs` | `2.89 µs` | 🥇 `Zinc` |

## Routing Realism

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `Param5` | `137.4 ns` | `153.5 ns` | `213.2 ns` | `573.4 ns` | 🥇 `Zinc` |
| `Param10` | `239.2 ns` | `321.4 ns` | `377.4 ns` | `1.23 µs` | 🥇 `Zinc` |
| `NestedGroupStatic` | `67.77 ns` | `185.9 ns` | `147.6 ns` | `856.9 ns` | 🥇 `Zinc` |
| `NestedGroupParam` | `130.4 ns` | `145.6 ns` | `192.3 ns` | `1.14 µs` | 🥇 `Zinc` |
| `NestedGroupNotFound` | `72.88 ns` | `117.8 ns` | `177.5 ns` | `1.07 µs` | 🥇 `Zinc` |
| `NestedGroupMethodMismatch` | `125.7 ns` | `177.3 ns` | `436.0 ns` | `1.16 µs` | 🥇 `Zinc` |
| `WildcardTail` | `82.29 ns` | `105.7 ns` | `135.9 ns` | `398.4 ns` | 🥇 `Zinc` |
| `WildcardTailNotFound` | `74.40 ns` | `106.5 ns` | `135.3 ns` | `200.6 ns` | 🥇 `Zinc` |

## Feature Paths

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `APIBindHeaderQueryJSON` | `2.37 µs` | `5.98 µs` | `3.09 µs` | `2.64 µs` | 🥇 `Zinc` |
| `APIBindInvalidJSON` | `762.2 ns` | `677.7 ns` | `805.6 ns` | `842.2 ns` | 🥇 `Gin` |
| `APIBindValidationFailure` | `701.8 ns` | `935.8 ns` | `890.1 ns` | `1.03 µs` | 🥇 `Zinc` |
| `APIBindMultipartHappyPath` | `11.2 µs` | `11.4 µs` | `11.5 µs` | `11.5 µs` | ≈ `Zinc` / `Gin` |
| `LargeJSONResponse` | `29.6 µs` | `32.9 µs` | `29.5 µs` | `29.7 µs` | ≈ `Echo` / `Zinc` |
| `LargeJSONBind` | `236.4 µs` | `248.9 µs` | `249.8 µs` | `249.0 µs` | 🥇 `Zinc` |
| `StaticFileHit` | `15.5 µs` | `25.8 µs` | `29.0 µs` | `15.4 µs` | ≈ `Chi` / `Zinc` |
| `StaticFileNotFound` | `1.55 µs` | `1.04 µs` | `2.86 µs` | `2.01 µs` | 🥇 `Gin` |
| `NestedGroupMiddlewareAPI` | `1.26 µs` | `3.27 µs` | `2.19 µs` | `2.88 µs` | 🥇 `Zinc` |
| `APIUnauthorizedReject` | `40.05 ns` | `49.39 ns` | `77.69 ns` | `211.3 ns` | 🥇 `Zinc` |

## Build-Time Highlights

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `RouteRegistrationStatic` | `60.9 µs` | `76.5 µs` | `343.0 µs` | `88.3 µs` | 🥇 `Zinc` |
| `RouteRegistrationParam` | `56.9 µs` | `52.5 µs` | `162.7 µs` | `70.8 µs` | 🥇 `Gin` |

## Scenario Build

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ScenarioBuild/Static157` | `46.3 µs` | `83.4 µs` | `205.1 µs` | `73.2 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GitHubAPI203` | `118.7 µs` | `146.3 µs` | `332.8 µs` | `173.7 µs` | 🥇 `Zinc` |
| `ScenarioBuild/GPlusAPI13` | `9.01 µs` | `8.85 µs` | `19.3 µs` | `12.9 µs` | ≈ `Gin` / `Zinc` |
| `ScenarioBuild/ParseAPI26` | `15.8 µs` | `15.2 µs` | `26.5 µs` | `16.9 µs` | 🥇 `Gin` |
| `ScenarioBuild/NestedAPI36` | `22.5 µs` | `24.1 µs` | `39.7 µs` | `25.3 µs` | 🥇 `Zinc` |
| `ScenarioBuild/ParamsAny24` | `24.8 µs` | `24.8 µs` | `44.7 µs` | `27.2 µs` | ≈ `Gin` / `Zinc` |

## Scenario Static

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ScenarioStatic/Static157` | `70.19 ns` | `125.4 ns` | `166.2 ns` | `273.8 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GitHubAPI203` | `72.24 ns` | `108.0 ns` | `152.0 ns` | `259.7 ns` | 🥇 `Zinc` |
| `ScenarioStatic/GPlusAPI13` | `85.37 ns` | `228.1 ns` | `160.1 ns` | `272.3 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParseAPI26` | `101.4 ns` | `104.2 ns` | `153.8 ns` | `286.8 ns` | 🥇 `Zinc` |
| `ScenarioStatic/NestedAPI36` | `74.92 ns` | `104.6 ns` | `150.7 ns` | `247.4 ns` | 🥇 `Zinc` |
| `ScenarioStatic/ParamsAny24` | `70.96 ns` | `107.2 ns` | `139.9 ns` | `247.4 ns` | 🥇 `Zinc` |

## Scenario Param

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ScenarioParam/GitHubAPI203` | `133.5 ns` | `167.5 ns` | `224.1 ns` | `595.5 ns` | 🥇 `Zinc` |
| `ScenarioParam/GPlusAPI13` | `105.4 ns` | `126.0 ns` | `172.5 ns` | `479.2 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParseAPI26` | `105.5 ns` | `127.9 ns` | `164.4 ns` | `491.7 ns` | 🥇 `Zinc` |
| `ScenarioParam/NestedAPI36` | `114.9 ns` | `139.9 ns` | `191.2 ns` | `534.0 ns` | 🥇 `Zinc` |
| `ScenarioParam/ParamsAny24` | `133.5 ns` | `167.2 ns` | `212.4 ns` | `569.7 ns` | 🥇 `Zinc` |

## Scenario Not Found

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ScenarioNotFound/Static157` | `101.0 ns` | `62.50 ns` | `625.4 ns` | `384.2 ns` | 🥇 `Gin` |
| `ScenarioNotFound/GitHubAPI203` | `73.62 ns` | `127.6 ns` | `622.8 ns` | `368.3 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/GPlusAPI13` | `89.84 ns` | `73.33 ns` | `655.0 ns` | `371.1 ns` | 🥇 `Gin` |
| `ScenarioNotFound/ParseAPI26` | `86.07 ns` | `109.8 ns` | `632.2 ns` | `365.9 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/NestedAPI36` | `75.08 ns` | `125.5 ns` | `626.1 ns` | `382.3 ns` | 🥇 `Zinc` |
| `ScenarioNotFound/ParamsAny24` | `114.0 ns` | `83.15 ns` | `623.6 ns` | `386.5 ns` | 🥇 `Gin` |

## Scenario 405

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `Scenario405/Static157` | `112.2 ns` | `113.6 ns` | `939.3 ns` | `369.2 ns` | ≈ `Zinc` / `Gin` |
| `Scenario405/GitHubAPI203` | `101.8 ns` | `178.7 ns` | `877.0 ns` | `339.2 ns` | 🥇 `Zinc` |
| `Scenario405/GPlusAPI13` | `111.6 ns` | `185.3 ns` | `910.6 ns` | `448.3 ns` | 🥇 `Zinc` |
| `Scenario405/ParseAPI26` | `101.0 ns` | `170.4 ns` | `885.0 ns` | `363.8 ns` | 🥇 `Zinc` |
| `Scenario405/NestedAPI36` | `111.8 ns` | `182.7 ns` | `877.8 ns` | `348.1 ns` | 🥇 `Zinc` |
| `Scenario405/ParamsAny24` | `111.0 ns` | `122.1 ns` | `899.3 ns` | `318.4 ns` | 🥇 `Zinc` |

## Scenario Mixed

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ScenarioAll/Static157` | `71.15 ns` | `134.2 ns` | `173.5 ns` | `293.9 ns` | 🥇 `Zinc` |
| `ScenarioAll/GitHubAPI203` | `129.8 ns` | `165.0 ns` | `198.8 ns` | `504.9 ns` | 🥇 `Zinc` |
| `ScenarioAll/GPlusAPI13` | `120.1 ns` | `181.4 ns` | `167.6 ns` | `421.2 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParseAPI26` | `105.4 ns` | `126.7 ns` | `160.7 ns` | `394.7 ns` | 🥇 `Zinc` |
| `ScenarioAll/NestedAPI36` | `113.3 ns` | `131.7 ns` | `170.3 ns` | `410.4 ns` | 🥇 `Zinc` |
| `ScenarioAll/ParamsAny24` | `143.1 ns` | `149.6 ns` | `185.5 ns` | `484.5 ns` | 🥇 `Zinc` |

## Parallel In-Process

| Benchmark | Zinc | Gin | Echo | Chi | Result |
|---|---:|---:|---:|---:|---|
| `ParallelStaticRoute` | `12.18 ns` | `42.69 ns` | `57.49 ns` | `157.6 ns` | 🥇 `Zinc` |
| `ParallelRouterParam` | `13.21 ns` | `55.74 ns` | `65.39 ns` | `301.3 ns` | 🥇 `Zinc` |
| `ParallelMiddlewareChain` | `58.91 ns` | `190.5 ns` | `259.3 ns` | `908.5 ns` | 🥇 `Zinc` |
| `ParallelAPIHappyPath` | `336.1 ns` | `848.1 ns` | `632.1 ns` | `1.30 µs` | 🥇 `Zinc` |
