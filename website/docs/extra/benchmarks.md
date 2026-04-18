---
id: benchmarks
title: Benchmarks
description: What Zinc measures, what the current peer-only suite says, and where to find the full data.
sidebar_position: 1
---

Zinc maintains a peer-only benchmark suite against:

- `Gin`
- `Echo`
- `Chi`

## Current headline

From the latest full local rerun:

- Zinc wins `65/85` comparable rows overall
- Zinc wins `65/77` non-throughput rows
- Zinc wins `0/8` throughput rows

The strongest Zinc results are on request-path and API-path latency. Throughput is currently the weakest category in the suite.

## What the suite covers

- core routing
- params and wildcard routes
- route misses and method mismatch
- build-time route registration
- realistic route-set scenarios
- binding and response feature paths
- static file hit and miss
- loopback throughput at multiple concurrency levels

## Full report

The complete benchmark report lives in the repository:

- [BENCKMARKS.md](https://github.com/0mjs/zinc/blob/main/BENCKMARKS.md)

That file includes:

- the raw scorecard
- remaining peer gaps
- per-benchmark tables
- benchmark environment details
