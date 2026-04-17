package jobs

import (
	"context"
	"sync"
	"time"
)

// Queue stores jobs, handlers, and schedules for a worker pool.
type Queue struct {
	mu        sync.Mutex
	notify    chan struct{}
	config    Config
	handlers  map[string]HandlerFunc
	pending   []queuedJob
	failed    []Job
	schedules map[string]*scheduleState
	running   bool
}

type queuedJob struct {
	job Job
}

type scheduleState struct {
	snapshot Schedule
	parsed   ScheduleSpec
	payload  []byte
}

// New creates a queue with DefaultConfig.
func New() *Queue {
	return NewWithConfig(DefaultConfig)
}

// NewWithConfig creates a queue with explicit defaults.
func NewWithConfig(config Config) *Queue {
	cfg := normalizeConfig(config)
	return &Queue{
		notify:    make(chan struct{}),
		config:    cfg,
		handlers:  map[string]HandlerFunc{},
		schedules: map[string]*scheduleState{},
	}
}

// Handle registers a handler for a named job.
func (q *Queue) Handle(name string, handler HandlerFunc) error {
	if name == "" {
		return ErrJobNameRequired
	}
	if handler == nil {
		return ErrHandlerRequired
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if _, exists := q.handlers[name]; exists {
		return ErrHandlerExists
	}
	q.handlers[name] = handler
	return nil
}

// Cron registers a named scheduled function.
func (q *Queue) Cron(name, spec string, handler CronFunc, config ...ScheduleConfig) (Schedule, error) {
	if handler == nil {
		return Schedule{}, ErrHandlerRequired
	}

	cfg := firstScheduleConfig(config)
	if cfg.ID == "" {
		cfg.ID = name
	}

	if err := q.Handle(name, func(ctx context.Context, _ Job) error {
		return handler(ctx)
	}); err != nil {
		return Schedule{}, err
	}

	schedule, err := q.Schedule(name, spec, nil, cfg)
	if err != nil {
		q.mu.Lock()
		delete(q.handlers, name)
		q.mu.Unlock()
		return Schedule{}, err
	}
	return schedule, nil
}

// Enqueue adds a job to the queue.
func (q *Queue) Enqueue(ctx context.Context, name string, payload any, config ...EnqueueConfig) (Job, error) {
	return q.enqueue(ctx, name, payload, firstEnqueueConfig(config))
}

// EnqueueIn adds a job that becomes runnable after delay.
func (q *Queue) EnqueueIn(ctx context.Context, name string, payload any, delay time.Duration, config ...EnqueueConfig) (Job, error) {
	cfg := firstEnqueueConfig(config)
	cfg.RunAt = q.now().Add(delay)
	return q.enqueue(ctx, name, payload, cfg)
}

// Schedule registers a recurring job using a cron expression or @every duration.
func (q *Queue) Schedule(name, spec string, payload any, config ...ScheduleConfig) (Schedule, error) {
	if name == "" {
		return Schedule{}, ErrScheduleNameRequired
	}

	parsed, err := ParseSchedule(spec)
	if err != nil {
		return Schedule{}, err
	}

	body, err := encodePayload(payload)
	if err != nil {
		return Schedule{}, err
	}

	cfg := firstScheduleConfig(config)
	cfg = q.normalizeScheduleConfig(cfg)
	if cfg.MaxAttempts <= 0 {
		return Schedule{}, ErrInvalidMaxAttempts
	}

	schedule := Schedule{
		ID:          cfg.ID,
		Name:        name,
		Spec:        spec,
		Queue:       cfg.Queue,
		NextRunAt:   parsed.Next(q.now()),
		MaxAttempts: cfg.MaxAttempts,
	}

	q.mu.Lock()
	if _, exists := q.schedules[schedule.ID]; exists {
		q.mu.Unlock()
		return Schedule{}, ErrScheduleExists
	}
	q.schedules[schedule.ID] = &scheduleState{
		snapshot: schedule,
		parsed:   parsed,
		payload:  body,
	}
	q.signalLocked()
	q.mu.Unlock()

	return schedule, nil
}

// Unschedule removes a recurring job by id.
func (q *Queue) Unschedule(id string) error {
	if id == "" {
		return ErrScheduleIDRequired
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	if _, ok := q.schedules[id]; !ok {
		return ErrScheduleNotFound
	}
	delete(q.schedules, id)
	q.signalLocked()
	return nil
}

// Pending returns the number of jobs waiting to run.
func (q *Queue) Pending() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pending)
}

// Failed returns failed jobs that exhausted their attempts.
func (q *Queue) Failed() []Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Job, 0, len(q.failed))
	for _, job := range q.failed {
		out = append(out, cloneJob(job))
	}
	return out
}

// Start launches workers and the scheduler until the context is canceled.
func (q *Queue) Start(ctx context.Context, workers int) (*Runner, error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkerCount
	}
	if ctx == nil {
		ctx = context.Background()
	}

	q.mu.Lock()
	if q.running {
		q.mu.Unlock()
		return nil, ErrQueueAlreadyRunning
	}
	q.running = true
	q.mu.Unlock()

	runCtx, cancel := context.WithCancel(ctx)
	runner := &Runner{
		cancel: cancel,
		done:   make(chan struct{}),
	}

	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go q.worker(runCtx, &wg)
	}

	wg.Add(1)
	go q.scheduler(runCtx, &wg)

	go func() {
		wg.Wait()
		q.mu.Lock()
		q.running = false
		q.signalLocked()
		q.mu.Unlock()
		close(runner.done)
	}()

	return runner, nil
}

// Runner controls a running queue.
type Runner struct {
	cancel context.CancelFunc
	done   chan struct{}
}

// Stop cancels the runner and waits for workers to exit.
func (r *Runner) Stop(ctx context.Context) error {
	if r == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	r.cancel()
	select {
	case <-r.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Wait blocks until the runner exits.
func (r *Runner) Wait() {
	if r == nil {
		return
	}
	<-r.done
}

func (q *Queue) enqueue(ctx context.Context, name string, payload any, cfg EnqueueConfig) (Job, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return Job{}, err
	}
	if name == "" {
		return Job{}, ErrJobNameRequired
	}

	body, err := encodePayload(payload)
	if err != nil {
		return Job{}, err
	}

	cfg = q.normalizeEnqueueConfig(cfg)
	if cfg.MaxAttempts <= 0 {
		return Job{}, ErrInvalidMaxAttempts
	}

	now := q.now()
	job := Job{
		ID:          cfg.ID,
		Name:        name,
		Queue:       cfg.Queue,
		Payload:     body,
		CreatedAt:   now,
		RunAt:       cfg.RunAt,
		MaxAttempts: cfg.MaxAttempts,
	}

	q.mu.Lock()
	q.pending = append(q.pending, queuedJob{job: job})
	q.signalLocked()
	q.mu.Unlock()

	q.emit(Event{Type: EventEnqueued, Job: cloneJob(job)})
	return cloneJob(job), nil
}

func (q *Queue) worker(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		item, ok := q.reserve(ctx)
		if !ok {
			return
		}
		q.perform(ctx, item)
	}
}

func (q *Queue) reserve(ctx context.Context) (queuedJob, bool) {
	for {
		if ctx.Err() != nil {
			return queuedJob{}, false
		}

		q.mu.Lock()
		index, wait := q.nextPendingLocked(q.now())
		if index >= 0 && wait <= 0 {
			item := q.pending[index]
			q.pending = append(q.pending[:index], q.pending[index+1:]...)
			q.mu.Unlock()
			return item, true
		}
		notify := q.notify
		q.mu.Unlock()

		if !waitFor(ctx, notify, wait) {
			return queuedJob{}, false
		}
	}
}

func (q *Queue) perform(ctx context.Context, item queuedJob) {
	q.mu.Lock()
	handler := q.handlers[item.job.Name]
	q.mu.Unlock()

	job := item.job
	job.Attempts++
	q.emit(Event{Type: EventStarted, Job: cloneJob(job)})

	var err error
	if handler == nil {
		err = ErrHandlerNotFound
	} else {
		err = handler(ctx, cloneJob(job))
	}

	if err == nil {
		q.emit(Event{Type: EventCompleted, Job: cloneJob(job)})
		return
	}

	job.LastError = err.Error()
	if job.Attempts < job.MaxAttempts {
		delay := q.config.Backoff(job.Attempts, err)
		if delay < 0 {
			delay = 0
		}
		job.RunAt = q.now().Add(delay)
		q.mu.Lock()
		q.pending = append(q.pending, queuedJob{job: job})
		q.signalLocked()
		q.mu.Unlock()
		q.emit(Event{Type: EventRetrying, Job: cloneJob(job), Error: err, NextRunAt: job.RunAt})
		return
	}

	q.mu.Lock()
	q.failed = append(q.failed, cloneJob(job))
	q.mu.Unlock()
	q.emit(Event{Type: EventFailed, Job: cloneJob(job), Error: err})
}

func (q *Queue) scheduler(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		if ctx.Err() != nil {
			return
		}

		now := q.now()
		due, wait := q.dueSchedules(now)
		for _, state := range due {
			job, err := q.enqueue(ctx, state.snapshot.Name, state.payload, EnqueueConfig{
				Queue:       state.snapshot.Queue,
				MaxAttempts: state.snapshot.MaxAttempts,
			})
			if err == nil {
				q.emit(Event{Type: EventScheduled, Job: job, NextRunAt: state.snapshot.NextRunAt})
			}
		}

		q.mu.Lock()
		notify := q.notify
		q.mu.Unlock()
		if !waitFor(ctx, notify, wait) {
			return
		}
	}
}

func (q *Queue) dueSchedules(now time.Time) ([]scheduleState, time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()

	var due []scheduleState
	var next time.Time

	for _, state := range q.schedules {
		if state.snapshot.NextRunAt.IsZero() {
			continue
		}
		if !state.snapshot.NextRunAt.After(now) {
			copyState := *state
			copyState.payload = append([]byte(nil), state.payload...)
			due = append(due, copyState)
			state.snapshot.NextRunAt = state.parsed.Next(now)
		}
		if !state.snapshot.NextRunAt.IsZero() && (next.IsZero() || state.snapshot.NextRunAt.Before(next)) {
			next = state.snapshot.NextRunAt
		}
	}

	if next.IsZero() {
		return due, -1
	}
	wait := next.Sub(now)
	if wait < 0 {
		wait = 0
	}
	return due, wait
}

func (q *Queue) nextPendingLocked(now time.Time) (int, time.Duration) {
	if len(q.pending) == 0 {
		return -1, -1
	}

	index := 0
	next := q.pending[0].job.RunAt
	for i := 1; i < len(q.pending); i++ {
		if q.pending[i].job.RunAt.Before(next) {
			index = i
			next = q.pending[i].job.RunAt
		}
	}

	wait := next.Sub(now)
	if wait < 0 {
		wait = 0
	}
	return index, wait
}

func (q *Queue) normalizeEnqueueConfig(cfg EnqueueConfig) EnqueueConfig {
	if cfg.ID == "" {
		cfg.ID = newID("job")
	}
	if cfg.Queue == "" {
		cfg.Queue = q.config.DefaultQueue
	}
	if cfg.RunAt.IsZero() {
		cfg.RunAt = q.now()
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = q.config.DefaultMaxAttempts
	}
	return cfg
}

func (q *Queue) normalizeScheduleConfig(cfg ScheduleConfig) ScheduleConfig {
	if cfg.ID == "" {
		cfg.ID = newID("schedule")
	}
	if cfg.Queue == "" {
		cfg.Queue = q.config.DefaultQueue
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = q.config.DefaultMaxAttempts
	}
	return cfg
}

func (q *Queue) now() time.Time {
	return q.config.Now()
}

func (q *Queue) emit(event Event) {
	if q.config.EventHandler != nil {
		q.config.EventHandler(event)
	}
}

func (q *Queue) signalLocked() {
	close(q.notify)
	q.notify = make(chan struct{})
}

func normalizeConfig(config Config) Config {
	cfg := config
	if cfg.DefaultQueue == "" {
		cfg.DefaultQueue = DefaultConfig.DefaultQueue
	}
	if cfg.DefaultMaxAttempts == 0 {
		cfg.DefaultMaxAttempts = DefaultConfig.DefaultMaxAttempts
	}
	if cfg.Backoff == nil {
		cfg.Backoff = DefaultConfig.Backoff
	}
	if cfg.Now == nil {
		cfg.Now = DefaultConfig.Now
	}
	return cfg
}

func firstEnqueueConfig(config []EnqueueConfig) EnqueueConfig {
	if len(config) == 0 {
		return EnqueueConfig{}
	}
	return config[0]
}

func firstScheduleConfig(config []ScheduleConfig) ScheduleConfig {
	if len(config) == 0 {
		return ScheduleConfig{}
	}
	return config[0]
}

func waitFor(ctx context.Context, notify <-chan struct{}, wait time.Duration) bool {
	if wait < 0 {
		select {
		case <-ctx.Done():
			return false
		case <-notify:
			return true
		}
	}
	if wait == 0 {
		select {
		case <-ctx.Done():
			return false
		default:
			return true
		}
	}

	timer := time.NewTimer(wait)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-notify:
		return true
	case <-timer.C:
		return true
	}
}
