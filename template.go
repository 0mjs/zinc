package zinc

import (
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// TemplateEngine handles template rendering with caching and static file serving
type TemplateEngine struct {
	templates    map[string]*template.Template
	directory    string
	staticDir    string
	extension    string
	mutex        sync.RWMutex
	funcMap      template.FuncMap
	cache        bool
	layout       string
	delimiters   [2]string
	fileServer   http.Handler
	contentTypes map[string]string
}

// TemplateOption is a function type for configuring the template engine
type TemplateOption func(*TemplateEngine)

// NewTemplateEngine creates a new template engine with the given directory and options
func NewTemplateEngine(directory string, options ...TemplateOption) *TemplateEngine {
	engine := &TemplateEngine{
		templates:    make(map[string]*template.Template),
		directory:    directory,
		staticDir:    directory, // Default static dir is the same as template dir
		extension:    ".html",
		funcMap:      make(template.FuncMap),
		cache:        true,
		delimiters:   [2]string{"{{", "}}"},
		contentTypes: getDefaultContentTypes(),
	}

	// Apply options
	for _, option := range options {
		option(engine)
	}

	// Initialize file server for static files
	engine.fileServer = http.FileServer(http.Dir(engine.staticDir))

	return engine
}

// WithFuncMap sets custom template functions
func WithFuncMap(funcMap template.FuncMap) TemplateOption {
	return func(e *TemplateEngine) {
		e.funcMap = funcMap
	}
}

// WithExtension sets the template file extension
func WithExtension(ext string) TemplateOption {
	return func(e *TemplateEngine) {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		e.extension = ext
	}
}

// WithLayout sets a default layout template
func WithLayout(layout string) TemplateOption {
	return func(e *TemplateEngine) {
		e.layout = layout
	}
}

// WithStaticDir sets the directory for static files
func WithStaticDir(dir string) TemplateOption {
	return func(e *TemplateEngine) {
		e.staticDir = dir
		e.fileServer = http.FileServer(http.Dir(dir))
	}
}

// WithNoCache disables template caching
func WithNoCache() TemplateOption {
	return func(e *TemplateEngine) {
		e.cache = false
	}
}

// WithDelimiters sets custom template delimiters
func WithDelimiters(left, right string) TemplateOption {
	return func(e *TemplateEngine) {
		e.delimiters = [2]string{left, right}
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
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	e.extension = ext
}

// SetLayout sets the default layout template
func (e *TemplateEngine) SetLayout(layout string) {
	e.layout = layout
}

// SetStaticDir sets the directory for static files
func (e *TemplateEngine) SetStaticDir(dir string) {
	e.staticDir = dir
	e.fileServer = http.FileServer(http.Dir(dir))
}

// DisableCache disables template caching
func (e *TemplateEngine) DisableCache() {
	e.cache = false
}

// AddContentType adds or overrides a content type mapping
func (e *TemplateEngine) AddContentType(ext, contentType string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	e.contentTypes[ext] = contentType
}

// getContentType returns the content type for a file extension
func (e *TemplateEngine) getContentType(path string) string {
	ext := filepath.Ext(path)
	if ct, ok := e.contentTypes[ext]; ok {
		return ct
	}
	// Try to detect from mime package
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}
	// Default
	return "application/octet-stream"
}

// Render renders a template with the given data
func (e *TemplateEngine) Render(w io.Writer, name string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return err
	}

	// If a layout is set, execute the named template within the layout
	if e.layout != "" {
		return tmpl.ExecuteTemplate(w, e.layout, data)
	}

	// Otherwise execute the template directly
	return tmpl.Execute(w, data)
}

// RenderWithLayout renders a template with the given layout and data
func (e *TemplateEngine) RenderWithLayout(w io.Writer, name, layout string, data interface{}) error {
	tmpl, err := e.getTemplate(name)
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, layout, data)
}

// ServeStatic serves a static file
func (e *TemplateEngine) ServeStatic(w http.ResponseWriter, r *http.Request, path string) error {
	fullPath := filepath.Join(e.staticDir, path)

	// Check if file exists
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return err
	}

	// Don't serve directories
	if info.IsDir() {
		return fmt.Errorf("cannot serve directory: %s", path)
	}

	// Set content type
	w.Header().Set("Content-Type", e.getContentType(path))

	// Serve the file
	http.ServeFile(w, r, fullPath)
	return nil
}

// Static configures automatic static file handling for an app
// This registers handlers for common static file types and directories
func (e *TemplateEngine) Static(app *App) {
	// Register handlers for common static file extensions
	app.Get("/css/*", func(c *Context) error {
		path := c.Param("*")
		c.Response.Header().Set("Content-Type", "text/css; charset=utf-8")
		http.ServeFile(c.Response, c.Request, filepath.Join(e.staticDir, "css", path))
		return nil
	})

	app.Get("/js/*", func(c *Context) error {
		path := c.Param("*")
		c.Response.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		http.ServeFile(c.Response, c.Request, filepath.Join(e.staticDir, "js", path))
		return nil
	})

	app.Get("/img/*", func(c *Context) error {
		path := c.Param("*")
		fullPath := filepath.Join(e.staticDir, "img", path)
		c.Response.Header().Set("Content-Type", e.getContentType(path))
		http.ServeFile(c.Response, c.Request, fullPath)
		return nil
	})

	// Handle specific files at the root level
	app.Get("/:file", func(c *Context) error {
		filename := c.Param("file")

		// Skip processing for common route patterns
		if !strings.Contains(filename, ".") || filename == "favicon.ico" {
			return c.Next()
		}

		fullPath := filepath.Join(e.staticDir, filename)

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return c.Next()
		}

		// Set the content type
		c.Response.Header().Set("Content-Type", e.getContentType(filename))

		// Serve the file
		http.ServeFile(c.Response, c.Request, fullPath)
		return nil
	})
}

// StaticFile registers a single static file with the correct MIME type
func (e *TemplateEngine) StaticFile(app *App, urlPath, filePath string) {
	if !strings.HasPrefix(urlPath, "/") {
		urlPath = "/" + urlPath
	}

	// If filePath is relative, make it relative to staticDir
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(e.staticDir, filePath)
	}

	app.Get(urlPath, func(c *Context) error {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return c.Status(http.StatusNotFound).Send("File not found")
		}

		// Set the content type
		c.Response.Header().Set("Content-Type", e.getContentType(filePath))

		// Serve the file
		http.ServeFile(c.Response, c.Request, filePath)
		return nil
	})
}

// getMountHandler returns a HandlerFunc for the specified path prefix
func (e *TemplateEngine) getMountHandler(prefix string) RouteHandler {
	// Strip the prefix for the file server
	fileServer := http.StripPrefix(prefix, e.fileServer)

	return func(c *Context) error {
		fileServer.ServeHTTP(c.Response, c.Request)
		return nil
	}
}

// Mount mounts the static file server at the given prefix
func (e *TemplateEngine) Mount(app *App, prefix string) {
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	// Register handler for the prefix path
	app.Get(prefix+"*", e.getMountHandler(prefix))
}

// getTemplate retrieves a template from cache or loads it from disk
func (e *TemplateEngine) getTemplate(name string) (*template.Template, error) {
	e.mutex.RLock()
	tmpl, exists := e.templates[name]
	e.mutex.RUnlock()

	if exists && e.cache {
		return tmpl, nil
	}

	// Check for layout template
	var layoutFile string
	if e.layout != "" {
		layoutFile = filepath.Join(e.directory, e.layout+e.extension)
	}

	// Load template from disk
	filename := filepath.Join(e.directory, name+e.extension)

	// Create template with custom delimiters if set
	var rootTemplate *template.Template
	if e.layout != "" {
		rootTemplate = template.New(filepath.Base(layoutFile)).Delims(e.delimiters[0], e.delimiters[1]).Funcs(e.funcMap)
	} else {
		rootTemplate = template.New(filepath.Base(filename)).Delims(e.delimiters[0], e.delimiters[1]).Funcs(e.funcMap)
	}

	var err error

	// Parse templates
	if e.layout != "" {
		// Parse layout first, then the content template
		_, err = rootTemplate.ParseFiles(layoutFile, filename)
	} else {
		// Just parse the content template
		_, err = rootTemplate.ParseFiles(filename)
	}

	if err != nil {
		return nil, err
	}

	if e.cache {
		e.mutex.Lock()
		e.templates[name] = rootTemplate
		e.mutex.Unlock()
	}

	return rootTemplate, nil
}

// LoadTemplates preloads all templates from the directory
func (e *TemplateEngine) LoadTemplates() error {
	// Track layout file for separate handling
	var layoutFile string
	if e.layout != "" {
		layoutFile = filepath.Join(e.directory, e.layout+e.extension)
	}

	// Get all template files
	pattern := filepath.Join(e.directory, "*"+e.extension)
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	e.mutex.Lock()
	defer e.mutex.Unlock()

	for _, file := range files {
		// Skip layout file as it's handled specially
		if e.layout != "" && file == layoutFile {
			continue
		}

		name := filepath.Base(file[:len(file)-len(e.extension)])
		var tmpl *template.Template

		if e.layout != "" {
			// Create template with layout
			tmpl = template.New(filepath.Base(layoutFile)).Delims(e.delimiters[0], e.delimiters[1]).Funcs(e.funcMap)
			_, err = tmpl.ParseFiles(layoutFile, file)
		} else {
			// Create template without layout
			tmpl = template.New(filepath.Base(file)).Delims(e.delimiters[0], e.delimiters[1]).Funcs(e.funcMap)
			_, err = tmpl.ParseFiles(file)
		}

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

// GetDirectory returns the template directory path
func (e *TemplateEngine) GetDirectory() string {
	return e.directory
}

// GetStaticDirectory returns the static directory path
func (e *TemplateEngine) GetStaticDirectory() string {
	return e.staticDir
}

// Context methods for templates and static files

// Render renders a template with the given data
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

// RenderWithLayout renders a template with the given layout and data
func (c *Context) RenderWithLayout(name, layout string, data interface{}) error {
	if app, ok := c.Get("app").(*App); ok {
		if app.templateEngine == nil {
			return fmt.Errorf("template engine not initialized")
		}
		c.Response.Header().Set("Content-Type", "text/html; charset=utf-8")
		return app.templateEngine.RenderWithLayout(c.Response, name, layout, data)
	}
	return fmt.Errorf("app context not found")
}

// ServeStatic serves a static file with the appropriate content type
func (c *Context) ServeStatic(path string) error {
	if app, ok := c.Get("app").(*App); ok {
		if app.templateEngine == nil {
			return fmt.Errorf("template engine not initialized")
		}
		return app.templateEngine.ServeStatic(c.Response, c.Request, path)
	}
	return fmt.Errorf("app context not found")
}

// Helper function to get default content types
func getDefaultContentTypes() map[string]string {
	return map[string]string{
		".html":  "text/html; charset=utf-8",
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript; charset=utf-8",
		".json":  "application/json; charset=utf-8",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".svg":   "image/svg+xml",
		".ico":   "image/x-icon",
		".txt":   "text/plain; charset=utf-8",
		".pdf":   "application/pdf",
		".woff":  "font/woff",
		".woff2": "font/woff2",
		".ttf":   "font/ttf",
		".eot":   "application/vnd.ms-fontobject",
		".otf":   "font/otf",
		".xml":   "application/xml",
		".zip":   "application/zip",
		".gz":    "application/gzip",
	}
}
