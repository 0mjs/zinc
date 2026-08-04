---
title: Installation
description: Add Zinc to a Go module and confirm your first build.
---

Zinc is a normal Go module.

## Requirements

- Go `1.25` or newer
- An existing Go module, or a new one created with `go mod init`

## Install

```bash
go get github.com/0mjs/zinc
```

Most applications also use Zinc's first-party middleware package:

```go
import (
	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/middleware"
)
```

The middleware package is part of the same module, so no extra install command is needed.

## Create a module

For a new project:

```bash
mkdir hello-zinc
cd hello-zinc
go mod init example.com/hello-zinc
go get github.com/0mjs/zinc
```

Then create `main.go` and continue with the [Quick Start](./quick-start).

## Check the version

Use the normal Go tooling to inspect the selected version:

```bash
go list -m github.com/0mjs/zinc
```

Zinc is currently pre-1.0, so read release notes before upgrading production services.
