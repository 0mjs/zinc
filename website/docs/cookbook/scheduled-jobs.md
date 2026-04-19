---
slug: /cookbook/scheduled-jobs
title: Scheduled Jobs
description: Run a cron scheduler and a Zinc HTTP API from one process, with graceful shutdown.
---

A compact service that runs background jobs on a cron schedule, accepts ad-hoc jobs through an HTTP endpoint, and shuts both down cleanly on `SIGINT` / `SIGTERM`.

This recipe covers:

- cron schedules with the `jobs` package (`@every`, cron expressions, named aliases)
- ad-hoc jobs enqueued from an HTTP handler with a typed payload
- a status endpoint that reports pending and failed counts
- shared context-based shutdown for the queue runner and the HTTP server

## Setup

```bash
go mod init zinc-jobs
go get github.com/0mjs/zinc
```

## Application

```go
package main

import (
	"context"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/jobs"
)

type Report struct {
	Since time.Time `json:"since"`
	Until time.Time `json:"until"`
}

type store struct {
	mu       sync.Mutex
	reports  []Report
	sessions int
}

func (s *store) addReport(r Report) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reports = append(s.reports, r)
}

func (s *store) clearExpiredSessions() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	cleared := s.sessions
	s.sessions = 0
	return cleared
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	data := &store{sessions: 42}
	queue := jobs.New()

	runReport := func(r Report) error {
		log.Printf("report %s → %s",
			r.Since.Format(time.RFC3339),
			r.Until.Format(time.RFC3339),
		)
		data.addReport(r)
		return nil
	}

	if _, err := queue.Cron("reports.daily", "0 0 * * *", func(ctx context.Context) error {
		return runReport(Report{
			Since: time.Now().Add(-24 * time.Hour),
			Until: time.Now(),
		})
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := queue.Cron("cache.refresh", "@every 2m", func(ctx context.Context) error {
		log.Println("refreshing cache")
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if _, err := queue.Cron("sessions.cleanup", "@every 15m", func(ctx context.Context) error {
		n := data.clearExpiredSessions()
		log.Printf("cleared %d expired sessions", n)
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := queue.Handle("reports.run", func(ctx context.Context, job jobs.Job) error {
		var r Report
		if err := job.Decode(&r); err != nil {
			return err
		}
		return runReport(r)
	}); err != nil {
		log.Fatal(err)
	}

	runner, err := queue.Start(ctx, 2)
	if err != nil {
		log.Fatal(err)
	}

	app := zinc.New()

	app.Get("/status", func(c *zinc.Context) error {
		return c.JSON(map[string]any{
			"pending": queue.Pending(),
			"failed":  len(queue.Failed()),
		})
	})

	app.Post("/reports/run", func(c *zinc.Context) error {
		var r Report
		if err := c.Bind().JSON(&r); err != nil {
			return err
		}
		if r.Since.IsZero() || r.Until.IsZero() {
			return zinc.ErrBadRequest.WithMessage("since and until are required")
		}
		if _, err := queue.Enqueue(c.Request().Context(), "reports.run", r); err != nil {
			return err
		}
		return c.Status(zinc.StatusAccepted).JSON(map[string]string{"status": "queued"})
	})

	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown: %v", err)
	}
	if err := runner.Stop(shutdownCtx); err != nil {
		log.Printf("queue shutdown: %v", err)
	}
}
```

## Try it

```bash
curl http://localhost:8080/status

curl -X POST http://localhost:8080/reports/run \
  -H "Content-Type: application/json" \
  -d '{"since":"2025-01-01T00:00:00Z","until":"2025-01-02T00:00:00Z"}'
```

Press `Ctrl+C` to trigger shutdown. The HTTP server stops accepting new connections first, then the queue runner drains and exits.

## How the pieces fit

**`queue.Cron(name, spec, fn)`** registers a handler and a schedule in one call. Use it when the job takes no input and the schedule is the only trigger. `spec` accepts five-field cron expressions (`0 0 * * *`), `@every <duration>`, `@hourly`, `@daily`, and `@weekly`.

**`queue.Handle(name, fn)`** registers a handler that receives a `Job`. Use it when the job has a payload or when the same handler is triggered by multiple sources. Pair it with `queue.Schedule(name, spec, payload)` for recurring jobs that need input.

**`queue.Enqueue(ctx, name, payload)`** dispatches the handler ad-hoc. The payload is JSON-encoded automatically, and handlers read it with `job.Decode(&v)`.

**`queue.Pending()` / `queue.Failed()`** are safe to call from HTTP handlers; they return snapshots and never block on worker state.

**Graceful shutdown.** The queue runner and the HTTP server share `ctx` from `signal.NotifyContext`, so a single signal cancels both. `runner.Stop` blocks until in-flight jobs finish or the shutdown context expires — a separate timeout context keeps that bounded.

## Going further

- Swap the in-memory store for SQLite or Postgres. The handler signature does not change.
- Attach `queue.Config.EventHandler` to ship `EventStarted` / `EventFailed` to your logger or metrics pipeline.
- Use `jobs.ExponentialBackoff(base, max)` in `queue.Config.Backoff` for jobs that hit flaky dependencies.
- Register a `ScheduleConfig{MaxAttempts: n}` on a cron to retry transient failures on the next tick.
