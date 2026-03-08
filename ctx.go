package zinc

import (
	"bytes"
	stdctx "context"
	"errors"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Context struct {
	writer      http.ResponseWriter
	request     *http.Request
	PathParams  params
	queryParams url.Values
	written     bool
	handlers    []HandlerFunc
	index       int
	store       map[any]any
	status      int
	app         *App
	routeInfo   routeMeta
	lastErr     error
	body        []byte
	bodyRead    bool
	bodyErr     error
	paramPath   string
	paramCount  int
}

type param struct {
	key   string
	value string
	start int32
	end   int32
}

type params [8]param

var emptyParam param

const directParamStart int32 = -1

var contextPool = sync.Pool{
	New: func() any {
		return &Context{
			store:  make(map[any]any, 8),
			status: http.StatusOK,
			index:  -1,
		}
	},
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.reset(w, r)
	return c
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	c.writer = w
	c.request = r
	c.queryParams = nil
	c.written = false
	c.handlers = nil
	c.index = -1
	c.status = http.StatusOK
	c.app = nil
	c.routeInfo = routeMeta{}
	c.lastErr = nil
	c.body = nil
	c.bodyRead = false
	c.bodyErr = nil
	c.paramPath = ""
	for i := 0; i < c.paramCount; i++ {
		c.PathParams[i] = emptyParam
	}
	c.paramCount = 0
	if len(c.store) > 0 {
		for key := range c.store {
			delete(c.store, key)
		}
	}
}

func (c *Context) release() {
	if c == nil {
		return
	}
	c.writer = nil
	c.request = nil
	c.handlers = nil
	c.queryParams = nil
	c.body = nil
	c.bodyRead = false
	c.bodyErr = nil
	c.paramPath = ""
	contextPool.Put(c)
}

func (c *Context) Writer() http.ResponseWriter {
	return c.writer
}

func (c *Context) SetWriter(w http.ResponseWriter) {
	c.writer = w
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) SetRequest(r *http.Request) {
	c.request = r
	c.queryParams = nil
	c.body = nil
	c.bodyRead = false
	c.bodyErr = nil
}

func (c *Context) Context() stdctx.Context {
	if c.request == nil {
		return stdctx.Background()
	}
	return c.request.Context()
}

func (c *Context) SetContext(ctx stdctx.Context) {
	if c.request != nil {
		c.request = c.request.WithContext(ctx)
	}
}

func (c *Context) Method() string {
	if c.request == nil {
		return ""
	}
	return c.request.Method
}

func (c *Context) Path() string {
	if c.request == nil || c.request.URL == nil {
		return ""
	}
	return c.request.URL.Path
}

func (c *Context) SetPath(path string) {
	if c.request == nil || c.request.URL == nil {
		return
	}
	c.request.URL.Path = path
	c.request.URL.RawPath = path
	c.request.RequestURI = cloneRequestURI(c.request.URL)
}

func (c *Context) OriginalURL() string {
	if c.request == nil || c.request.URL == nil {
		return ""
	}
	return c.request.URL.RequestURI()
}

func (c *Context) Next() error {
	c.index++
	if c.index < len(c.handlers) {
		return c.handlers[c.index](c)
	}
	return nil
}

func (c *Context) setHandlers(handlers []HandlerFunc) {
	c.handlers = handlers
	c.index = -1
}

func (c *Context) Set(key any, value any) {
	if c.store == nil {
		c.store = make(map[any]any, 8)
	}
	c.store[key] = value
}

func (c *Context) Get(key any) (any, bool) {
	value, ok := c.store[key]
	return value, ok
}

func (c *Context) MustGet(key any) any {
	value, ok := c.Get(key)
	if !ok {
		panic("zinc: context key not found")
	}
	return value
}

func (c *Context) Status(code int) *Context {
	c.status = code
	return c
}

func (c *Context) Param(name string) string {
	for i := 0; i < c.paramCount; i++ {
		if c.PathParams[i].key == name {
			return c.pathParamValueAt(i)
		}
	}
	return ""
}

func (c *Context) ParamOr(name, fallback string) string {
	if value := c.Param(name); value != "" {
		return value
	}
	return fallback
}

func (c *Context) Query(name string) string {
	return c.QueryValues().Get(name)
}

func (c *Context) QueryOr(name, fallback string) string {
	if value := c.Query(name); value != "" {
		return value
	}
	return fallback
}

func (c *Context) QueryValues() url.Values {
	if c.queryParams == nil {
		if c.request == nil || c.request.URL == nil {
			return url.Values{}
		}
		c.queryParams = c.request.URL.Query()
	}
	return c.queryParams
}

func (c *Context) FormValue(name string) string {
	if c.request == nil {
		return ""
	}
	return c.request.FormValue(name)
}

func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.request == nil {
		return nil, errors.New("request is nil")
	}
	file, header, err := c.request.FormFile(name)
	if err != nil {
		return nil, err
	}
	_ = file.Close()
	return header, nil
}

func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}
	files := form.File[name]
	if len(files) == 0 {
		return nil, http.ErrMissingFile
	}
	return files, nil
}

func (c *Context) MultipartForm() (*multipart.Form, error) {
	if c.request == nil {
		return nil, errors.New("request is nil")
	}
	if c.request.MultipartForm != nil {
		return c.request.MultipartForm, nil
	}
	if err := c.request.ParseMultipartForm(32 << 20); err != nil {
		return nil, err
	}
	return c.request.MultipartForm, nil
}

func (c *Context) SaveFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func (c *Context) GetHeader(key string) string {
	if c.request == nil {
		return ""
	}
	return c.request.Header.Get(key)
}

func (c *Context) Cookie(name string) (*http.Cookie, error) {
	if c.request == nil {
		return nil, http.ErrNoCookie
	}
	return c.request.Cookie(name)
}

func (c *Context) Cookies() []*http.Cookie {
	if c.request == nil {
		return nil
	}
	return c.request.Cookies()
}

func (c *Context) BodyBytes() ([]byte, error) {
	body, err := c.bodyBytes()
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), body...), nil
}

func (c *Context) BodyString() (string, error) {
	body, err := c.bodyBytes()
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *Context) bodyBytes() ([]byte, error) {
	if c.bodyRead {
		return c.body, c.bodyErr
	}
	c.bodyRead = true
	if c.request == nil || c.request.Body == nil {
		return nil, nil
	}

	reader := io.Reader(c.request.Body)
	if c.app != nil && c.app.config.BodyLimit > 0 {
		reader = io.LimitReader(reader, c.app.config.BodyLimit+1)
	}
	body, err := io.ReadAll(reader)
	if err == nil && c.app != nil && c.app.config.BodyLimit > 0 && int64(len(body)) > c.app.config.BodyLimit {
		err = ErrRequestEntityTooLarge
		body = nil
	}
	if closeErr := c.request.Body.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	c.body = body
	c.bodyErr = err
	if err == nil {
		c.request.Body = io.NopCloser(bytes.NewReader(body))
	}
	return c.body, c.bodyErr
}

func (c *Context) Scheme() string {
	if c.request == nil {
		return "http"
	}
	if c.request.TLS != nil {
		return "https"
	}
	if c.trustProxy() {
		if proto := c.GetHeader("X-Forwarded-Proto"); proto != "" {
			if idx := strings.IndexByte(proto, ','); idx >= 0 {
				return strings.TrimSpace(proto[:idx])
			}
			return strings.TrimSpace(proto)
		}
	}
	return "http"
}

func (c *Context) IP() string {
	ips := c.IPs()
	if len(ips) == 0 {
		return ""
	}
	return ips[0]
}

func (c *Context) IPs() []string {
	remote := c.RemoteIP()
	if remote == "" {
		return nil
	}
	if !c.trustProxy() {
		return []string{remote}
	}
	header := c.app.config.ProxyHeader
	if header == "" {
		header = DefaultConfig.ProxyHeader
	}
	raw := c.GetHeader(header)
	if raw == "" {
		return []string{remote}
	}
	parts := strings.Split(raw, ",")
	ips := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ips = append(ips, part)
		}
	}
	if len(ips) == 0 {
		return []string{remote}
	}
	return ips
}

func (c *Context) RemoteIP() string {
	if c.request == nil {
		return ""
	}
	ip, _, err := net.SplitHostPort(c.request.RemoteAddr)
	if err != nil {
		return c.request.RemoteAddr
	}
	return ip
}

func (c *Context) Secure() bool {
	return c.Scheme() == "https"
}

func (c *Context) IsWebSocket() bool {
	connection := strings.ToLower(c.GetHeader(HeaderConnection))
	upgrade := strings.ToLower(c.GetHeader(HeaderUpgrade))
	return strings.Contains(connection, "upgrade") && upgrade == "websocket"
}

func (c *Context) IsPreflight() bool {
	return c.Method() == MethodOptions && c.GetHeader(HeaderAccessControlRequestMethod) != ""
}

func (c *Context) RequestID() string {
	return c.GetHeader(HeaderXRequestID)
}

func (c *Context) FullPath() string {
	return c.routeInfo.path
}

func (c *Context) LastError() error {
	return c.lastErr
}

func (c *Context) Error(err error) {
	if err == nil {
		return
	}
	c.lastErr = err
	if c.app != nil {
		c.app.handleError(c, err)
	}
}

func (c *Context) Route() RouteInfo {
	return c.routeInfo.export()
}

func (c *Context) setRoute(info routeMeta) {
	c.routeInfo = info
}

func (c *Context) setParam(key, value string) {
	for i := range c.PathParams {
		if c.PathParams[i].key == "" {
			c.PathParams[i] = param{key: key, value: value, start: directParamStart}
			if c.paramCount < len(c.PathParams) {
				c.paramCount++
			}
			return
		}
	}
}

func (c *Context) applyRouteParams(path string, route *radixRoute, values [8]paramRange) {
	count := route.paramCount
	if count > len(c.PathParams) {
		count = len(c.PathParams)
	}
	if c.request != nil && c.request.URL != nil && c.request.URL.Path == path {
		c.paramPath = ""
	} else {
		c.paramPath = path
	}
	previousCount := c.paramCount
	for i := 0; i < count; i++ {
		valueRange := values[i]
		c.PathParams[i] = param{
			key:   route.paramNames[i],
			start: int32(valueRange.start),
			end:   int32(valueRange.end),
		}
	}
	for i := count; i < previousCount; i++ {
		c.PathParams[i] = emptyParam
	}
	c.paramCount = count
}

func (c *Context) truncateParams(count int) {
	if count < 0 {
		count = 0
	}
	if count > c.paramCount {
		count = c.paramCount
	}
	for i := count; i < c.paramCount; i++ {
		c.PathParams[i] = emptyParam
	}
	c.paramCount = count
}

func (c *Context) pathParamValueAt(i int) string {
	p := c.PathParams[i]
	if p.start == directParamStart {
		return p.value
	}
	path := c.paramPath
	if path == "" && c.request != nil && c.request.URL != nil {
		path = c.request.URL.Path
	}
	start := int(p.start)
	end := int(p.end)
	if start < 0 || end < start || end > len(path) {
		return ""
	}
	return path[start:end]
}

func (c *Context) trustProxy() bool {
	if c.app == nil || len(c.app.config.TrustedProxies) == 0 {
		return false
	}
	return isTrustedProxy(c.RemoteIP(), c.app.config.TrustedProxies)
}

func isTrustedProxy(ip string, trusted []string) bool {
	if len(trusted) == 0 || ip == "" {
		return false
	}
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return false
	}
	for _, candidate := range trusted {
		if strings.Contains(candidate, "/") {
			_, network, err := net.ParseCIDR(candidate)
			if err == nil && network.Contains(parsed) {
				return true
			}
			continue
		}
		if parsed.Equal(net.ParseIP(candidate)) {
			return true
		}
	}
	return false
}

func cloneURL(u *url.URL) *url.URL {
	if u == nil {
		return &url.URL{}
	}
	clone := *u
	return &clone
}

func cloneRequestURI(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.RequestURI()
}
