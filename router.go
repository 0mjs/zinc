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
	handler   HandlerFunc
	infoIndex uint32
}

type radixNodeKind uint8

const (
	radixRoot radixNodeKind = iota
	radixStatic
	radixParam
	radixCatchAll
)

type radixRoute struct {
	handler          HandlerFunc
	extraParamNames  []string
	inlineParamNames [2]string
	infoIndex        uint32
	paramCount       uint8
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
	trees             map[string]*radixNode
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
		size: size,
	}
}

func (r *Router) Add(method, path string, handlers ...HandlerFunc) error {
	if len(handlers) == 0 {
		return fmt.Errorf("no handler provided for %s %s", method, path)
	}

	path = r.normalizePath(path)
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

	infoIndex := uint32(len(r.routeInfos))
	info := newRouteMeta(method, path, finalHandler)

	if !isDynamic {
		if r.routes == nil {
			r.routes = make(map[string]map[string]*Route)
		}
		if _, ok := r.routes[method]; !ok {
			r.routes[method] = make(map[string]*Route)
		}
		route := &Route{
			handler:   precomposed,
			infoIndex: infoIndex,
		}
		methodRoutes := r.routes[method]
		strictRouting := r.config != nil && r.config.StrictRouting
		lowerPath := path
		lowerChanged := false
		if r.config != nil && !r.config.CaseSensitive {
			if lower, changed := lowercasePath(path); changed {
				lowerPath = lower
				lowerChanged = true
			}
		}

		var routeCandidates [4]string
		routeCandidateCount := 0
		addRouteCandidate := func(candidate string) {
			if candidate == "" {
				return
			}
			for i := 0; i < routeCandidateCount; i++ {
				if routeCandidates[i] == candidate {
					return
				}
			}
			routeCandidates[routeCandidateCount] = candidate
			routeCandidateCount++
		}

		addRouteCandidate(path)
		if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
			addRouteCandidate(path[:len(path)-1])
		}
		if lowerChanged {
			addRouteCandidate(lowerPath)
			if !strictRouting && len(lowerPath) > 1 && lowerPath[len(lowerPath)-1] == '/' {
				addRouteCandidate(lowerPath[:len(lowerPath)-1])
			}
		}
		for i := 0; i < routeCandidateCount; i++ {
			methodRoutes[routeCandidates[i]] = route
		}

		r.routeInfos = append(r.routeInfos, info)
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
	if err := root.add(path, newRadixRoute(precomposed, infoIndex, paramNames, paramCount)); err != nil {
		return err
	}
	r.routeInfos = append(r.routeInfos, info)
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

func (r *Router) routeMetaAt(index uint32) routeMeta {
	return r.routeInfos[index]
}

func (r *Router) findInto(method, path string, ctx *Context) HandlerFunc {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	caseSensitive := r.config != nil && r.config.CaseSensitive
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if routes, ok := r.routes[method]; ok {
		if route := lookupStaticRouteIn(routes, originalPath, path, caseSensitive); route != nil {
			ctx.setRoute(r.routeMetaAt(route.infoIndex))
			return route.handler
		}
	}
	if caseSensitive {
		return r.findDynamicInto(method, path, ctx)
	}
	if handler := r.findDynamicInto(method, path, ctx); handler != nil {
		return handler
	}
	if lower, changed := lowercasePath(path); changed {
		return r.findDynamicInto(method, lower, ctx)
	}
	return nil
}

func (r *Router) dispatchInto(method, path string, ctx *Context) (bool, error) {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	caseSensitive := r.config != nil && r.config.CaseSensitive
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if routes, ok := r.routes[method]; ok {
		if route := lookupStaticRouteIn(routes, originalPath, path, caseSensitive); route != nil {
			ctx.setRouteIndex(route.infoIndex)
			return true, route.handler(ctx)
		}
	}
	if caseSensitive {
		return r.dispatchDynamicInto(method, path, ctx)
	}
	if handled, err := r.dispatchDynamicInto(method, path, ctx); handled {
		return true, err
	}
	if lower, changed := lowercasePath(path); changed {
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
			ctx.setRoute(r.routeMetaAt(entry.route.infoIndex))
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
	ctx.setRoute(r.routeMetaAt(matched.infoIndex))
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
			ctx.setRouteIndex(entry.route.infoIndex)
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
	ctx.setRouteIndex(matched.infoIndex)
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

func (r *Router) hasDynamicRoutes() bool {
	return len(r.trees) > 0
}

func (r *Router) staticPathKnown(path string) bool {
	if len(r.routes) == 0 {
		return false
	}

	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	caseSensitive := r.config != nil && r.config.CaseSensitive
	for _, methodRoutes := range r.routes {
		if lookupStaticRouteIn(methodRoutes, originalPath, path, caseSensitive) != nil {
			return true
		}
	}
	return false
}

func lookupStaticRouteIn(methodRoutes map[string]*Route, originalPath, path string, caseSensitive bool) *Route {
	if len(methodRoutes) == 0 {
		return nil
	}
	if route := methodRoutes[originalPath]; route != nil {
		return route
	}
	if route := methodRoutes[path]; route != nil {
		return route
	}
	if caseSensitive {
		return nil
	}
	if lower, changed := lowercasePath(originalPath); changed {
		if route := methodRoutes[lower]; route != nil {
			return route
		}
	}
	if path != originalPath {
		if lower, changed := lowercasePath(path); changed {
			if route := methodRoutes[lower]; route != nil {
				return route
			}
		}
	}
	return nil
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

func newRadixRoute(handler HandlerFunc, infoIndex uint32, names [8]string, count int) *radixRoute {
	route := &radixRoute{
		handler:   handler,
		infoIndex: infoIndex,
	}
	if count < 0 {
		count = 0
	}
	if count > len(names) {
		count = len(names)
	}
	route.paramCount = uint8(count)
	inlineCount := count
	if inlineCount > len(route.inlineParamNames) {
		inlineCount = len(route.inlineParamNames)
	}
	for i := 0; i < inlineCount; i++ {
		route.inlineParamNames[i] = names[i]
	}
	if count > len(route.inlineParamNames) {
		route.extraParamNames = append([]string(nil), names[len(route.inlineParamNames):count]...)
	}
	return route
}

func (r *radixRoute) paramNameAt(index int) string {
	if index < len(r.inlineParamNames) {
		return r.inlineParamNames[index]
	}
	return r.extraParamNames[index-len(r.inlineParamNames)]
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
				return fmt.Errorf("wildcard must be final in path %q", path)
			}
			current = current.addCatchAllChild()
			remaining = ""
		default:
			return fmt.Errorf("invalid route segment in path %q", path)
		}
	}
	if current.route != nil {
		return fmt.Errorf("route already registered for %s", path)
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
	if n.route == nil || captured != int(n.route.paramCount) {
		return nil
	}
	return n.route
}

func (rc *RouteCache) get(key routeCacheKey) (routeCacheEntry, bool) {
	if rc == nil || rc.cache == nil {
		return routeCacheEntry{}, false
	}
	rc.mu.RLock()
	entry, ok := rc.cache[key]
	rc.mu.RUnlock()
	return entry, ok
}

func (rc *RouteCache) set(key routeCacheKey, entry routeCacheEntry) {
	if rc == nil || rc.size <= 0 {
		return
	}
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.cache == nil {
		rc.cache = make(map[routeCacheKey]routeCacheEntry, rc.size)
		rc.keys = make([]routeCacheKey, 0, rc.size)
	}
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

func handlerPC(handler HandlerFunc) uintptr {
	if handler == nil {
		return 0
	}
	value := reflect.ValueOf(handler)
	if !value.IsValid() || value.IsNil() {
		return 0
	}
	return value.Pointer()
}

func handlerName(handler HandlerFunc) string {
	return handlerNameFromPC(handlerPC(handler))
}

func handlerNameFromPC(pc uintptr) string {
	if pc == 0 {
		return ""
	}
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
