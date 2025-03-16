package zinc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Test structures
type TestUser struct {
	ID    int    `json:"id" xml:"id" form:"id" query:"id" validate:"required"`
	Name  string `json:"name" xml:"name" form:"name" query:"name" validate:"required"`
	Email string `json:"email" xml:"email" form:"email" query:"email" validate:"required,email"`
	Age   int    `json:"age" xml:"age" form:"age" query:"age" validate:"min=18"`
}

func TestBindJSON(t *testing.T) {
	app := New()

	// Create test data
	user := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	jsonData, _ := json.Marshal(user)

	// Create request with JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding
	var result TestUser
	err := c.BindJSON(&result)

	// Check results
	if err != nil {
		t.Fatalf("BindJSON failed: %v", err)
	}

	if result.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, result.ID)
	}

	if result.Name != user.Name {
		t.Errorf("Expected Name %s, got %s", user.Name, result.Name)
	}

	if result.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, result.Email)
	}

	if result.Age != user.Age {
		t.Errorf("Expected Age %d, got %d", user.Age, result.Age)
	}
}

func TestBindJSONValidationError(t *testing.T) {
	app := New()

	// Create test data with invalid email
	user := map[string]interface{}{
		"id":    1,
		"name":  "John Doe",
		"email": "not-an-email", // Invalid email format
		"age":   25,
	}

	jsonData, _ := json.Marshal(user)

	// Create request with JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding
	var result TestUser
	err := c.BindJSON(&result)

	// Should fail validation
	if err == nil {
		t.Error("Expected validation error for invalid email, got nil")
	}
}

func TestBindXML(t *testing.T) {
	app := New()

	// Create test data
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
	<TestUser>
		<id>1</id>
		<name>John Doe</name>
		<email>john@example.com</email>
		<age>25</age>
	</TestUser>`

	// Create request with XML data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(xmlData))
	req.Header.Set("Content-Type", "application/xml")
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding
	var result TestUser
	err := c.BindXML(&result)

	// Check results
	if err != nil {
		t.Fatalf("BindXML failed: %v", err)
	}

	expected := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	if result.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, result.ID)
	}

	if result.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, result.Name)
	}

	if result.Email != expected.Email {
		t.Errorf("Expected Email %s, got %s", expected.Email, result.Email)
	}

	if result.Age != expected.Age {
		t.Errorf("Expected Age %d, got %d", expected.Age, result.Age)
	}
}

func TestBindForm(t *testing.T) {
	app := New()

	// Create form data
	form := url.Values{}
	form.Add("id", "1")
	form.Add("name", "John Doe")
	form.Add("email", "john@example.com")
	form.Add("age", "25")

	// Create request with form data
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(form.Encode())))
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding
	var result TestUser
	err := c.BindForm(&result)

	// Check results
	if err != nil {
		t.Fatalf("BindForm failed: %v", err)
	}

	expected := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	if result.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, result.ID)
	}

	if result.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, result.Name)
	}

	if result.Email != expected.Email {
		t.Errorf("Expected Email %s, got %s", expected.Email, result.Email)
	}

	if result.Age != expected.Age {
		t.Errorf("Expected Age %d, got %d", expected.Age, result.Age)
	}
}

func TestBindQuery(t *testing.T) {
	app := New()

	// Create request with query params
	req := httptest.NewRequest(http.MethodGet, "/test?id=1&name=John+Doe&email=john@example.com&age=25", nil)
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding
	var result TestUser
	err := c.BindQuery(&result)

	// Check results
	if err != nil {
		t.Fatalf("BindQuery failed: %v", err)
	}

	expected := TestUser{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   25,
	}

	if result.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, result.ID)
	}

	if result.Name != expected.Name {
		t.Errorf("Expected Name %s, got %s", expected.Name, result.Name)
	}

	if result.Email != expected.Email {
		t.Errorf("Expected Email %s, got %s", expected.Email, result.Email)
	}

	if result.Age != expected.Age {
		t.Errorf("Expected Age %d, got %d", expected.Age, result.Age)
	}
}

func TestAutoBind(t *testing.T) {
	app := New()

	// Test cases with different content types
	testCases := []struct {
		name        string
		contentType string
		body        io.Reader
		queryString string
	}{
		{
			name:        "JSON Binding",
			contentType: "application/json",
			body:        bytes.NewBufferString(`{"id":1,"name":"John Doe","email":"john@example.com","age":25}`),
		},
		{
			name:        "XML Binding",
			contentType: "application/xml",
			body: bytes.NewBufferString(`<?xml version="1.0" encoding="UTF-8"?>
			<TestUser>
				<id>1</id>
				<name>John Doe</name>
				<email>john@example.com</email>
				<age>25</age>
			</TestUser>`),
		},
		{
			name:        "Form Binding",
			contentType: "application/x-www-form-urlencoded",
			body:        bytes.NewBufferString("id=1&name=John+Doe&email=john%40example.com&age=25"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/test", tc.body)
			req.Header.Set("Content-Type", tc.contentType)
			w := httptest.NewRecorder()

			c := NewContext(w, req)
			c.Set("app", app)

			var result TestUser
			err := c.Bind(&result)

			if err != nil {
				t.Fatalf("Bind failed for %s: %v", tc.contentType, err)
			}

			expected := TestUser{
				ID:    1,
				Name:  "John Doe",
				Email: "john@example.com",
				Age:   25,
			}

			if result.ID != expected.ID {
				t.Errorf("Expected ID %d, got %d", expected.ID, result.ID)
			}

			if result.Name != expected.Name {
				t.Errorf("Expected Name %s, got %s", expected.Name, result.Name)
			}

			if result.Email != expected.Email {
				t.Errorf("Expected Email %s, got %s", expected.Email, result.Email)
			}

			if result.Age != expected.Age {
				t.Errorf("Expected Age %d, got %d", expected.Age, result.Age)
			}
		})
	}
}

func TestBindOptions(t *testing.T) {
	app := New()

	// Create test data with invalid email but disable validation
	user := map[string]interface{}{
		"id":    1,
		"name":  "John Doe",
		"email": "not-an-email", // Invalid email format
		"age":   25,
	}

	jsonData, _ := json.Marshal(user)

	// Create request with JSON data
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)
	c.Set("app", app)

	// Test binding with disabled validation
	var result TestUser
	bindOptions := &BindOptions{
		DisableValidation: true,
	}
	err := c.BindJSON(&result, bindOptions)

	// Should succeed despite invalid email due to disabled validation
	if err != nil {
		t.Errorf("BindJSON with DisableValidation failed: %v", err)
	}
}

func TestFormFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test_upload_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpFilePath := tmpFile.Name()
	defer os.Remove(tmpFilePath)

	content := []byte("Hello, World!")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Prepare multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add file field
	fileWriter, err := writer.CreateFormFile("file", filepath.Base(tmpFilePath))
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Open the file for reading
	file, err := os.Open(tmpFilePath)
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	// Copy the file content to the form
	if _, err = io.Copy(fileWriter, file); err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	// Close the multipart writer
	writer.Close()

	// Create a new request with the form
	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)

	// Test FormFile
	fileHeader, err := c.FormFile("file")
	if err != nil {
		t.Fatalf("FormFile failed: %v", err)
	}

	if fileHeader.Filename != filepath.Base(tmpFilePath) {
		t.Errorf("Expected filename %s, got %s", filepath.Base(tmpFilePath), fileHeader.Filename)
	}

	// Test SaveFile
	uploadDir, err := os.MkdirTemp("", "upload_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(uploadDir)

	savedPath := filepath.Join(uploadDir, fileHeader.Filename)
	if err := c.SaveFile(fileHeader, savedPath); err != nil {
		t.Fatalf("SaveFile failed: %v", err)
	}

	// Verify saved file
	savedContent, err := os.ReadFile(savedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if !bytes.Equal(savedContent, content) {
		t.Errorf("Saved file content doesn't match original content")
	}
}

func TestFormFiles(t *testing.T) {
	// Create temporary files
	fileContents := []string{"First file content", "Second file content"}
	filenames := make([]string, len(fileContents))
	tempFiles := make([]*os.File, len(fileContents))

	for i, content := range fileContents {
		tmpFile, err := os.CreateTemp("", fmt.Sprintf("test_upload_%d_*.txt", i))
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tempFiles[i] = tmpFile
		filenames[i] = filepath.Base(tmpFile.Name())
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(content); err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}
		if err := tmpFile.Close(); err != nil {
			t.Fatalf("Failed to close temp file: %v", err)
		}
	}

	// Prepare multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add multiple files with the same field name
	for i, tmpFile := range tempFiles {
		file, err := os.Open(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to open temp file: %v", err)
		}

		fw, err := writer.CreateFormFile("files", filenames[i])
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		if _, err = io.Copy(fw, file); err != nil {
			t.Fatalf("Failed to copy file content: %v", err)
		}

		file.Close()
	}

	// Close the multipart writer
	writer.Close()

	// Create a new request with the form
	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	// Create context
	c := NewContext(w, req)

	// Test FormFiles
	fileHeaders, err := c.FormFiles("files")
	if err != nil {
		t.Fatalf("FormFiles failed: %v", err)
	}

	if len(fileHeaders) != len(fileContents) {
		t.Fatalf("Expected %d files, got %d", len(fileContents), len(fileHeaders))
	}

	// Verify filenames
	for i, fh := range fileHeaders {
		found := false
		for _, expected := range filenames {
			if fh.Filename == expected {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("File %d: filename %s not found in expected filenames", i, fh.Filename)
		}
	}
}
