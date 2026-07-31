# Contributing to Zinc

Thanks for helping improve Zinc. Bug fixes, tests, documentation, and focused feature proposals are all welcome.

## Before you start

Search the [open issues](https://github.com/0mjs/zinc/issues) before opening a new one.

Small fixes can go straight to a pull request. Please open an issue first when a change:

- alters the public API
- changes routing or middleware semantics
- adds a dependency
- makes a broad internal refactor
- is primarily a performance optimization

This gives us a chance to agree on the problem and the shape of the solution before either side spends much time on it.

## Development setup

Zinc requires Go 1.25 or newer.

Fork the repository, then create a branch from `dev`:

```sh
git clone https://github.com/YOUR_NAME/zinc.git
cd zinc
git remote add upstream https://github.com/0mjs/zinc.git
git fetch upstream
git switch -c your-change upstream/dev
```

Run the test suite once before making changes:

```sh
go test ./...
```

## Working on Zinc

Keep changes small enough to review. A pull request should solve one problem and avoid unrelated cleanup.

When changing the public API, keep Zinc's direction in mind:

- build on standard `net/http` contracts
- keep the common path short and explicit
- prefer compile-time types over runtime handler adaptation
- preserve access to standard requests, writers, handlers, contexts, and servers
- treat performance as a constraint without trading away clear Go code

New behavior needs tests. A bug fix should include a test that fails before the fix and passes afterwards.

Format Go code before committing:

```sh
gofmt -w path/to/changed.go
```

## Checks

Run these for Go changes:

```sh
go test ./...
go test -race ./...
go vet ./...
```

If you change the documentation site:

```sh
cd website
npm ci
npm run build
```

Do not commit generated site output, local benchmark binaries, coverage files, or editor settings.

## Performance changes

The comparison suite is a separate Go module under `benchmarks/`.

Run the main suite with:

```sh
cd benchmarks
go test -run=^$ -bench . -benchmem
```

The RPS benchmark is intentionally optional:

```sh
go test -tags rps -run=^$ -bench '^BenchmarkRequestsPerSecond$'
```

For a performance pull request, include repeated before-and-after results for the affected benchmarks. Use at least ten samples and `benchstat` when the claimed improvement is small. A faster result is not enough if it adds request-time allocations or regresses another common route shape.

## Pull requests

Open pull requests against `dev` and include:

- the problem being solved
- a short explanation of the approach
- the tests you ran
- benchmark results when a hot path changed
- documentation or migration notes when public behavior changed

Keep commits understandable, but do not spend time manufacturing a perfect history. Maintainers may squash a pull request when merging.

## License

By contributing, you agree that your contribution is licensed under Zinc's [MIT License](./LICENSE).
