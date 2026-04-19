package jobs

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueueRunsEnqueuedJob(t *testing.T) {
	queue := New()
	done := make(chan string, 1)

	if err := queue.Handle("email.send", func(_ context.Context, job Job) error {
		var payload struct {
			To string `json:"to"`
		}
		if err := job.Decode(&payload); err != nil {
			return err
		}
		done <- payload.To
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	if _, err := queue.Enqueue(ctx, "email.send", map[string]string{"to": "sam@example.com"}); err != nil {
		t.Fatal(err)
	}

	select {
	case got := <-done:
		if got != "sam@example.com" {
			t.Fatalf("payload=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for job")
	}
}

func TestQueueRetriesFailedJob(t *testing.T) {
	var attempts int32
	queue := NewWithConfig(Config{
		DefaultMaxAttempts: 1,
		Backoff:            FixedBackoff(time.Millisecond),
	})
	done := make(chan struct{}, 1)

	if err := queue.Handle("flaky", func(_ context.Context, job Job) error {
		got := atomic.AddInt32(&attempts, 1)
		if job.Attempts != int(got) {
			t.Fatalf("job attempts=%d got=%d", job.Attempts, got)
		}
		if got == 1 {
			return errors.New("temporary failure")
		}
		done <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	if _, err := queue.Enqueue(ctx, "flaky", nil, EnqueueConfig{MaxAttempts: 2}); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for retry")
	}

	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Fatalf("attempts=%d", got)
	}
	if failed := queue.Failed(); len(failed) != 0 {
		t.Fatalf("failed jobs=%d", len(failed))
	}
}

func TestExponentialBackoff(t *testing.T) {
	backoff := ExponentialBackoff(10*time.Millisecond, 35*time.Millisecond)
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 0, want: 10 * time.Millisecond},
		{attempt: 1, want: 10 * time.Millisecond},
		{attempt: 2, want: 20 * time.Millisecond},
		{attempt: 3, want: 35 * time.Millisecond},
		{attempt: 4, want: 35 * time.Millisecond},
	}
	for _, tc := range cases {
		if got := backoff(tc.attempt, errors.New("boom")); got != tc.want {
			t.Fatalf("attempt %d backoff=%s want %s", tc.attempt, got, tc.want)
		}
	}

	if got := ExponentialBackoff(-time.Second, -time.Second)(3, nil); got != 0 {
		t.Fatalf("negative backoff=%s", got)
	}
}

func TestQueuePendingAndRunnerWait(t *testing.T) {
	queue := New()
	if got := queue.Pending(); got != 0 {
		t.Fatalf("initial pending=%d", got)
	}
	if _, err := queue.Enqueue(context.Background(), "missing.handler", nil); err != nil {
		t.Fatal(err)
	}
	if got := queue.Pending(); got != 1 {
		t.Fatalf("pending=%d", got)
	}

	runner := &Runner{done: make(chan struct{})}
	waited := make(chan struct{})
	go func() {
		runner.Wait()
		close(waited)
	}()

	select {
	case <-waited:
		t.Fatal("Wait returned before runner was done")
	case <-time.After(10 * time.Millisecond):
	}

	close(runner.done)
	select {
	case <-waited:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for Wait")
	}

	var nilRunner *Runner
	nilRunner.Wait()
}

func TestQueueStoresFailedJob(t *testing.T) {
	failed := make(chan Job, 1)
	queue := NewWithConfig(Config{
		Backoff: FixedBackoff(time.Millisecond),
		EventHandler: func(event Event) {
			if event.Type == EventFailed {
				failed <- event.Job
			}
		},
	})

	if err := queue.Handle("always.fail", func(context.Context, Job) error {
		return errors.New("boom")
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	if _, err := queue.Enqueue(ctx, "always.fail", nil); err != nil {
		t.Fatal(err)
	}

	select {
	case job := <-failed:
		if job.LastError != "boom" {
			t.Fatalf("last error=%q", job.LastError)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for failed job")
	}

	if failed := queue.Failed(); len(failed) != 1 {
		t.Fatalf("failed jobs=%d", len(failed))
	}
}

func TestEnqueueInDelaysJob(t *testing.T) {
	queue := New()
	done := make(chan time.Time, 1)

	if err := queue.Handle("delayed", func(context.Context, Job) error {
		done <- time.Now()
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	start := time.Now()
	if _, err := queue.EnqueueIn(ctx, "delayed", nil, 30*time.Millisecond); err != nil {
		t.Fatal(err)
	}

	select {
	case <-done:
		t.Fatal("job ran before delay")
	case <-time.After(10 * time.Millisecond):
	}

	select {
	case ranAt := <-done:
		if ranAt.Sub(start) < 25*time.Millisecond {
			t.Fatalf("job ran too early after %s", ranAt.Sub(start))
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for delayed job")
	}
}

func TestScheduleEveryEnqueuesJobs(t *testing.T) {
	queue := New()
	done := make(chan struct{}, 2)

	if err := queue.Handle("tick", func(context.Context, Job) error {
		done <- struct{}{}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := queue.Schedule("tick", "@every 10ms", nil); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	for i := 0; i < 2; i++ {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for scheduled job")
		}
	}
}

func TestCronSchedulesFunction(t *testing.T) {
	queue := New()
	done := make(chan struct{}, 1)

	schedule, err := queue.Cron("tick", "10ms", func(context.Context) error {
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if schedule.ID != "tick" {
		t.Fatalf("schedule id=%q", schedule.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := queue.Start(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	defer runner.Stop(context.Background())

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for cron function")
	}
}

func TestCronRollsBackHandlerWhenScheduleFails(t *testing.T) {
	queue := New()

	if _, err := queue.Cron("tick", "bad", func(context.Context) error {
		return nil
	}); !errors.Is(err, ErrInvalidSchedule) {
		t.Fatalf("expected ErrInvalidSchedule, got %v", err)
	}

	if err := queue.Handle("tick", func(context.Context, Job) error {
		return nil
	}); err != nil {
		t.Fatalf("handler was not rolled back: %v", err)
	}
}

func TestUnscheduleStopsFutureRuns(t *testing.T) {
	queue := New()
	if err := queue.Handle("tick", func(context.Context, Job) error {
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	schedule, err := queue.Schedule("tick", "@every 10ms", nil, ScheduleConfig{ID: "tick"})
	if err != nil {
		t.Fatal(err)
	}
	if schedule.ID != "tick" {
		t.Fatalf("schedule id=%q", schedule.ID)
	}
	if err := queue.Unschedule("tick"); err != nil {
		t.Fatal(err)
	}
	if err := queue.Unschedule("tick"); !errors.Is(err, ErrScheduleNotFound) {
		t.Fatalf("expected ErrScheduleNotFound, got %v", err)
	}
}

func TestScheduleRejectsDuplicateID(t *testing.T) {
	queue := New()
	if _, err := queue.Schedule("tick", "@every 10ms", nil, ScheduleConfig{ID: "tick"}); err != nil {
		t.Fatal(err)
	}
	if _, err := queue.Schedule("tick", "@every 10ms", nil, ScheduleConfig{ID: "tick"}); !errors.Is(err, ErrScheduleExists) {
		t.Fatalf("expected ErrScheduleExists, got %v", err)
	}
}

func TestParseCronScheduleNext(t *testing.T) {
	spec, err := ParseSchedule("*/15 9-17 * * mon-fri")
	if err != nil {
		t.Fatal(err)
	}

	after := time.Date(2026, 4, 17, 9, 7, 0, 0, time.UTC)
	want := time.Date(2026, 4, 17, 9, 15, 0, 0, time.UTC)
	if got := spec.Next(after); !got.Equal(want) {
		t.Fatalf("next=%s want=%s", got, want)
	}
}

func TestParseCronDoesNotMatchNormalizedWeekdayInMinuteField(t *testing.T) {
	spec, err := ParseSchedule("0 * * * *")
	if err != nil {
		t.Fatal(err)
	}

	after := time.Date(2026, 4, 17, 9, 7, 0, 0, time.UTC)
	want := time.Date(2026, 4, 17, 10, 0, 0, 0, time.UTC)
	if got := spec.Next(after); !got.Equal(want) {
		t.Fatalf("next=%s want=%s", got, want)
	}
}

func TestParseScheduleAcceptsDuration(t *testing.T) {
	spec, err := ParseSchedule("10s")
	if err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, 4, 17, 9, 0, 0, 0, time.UTC)
	want := now.Add(10 * time.Second)
	if got := spec.Next(now); !got.Equal(want) {
		t.Fatalf("next=%s want=%s", got, want)
	}
}

func TestParseScheduleRejectsBadSpecs(t *testing.T) {
	for _, spec := range []string{"", "* * *", "@every 0s", "61 * * * *"} {
		if _, err := ParseSchedule(spec); !errors.Is(err, ErrInvalidSchedule) {
			t.Fatalf("ParseSchedule(%q) error=%v", spec, err)
		}
	}
}
