package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
	// Define template functions
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
	}

	// Parse templates with function map
	tmpl := template.New("").Funcs(funcMap)
	templatePaths := template.Must(tmpl.ParseGlob("templates/*.html"))

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
			Subtitle: "A focused net/http framework for routing, middleware, binding, and explicit server lifecycle.",
			Content:  introductionContent,
		},
		"getting-started": {
			Title:    "Getting Started",
			Subtitle: "Install Zinc, register routes, and run a small API with the current core APIs.",
			Content:  gettingStartedContent,
		},
		"guide": {
			Title:    "Guide",
			Subtitle: "Lifecycle, routing, middleware, binding, responses, static files, and configuration.",
			Content:  guideContent,
		},
		"concepts": {
			Title:    "Concepts",
			Subtitle: "How Zinc approaches handlers, routing, request state, proxy trust, and HTTP semantics.",
			Content:  conceptsContent,
		},
	}

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
