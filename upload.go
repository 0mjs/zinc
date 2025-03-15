package zinc

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// FileUpload represents file upload configuration
type FileUpload struct {
	MaxSize    int64
	AllowTypes []string
	Directory  string
}

// NewFileUpload creates a new file upload handler
func NewFileUpload(directory string) *FileUpload {
	return &FileUpload{
		MaxSize:    10 << 20, // 10MB default
		AllowTypes: []string{},
		Directory:  directory,
	}
}

// SetMaxSize sets the maximum file size in bytes
func (f *FileUpload) SetMaxSize(size int64) {
	f.MaxSize = size
}

// SetAllowedTypes sets allowed file types
func (f *FileUpload) SetAllowedTypes(types []string) {
	f.AllowTypes = types
}

// SaveFile saves an uploaded file
func (f *FileUpload) SaveFile(file *multipart.FileHeader) (string, error) {
	// Check file size
	if file.Size > f.MaxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", f.MaxSize)
	}

	// Check file type if restrictions are set
	if len(f.AllowTypes) > 0 {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !slices.Contains(f.AllowTypes, ext) {
			return "", fmt.Errorf("file type %s not allowed", ext)
		}
	}

	// Create upload directory if it doesn't exist
	if err := os.MkdirAll(f.Directory, 0755); err != nil {
		return "", err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
	filepath := filepath.Join(f.Directory, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return filename, nil
}

// Add file upload methods to Context
func (c *Context) SaveFile(file *multipart.FileHeader) (string, error) {
	if app, ok := c.Get("app").(*App); ok {
		if app.fileUpload == nil {
			return "", fmt.Errorf("file upload handler not initialized")
		}
		return app.fileUpload.SaveFile(file)
	}
	return "", fmt.Errorf("app context not found")
}

// File gets an uploaded file from the request
func (c *Context) File(name string) (*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	file, header, err := c.Request.FormFile(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return header, nil
}

// Files gets all uploaded files for a form field
func (c *Context) Files(name string) ([]*multipart.FileHeader, error) {
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	return c.Request.MultipartForm.File[name], nil
}
