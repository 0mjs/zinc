---
title: Health and Readiness Checks
description: Separate process liveness from dependency readiness for deployments and load balancers.
---

Liveness answers “is this process running?” Readiness answers “can this process
serve real traffic?” Keeping them separate prevents a temporary database outage
from restarting a healthy process.

## Setup

```bash
mkdir zinc-health
cd zinc-health
go mod init example.com/zinc-health
go get github.com/0mjs/zinc
go get github.com/jackc/pgx/v5/stdlib
```

## Application

```go
package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/0mjs/zinc"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	db, err := sql.Open("pgx", "postgres://localhost/app?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := zinc.New()

	app.Get("/live", func(c *zinc.Context) error {
		return c.JSON(zinc.Map{"status": "up"})
	})

	app.Get("/ready", func(c *zinc.Context) error {
		ctx, cancel := context.WithTimeout(c.Context(), 500*time.Millisecond)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			return c.Status(zinc.StatusServiceUnavailable).JSON(zinc.Map{
				"status":   "unavailable",
				"database": "down",
			})
		}

		return c.JSON(zinc.Map{
			"status":   "ready",
			"database": "up",
		})
	})

	log.Fatal(app.Listen(":8080"))
}
```

Swap the PostgreSQL driver and connection string for the database your service
already uses; the Zinc handlers are unchanged.

## Probe it

```bash
curl -i http://localhost:8080/live
curl -i http://localhost:8080/ready
```

Point an orchestrator's liveness probe at `/live` and readiness probe at
`/ready`. Keep readiness checks bounded with short timeouts, and include only
dependencies required to serve traffic.
