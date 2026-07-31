---
id: benchmarks
title: Benchmarks
description: Current Zinc benchmark results against Gin, Echo, and Chi.
sidebar_position: 1
---

Zinc maintains an in-process comparison suite against Gin, Echo, and Chi. Each framework performs equivalent work through its idiomatic API.

## Current results

The latest full run on an Apple M1 Pro recorded:

- lowest latency for Zinc in `62/77` comparable rows
- lowest latency, or a result within 2% of it, in `63/77` rows
- zero request-time allocations on Zinc's primary static, parameter, and not-found dispatch paths

Lower values are better. Results vary by workload and machine, and small differences should be confirmed with repeated samples.

## What the suite covers

- core routing
- params and wildcard routes
- route misses and method mismatch
- build-time route registration
- realistic route-set scenarios
- binding and response feature paths
- static file hit and miss

## Full report

The [complete benchmark report](https://github.com/0mjs/zinc/blob/main/BENCKMARKS.md) contains the environment, command, scorecard, and all 77 comparable rows from the current run.

To run the suite locally:

```sh
cd benchmarks
go test -run=^$ -bench=. -benchmem -count=1
```
