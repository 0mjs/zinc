package zinc

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"sync"
)

// TemplateEngine handles template rendering with caching
type TemplateEngine struct {
	templates map[string]*template.Template
	directory string
	extension string
	mutex     sync.RWMutex
	funcMap   template.FuncMap
	cache     bool
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine(directory string) *TemplateEngine {
	return &TemplateEngine{
		templates: make(map[string]*template.Template),
		directory: directory,
		extension: ".html",
		funcMap:   make(template.FuncMap),
		cache:     true,
	}
}

// SetFuncMap sets custom template functions
func (e *TemplateEngine) SetFuncMap(funcMap template.FuncMap) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.funcMap = funcMap
}

// AddFunc adds a single template function
func (e *TemplateEngine) AddFunc(name string, fn interface{}) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	if e.funcMap == nil {
		e.funcMap = make(template.FuncMap)
	}
	e.funcMap[name] = fn
}

// SetExtension sets the template file extension
func (e *TemplateEngine) SetExtension(ext string) {
	e.extension = ext
}

// DisableCache disables template caching
func (e *TemplateEngine) DisableCache() {
	e.cache = false
}

// Render renders a template with the given data
func (e *TemplateEngine) Render(w io.Writer, name string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, data)
}

// getTemplate retrieves a template from cache or loads it from disk
func (e *TemplateEngine) getTemplate(name string) (*template.Template, error) {
	e.mutex.RLock()
	tmpl, exists := e.templates[name]
	e.mutex.RUnlock()

	if exists && e.cache {
		return tmpl, nil
	}

	// Load template from disk
	filename := filepath.Join(e.directory, name+e.extension)
	fmt.Printf("Loading template from: %s\n", filename)
	tmpl, err := template.New(filepath.Base(filename)).Funcs(e.funcMap).ParseFiles(filename)
	if err != nil {
		fmt.Printf("Error loading template: %v\n", err)
		return nil, err
	}

	if e.cache {
		e.mutex.Lock()
		e.templates[name] = tmpl
		e.mutex.Unlock()
	}

	return tmpl, nil
}

// LoadTemplates preloads all templates from the directory
func (e *TemplateEngine) LoadTemplates() error {
	pattern := filepath.Join(e.directory, "*"+e.extension)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, file := range files {
		name := filepath.Base(file[:len(file)-len(e.extension)])
		tmpl, err := template.New(filepath.Base(file)).Funcs(e.funcMap).ParseFiles(file)
		if err != nil {
			return err
		}
		e.templates[name] = tmpl
	}

	return nil
}

// ClearCache clears the template cache
func (e *TemplateEngine) ClearCache() {
	e.mutex.Lock()
	e.templates = make(map[string]*template.Template)
	e.mutex.Unlock()
}

// Add template rendering methods to Context
func (c *Context) Render(name string, data interface{}) error {
	if app, ok := c.Get("app").(*App); ok {
		if app.templateEngine == nil {
			return fmt.Errorf("template engine not initialized")
		}
		c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return app.templateEngine.Render(c.Response, name, data)
	}
	return fmt.Errorf("app context not found")
}

// GetDirectory returns the template directory path
func (e *TemplateEngine) GetDirectory() string {
	return e.directory
}
