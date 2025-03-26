package zinc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestContextParam(t *testing.T) {
	c := &Context{}

	// Setup some path parameters
	c.PathParams[0] = param{key: "id", value: "123"}
	c.PathParams[1] = param{key: "name", value: "test"}
	c.PathParams[2] = param{key: "a", value: "1"}

	// Test getting parameters
	tests := []struct {
		paramName string
		want      string
	}{
		{"id", "123"},
		{"name", "test"},
		{"a", "1"},
		{"notfound", ""}, // Non-existent parameter should return empty string
		{"", ""},         // Empty parameter name should return empty string
	}

	for _, tt := range tests {
		t.Run(tt.paramName, func(t *testing.T) {
			got := c.Param(tt.paramName)
			if got != tt.want {
				t.Errorf("Param(%q) = %q, want %q", tt.paramName, got, tt.want)
			}
		})
	}

	// Test special optimization for single character parameters
	c.PathParams[3] = param{key: "b", value: "2"}
	if got := c.Param("b"); got != "2" {
		t.Errorf("Param(\"b\") = %q, want %q", got, "2")
	}
}

func TestContextQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?name=john&age=25&empty=", nil)
	c := &Context{Request: r}

	// Test query parameters before being accessed (lazy loading)
	if c.QueryParams != nil {
		t.Error("QueryParams should be nil before first access")
	}

	// First access should parse query parameters
	if got := c.Query("name"); got != "john" {
		t.Errorf("Query(\"name\") = %q, want %q", got, "john")
	}

	// Query parameters should be cached after first access
	if c.QueryParams == nil {
		t.Error("QueryParams should not be nil after first access")
	}

	// Test other parameters
	if got := c.Query("age"); got != "25" {
		t.Errorf("Query(\"age\") = %q, want %q", got, "25")
	}

	if got := c.Query("empty"); got != "" {
		t.Errorf("Query(\"empty\") = %q, want %q", got, "")
	}

	if got := c.Query("notfound"); got != "" {
		t.Errorf("Query(\"notfound\") = %q, want %q", got, "")
	}
}

func TestContextHasQuery(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/?name=john&empty=", nil)
	c := &Context{Request: r}

	// Test if query parameters exist
	if !c.HasQuery("name") {
		t.Error("HasQuery(\"name\") = false, want true")
	}

	if !c.HasQuery("empty") {
		t.Error("HasQuery(\"empty\") = false, want true")
	}

	if c.HasQuery("notfound") {
		t.Error("HasQuery(\"notfound\") = true, want false")
	}
}

func TestContextSetParam(t *testing.T) {
	c := &Context{}

	// Set parameters
	c.setParam("id", "123")
	c.setParam("name", "test")
	c.setParam("a", "1")
	c.setParam("b", "2")

	// Test if parameters were set correctly
	if got := c.Param("id"); got != "123" {
		t.Errorf("After setParam, Param(\"id\") = %q, want %q", got, "123")
	}

	if got := c.Param("name"); got != "test" {
		t.Errorf("After setParam, Param(\"name\") = %q, want %q", got, "test")
	}

	// Test that the parameters were set in order
	expectedParams := []param{
		{key: "id", value: "123"},
		{key: "name", value: "test"},
		{key: "a", value: "1"},
		{key: "b", value: "2"},
	}

	for i, expected := range expectedParams {
		if c.PathParams[i].key != expected.key || c.PathParams[i].value != expected.value {
			t.Errorf("PathParams[%d] = {%q, %q}, want {%q, %q}",
				i, c.PathParams[i].key, c.PathParams[i].value, expected.key, expected.value)
		}
	}

	// Test setting more parameters than the fixed size array can hold
	c.setParam("extra", "value")
	// This should be ignored in the current implementation,
	// or the implementation should handle it gracefully
	// But we don't test for a specific behavior as that may change
}

func TestContextTestStore(t *testing.T) {
	c := &Context{Store: make(map[string]interface{})}

	// Test setting and getting values
	c.Set("string", "value")
	c.Set("int", 123)
	c.Set("bool", true)

	if got := c.Get("string"); got != "value" {
		t.Errorf("Get(\"string\") = %v, want %v", got, "value")
	}

	if got := c.Get("int"); got != 123 {
		t.Errorf("Get(\"int\") = %v, want %v", got, 123)
	}

	if got := c.Get("bool"); got != true {
		t.Errorf("Get(\"bool\") = %v, want %v", got, true)
	}

	// Non-existent key should return nil
	if got := c.Get("notfound"); got != nil {
		t.Errorf("Get(\"notfound\") = %v, want nil", got)
	}
}

func TestContextStatus(t *testing.T) {
	c := &Context{status: http.StatusOK}

	// Test default status
	if c.status != http.StatusOK {
		t.Errorf("Default status = %d, want %d", c.status, http.StatusOK)
	}

	// Test setting status
	c.Status(http.StatusCreated)
	if c.status != http.StatusCreated {
		t.Errorf("After Status(201), status = %d, want %d", c.status, http.StatusCreated)
	}

	// Test method chaining
	if c.Status(http.StatusAccepted) != c {
		t.Error("Status method should return the context for chaining")
	}
}

func TestContextSend(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Test sending string
	err := c.Send("Hello World!")
	if err != nil {
		t.Errorf("Send(\"Hello World!\") returned error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("Response status = %d, want %d", w.Code, http.StatusOK)
	}

	if got := w.Body.String(); got != "Hello World!" {
		t.Errorf("Response body = %q, want %q", got, "Hello World!")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/plain; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/plain; charset=utf-8")
	}
}

func TestContextSendBytes(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Test sending bytes
	data := []byte("Binary Data")
	err := c.Send(data)
	if err != nil {
		t.Errorf("Send([]byte) returned error: %v", err)
	}

	if got := w.Body.String(); got != "Binary Data" {
		t.Errorf("Response body = %q, want %q", got, "Binary Data")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/octet-stream")
	}
}

func TestContextSendNull(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Test sending nil
	err := c.Send(nil)
	if err != nil {
		t.Errorf("Send(nil) returned error: %v", err)
	}

	if got := w.Body.String(); got != "null" {
		t.Errorf("Response body = %q, want %q", got, "null")
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
	}
}

func TestContextJSON(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Test sending struct as JSON
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	user := User{Name: "John Doe", Email: "john@example.com"}
	err := c.JSON(user)
	if err != nil {
		t.Errorf("JSON() returned error: %v", err)
	}

	var result User
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if result.Name != "John Doe" || result.Email != "john@example.com" {
		t.Errorf("JSON response = %+v, want {Name:\"John Doe\", Email:\"john@example.com\"}", result)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json; charset=utf-8")
	}
}

func TestContextHTML(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Test sending HTML
	html := "<h1>Hello World!</h1>"
	err := c.HTML(html)
	if err != nil {
		t.Errorf("HTML() returned error: %v", err)
	}

	if got := w.Body.String(); got != html {
		t.Errorf("Response body = %q, want %q", got, html)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", contentType, "text/html; charset=utf-8")
	}
}

func TestContextMultipleResponses(t *testing.T) {
	w := httptest.NewRecorder()
	c := &Context{Response: w, status: http.StatusOK}

	// Send first response
	err := c.Send("First response")
	if err != nil {
		t.Errorf("First Send() returned error: %v", err)
	}

	// Send second response (should return error)
	err = c.Send("Second response")
	if err != ErrResponseAlreadySent {
		t.Errorf("Second Send() returned error = %v, want %v", err, ErrResponseAlreadySent)
	}

	// Check that only the first response was sent
	if got := w.Body.String(); got != "First response" {
		t.Errorf("Response body = %q, want %q", got, "First response")
	}
}

func TestContextBody(t *testing.T) {
	// Create a request with JSON body
	jsonData := `{"name":"John Doe","email":"john@example.com"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonData))
	r.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	c := &Context{Response: w, Request: r, Store: make(map[string]interface{})}

	// Set app with config for testing
	app := &App{config: &Config{BodyLimit: 1024 * 1024}} // 1MB limit
	c.Set("app", app)

	// Test getting request body as string
	body, err := c.Body()
	if err != nil {
		t.Errorf("Body() returned error: %v", err)
	}

	if body != jsonData {
		t.Errorf("Body() = %q, want %q", body, jsonData)
	}

	// Test parsing with BodyParser
	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	r2 := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(jsonData))
	r2.Header.Set("Content-Type", "application/json")
	c2 := &Context{Response: w, Request: r2, Store: make(map[string]interface{})}
	c2.Set("app", app)

	var user User
	err = c2.BodyParser(&user)
	if err != nil {
		t.Errorf("BodyParser() returned error: %v", err)
	}

	if user.Name != "John Doe" || user.Email != "john@example.com" {
		t.Errorf("Parsed body = %+v, want {Name:\"John Doe\", Email:\"john@example.com\"}", user)
	}
}

func TestContextNext(t *testing.T) {
	c := &Context{}

	// Create middleware chain
	calls := []string{}

	middleware1 := func(c *Context) error {
		calls = append(calls, "middleware1 before")
		err := c.Next()
		calls = append(calls, "middleware1 after")
		return err
	}

	middleware2 := func(c *Context) error {
		calls = append(calls, "middleware2 before")
		err := c.Next()
		calls = append(calls, "middleware2 after")
		return err
	}

	handler := func(c *Context) error {
		calls = append(calls, "handler")
		return nil
	}

	// Set up the handlers
	c.handlers = []Middleware{middleware1, middleware2, handler}
	c.index = -1

	// Call Next to start middleware chain
	err := c.Next()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify middleware order
	expected := []string{
		"middleware1 before",
		"middleware2 before",
		"handler",
		"middleware2 after",
		"middleware1 after",
	}

	if len(calls) != len(expected) {
		t.Errorf("Got %d calls, want %d calls", len(calls), len(expected))
	}

	for i := 0; i < len(calls) && i < len(expected); i++ {
		if calls[i] != expected[i] {
			t.Errorf("calls[%d] = %q, want %q", i, calls[i], expected[i])
		}
	}
}

func TestContextReset(t *testing.T) {
	c := &Context{
		Store:      make(map[string]interface{}),
		PathParams: params{},
	}

	// Set up some initial state
	c.Store["key"] = "value"
	c.PathParams[0] = param{key: "id", value: "123"}
	c.written = true
	c.status = http.StatusCreated

	// Create a new request and response
	oldReq := httptest.NewRequest(http.MethodGet, "/old", nil)
	oldRes := httptest.NewRecorder()
	c.Request = oldReq
	c.Response = oldRes

	// Reset the context with new request and response
	newReq := httptest.NewRequest(http.MethodPost, "/new", nil)
	newRes := httptest.NewRecorder()
	c.reset(newRes, newReq)

	// Verify request and response are updated
	if c.Request != newReq {
		t.Error("Request was not reset properly")
	}

	if c.Response != newRes {
		t.Error("Response was not reset properly")
	}

	// Verify method is updated
	if c.Method != http.MethodPost {
		t.Errorf("Method = %q, want %q", c.Method, http.MethodPost)
	}

	// Verify written flag is reset
	if c.written {
		t.Error("written flag was not reset")
	}

	// Verify status is reset
	if c.status != http.StatusOK {
		t.Errorf("status = %d, want %d", c.status, http.StatusOK)
	}

	// Verify handlers are reset
	if c.handlers != nil {
		t.Error("handlers were not reset")
	}

	// Verify index is reset
	if c.index != -1 {
		t.Errorf("index = %d, want %d", c.index, -1)
	}

	// Verify path params are cleared
	for i := range c.PathParams {
		if c.PathParams[i] != emptyParam {
			t.Errorf("PathParams[%d] = %v, want %v", i, c.PathParams[i], emptyParam)
		}
	}

	// Verify store is cleared
	if len(c.Store) != 0 {
		t.Errorf("Store has %d entries, want 0", len(c.Store))
	}
}

func TestContextRelease(t *testing.T) {
	// Create a context with values
	c := &Context{
		Response:   httptest.NewRecorder(),
		Request:    httptest.NewRequest(http.MethodGet, "/", nil),
		handlers:   []Middleware{func(c *Context) error { return nil }},
		PathParams: params{},
		Store:      make(map[string]interface{}),
	}

	// Release the context
	c.release()

	// Verify references are cleared
	if c.Response != nil {
		t.Error("Response was not cleared")
	}

	if c.Request != nil {
		t.Error("Request was not cleared")
	}

	if c.handlers != nil {
		t.Error("handlers were not cleared")
	}
}

func BenchmarkContextParam(b *testing.B) {
	c := &Context{}
	c.PathParams[0] = param{key: "id", value: "123"}
	c.PathParams[1] = param{key: "name", value: "test"}
	c.PathParams[2] = param{key: "long-param-name", value: "test"}
	c.PathParams[3] = param{key: "x", value: "1"}

	b.ResetTimer()

	tests := []struct {
		name  string
		param string
	}{
		{"short param", "id"},
		{"single char param", "x"},
		{"long param", "long-param-name"},
		{"missing param", "notfound"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				c.Param(tt.param)
			}
		})
	}
}

func BenchmarkContextSend(b *testing.B) {
	tests := []struct {
		name string
		data interface{}
	}{
		{"string", "Hello World!"},
		{"bytes", []byte("Hello World!")},
		{"nil", nil},
		{"struct", struct{ Name string }{"John"}},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				w := httptest.NewRecorder()
				c := &Context{Response: w, status: http.StatusOK}
				c.Send(tt.data)
			}
		})
	}
}
