# Benchmark Process

This document exists to stop benchmark work from drifting, stacking, and regressing.

The target peer set for performance work is:

- `Gin`
- `Echo`
- `Chi`

`HttpRouter` and `ServeMux` are useful reference points, but they are not the optimization target for acceptance decisions.

## What Has Been Going Wrong

We have repeatedly made the same process mistakes:

- working on a dirty tree and then treating that tree as the new baseline
- accepting targeted benchmark wins before rerunning the full relevant suite
- using `count=1` full-suite output to make keep/revert decisions on near-tie rows
- optimizing multiple coupled concerns at once
  - registration
  - `404/405`
  - cold param lookup
  - throughput
- moving cost instead of deleting it
  - less registration work often means more miss-path work
  - more precomputed allow data often means slower `Add`
- updating benchmark docs before the current tree is clearly better than the last stable snapshot

This is why progress has felt circular. The remaining Gin wins are concentrated in tightly coupled paths, so local tweaks can easily help one row and hurt another.

## Core Rules

1. Always start from a known stable baseline.
2. Change one benchmark cluster at a time.
3. Do not keep a change because one targeted row improved.
4. Do not treat a `count=1` near-tie as a real win or loss.
5. Do not update [BENCKMARKS.md](/Users/matt/dev/oss/zinc/BENCKMARKS.md) or [README.md](/Users/matt/dev/oss/zinc/README.md) until the full relevant suite has been rerun on the kept tree.

## Stable Baseline

Before any optimization pass:

1. Record the git commit or tag being used as the baseline.
2. Save the raw benchmark output for that baseline.
3. Record the current focus score:
   - non-throughput: Zinc vs `Gin/Echo/Chi`
   - throughput: Zinc vs `Gin/Echo/Chi`
4. Write down the exact rows being targeted.

If the tree is dirty, do not treat it as the baseline until it has passed the full suite and been explicitly accepted.

## Benchmark Groups

Work only one group at a time:

- static registration
- static miss / `405`
- cold param lookup
- deep param routing
- binder / JSON path
- throughput

Throughput is last unless the current task is explicitly throughput-only.

## Required Workflow Per Pass

### 1. Pick One Hypothesis

Each pass must have a single sentence hypothesis.

Example:

`Make exact static misses cheaper without adding registration-time bookkeeping.`

If the change needs more than one hypothesis, split it into multiple passes.

### 2. Define Target Rows

Each pass must declare the rows it is supposed to improve.

Example static miss / `405` target rows:

- `BenchmarkNotFound`
- `BenchmarkLargeRouteSetNotFound`
- `BenchmarkLargeRouteSetMethodMismatch`
- `BenchmarkScenarioRouteSetNotFound/*`
- `BenchmarkScenarioRouteSetMethodMismatch/*`

### 3. Define Safety Rows

Each pass must also declare rows that must not regress.

Example safety rows:

- `BenchmarkParam10`
- `BenchmarkNestedGroupParam`
- `BenchmarkRouteRegistrationStatic`
- `BenchmarkRouteRegistrationParam`
- `BenchmarkScenarioRouteSetAll/*`

### 4. Run Targeted Validation First

Run the target slice and the safety slice with `-count=3` before touching the full suite.

Rules:

- target rows must improve clearly, not marginally
- safety rows must stay flat or better
- if a target row improves but safety rows regress, reject the change

### 5. Promotion Gate

Only after the targeted gate passes:

- rerun the full non-throughput suite
- compare it against the last accepted non-throughput snapshot

If the pass is throughput-related, rerun throughput after the non-throughput check.

## Acceptance Rules

A pass is keepable only if all of the following are true:

- `go test ./...` passes
- the target slice improves on `count=3`
- the safety slice does not materially regress on `count=3`
- the full relevant suite does not reduce Zinc’s accepted score against `Gin/Echo/Chi`

If the full score gets worse, revert the pass unless there is an explicit decision to trade score for a strategic win.

## Near-Tie Policy

Do not treat tiny differences as meaningful.

For benchmark decisions:

- under `2%` difference: near tie
- `2%` to `5%`: weak signal, needs repeat confirmation
- over `5%`: meaningful signal

This matters because several recent row flips were only around `0.4%` to `1.9%`, which is not a strong enough basis for accepting broader regressions.

## Full-Suite Policy

There are only two suite-level checkpoints that matter:

- full non-throughput
- full throughput

Do not rerun throughput during router work unless the router pass is already accepted on non-throughput or the user explicitly asks for it.

## Documentation Policy

Do not update docs after a targeted benchmark win.

Only update:

- [BENCKMARKS.md](/Users/matt/dev/oss/zinc/BENCKMARKS.md)
- [README.md](/Users/matt/dev/oss/zinc/README.md)

after:

- the pass is accepted
- the full relevant suite has been rerun
- the raw output has been saved

## Experiment Log Template

Every pass should write down this checklist before implementation:

```md
Hypothesis:
Target group:
Files touched:
Target rows:
Safety rows:
Baseline raw output:
Decision rule:
```

And after validation:

```md
Target result:
Safety result:
Full-suite result:
Keep or revert:
Reason:
```

## Current Practical Guidance

For Zinc right now, the safe order remains:

1. static miss / `405`
2. cold param lookup
3. throughput

Do not reopen registration again unless the goal is explicitly registration, because that path has already shown how easy it is to trade one win for another regression.

## Bottom Line

The problem is not that Zinc cannot beat Gin.

The problem is that the remaining Gin wins live in a small set of coupled paths, and we have been letting partial wins override full-suite discipline.

From here on:

- one hypothesis
- one cluster
- fixed target rows
- fixed safety rows
- `count=3` before full-suite
- revert anything that lowers the accepted score
