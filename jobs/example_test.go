package jobs_test

import (
	"context"
	"fmt"

	"github.com/0mjs/zinc/jobs"
)

func ExampleQueue_Cron() {
	queue := jobs.New()

	schedule, err := queue.Cron("cache.refresh", "@every 15m", func(context.Context) error {
		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(schedule.Name, schedule.Spec)
	// Output: cache.refresh @every 15m
}
