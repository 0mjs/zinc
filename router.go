package zinc

import (
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"sync"
)

type RouteHandlerMap map[string]HandlerFunc

type RouteMap map[string]map[string]*Route

type Route struct {
	handler HandlerFunc
	method  string
	path    string
	info    routeMeta
}

type radixNodeKind uint8

const (
	radixRoot radixNodeKind = iota
	radixStatic
	radixParam
	radixCatchAll
)

type radixRoute struct {
	handler    HandlerFunc
	path       string
	paramNames [8]string
	paramCount int
	info       routeMeta
}

type radixNode struct {
	kind          radixNodeKind
	prefix        string
	route         *radixRoute
	indices       []byte
	indexTable    *[256]uint16
	children      []*radixNode
	paramChild    *radixNode
	catchAllChild *radixNode
}

type Router struct {
	cache             *RouteCache
	config            *Config
	routes            RouteMap
	staticFold        map[string]map[uint64][]staticFoldEntry
	trees             map[string]*radixNode
	pathMethods       map[string]uint16
	routeInfos        []routeMeta
	dynamicRouteCount int
}

type routeCacheKey struct {
	method string
	path   string
}

type routeCacheEntry struct {
	route  *radixRoute
	values [8]paramRange
}

type paramRange struct {
	start uint32
	end   uint32
}

type staticFoldEntry struct {
	path  string
	route *Route
}

type RouteCache struct {
	cache map[routeCacheKey]routeCacheEntry
	mu    sync.RWMutex
	size  int
	keys  []routeCacheKey
}

var pathBuilderPool = sync.Pool{
	New: func() any {
		return new(strings.Builder)
	},
}

var handlerNameCache sync.Map

var routeMethods = []string{
	MethodGet,
	MethodHead,
	MethodPost,
	MethodPut,
	MethodPatch,
	MethodDelete,
	MethodOptions,
	MethodConnect,
	MethodTrace,
}

const routeCacheMinDynamicRoutes = 64

func NewRouteCache(size int) *RouteCache {
	return &RouteCache{
		cache: make(map[routeCacheKey]routeCacheEntry, size),
		size:  size,
		keys:  make([]routeCacheKey, 0, size),
	}
}

func (r *Router) Add(method, path string, handlers ...HandlerFunc) error {
	if len(handlers) == 0 {
		return fmt.Errorf("no handler provided for %s %s", method, path)
	}

	path = r.normalizePath(path)
	originalPath := path
	lowerPath := path
	if r.config != nil && !r.config.CaseSensitive {
		lowerPath = strings.ToLower(path)
	}
	pathWithoutSlash := path
	if len(path) > 1 && path[len(path)-1] == '/' {
		pathWithoutSlash = path[:len(path)-1]
	}
	pathWithSlash := path
	if len(path) > 0 && path[len(path)-1] != '/' {
		pathWithSlash = path + "/"
	}

	paramNames, paramCount, isDynamic, err := collectRouteParams(path)
	if err != nil {
		return err
	}

	finalHandler := handlers[len(handlers)-1]
	precomposed := finalHandler
	if len(handlers) > 1 {
		chain := append([]HandlerFunc(nil), handlers...)
		precomposed = func(c *Context) error {
			c.setHandlers(chain)
			return c.Next()
		}
	}

	info := routeMeta{method: method, path: path, handler: finalHandler}
	r.routeInfos = append(r.routeInfos, info)

	route := &Route{
		path:    path,
		handler: precomposed,
		method:  method,
		info:    info,
	}

	if !isDynamic {
		if r.routes == nil {
			r.routes = make(map[string]map[string]*Route)
		}
		if _, ok := r.routes[method]; !ok {
			r.routes[method] = make(map[string]*Route)
		}
		methodRoutes := r.routes[method]
		addCandidate := func(candidate string) {
			if candidate == "" {
				return
			}
			methodRoutes[candidate] = route
			r.trackStaticFoldCandidate(method, candidate, route)
			r.trackStaticPathMethod(candidate, method)
		}

		addCandidate(originalPath)
		if r.config != nil && !r.config.CaseSensitive && lowerPath != originalPath {
			addCandidate(lowerPath)
		}
		if r.config == nil || !r.config.StrictRouting {
			addCandidate(pathWithoutSlash)
			addCandidate(pathWithSlash)
		}
		return nil
	}

	if r.trees == nil {
		r.trees = make(map[string]*radixNode, 4)
	}
	root := r.trees[method]
	if root == nil {
		root = &radixNode{kind: radixRoot}
		r.trees[method] = root
	}
	if err := root.add(path, &radixRoute{
		handler:    precomposed,
		path:       path,
		paramNames: paramNames,
		paramCount: paramCount,
		info:       info,
	}); err != nil {
		return err
	}
	r.dynamicRouteCount++
	return nil
}

func (r *Router) Routes() []RouteInfo {
	out := make([]RouteInfo, len(r.routeInfos))
	for i, info := range r.routeInfos {
		out[i] = info.export()
	}
	return out
}

func (r *Router) Find(method, path string) (HandlerFunc, *Context) {
	ctx := &Context{}
	handler := r.findInto(method, path, ctx)
	if handler == nil {
		return nil, nil
	}
	return handler, ctx
}

func (r *Router) findInto(method, path string, ctx *Context) HandlerFunc {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if routes, ok := r.routes[method]; ok {
		if route, ok := routes[originalPath]; ok {
			ctx.setRoute(route.info)
			return route.handler
		}
		if route, ok := routes[path]; ok {
			ctx.setRoute(route.info)
			return route.handler
		}
	}
	if r.config == nil || r.config.CaseSensitive {
		return r.findDynamicInto(method, path, ctx)
	}
	if route := r.findStaticFoldRoute(method, originalPath); route != nil {
		ctx.setRoute(route.info)
		return route.handler
	}
	if path != originalPath {
		if route := r.findStaticFoldRoute(method, path); route != nil {
			ctx.setRoute(route.info)
			return route.handler
		}
	}
	if handler := r.findDynamicInto(method, path, ctx); handler != nil {
		return handler
	}
	if lower, changed := lowercasePath(path); changed {
		if routes, ok := r.routes[method]; ok {
			if route, ok := routes[lower]; ok {
				ctx.setRoute(route.info)
				return route.handler
			}
		}
		return r.findDynamicInto(method, lower, ctx)
	}
	return nil
}

func (r *Router) dispatchInto(method, path string, ctx *Context) (bool, error) {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if routes, ok := r.routes[method]; ok {
		if route, ok := routes[originalPath]; ok {
			ctx.setRoute(route.info)
			return true, route.handler(ctx)
		}
		if route, ok := routes[path]; ok {
			ctx.setRoute(route.info)
			return true, route.handler(ctx)
		}
	}
	if r.config == nil || r.config.CaseSensitive {
		return r.dispatchDynamicInto(method, path, ctx)
	}
	if route := r.findStaticFoldRoute(method, originalPath); route != nil {
		ctx.setRoute(route.info)
		return true, route.handler(ctx)
	}
	if path != originalPath {
		if route := r.findStaticFoldRoute(method, path); route != nil {
			ctx.setRoute(route.info)
			return true, route.handler(ctx)
		}
	}
	if handled, err := r.dispatchDynamicInto(method, path, ctx); handled {
		return true, err
	}
	if lower, changed := lowercasePath(path); changed {
		if routes, ok := r.routes[method]; ok {
			if route, ok := routes[lower]; ok {
				ctx.setRoute(route.info)
				return true, route.handler(ctx)
			}
		}
		return r.dispatchDynamicInto(method, lower, ctx)
	}
	return false, nil
}

func (r *Router) findDynamicInto(method, path string, ctx *Context) HandlerFunc {
	cacheEnabled := r.dynamicCacheEnabled()
	if cacheEnabled {
		key := routeCacheKey{method: method, path: path}
		if entry, ok := r.cache.get(key); ok {
			if entry.route == nil {
				return nil
			}
			ctx.applyRouteParams(path, entry.route, entry.values)
			ctx.setRoute(entry.route.info)
			return entry.route.handler
		}
	}
	root := r.trees[method]
	if root == nil {
		return nil
	}
	var captured [8]paramRange
	matched := root.lookup(path, 0, &captured, 0)
	if matched == nil {
		return nil
	}
	ctx.applyRouteParams(path, matched, captured)
	ctx.setRoute(matched.info)
	if cacheEnabled {
		entry := routeCacheEntry{
			route:  matched,
			values: captured,
		}
		key := routeCacheKey{method: method, path: path}
		r.cache.set(key, entry)
	}
	return matched.handler
}

func (r *Router) dispatchDynamicInto(method, path string, ctx *Context) (bool, error) {
	cacheEnabled := r.dynamicCacheEnabled()
	if cacheEnabled {
		key := routeCacheKey{method: method, path: path}
		if entry, ok := r.cache.get(key); ok {
			if entry.route == nil {
				return false, nil
			}
			ctx.applyRouteParams(path, entry.route, entry.values)
			ctx.setRoute(entry.route.info)
			return true, entry.route.handler(ctx)
		}
	}
	root := r.trees[method]
	if root == nil {
		return false, nil
	}
	var captured [8]paramRange
	matched := root.lookup(path, 0, &captured, 0)
	if matched == nil {
		return false, nil
	}
	ctx.applyRouteParams(path, matched, captured)
	ctx.setRoute(matched.info)
	if cacheEnabled {
		entry := routeCacheEntry{
			route:  matched,
			values: captured,
		}
		key := routeCacheKey{method: method, path: path}
		r.cache.set(key, entry)
	}
	return true, matched.handler(ctx)
}

func (r *Router) dynamicCacheEnabled() bool {
	return r.cache != nil && r.dynamicRouteCount >= routeCacheMinDynamicRoutes
}

func (r *Router) allowedMethods(path string, autoHead, autoOptions bool) []string {
	allowed := make([]string, 0, len(routeMethods)+1)
	hasGet := false
	hasOptions := false
	ctx := &Context{}
	for _, method := range routeMethods {
		ctx.truncateParams(0)
		if r.findInto(method, path, ctx) == nil {
			continue
		}
		allowed = append(allowed, method)
		if method == MethodGet {
			hasGet = true
		}
		if method == MethodOptions {
			hasOptions = true
		}
	}
	if autoHead && hasGet && !containsMethod(allowed, MethodHead) {
		allowed = append(allowed, MethodHead)
	}
	if autoOptions && len(allowed) > 0 && !hasOptions {
		allowed = append(allowed, MethodOptions)
	}
	sort.SliceStable(allowed, func(i, j int) bool {
		return methodRank(allowed[i]) < methodRank(allowed[j])
	})
	return allowed
}

func containsMethod(methods []string, target string) bool {
	for _, method := range methods {
		if method == target {
			return true
		}
	}
	return false
}

func methodRank(method string) int {
	for i, candidate := range routeMethods {
		if candidate == method {
			return i
		}
	}
	return len(routeMethods)
}

func methodBit(method string) uint16 {
	for i, candidate := range routeMethods {
		if candidate == method {
			return 1 << i
		}
	}
	return 0
}

func (r *Router) trackStaticPathMethod(path, method string) {
	bit := methodBit(method)
	if bit == 0 {
		return
	}
	if r.pathMethods == nil {
		r.pathMethods = make(map[string]uint16, 32)
	}
	r.pathMethods[path] |= bit
}

func (r *Router) trackStaticFoldCandidate(method, path string, route *Route) {
	if r.config == nil || r.config.CaseSensitive {
		return
	}
	hash, ok := foldHashASCII(path)
	if !ok {
		return
	}
	if r.staticFold == nil {
		r.staticFold = make(map[string]map[uint64][]staticFoldEntry, 4)
	}
	methodEntries := r.staticFold[method]
	if methodEntries == nil {
		methodEntries = make(map[uint64][]staticFoldEntry, 32)
		r.staticFold[method] = methodEntries
	}
	methodEntries[hash] = append(methodEntries[hash], staticFoldEntry{path: path, route: route})
}

func (r *Router) findStaticFoldRoute(method, path string) *Route {
	if len(path) == 0 || r.staticFold == nil {
		return nil
	}
	hash, ok := foldHashASCII(path)
	if !ok {
		return nil
	}
	methodEntries := r.staticFold[method]
	if len(methodEntries) == 0 {
		return nil
	}
	for _, entry := range methodEntries[hash] {
		if equalFoldASCII(entry.path, path) {
			return entry.route
		}
	}
	return nil
}

func (r *Router) hasDynamicRoutes() bool {
	return len(r.trees) > 0
}

func (r *Router) staticPathKnown(path string) bool {
	if len(r.pathMethods) == 0 {
		return false
	}

	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	if r.config != nil && !r.config.CaseSensitive {
		if lower, changed := lowercasePath(path); changed {
			path = lower
		}
	}
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	if _, ok := r.pathMethods[originalPath]; ok {
		return true
	}
	if _, ok := r.pathMethods[path]; ok {
		return true
	}

	if !strictRouting && path != "/" {
		if _, ok := r.pathMethods[path+"/"]; ok {
			return true
		}
	}

	return false
}

func (r *Router) normalizePath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] == '/' {
		return path
	}
	builder := pathBuilderPool.Get().(*strings.Builder)
	builder.Reset()
	builder.WriteByte('/')
	builder.WriteString(path)
	result := builder.String()
	pathBuilderPool.Put(builder)
	return result
}

func collectRouteParams(path string) ([8]string, int, bool, error) {
	var names [8]string
	count := 0
	dynamic := false
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case ParamIdentifier:
			dynamic = true
			start := i + 1
			if start >= len(path) || path[start] == '/' {
				return names, count, dynamic, fmt.Errorf("invalid parameter in path %q", path)
			}
			end := start
			for end < len(path) && path[end] != '/' {
				if path[end] == ParamIdentifier || path[end] == WildcardIdentifier {
					return names, count, dynamic, fmt.Errorf("invalid parameter in path %q", path)
				}
				end++
			}
			if count < len(names) {
				names[count] = path[start:end]
			}
			count++
			i = end - 1
		case WildcardIdentifier:
			dynamic = true
			start := i + 1
			if start >= len(path) || path[start] == '/' {
				return names, count, dynamic, fmt.Errorf("invalid wildcard in path %q", path)
			}
			end := start
			for end < len(path) && path[end] != '/' {
				if path[end] == ParamIdentifier || path[end] == WildcardIdentifier {
					return names, count, dynamic, fmt.Errorf("invalid wildcard in path %q", path)
				}
				end++
			}
			if end != len(path) {
				return names, count, dynamic, fmt.Errorf("wildcard must be final in path %q", path)
			}
			if count < len(names) {
				names[count] = "*"
			}
			count++
			return names, count, dynamic, nil
		}
	}
	return names, count, dynamic, nil
}

func commonPrefixLen(a, b string) int {
	limit := len(a)
	if len(b) < limit {
		limit = len(b)
	}
	for i := 0; i < limit; i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return limit
}

func nextSlash(path string) int {
	return strings.IndexByte(path, '/')
}

func (n *radixNode) add(path string, route *radixRoute) error {
	current := n
	remaining := path
	for len(remaining) > 0 {
		wildIndex := strings.IndexAny(remaining, ":*")
		if wildIndex < 0 {
			current = current.addStaticPath(remaining)
			remaining = ""
			break
		}
		if wildIndex > 0 {
			current = current.addStaticPath(remaining[:wildIndex])
			remaining = remaining[wildIndex:]
		}
		switch remaining[0] {
		case ParamIdentifier:
			end := 1
			for end < len(remaining) && remaining[end] != '/' {
				end++
			}
			current = current.addParamChild()
			remaining = remaining[end:]
		case WildcardIdentifier:
			end := 1
			for end < len(remaining) && remaining[end] != '/' {
				end++
			}
			if end != len(remaining) {
				return fmt.Errorf("wildcard must be final in path %q", route.path)
			}
			current = current.addCatchAllChild()
			remaining = ""
		default:
			return fmt.Errorf("invalid route segment in path %q", route.path)
		}
	}
	if current.route != nil {
		return fmt.Errorf("route already registered for %s", route.path)
	}
	current.route = route
	return nil
}

func (n *radixNode) addStaticPath(path string) *radixNode {
	current := n
	remaining := path
	for len(remaining) > 0 {
		idx := current.staticChildIndex(remaining[0])
		if idx < 0 {
			child := &radixNode{kind: radixStatic, prefix: remaining}
			current.addStaticChild(child)
			return child
		}
		child := current.children[idx]
		common := commonPrefixLen(child.prefix, remaining)
		if common == len(child.prefix) {
			current = child
			remaining = remaining[common:]
			continue
		}
		existing := &radixNode{
			kind:          radixStatic,
			prefix:        child.prefix[common:],
			route:         child.route,
			indices:       child.indices,
			indexTable:    child.indexTable,
			children:      child.children,
			paramChild:    child.paramChild,
			catchAllChild: child.catchAllChild,
		}
		child.prefix = child.prefix[:common]
		child.route = nil
		child.indices = nil
		child.indexTable = nil
		child.children = nil
		child.paramChild = nil
		child.catchAllChild = nil
		child.addStaticChild(existing)
		if common == len(remaining) {
			return child
		}
		inserted := &radixNode{kind: radixStatic, prefix: remaining[common:]}
		child.addStaticChild(inserted)
		return inserted
	}
	return current
}

func (n *radixNode) addParamChild() *radixNode {
	if n.paramChild != nil {
		return n.paramChild
	}
	child := &radixNode{kind: radixParam}
	n.paramChild = child
	return child
}

func (n *radixNode) addCatchAllChild() *radixNode {
	if n.catchAllChild != nil {
		return n.catchAllChild
	}
	child := &radixNode{kind: radixCatchAll}
	n.catchAllChild = child
	return child
}

func (n *radixNode) addStaticChild(child *radixNode) {
	n.indices = append(n.indices, child.prefix[0])
	n.children = append(n.children, child)
	if n.indexTable != nil {
		n.indexTable[child.prefix[0]] = uint16(len(n.children))
		return
	}
	if len(n.children) < 8 {
		return
	}
	table := new([256]uint16)
	for i, index := range n.indices {
		table[index] = uint16(i + 1)
	}
	n.indexTable = table
}

func (n *radixNode) staticChildIndex(b byte) int {
	if n.indexTable != nil {
		if idx := n.indexTable[b]; idx > 0 {
			return int(idx) - 1
		}
		return -1
	}
	for i, index := range n.indices {
		if index == b {
			return i
		}
	}
	return -1
}

func (n *radixNode) lookup(path string, offset int, values *[8]paramRange, captured int) *radixRoute {
	switch n.kind {
	case radixStatic:
		if len(path) < len(n.prefix) || path[:len(n.prefix)] != n.prefix {
			return nil
		}
		offset += len(n.prefix)
		path = path[len(n.prefix):]
	case radixParam:
		if len(path) == 0 || path[0] == '/' {
			return nil
		}
		end := nextSlash(path)
		if end < 0 {
			end = len(path)
		}
		if captured < len(values) {
			values[captured] = paramRange{
				start: uint32(offset),
				end:   uint32(offset + end),
			}
		}
		captured++
		offset += end
		path = path[end:]
	case radixCatchAll:
		start := offset
		if len(path) > 0 && path[0] == '/' {
			start++
		}
		if captured < len(values) {
			values[captured] = paramRange{
				start: uint32(start),
				end:   uint32(offset + len(path)),
			}
		}
		captured++
		return n.matchRoute(captured)
	}
	if len(path) == 0 {
		return n.matchRoute(captured)
	}
	if idx := n.staticChildIndex(path[0]); idx >= 0 {
		if matched := n.children[idx].lookup(path, offset, values, captured); matched != nil {
			return matched
		}
	}
	if n.paramChild != nil {
		if matched := n.paramChild.lookup(path, offset, values, captured); matched != nil {
			return matched
		}
	}
	if n.catchAllChild != nil {
		if matched := n.catchAllChild.lookup(path, offset, values, captured); matched != nil {
			return matched
		}
	}
	return nil
}

func (n *radixNode) matchRoute(captured int) *radixRoute {
	if n.route == nil || captured != n.route.paramCount {
		return nil
	}
	return n.route
}

func (rc *RouteCache) get(key routeCacheKey) (routeCacheEntry, bool) {
	rc.mu.RLock()
	entry, ok := rc.cache[key]
	rc.mu.RUnlock()
	return entry, ok
}

func (rc *RouteCache) set(key routeCacheKey, entry routeCacheEntry) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if _, exists := rc.cache[key]; exists {
		rc.cache[key] = entry
		return
	}
	if len(rc.cache) >= rc.size && len(rc.keys) > 0 {
		delete(rc.cache, rc.keys[0])
		rc.keys = rc.keys[1:]
	}
	rc.cache[key] = entry
	rc.keys = append(rc.keys, key)
}

func lowercasePath(path string) (string, bool) {
	for i := 0; i < len(path); i++ {
		c := path[i]
		if c >= 'A' && c <= 'Z' {
			return strings.ToLower(path), true
		}
		if c >= 0x80 {
			lower := strings.ToLower(path)
			return lower, lower != path
		}
	}
	return path, false
}

func foldASCIIByte(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + ('a' - 'A')
	}
	return b
}

func equalFoldASCII(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if foldASCIIByte(a[i]) != foldASCIIByte(b[i]) {
			return false
		}
	}
	return true
}

func foldHashASCII(path string) (uint64, bool) {
	var hash uint64 = 1469598103934665603
	for i := 0; i < len(path); i++ {
		b := path[i]
		if b >= 0x80 {
			return 0, false
		}
		hash ^= uint64(foldASCIIByte(b))
		hash *= 1099511628211
	}
	return hash, true
}

func handlerName(handler HandlerFunc) string {
	if handler == nil {
		return ""
	}
	value := reflect.ValueOf(handler)
	if !value.IsValid() || value.IsNil() {
		return ""
	}
	pc := value.Pointer()
	if cached, ok := handlerNameCache.Load(pc); ok {
		return cached.(string)
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ""
	}
	name := fn.Name()
	handlerNameCache.Store(pc, name)
	return name
}

func normalizeGetHandlers(handlers ...any) ([]HandlerFunc, error) {
	if len(handlers) == 0 {
		return nil, nil
	}
	out := make([]HandlerFunc, 0, len(handlers))
	for idx, handler := range handlers {
		switch h := handler.(type) {
		case HandlerFunc:
			if h == nil {
				return nil, fmt.Errorf("handler at index %d is nil", idx)
			}
			out = append(out, h)
		case func(*Context) error:
			if h == nil {
				return nil, fmt.Errorf("handler at index %d is nil", idx)
			}
			out = append(out, HandlerFunc(h))
		case string:
			out = append(out, StringHandler(h))
		case nil:
			return nil, fmt.Errorf("handler at index %d is nil", idx)
		default:
			return nil, fmt.Errorf("unsupported GET handler type %T at index %d", handler, idx)
		}
	}
	return out, nil
}

func (a *App) Add(method, path string, handlers ...HandlerFunc) error {
	return a.router.Add(method, path, handlers...)
}

func (a *App) Get(path string, handlers ...any) error {
	normalized, err := normalizeGetHandlers(handlers...)
	if err != nil {
		return err
	}
	return a.Add(MethodGet, path, normalized...)
}
func (a *App) Post(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodPost, path, handlers...)
}
func (a *App) Put(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodPut, path, handlers...)
}
func (a *App) Delete(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodDelete, path, handlers...)
}
func (a *App) Patch(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodPatch, path, handlers...)
}
func (a *App) Head(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodHead, path, handlers...)
}
func (a *App) Options(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodOptions, path, handlers...)
}
func (a *App) Connect(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodConnect, path, handlers...)
}
func (a *App) Trace(path string, handlers ...HandlerFunc) error {
	return a.Add(MethodTrace, path, handlers...)
}

func (a *App) Match(methods []string, path string, handlers ...HandlerFunc) error {
	for _, method := range methods {
		if err := a.Add(method, path, handlers...); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) All(path string, handlers ...HandlerFunc) error {
	return a.Match(routeMethods, path, handlers...)
}

func (a *App) Any(path string, handlers ...HandlerFunc) error {
	return a.All(path, handlers...)
}

// SSE registers a GET route that serves server-sent events.
func (a *App) SSE(path string, handler SSEHandler) error {
	if handler == nil {
		return ErrSSEHandlerNil
	}
	return a.Get(path, func(c *Context) error {
		return c.SSE(handler)
	})
}

// WS registers a GET route that upgrades to websocket.
func (a *App) WS(path string, handler WebSocketHandler, config ...WebSocketConfig) error {
	if handler == nil {
		return ErrWebSocketHandlerNil
	}
	cfg := firstWebSocketConfig(config)
	return a.Get(path, func(c *Context) error {
		return c.WebSocket(handler, cfg)
	})
}

func StringHandler(str string) HandlerFunc {
	return func(c *Context) error {
		return c.String(str)
	}
}
