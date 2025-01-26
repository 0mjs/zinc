package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	p "github.com/0mjs/zinc/docs/pages"
)

func dataLoader(filename string) (template.HTML, error) {
	path := filepath.Join("templates", filename)
	content, err := os.ReadFile(path)
	if err != nil || content == nil {
		return "", err
	}
	return template.HTML(content), nil
}

func main() {
	templatePaths := template.Must(template.ParseGlob("templates/*.html"))

	http.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("static"))),
	)

	pages := map[string]p.Page{}
	introductionContent, _ := dataLoader("introduction.html")
	gettingStartedContent, _ := dataLoader("getting-started.html")
	guideContent, _ := dataLoader("guide.html")
	conceptsContent, _ := dataLoader("concepts.html")

	pages = map[string]p.Page{
		"introduction": {
			Title:    "Introduction",
			Subtitle: "Welcome to Zinc - A minimalistic Go framework for building robust API's with ease.",
			Content:  introductionContent,
		},
		"getting-started": {
			Title:    "Getting Started",
			Subtitle: "A lightweight Go web framework focused on simplicity",
			Content:  gettingStartedContent,
		},
		"guide": {
			Title:    "Guide",
			Subtitle: "Comprehensive Guide to Zinc Features",
			Content:  guideContent,
		},
		"concepts": {
			Title:    "Concepts",
			Subtitle: "Core Principles and Philosophy",
			Content:  conceptsContent,
		},
	}

	// Create route handlers for each page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		templatePaths.ExecuteTemplate(w, "layout.html", pages["introduction"])
	})

	http.HandleFunc("/getting-started", func(w http.ResponseWriter, r *http.Request) {
		templatePaths.ExecuteTemplate(w, "layout.html", pages["getting-started"])
	})

	http.HandleFunc("/guide", func(w http.ResponseWriter, r *http.Request) {
		templatePaths.ExecuteTemplate(w, "layout.html", pages["guide"])
	})

	http.HandleFunc("/concepts", func(w http.ResponseWriter, r *http.Request) {
		templatePaths.ExecuteTemplate(w, "layout.html", pages["concepts"])
	})

	http.ListenAndServe(":3000", nil)
}
