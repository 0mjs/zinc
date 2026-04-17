package jobs

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrJobNameRequired      = errors.New("zincjobs: job name is required")
	ErrHandlerRequired      = errors.New("zincjobs: handler is required")
	ErrHandlerExists        = errors.New("zincjobs: handler already registered")
	ErrHandlerNotFound      = errors.New("zincjobs: handler not found")
	ErrInvalidWorkerCount   = errors.New("zincjobs: worker count must be greater than zero")
	ErrInvalidMaxAttempts   = errors.New("zincjobs: max attempts must be greater than zero")
	ErrInvalidSchedule      = errors.New("zincjobs: invalid schedule")
	ErrQueueAlreadyRunning  = errors.New("zincjobs: queue is already running")
	ErrScheduleExists       = errors.New("zincjobs: schedule already registered")
	ErrScheduleNotFound     = errors.New("zincjobs: schedule not found")
	ErrScheduleIDRequired   = errors.New("zincjobs: schedule id is required")
	ErrScheduleNameRequired = errors.New("zincjobs: schedule name is required")
)

// HandlerFunc processes a job attempt.
type HandlerFunc func(ctx context.Context, job Job) error

// CronFunc processes a scheduled function job.
type CronFunc func(ctx context.Context) error

// BackoffFunc returns the delay before the next attempt after a failed job.
type BackoffFunc func(attempt int, err error) time.Duration

// EventHandler observes job lifecycle events emitted by workers and schedules.
type EventHandler func(Event)

// EventType identifies a job lifecycle event.
type EventType string

const (
	EventEnqueued  EventType = "enqueued"
	EventStarted   EventType = "started"
	EventCompleted EventType = "completed"
	EventRetrying  EventType = "retrying"
	EventFailed    EventType = "failed"
	EventScheduled EventType = "scheduled"
)

// Event describes a job lifecycle transition.
type Event struct {
	Type      EventType
	Job       Job
	Error     error
	NextRunAt time.Time
}

// Job is the immutable job snapshot passed to handlers.
type Job struct {
	ID          string
	Name        string
	Queue       string
	Payload     []byte
	CreatedAt   time.Time
	RunAt       time.Time
	Attempts    int
	MaxAttempts int
	LastError   string
}

// Decode unmarshals the JSON job payload into v.
func (j Job) Decode(v any) error {
	if len(j.Payload) == 0 {
		return nil
	}
	return json.Unmarshal(j.Payload, v)
}

// Config controls queue defaults.
type Config struct {
	DefaultQueue       string
	DefaultMaxAttempts int
	Backoff            BackoffFunc
	EventHandler       EventHandler
	Now                func() time.Time
}

// DefaultConfig is used by New.
var DefaultConfig = Config{
	DefaultQueue:       "default",
	DefaultMaxAttempts: 1,
	Backoff:            FixedBackoff(time.Second),
	Now:                time.Now,
}

// EnqueueConfig customizes a single enqueued job.
type EnqueueConfig struct {
	ID          string
	Queue       string
	RunAt       time.Time
	MaxAttempts int
}

// ScheduleConfig customizes jobs created by a schedule.
type ScheduleConfig struct {
	ID          string
	Queue       string
	MaxAttempts int
}

// Schedule describes a registered recurring job.
type Schedule struct {
	ID          string
	Name        string
	Spec        string
	Queue       string
	NextRunAt   time.Time
	MaxAttempts int
}

// FixedBackoff returns a backoff function that always uses delay.
func FixedBackoff(delay time.Duration) BackoffFunc {
	if delay < 0 {
		delay = 0
	}
	return func(int, error) time.Duration {
		return delay
	}
}

// ExponentialBackoff returns a backoff function that doubles from base up to max.
func ExponentialBackoff(base, max time.Duration) BackoffFunc {
	if base < 0 {
		base = 0
	}
	if max < 0 {
		max = 0
	}
	return func(attempt int, _ error) time.Duration {
		if attempt <= 1 || base == 0 {
			return base
		}
		delay := base
		for i := 1; i < attempt; i++ {
			if max > 0 && delay >= max/2 {
				return max
			}
			delay *= 2
		}
		if max > 0 && delay > max {
			return max
		}
		return delay
	}
}

func encodePayload(payload any) ([]byte, error) {
	switch value := payload.(type) {
	case nil:
		return nil, nil
	case []byte:
		return append([]byte(nil), value...), nil
	case json.RawMessage:
		return append([]byte(nil), value...), nil
	default:
		return json.Marshal(payload)
	}
}

func cloneJob(job Job) Job {
	job.Payload = append([]byte(nil), job.Payload...)
	return job
}

func newID(prefix string) string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err == nil {
		return prefix + "_" + hex.EncodeToString(b[:])
	}
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
