package zinc

import (
	"bytes"
	stdctx "context"
	"errors"
	"io"
	"mime"
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
	writer       http.ResponseWriter
	request      *http.Request
	PathParams   params
	inlineParams [inlineParamSlotCount]param
	queryParams  url.Values
	written      bool
	handlers     []HandlerFunc
	index        int
	store        map[any]any
	status       int
	app          *App
	routeInfo    routeMeta
	routeIndex   int32
	routeIndexed bool
	lastErr      error
	body         []byte
	bodyRead     bool
	bodyErr      error
	sameSite     http.SameSite
	paramPath    string
	paramCount   int
	paramRanges  paramRanges
	matchRanges  paramRanges
	paramRoute   *radixRoute
}

type param struct {
	key   string
	value string
	start int32
	end   int32
}

type params []param

var emptyParam param

const inlineParamSlotCount = 8

const directParamStart int32 = -1

var contextPool = sync.Pool{
	New: func() any {
		c := &Context{
			status:     http.StatusOK,
			index:      -1,
			routeIndex: -1,
		}
		c.PathParams = c.inlineParams[:inlineParamSlotCount]
		return c
	},
}

func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	c := contextPool.Get().(*Context)
	c.reset(w, r)
	return c
}

func (c *Context) Copy() *Context {
	if c == nil {
		return nil
	}

	clone := &Context{
		writer:       c.writer,
		request:      cloneRequest(c.request, c.body),
		written:      c.written,
		index:        c.index,
		status:       c.status,
		app:          c.app,
		lastErr:      c.lastErr,
		bodyRead:     c.bodyRead,
		bodyErr:      c.bodyErr,
		sameSite:     c.sameSite,
		paramPath:    c.paramPath,
		paramCount:   c.paramCount,
		paramRanges:  c.paramRanges,
		matchRanges:  c.matchRanges,
		paramRoute:   c.paramRoute,
		routeIndex:   -1,
		routeIndexed: false,
	}
	clone.PathParams = clone.inlineParams[:inlineParamSlotCount]
	clone.ensurePathParamCapacity(c.paramCount)
	copy(clone.PathParams[:c.paramCount], c.PathParams[:c.paramCount])
	if len(c.queryParams) > 0 {
		clone.queryParams = cloneValues(c.queryParams)
	}
	if len(c.body) > 0 {
		clone.body = append([]byte(nil), c.body...)
	}
	if len(c.handlers) > 0 {
		clone.handlers = append([]HandlerFunc(nil), c.handlers...)
	}
	if len(c.store) > 0 {
		clone.store = make(map[any]any, len(c.store))
		for key, value := range c.store {
			clone.store[key] = value
		}
	}
	if c.routeIndexed && c.app != nil && c.app.router != nil {
		clone.routeInfo = c.app.router.routeMetaAt(uint32(c.routeIndex))
	} else {
		clone.routeInfo = c.routeInfo
	}
	return clone
}

func (c *Context) reset(w http.ResponseWriter, r *http.Request) {
	// release clears request-owned references before pooling. Keep reset focused
	// on state that handlers mutate without release-time retention concerns.
	c.initPathParams()
	c.writer = w
	c.request = r
	c.written = false
	c.index = -1
	c.status = http.StatusOK
	c.app = nil
	c.routeInfo = routeMeta{}
	c.routeIndex = -1
	c.routeIndexed = false
	c.lastErr = nil
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
	c.sameSite = 0
	c.paramPath = ""
	c.paramRoute = nil
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

func (c *Context) GetString(key any) string {
	value, _ := c.Get(key)
	result, _ := value.(string)
	return result
}

func (c *Context) GetBool(key any) bool {
	value, _ := c.Get(key)
	result, _ := value.(bool)
	return result
}

func (c *Context) GetInt(key any) int {
	value, _ := c.Get(key)
	result, _ := value.(int)
	return result
}

func (c *Context) GetInt64(key any) int64 {
	value, _ := c.Get(key)
	result, _ := value.(int64)
	return result
}

func (c *Context) GetFloat64(key any) float64 {
	value, _ := c.Get(key)
	result, _ := value.(float64)
	return result
}

func (c *Context) GetStringSlice(key any) []string {
	value, _ := c.Get(key)
	result, _ := value.([]string)
	return result
}

func (c *Context) GetStringMap(key any) map[string]any {
	value, _ := c.Get(key)
	switch result := value.(type) {
	case map[string]any:
		return result
	case Map:
		return map[string]any(result)
	default:
		return nil
	}
}

func (c *Context) GetStringMapString(key any) map[string]string {
	value, _ := c.Get(key)
	result, _ := value.(map[string]string)
	return result
}

func (c *Context) GetStringMapStringSlice(key any) map[string][]string {
	value, _ := c.Get(key)
	result, _ := value.(map[string][]string)
	return result
}

func (c *Context) Status(code int) *Context {
	c.status = code
	return c
}

func (c *Context) Param(name string) string {
	if route := c.paramRoute; route != nil {
		if c.paramPath != "" && c.paramCount > 1 && len(route.paramIndices) > 0 {
			c.materializePathParams()
		}
		if index, ok := route.paramIndex(name); ok {
			if index >= c.paramCount {
				return ""
			}
			if c.PathParams[index].start == directParamStart {
				return c.PathParams[index].value
			}
			return c.pathParamValueAt(index)
		}
		return ""
	}
	if c.paramPath != "" && c.paramCount > 1 {
		c.materializePathParams()
	}
	for i := 0; i < c.paramCount; i++ {
		if c.PathParams[i].key == name {
			if c.PathParams[i].start == directParamStart {
				return c.PathParams[i].value
			}
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

func (c *Context) QueryArray(name string) []string {
	values := c.QueryValues()[name]
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func (c *Context) QueryMap(name string) map[string]string {
	return valuesMap(c.QueryValues(), name)
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

func (c *Context) PostForm(name string) string {
	return c.postFormValues().Get(name)
}

func (c *Context) PostFormOr(name, fallback string) string {
	if value := c.PostForm(name); value != "" {
		return value
	}
	return fallback
}

func (c *Context) PostFormArray(name string) []string {
	values := c.postFormValues()[name]
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func (c *Context) PostFormMap(name string) map[string]string {
	return valuesMap(c.postFormValues(), name)
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

func (c *Context) ContentType() string {
	return mediaTypeOnly(c.GetHeader(HeaderContentType))
}

func (c *Context) IsWebSocket() bool {
	return headerHasToken(c.GetHeader(HeaderConnection), "upgrade") &&
		strings.EqualFold(strings.TrimSpace(c.GetHeader(HeaderUpgrade)), "websocket")
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

func (c *Context) postFormValues() url.Values {
	if c.request == nil {
		return url.Values{}
	}
	if c.request.PostForm != nil {
		return c.request.PostForm
	}
	if c.ContentType() == "multipart/form-data" {
		if err := c.request.ParseMultipartForm(32 << 20); err != nil {
			return url.Values{}
		}
		if c.request.PostForm != nil {
			return c.request.PostForm
		}
		if c.request.MultipartForm != nil {
			return c.request.MultipartForm.Value
		}
		return url.Values{}
	}
	if err := c.request.ParseForm(); err != nil {
		return url.Values{}
	}
	if c.request.PostForm == nil {
		return url.Values{}
	}
	return c.request.PostForm
}

func valuesMap(values url.Values, name string) map[string]string {
	out := map[string]string{}
	if name == "" {
		return out
	}
	prefix := name + "["
	for key, values := range values {
		if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, "]") {
			continue
		}
		mapKey := key[len(prefix) : len(key)-1]
		if mapKey == "" {
			continue
		}
		if len(values) == 0 {
			out[mapKey] = ""
			continue
		}
		out[mapKey] = values[0]
	}
	return out
}

func mediaTypeOnly(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if mediaType, _, err := mime.ParseMediaType(value); err == nil {
		return strings.ToLower(mediaType)
	}
	if mediaType, _, ok := strings.Cut(value, ";"); ok {
		return strings.ToLower(strings.TrimSpace(mediaType))
	}
	return strings.ToLower(value)
}

func headerHasToken(value, token string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}

func (c *Context) bodyBytes() ([]byte, error) {
	if c.bodyRead {
		return c.body, c.bodyErr
	}
	return c.readAndCacheBodyBytes()
}

func (c *Context) readAndCacheBodyBytes() ([]byte, error) {
	if c.bodyRead {
		return c.body, c.bodyErr
	}
	c.bodyRead = true
	if !requestHasBody(c.request) {
		return nil, nil
	}

	reader := io.Reader(c.request.Body)
	if c.app != nil && c.app.config.BodyLimit > 0 {
		reader = io.LimitReader(reader, c.app.config.BodyLimit+1)
	}

	body, readErr := io.ReadAll(reader)
	if readErr == nil && c.app != nil && c.app.config.BodyLimit > 0 && int64(len(body)) > c.app.config.BodyLimit {
		readErr = ErrRequestEntityTooLarge
		body = nil
	}
	if closeErr := c.request.Body.Close(); readErr == nil && closeErr != nil {
		readErr = closeErr
	}

	c.body = body
	c.bodyErr = readErr
	if readErr == nil {
		c.request.Body = io.NopCloser(bytes.NewReader(body))
	}
	return body, readErr
}

func (c *Context) readAndCacheJSONBody(codec JSONCodec, v any) (int, error, error) {
	body, readErr := c.readAndCacheBodyBytes()
	if readErr != nil {
		return len(body), readErr, nil
	}
	if len(body) == 0 {
		return 0, nil, nil
	}
	return len(body), nil, decodeJSONBody(codec, body, v)
}

func (c *Context) readAndCacheBody(decode func(io.Reader) error) (int, error, error) {
	if c.bodyRead {
		if c.bodyErr != nil {
			return len(c.body), c.bodyErr, nil
		}
		if decode == nil {
			return len(c.body), nil, nil
		}
		return len(c.body), nil, decode(bytes.NewReader(c.body))
	}
	c.bodyRead = true
	if !requestHasBody(c.request) {
		return 0, nil, nil
	}

	reader := io.Reader(c.request.Body)
	if c.app != nil && c.app.config.BodyLimit > 0 {
		reader = io.LimitReader(reader, c.app.config.BodyLimit+1)
	}

	capture := newBodyCaptureReader(reader, c.request.ContentLength)
	var decodeErr error
	if decode != nil {
		decodeErr = decode(capture)
	}
	_, drainErr := io.Copy(io.Discard, capture)

	body := capture.Bytes()
	readErr := capture.readErr
	if readErr == nil {
		readErr = drainErr
	}
	if readErr == nil && c.app != nil && c.app.config.BodyLimit > 0 && int64(len(body)) > c.app.config.BodyLimit {
		readErr = ErrRequestEntityTooLarge
		body = nil
	}
	if closeErr := c.request.Body.Close(); readErr == nil && closeErr != nil {
		readErr = closeErr
	}

	c.body = body
	c.bodyErr = readErr
	if readErr == nil {
		c.request.Body = io.NopCloser(bytes.NewReader(body))
	}
	return len(body), readErr, decodeErr
}

func requestHasBody(req *http.Request) bool {
	if req == nil || req.Body == nil || req.Body == http.NoBody {
		return false
	}
	return true
}

type bodyCaptureReader struct {
	reader  io.Reader
	buffer  bytes.Buffer
	readErr error
}

func newBodyCaptureReader(reader io.Reader, contentLength int64) *bodyCaptureReader {
	capture := &bodyCaptureReader{reader: reader}
	if contentLength > 0 && contentLength <= int64(^uint(0)>>1) {
		capture.buffer.Grow(int(contentLength))
	}
	return capture
}

func (r *bodyCaptureReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		_, _ = r.buffer.Write(p[:n])
	}
	if err != nil && err != io.EOF && r.readErr == nil {
		r.readErr = err
	}
	return n, err
}

func (r *bodyCaptureReader) Bytes() []byte {
	return r.buffer.Bytes()
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

func (c *Context) IsPreflight() bool {
	return c.Method() == MethodOptions && c.GetHeader(HeaderAccessControlRequestMethod) != ""
}

func (c *Context) RequestID() string {
	return c.GetHeader(HeaderXRequestID)
}

func (c *Context) FullPath() string {
	if c.routeIndexed && c.app != nil && c.app.router != nil {
		return c.app.router.routeMetaAt(uint32(c.routeIndex)).path
	}
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

func (c *Context) AbortWithStatus(code int) error {
	return NewError(code)
}

func (c *Context) AbortWithJSON(code int, v any) error {
	return c.Status(code).JSON(v)
}

func (c *Context) Fail(err error) error {
	return err
}

func (c *Context) Route() RouteInfo {
	if c.routeIndexed && c.app != nil && c.app.router != nil {
		return c.app.router.routeMetaAt(uint32(c.routeIndex)).export()
	}
	return c.routeInfo.export()
}

func (c *Context) setRoute(info routeMeta) {
	c.routeInfo = info
	c.routeIndex = -1
	c.routeIndexed = false
}

func (c *Context) setRouteIndex(index uint32) {
	c.routeInfo = routeMeta{}
	c.routeIndex = int32(index)
	c.routeIndexed = true
}

func (c *Context) initPathParams() {
	if c.PathParams == nil {
		c.PathParams = c.inlineParams[:inlineParamSlotCount]
	}
}

func (c *Context) paramRangesScratch() *paramRanges {
	if c == nil {
		return nil
	}
	return &c.paramRanges
}

func (c *Context) ensurePathParamCapacity(count int) {
	c.initPathParams()
	if count <= len(c.PathParams) {
		return
	}
	size := len(c.PathParams)
	if size < inlineParamSlotCount {
		size = inlineParamSlotCount
	}
	if size == 0 {
		size = inlineParamSlotCount
	}
	for size < count {
		size *= 2
	}
	grown := make(params, size)
	copy(grown, c.PathParams[:c.paramCount])
	c.PathParams = grown
}

func (c *Context) setParam(key, value string) {
	c.ensurePathParamCapacity(c.paramCount + 1)
	c.PathParams[c.paramCount] = param{key: key, value: value, start: directParamStart}
	c.paramCount++
	c.paramRoute = nil
}

func (c *Context) applyRouteParams(path string, route *radixRoute, values paramRanges) {
	count := int(route.paramCount)
	c.ensurePathParamCapacity(count)
	c.paramPath = path
	c.paramRoute = route
	previousCount := c.paramCount
	for i := 0; i < count; i++ {
		valueRange := values.at(i)
		c.PathParams[i] = param{
			key:   route.paramNameAt(i),
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
	if count == 0 {
		c.paramRoute = nil
	}
}

func (c *Context) pathParamValueAt(i int) string {
	p := &c.PathParams[i]
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
	p.value = path[start:end]
	p.start = directParamStart
	p.end = 0
	return p.value
}

func (c *Context) materializePathParams() {
	path := c.paramPath
	if path == "" {
		if c.request == nil || c.request.URL == nil {
			return
		}
		path = c.request.URL.Path
	}
	for i := 0; i < c.paramCount; i++ {
		p := &c.PathParams[i]
		if p.start == directParamStart {
			continue
		}
		start := int(p.start)
		end := int(p.end)
		if start < 0 || end < start || end > len(path) {
			p.value = ""
			p.start = directParamStart
			p.end = 0
			continue
		}
		p.value = path[start:end]
		p.start = directParamStart
		p.end = 0
	}
	c.paramPath = ""
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

func cloneRequest(r *http.Request, body []byte) *http.Request {
	if r == nil {
		return nil
	}
	clone := r.Clone(r.Context())
	if body != nil {
		clone.Body = io.NopCloser(bytes.NewReader(body))
		clone.ContentLength = int64(len(body))
	}
	return clone
}

func cloneValues(values url.Values) url.Values {
	if len(values) == 0 {
		return nil
	}
	clone := make(url.Values, len(values))
	for key, list := range values {
		clone[key] = append([]string(nil), list...)
	}
	return clone
}

func cloneRequestURI(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.RequestURI()
}
