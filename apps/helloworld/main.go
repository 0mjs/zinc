package main

import "github.com/0mjs/zinc"

func main() {
	app := zinc.New()
	app.Get("/", "Hello, World!")
	app.Serve()
}
