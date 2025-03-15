package zinc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// CronJob represents a single scheduled job
type CronJob struct {
	ID        string
	Schedule  string
	Handler   func() error
	cronID    cron.EntryID
	isRunning bool
}

// CronScheduler manages all cron jobs
type CronScheduler struct {
	jobs      map[string]*CronJob
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	cron      *cron.Cron
	isRunning bool
}

// newCronScheduler creates a new CronScheduler
func newCronScheduler() *CronScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	// Create cron instance with second precision
	cronParser := cron.New(
		cron.WithSeconds(),
		cron.WithLogger(cron.DefaultLogger),
	)

	return &CronScheduler{
		jobs:      make(map[string]*CronJob),
		ctx:       ctx,
		cancel:    cancel,
		cron:      cronParser,
		isRunning: false,
	}
}

// AddJob adds a new job to the scheduler
func (cs *CronScheduler) AddJob(id, schedule string, handler func() error) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Create job instance
	job := &CronJob{
		ID:        id,
		Schedule:  schedule,
		Handler:   handler,
		isRunning: false,
	}

	// Handle special macros before parsing as cron expression
	parsedSchedule := schedule
	switch schedule {
	case "@hourly":
		parsedSchedule = "0 0 * * * *" // At minute 0 of every hour
	case "@daily":
		parsedSchedule = "0 0 0 * * *" // At midnight every day
	case "@weekly":
		parsedSchedule = "0 0 0 * * 0" // At midnight on Sunday
	case "@monthly":
		parsedSchedule = "0 0 0 1 * *" // At midnight on the 1st of every month
	case "@yearly", "@annually":
		parsedSchedule = "0 0 0 1 1 *" // At midnight on January 1
	case "@every":
		// Try to handle simple duration formats
		if duration, err := time.ParseDuration(schedule[6:]); err == nil {
			seconds := int(duration.Seconds())
			parsedSchedule = fmt.Sprintf("@every %ds", seconds)
		}
	}

	// Add job to cron
	jobFunc := func() {
		// Only execute if scheduler is running
		if cs.isRunning {
			cs.executeJob(job)
		}
	}

	// Add the job to the cron scheduler
	entryID, err := cs.cron.AddFunc(parsedSchedule, jobFunc)
	if err != nil {
		return fmt.Errorf("invalid cron expression '%s': %v", schedule, err)
	}

	// Store job ID for later reference
	job.cronID = entryID

	// Add to jobs map
	cs.jobs[id] = job

	return nil
}

// RemoveJob removes a job from the scheduler
func (cs *CronScheduler) RemoveJob(id string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if job, exists := cs.jobs[id]; exists {
		cs.cron.Remove(job.cronID)
		delete(cs.jobs, id)
	}
}

// Start starts the scheduler
func (cs *CronScheduler) Start() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.isRunning {
		return
	}

	cs.isRunning = true
	cs.cron.Start()
}

// Stop stops the scheduler
func (cs *CronScheduler) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.isRunning {
		return
	}

	cs.isRunning = false
	cs.cron.Stop()
}

// executeJob runs a job and handles any errors
func (cs *CronScheduler) executeJob(job *CronJob) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Cron job %s panicked: %v\n", job.ID, r)
		}
	}()

	job.isRunning = true
	err := job.Handler()
	job.isRunning = false

	if err != nil {
		fmt.Printf("Cron job %s failed: %v\n", job.ID, err)
	}
}

// Custom logger adapter to integrate with zinc's logging
type loggerAdapter struct{}

func (l loggerAdapter) Info(msg string, keysAndValues ...interface{}) {
	fmt.Printf("CRON INFO: %s %v\n", msg, keysAndValues)
}

func (l loggerAdapter) Error(err error, msg string, keysAndValues ...interface{}) {
	fmt.Printf("CRON ERROR: %s - %v %v\n", msg, err, keysAndValues)
}
