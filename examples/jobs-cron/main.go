package main

import (
	"context"
	"log"

	"github.com/0mjs/zinc"
	"github.com/0mjs/zinc/jobs"
)

func main() {
	queue := jobs.New()

	queue.Cron("log.zinc", "3s", func(ctx context.Context) error {
		log.Println("I'm running every 3 seconds")
		return nil
	})

	queue.Start(context.Background(), 1)

	app := zinc.New()
	app.Get("/", "ok")

	log.Fatal(app.Listen(":8080"))
}
