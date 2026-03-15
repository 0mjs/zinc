package zinc

import (
	"fmt"
	"math/bits"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

type RouteHandlerMap map[string]HandlerFunc

type RouteMap map[string]map[string]*Route

type Route struct {
	handler   HandlerFunc
	infoIndex uint32
}

type dynamicMethodTree struct {
	method string
	root   *radixNode
}

type dynamicMethodTrees []dynamicMethodTree

type radixNodeKind uint8

const (
	radixRoot radixNodeKind = iota
	radixStatic
	radixParam
	radixCatchAll
)

type radixRoute struct {
	handler          HandlerFunc
	methodName       string
	extraParamNames  []string
	inlineParamNames [2]string
	paramIndices     map[string]uint8
	infoIndex        uint32
	paramCount       uint16
	method           methodMask
}

type radixNode struct {
	kind           radixNodeKind
	prefix         string
	route          *radixRoute
	routeMethod    methodMask
	routesByMethod *[routeMethodCount]*radixRoute
	extraRoutes    map[string]*radixRoute
	allowMethods   methodMask
	allowExtra     []string
	allowParamCnt  uint16
	indices        []byte
	indexTable     *[256]uint16
	children       []*radixNode
	paramChild     *radixNode
	catchAllChild  *radixNode
}

type Router struct {
	cache             *RouteCache
	config            *Config
	routes            RouteMap
	staticAllowed     map[string]allowedMethodSet
	dynamicRoots      [routeMethodCount]*radixNode
	dynamicTrees      dynamicMethodTrees
	routeTree         *radixNode
	routeInfos        []routeMeta
	dynamicRouteCount int
}

type routeCacheKey struct {
	method string
	path   string
}

type routeCacheEntry struct {
	route   *radixRoute
	values  paramRanges
	allowed allowedMethodSet
}

type paramRange struct {
	start uint32
	end   uint32
}

type allowedMethodSet struct {
	mask  methodMask
	extra []string
}

type paramRanges struct {
	inline [inlineParamSlotCount]paramRange
	extra  []paramRange
}

type RouteCache struct {
	cache map[routeCacheKey]routeCacheEntry
	mu    sync.RWMutex
	size  int
	keys  []routeCacheKey
	count uint32
	hot   atomic.Pointer[routeCacheHotEntry]
}

type routeCacheHotEntry struct {
	key   routeCacheKey
	entry routeCacheEntry
}

func (trees dynamicMethodTrees) get(method string) *radixNode {
	for i := range trees {
		if trees[i].method == method {
			return trees[i].root
		}
	}
	return nil
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

const routeMethodCount = 9
const indexedParamThreshold = 10

type methodMask uint16

const (
	methodMaskGet methodMask = 1 << iota
	methodMaskHead
	methodMaskPost
	methodMaskPut
	methodMaskPatch
	methodMaskDelete
	methodMaskOptions
	methodMaskConnect
	methodMaskTrace
)

const allowHeaderTableSize = 1 << 9

const routeCacheMinRoutes = 64

var allowHeaderByMask [allowHeaderTableSize]string

func init() {
	for raw := 0; raw < len(allowHeaderByMask); raw++ {
		allowHeaderByMask[raw] = buildAllowHeader(methodMask(raw))
	}
}

func NewRouteCache(size int) *RouteCache {
	return &RouteCache{
		size: size,
	}
}

func (r *Router) Add(method, path string, handlers ...HandlerFunc) error {
	if len(handlers) == 0 {
		return fmt.Errorf("no handler provided for %s %s", method, path)
	}
	if r.cache != nil {
		r.cache.clear()
	}

	path = r.normalizePath(path)
	var (
		paramNames []string
		isDynamic  bool
		err        error
	)
	if strings.IndexAny(path, ":*") >= 0 {
		paramNames, isDynamic, err = collectRouteParams(path)
		if err != nil {
			return err
		}
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
		methodRoutes := r.routes[method]
		if methodRoutes == nil {
			methodRoutes = make(map[string]*Route)
			r.routes[method] = methodRoutes
		}
		strictRouting := r.config != nil && r.config.StrictRouting
		caseSensitive := r.config == nil || r.config.CaseSensitive
		routeCandidates, routeCandidateCount := staticRouteCandidates(path, strictRouting, caseSensitive)
		for i := 0; i < routeCandidateCount; i++ {
			if methodRoutes[routeCandidates[i]] != nil {
				return fmt.Errorf("route already registered for %s", path)
			}
		}
		route := &Route{
			handler:   precomposed,
			infoIndex: infoIndex,
		}
		if r.staticAllowed == nil {
			r.staticAllowed = make(map[string]allowedMethodSet, routeCandidateCount)
		}
		for i := 0; i < routeCandidateCount; i++ {
			candidate := routeCandidates[i]
			methodRoutes[candidate] = route
			allowed := r.staticAllowed[candidate]
			allowed.addMethod(method)
			r.staticAllowed[candidate] = allowed
		}
		r.routeInfos = append(r.routeInfos, info)
		return nil
	}

	mask := methodMaskFor(method)
	route := newRadixRoute(method, precomposed, infoIndex, paramNames)
	if err := r.ensureDynamicTree(method, mask).add(path, route); err != nil {
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
		if route := lookupStaticRouteExact(routes, originalPath, path); route != nil {
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
	if routes, ok := r.routes[method]; ok {
		if route := lookupStaticRouteLower(routes, originalPath, path); route != nil {
			ctx.setRoute(r.routeMetaAt(route.infoIndex))
			return route.handler
		}
	}
	if lower, changed := lowercasePath(path); changed {
		return r.findDynamicInto(method, lower, ctx)
	}
	return nil
}

func (r *Router) dispatchInto(method, path string, needAllowed bool, ctx *Context) (bool, allowedMethodSet, error) {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	caseSensitive := r.config != nil && r.config.CaseSensitive
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	if routes, ok := r.routes[method]; ok {
		if route := lookupStaticRouteExact(routes, originalPath, path); route != nil {
			ctx.setRouteIndex(route.infoIndex)
			return true, allowedMethodSet{}, route.handler(ctx)
		}
	}
	dispatchCacheEnabled := r.dispatchCacheEnabled()
	if dispatchCacheEnabled {
		key := routeCacheKey{method: method, path: originalPath}
		if entry, ok := r.cache.get(key); ok {
			if entry.route == nil {
				return false, entry.allowed, nil
			}
			ctx.applyRouteParams(originalPath, entry.route, entry.values)
			ctx.setRouteIndex(entry.route.infoIndex)
			return true, allowedMethodSet{}, entry.route.handler(ctx)
		}
	}
	captured := ctx.paramRangesScratch()
	entry := r.lookupDynamicDispatch(method, path, needAllowed, captured)
	if !caseSensitive && entry.route == nil {
		if routes, ok := r.routes[method]; ok {
			if route := lookupStaticRouteLower(routes, originalPath, path); route != nil {
				ctx.setRouteIndex(route.infoIndex)
				return true, allowedMethodSet{}, route.handler(ctx)
			}
		}
		if lower, changed := lowercasePath(path); changed {
			lowerEntry := r.lookupDynamicDispatch(method, lower, needAllowed, captured)
			if lowerEntry.route != nil || !lowerEntry.allowed.empty() {
				entry = lowerEntry
			}
		}
	}
	if needAllowed && entry.route == nil {
		entry.allowed.merge(r.lookupStaticAllowedMethods(originalPath, path, caseSensitive))
	}
	if dispatchCacheEnabled {
		if entry.route != nil {
			entry.values = cloneParamRangesForCache(entry.values, int(entry.route.paramCount))
		}
		r.cache.set(routeCacheKey{method: method, path: originalPath}, entry)
	}
	if entry.route == nil {
		return false, entry.allowed, nil
	}
	ctx.applyRouteParams(originalPath, entry.route, entry.values)
	ctx.setRouteIndex(entry.route.infoIndex)
	return true, allowedMethodSet{}, entry.route.handler(ctx)
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
	mask := methodMaskFor(method)
	root := r.dynamicTree(method, mask)
	if root == nil {
		return nil
	}
	captured := ctx.paramRangesScratch()
	if captured == nil {
		captured = &paramRanges{}
	}
	matched := lookupDynamicRoute(root, method, mask, path, captured)
	if matched == nil {
		return nil
	}
	ctx.applyRouteParams(path, matched, *captured)
	ctx.setRoute(r.routeMetaAt(matched.infoIndex))
	if cacheEnabled {
		entry := routeCacheEntry{
			route:  matched,
			values: cloneParamRangesForCache(*captured, int(matched.paramCount)),
		}
		key := routeCacheKey{method: method, path: path}
		r.cache.set(key, entry)
	}
	return matched.handler
}

func (r *Router) dynamicCacheEnabled() bool {
	return r.cache != nil && r.dynamicRouteCount >= routeCacheMinRoutes
}

func (r *Router) dispatchCacheEnabled() bool {
	return r.cache != nil
}

func (r *Router) lookupDynamicDispatch(method, path string, needAllowed bool, captured *paramRanges) routeCacheEntry {
	if r.dynamicRouteCount == 0 {
		return routeCacheEntry{}
	}
	mask := methodMaskFor(method)
	if captured == nil {
		captured = &paramRanges{}
	}
	if root := r.dynamicTree(method, mask); root != nil {
		if route := lookupDynamicRoute(root, method, mask, path, captured); route != nil {
			return routeCacheEntry{
				route:  route,
				values: *captured,
			}
		}
	}
	if !needAllowed {
		return routeCacheEntry{}
	}
	return routeCacheEntry{allowed: r.lookupAllowedInDynamicTrees(path, method)}
}

func (r *Router) allowedMethods(path string, autoHead, autoOptions bool) []string {
	return r.lookupAllowedMethods(path).methods(autoHead, autoOptions)
}

func (r *Router) allowedMethodHeader(path string, autoHead, autoOptions bool) string {
	return r.lookupAllowedMethods(path).header(autoHead, autoOptions)
}

func (r *Router) lookupAllowedMethods(path string) allowedMethodSet {
	originalPath := path
	strictRouting := r.config != nil && r.config.StrictRouting
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	caseSensitive := r.config != nil && r.config.CaseSensitive

	allowed := r.lookupStaticAllowedMethods(originalPath, path, caseSensitive)
	if r.hasDynamicTrees() {
		allowed.merge(r.lookupAllowedDynamicMethods(originalPath, path, caseSensitive))
	}
	return allowed
}

func (r *Router) hasDynamicTrees() bool {
	return r.dynamicRouteCount > 0
}

func (r *Router) dynamicTree(method string, mask methodMask) *radixNode {
	if slot := singleBitIndex(mask); slot >= 0 {
		return r.dynamicRoots[slot]
	}
	return r.dynamicTrees.get(method)
}

func (r *Router) ensureDynamicTree(method string, mask methodMask) *radixNode {
	if slot := singleBitIndex(mask); slot >= 0 {
		if r.dynamicRoots[slot] == nil {
			r.dynamicRoots[slot] = &radixNode{kind: radixRoot}
		}
		return r.dynamicRoots[slot]
	}
	if root := r.dynamicTrees.get(method); root != nil {
		return root
	}
	root := &radixNode{kind: radixRoot}
	r.dynamicTrees = append(r.dynamicTrees, dynamicMethodTree{method: method, root: root})
	return root
}

func lookupDynamicRoute(root *radixNode, method string, mask methodMask, path string, captured *paramRanges) *radixRoute {
	if root == nil {
		return nil
	}
	if mask != 0 {
		return root.lookup(path, 0, captured, 0)
	}
	return root.lookupByMethod(path, 0, captured, 0, method, mask)
}

func (r *Router) lookupAllowedDynamicMethods(originalPath, path string, caseSensitive bool) allowedMethodSet {
	tryLookup := func(candidate string) allowedMethodSet {
		if candidate == "" {
			return allowedMethodSet{}
		}
		return r.lookupAllowedInDynamicTrees(candidate, "")
	}

	if allowed := tryLookup(originalPath); !allowed.empty() {
		return allowed
	}
	if path != originalPath {
		if allowed := tryLookup(path); !allowed.empty() {
			return allowed
		}
	}
	if caseSensitive {
		return allowedMethodSet{}
	}
	if lower, changed := lowercasePath(originalPath); changed {
		if allowed := tryLookup(lower); !allowed.empty() {
			return allowed
		}
	}
	if path != originalPath {
		if lower, changed := lowercasePath(path); changed {
			if allowed := tryLookup(lower); !allowed.empty() {
				return allowed
			}
		}
	}
	return allowedMethodSet{}
}

func (r *Router) lookupAllowedInDynamicTrees(path, excludeMethod string) allowedMethodSet {
	if !r.hasAlternateDynamicMethods(excludeMethod) {
		return allowedMethodSet{}
	}
	if r.routeTree != nil {
		allowed := r.routeTree.lookupAllowed(path, 0)
		allowed.removeMethod(excludeMethod)
		return allowed
	}
	return r.lookupAllowedInDynamicRoots(path, excludeMethod)
}

func (r *Router) hasAlternateDynamicMethods(excludeMethod string) bool {
	for slot, root := range r.dynamicRoots {
		if root == nil {
			continue
		}
		if routeMethods[slot] != excludeMethod {
			return true
		}
	}
	for i := range r.dynamicTrees {
		tree := r.dynamicTrees[i]
		if tree.root != nil && tree.method != excludeMethod {
			return true
		}
	}
	return false
}

func (r *Router) lookupAllowedInDynamicRoots(path, excludeMethod string) allowedMethodSet {
	var allowed allowedMethodSet
	for slot, root := range r.dynamicRoots {
		if root == nil {
			continue
		}
		method := routeMethods[slot]
		if method == excludeMethod {
			continue
		}
		if root.matchesPath(path, 0) {
			allowed.addMethod(method)
		}
	}
	for i := range r.dynamicTrees {
		tree := r.dynamicTrees[i]
		if tree.method == excludeMethod || tree.root == nil {
			continue
		}
		if tree.root.matchesPath(path, 0) {
			allowed.addMethod(tree.method)
		}
	}
	return allowed
}

func (r *Router) lookupStaticAllowedMethods(originalPath, path string, caseSensitive bool) allowedMethodSet {
	if len(r.staticAllowed) == 0 {
		return allowedMethodSet{}
	}
	return lookupStaticAllowed(r.staticAllowed, originalPath, path, caseSensitive)
}

func lookupStaticRouteExact(methodRoutes map[string]*Route, originalPath, path string) *Route {
	if len(methodRoutes) == 0 {
		return nil
	}
	if route := methodRoutes[originalPath]; route != nil {
		return route
	}
	if path != originalPath {
		if route := methodRoutes[path]; route != nil {
			return route
		}
	}
	return nil
}

func lookupStaticRouteLower(methodRoutes map[string]*Route, originalPath, path string) *Route {
	if len(methodRoutes) == 0 {
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

func lookupStaticAllowed(staticAllowed map[string]allowedMethodSet, originalPath, path string, caseSensitive bool) allowedMethodSet {
	if len(staticAllowed) == 0 {
		return allowedMethodSet{}
	}
	if allowed := staticAllowed[originalPath]; !allowed.empty() {
		return allowed
	}
	if path != originalPath {
		if allowed := staticAllowed[path]; !allowed.empty() {
			return allowed
		}
	}
	if caseSensitive {
		return allowedMethodSet{}
	}
	if lower, changed := lowercasePath(originalPath); changed {
		if allowed := staticAllowed[lower]; !allowed.empty() {
			return allowed
		}
	}
	if path != originalPath {
		if lower, changed := lowercasePath(path); changed {
			if allowed := staticAllowed[lower]; !allowed.empty() {
				return allowed
			}
		}
	}
	return allowedMethodSet{}
}

func staticRouteCandidates(path string, strictRouting, caseSensitive bool) ([4]string, int) {
	var candidates [4]string
	count := 0
	addCandidate := func(candidate string) {
		if candidate == "" {
			return
		}
		for i := 0; i < count; i++ {
			if candidates[i] == candidate {
				return
			}
		}
		candidates[count] = candidate
		count++
	}

	addCandidate(path)
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		addCandidate(path[:len(path)-1])
	}
	if caseSensitive {
		return candidates, count
	}
	if lower, changed := lowercasePath(path); changed {
		addCandidate(lower)
		if !strictRouting && len(lower) > 1 && lower[len(lower)-1] == '/' {
			addCandidate(lower[:len(lower)-1])
		}
	}
	return candidates, count
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

func collectRouteParams(path string) ([]string, bool, error) {
	var names []string
	dynamic := false
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case ParamIdentifier:
			dynamic = true
			start := i + 1
			if start >= len(path) || path[start] == '/' {
				return names, dynamic, fmt.Errorf("invalid parameter in path %q", path)
			}
			end := start
			for end < len(path) && path[end] != '/' {
				if path[end] == ParamIdentifier || path[end] == WildcardIdentifier {
					return names, dynamic, fmt.Errorf("invalid parameter in path %q", path)
				}
				end++
			}
			names = append(names, path[start:end])
			i = end - 1
		case WildcardIdentifier:
			dynamic = true
			start := i + 1
			if start >= len(path) || path[start] == '/' {
				return names, dynamic, fmt.Errorf("invalid wildcard in path %q", path)
			}
			end := start
			for end < len(path) && path[end] != '/' {
				if path[end] == ParamIdentifier || path[end] == WildcardIdentifier {
					return names, dynamic, fmt.Errorf("invalid wildcard in path %q", path)
				}
				end++
			}
			if end != len(path) {
				return names, dynamic, fmt.Errorf("wildcard must be final in path %q", path)
			}
			names = append(names, "*")
			return names, dynamic, nil
		}
	}
	return names, dynamic, nil
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

func newRadixRoute(method string, handler HandlerFunc, infoIndex uint32, names []string) *radixRoute {
	route := &radixRoute{
		handler:    handler,
		methodName: method,
		infoIndex:  infoIndex,
		method:     methodMaskFor(method),
	}
	count := len(names)
	route.paramCount = uint16(count)
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
	if count >= indexedParamThreshold {
		route.paramIndices = make(map[string]uint8, count)
		for i, name := range names {
			route.paramIndices[name] = uint8(i)
		}
	}
	return route
}

func (r *radixRoute) paramNameAt(index int) string {
	if index < len(r.inlineParamNames) {
		return r.inlineParamNames[index]
	}
	return r.extraParamNames[index-len(r.inlineParamNames)]
}

func (r *radixRoute) paramIndex(name string) (int, bool) {
	if r == nil || len(r.paramIndices) == 0 {
		return 0, false
	}
	index, ok := r.paramIndices[name]
	return int(index), ok
}

func singleBitIndex(mask methodMask) int {
	raw := uint16(mask)
	if raw == 0 || raw&(raw-1) != 0 {
		return -1
	}
	index := bits.TrailingZeros16(raw)
	if index >= routeMethodCount {
		return -1
	}
	return index
}

func (s *allowedMethodSet) addMethod(method string) {
	if method == "" {
		return
	}
	if mask := methodMaskFor(method); mask != 0 {
		s.mask |= mask
		return
	}
	for _, existing := range s.extra {
		if existing == method {
			return
		}
	}
	s.extra = append(s.extra, method)
}

func (s *allowedMethodSet) merge(other allowedMethodSet) {
	s.mask |= other.mask
	for _, method := range other.extra {
		s.addMethod(method)
	}
}

func (s *allowedMethodSet) removeMethod(method string) {
	if method == "" {
		return
	}
	if mask := methodMaskFor(method); mask != 0 {
		s.mask &^= mask
		return
	}
	for i, existing := range s.extra {
		if existing != method {
			continue
		}
		copy(s.extra[i:], s.extra[i+1:])
		s.extra = s.extra[:len(s.extra)-1]
		return
	}
}

func (s allowedMethodSet) empty() bool {
	return s.mask == 0 && len(s.extra) == 0
}

func (s allowedMethodSet) withAutomatic(autoHead, autoOptions bool) allowedMethodSet {
	if s.empty() {
		return allowedMethodSet{}
	}
	if autoHead && s.mask&methodMaskGet != 0 {
		s.mask |= methodMaskHead
	}
	if autoOptions {
		s.mask |= methodMaskOptions
	}
	return s
}

func (s allowedMethodSet) methods(autoHead, autoOptions bool) []string {
	s = s.withAutomatic(autoHead, autoOptions)
	if s.empty() {
		return nil
	}
	allowed := make([]string, 0, len(routeMethods)+len(s.extra))
	for _, method := range routeMethods {
		if s.mask&methodMaskFor(method) == 0 {
			continue
		}
		allowed = append(allowed, method)
	}
	allowed = append(allowed, s.extra...)
	return allowed
}

func (s allowedMethodSet) header(autoHead, autoOptions bool) string {
	s = s.withAutomatic(autoHead, autoOptions)
	if s.empty() {
		return ""
	}
	if len(s.extra) == 0 {
		return allowHeader(s.mask)
	}
	return buildAllowHeaderWithExtra(s.mask, s.extra)
}

func (p *paramRanges) set(index int, value paramRange) {
	if index < len(p.inline) {
		p.inline[index] = value
		return
	}
	extraIndex := index - len(p.inline)
	if extraIndex >= len(p.extra) {
		grown := make([]paramRange, extraIndex+1)
		copy(grown, p.extra)
		p.extra = grown
	}
	p.extra[extraIndex] = value
}

func (p paramRanges) at(index int) paramRange {
	if index < len(p.inline) {
		return p.inline[index]
	}
	return p.extra[index-len(p.inline)]
}

func (p *paramRanges) cloneFrom(other *paramRanges, count int) {
	p.inline = other.inline
	extraCount := count - len(p.inline)
	if extraCount <= 0 {
		p.extra = nil
		return
	}
	if cap(p.extra) < extraCount {
		p.extra = make([]paramRange, extraCount)
	} else {
		p.extra = p.extra[:extraCount]
	}
	copy(p.extra, other.extra[:extraCount])
}

func cloneParamRangesForCache(values paramRanges, count int) paramRanges {
	var cloned paramRanges
	cloned.cloneFrom(&values, count)
	return cloned
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
	if !current.trySetRoute(route) {
		return fmt.Errorf("route already registered for %s", path)
	}
	return nil
}

func (n *radixNode) trySetRoute(route *radixRoute) bool {
	if route == nil {
		return false
	}
	if !n.hasRoutes() {
		n.allowParamCnt = route.paramCount
	} else if n.allowParamCnt != route.paramCount {
		return false
	}

	if route.method == 0 {
		if n.extraRoutes == nil {
			n.extraRoutes = make(map[string]*radixRoute, 1)
		}
		if n.extraRoutes[route.methodName] != nil {
			return false
		}
		n.extraRoutes[route.methodName] = route
		n.allowExtra = append(n.allowExtra, route.methodName)
		return true
	}

	if n.allowMethods&route.method != 0 {
		return false
	}
	if n.route == nil {
		n.route = route
		n.routeMethod = route.method
		n.allowMethods |= route.method
		return true
	}
	slot := singleBitIndex(route.method)
	if slot < 0 {
		return false
	}
	if n.routesByMethod == nil {
		table := new([routeMethodCount]*radixRoute)
		if existingSlot := singleBitIndex(n.routeMethod); existingSlot >= 0 {
			table[existingSlot] = n.route
		}
		n.routesByMethod = table
	}
	if n.routesByMethod[slot] != nil {
		return false
	}
	n.routesByMethod[slot] = route
	n.allowMethods |= route.method
	return true
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
			kind:           radixStatic,
			prefix:         child.prefix[common:],
			route:          child.route,
			routeMethod:    child.routeMethod,
			routesByMethod: child.routesByMethod,
			extraRoutes:    child.extraRoutes,
			allowMethods:   child.allowMethods,
			allowExtra:     child.allowExtra,
			allowParamCnt:  child.allowParamCnt,
			indices:        child.indices,
			indexTable:     child.indexTable,
			children:       child.children,
			paramChild:     child.paramChild,
			catchAllChild:  child.catchAllChild,
		}
		child.prefix = child.prefix[:common]
		child.route = nil
		child.routeMethod = 0
		child.routesByMethod = nil
		child.extraRoutes = nil
		child.allowMethods = 0
		child.allowExtra = nil
		child.allowParamCnt = 0
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

func (n *radixNode) lookup(path string, offset int, values *paramRanges, captured int) *radixRoute {
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
		values.set(captured, paramRange{
			start: uint32(offset),
			end:   uint32(offset + end),
		})
		captured++
		offset += end
		path = path[end:]
	case radixCatchAll:
		start := offset
		if len(path) > 0 && path[0] == '/' {
			start++
		}
		values.set(captured, paramRange{
			start: uint32(start),
			end:   uint32(offset + len(path)),
		})
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

func (n *radixNode) lookupByMethod(path string, offset int, values *paramRanges, captured int, methodName string, method methodMask) *radixRoute {
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
		values.set(captured, paramRange{
			start: uint32(offset),
			end:   uint32(offset + end),
		})
		captured++
		offset += end
		path = path[end:]
	case radixCatchAll:
		start := offset
		if len(path) > 0 && path[0] == '/' {
			start++
		}
		values.set(captured, paramRange{
			start: uint32(start),
			end:   uint32(offset + len(path)),
		})
		captured++
		return n.matchRouteFor(methodName, method, captured)
	}
	if len(path) == 0 {
		return n.matchRouteFor(methodName, method, captured)
	}
	if idx := n.staticChildIndex(path[0]); idx >= 0 {
		if matched := n.children[idx].lookupByMethod(path, offset, values, captured, methodName, method); matched != nil {
			return matched
		}
	}
	if n.paramChild != nil {
		if matched := n.paramChild.lookupByMethod(path, offset, values, captured, methodName, method); matched != nil {
			return matched
		}
	}
	if n.catchAllChild != nil {
		if matched := n.catchAllChild.lookupByMethod(path, offset, values, captured, methodName, method); matched != nil {
			return matched
		}
	}
	return nil
}

func (n *radixNode) matchesPath(path string, captured int) bool {
	switch n.kind {
	case radixStatic:
		if len(path) < len(n.prefix) || path[:len(n.prefix)] != n.prefix {
			return false
		}
		path = path[len(n.prefix):]
	case radixParam:
		if len(path) == 0 || path[0] == '/' {
			return false
		}
		end := nextSlash(path)
		if end < 0 {
			end = len(path)
		}
		captured++
		path = path[end:]
	case radixCatchAll:
		captured++
		return n.hasPathRoute(captured)
	}

	if len(path) == 0 && n.hasPathRoute(captured) {
		return true
	}
	if len(path) == 0 {
		return false
	}
	if idx := n.staticChildIndex(path[0]); idx >= 0 && n.children[idx].matchesPath(path, captured) {
		return true
	}
	if n.paramChild != nil && n.paramChild.matchesPath(path, captured) {
		return true
	}
	if n.catchAllChild != nil && n.catchAllChild.matchesPath(path, captured) {
		return true
	}
	return false
}

func (n *radixNode) lookupAllowed(path string, captured int) allowedMethodSet {
	switch n.kind {
	case radixStatic:
		if len(path) < len(n.prefix) || path[:len(n.prefix)] != n.prefix {
			return allowedMethodSet{}
		}
		path = path[len(n.prefix):]
	case radixParam:
		if len(path) == 0 || path[0] == '/' {
			return allowedMethodSet{}
		}
		end := nextSlash(path)
		if end < 0 {
			end = len(path)
		}
		captured++
		path = path[end:]
	case radixCatchAll:
		captured++
		return n.allowedPathMethods(captured)
	}

	if len(path) == 0 {
		return n.allowedPathMethods(captured)
	}
	if idx := n.staticChildIndex(path[0]); idx >= 0 {
		if allowed := n.children[idx].lookupAllowed(path, captured); !allowed.empty() {
			return allowed
		}
	}
	if n.paramChild != nil {
		if allowed := n.paramChild.lookupAllowed(path, captured); !allowed.empty() {
			return allowed
		}
	}
	if n.catchAllChild != nil {
		if allowed := n.catchAllChild.lookupAllowed(path, captured); !allowed.empty() {
			return allowed
		}
	}
	return allowedMethodSet{}
}

func (n *radixNode) hasPathRoute(captured int) bool {
	return n.hasRoutes() && captured == int(n.allowParamCnt)
}

func (n *radixNode) allowedPathMethods(captured int) allowedMethodSet {
	if !n.hasPathRoute(captured) {
		return allowedMethodSet{}
	}
	allowed := allowedMethodSet{mask: n.allowMethods}
	if len(n.allowExtra) != 0 {
		allowed.extra = append([]string(nil), n.allowExtra...)
	}
	return allowed
}

func (n *radixNode) matchRoute(captured int) *radixRoute {
	if n.route == nil || captured != int(n.route.paramCount) {
		return nil
	}
	return n.route
}

func (n *radixNode) matchRouteFor(methodName string, method methodMask, captured int) *radixRoute {
	if captured != int(n.allowParamCnt) {
		return nil
	}
	if method != 0 {
		if n.allowMethods&method == 0 {
			return nil
		}
		if n.route != nil && n.routeMethod == method {
			return n.route
		}
		if n.routesByMethod == nil {
			return nil
		}
		slot := singleBitIndex(method)
		if slot < 0 {
			return nil
		}
		return n.routesByMethod[slot]
	}
	if len(n.extraRoutes) == 0 || methodName == "" {
		return nil
	}
	return n.extraRoutes[methodName]
}

func (n *radixNode) hasRoutes() bool {
	return n.route != nil || n.routesByMethod != nil || len(n.extraRoutes) > 0
}

func methodMaskFor(method string) methodMask {
	switch method {
	case MethodGet:
		return methodMaskGet
	case MethodHead:
		return methodMaskHead
	case MethodPost:
		return methodMaskPost
	case MethodPut:
		return methodMaskPut
	case MethodPatch:
		return methodMaskPatch
	case MethodDelete:
		return methodMaskDelete
	case MethodOptions:
		return methodMaskOptions
	case MethodConnect:
		return methodMaskConnect
	case MethodTrace:
		return methodMaskTrace
	default:
		return 0
	}
}

func allowHeader(mask methodMask) string {
	if mask == 0 {
		return ""
	}
	if int(mask) < len(allowHeaderByMask) {
		return allowHeaderByMask[int(mask)]
	}
	return buildAllowHeader(mask)
}

func buildAllowHeader(mask methodMask) string {
	if mask == 0 {
		return ""
	}
	return buildAllowHeaderWithExtra(mask, nil)
}

func buildAllowHeaderWithExtra(mask methodMask, extra []string) string {
	if mask == 0 && len(extra) == 0 {
		return ""
	}
	var builder strings.Builder
	for _, method := range routeMethods {
		if mask&methodMaskFor(method) == 0 {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(method)
	}
	for _, method := range extra {
		if builder.Len() > 0 {
			builder.WriteString(", ")
		}
		builder.WriteString(method)
	}
	return builder.String()
}

func (rc *RouteCache) get(key routeCacheKey) (routeCacheEntry, bool) {
	if rc == nil {
		return routeCacheEntry{}, false
	}
	if hot := rc.hot.Load(); hot != nil && hot.key == key {
		return hot.entry, true
	}
	if rc.cache == nil {
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
		rc.hot.Store(&routeCacheHotEntry{key: key, entry: entry})
		return
	}
	if len(rc.cache) >= rc.size && len(rc.keys) > 0 {
		delete(rc.cache, rc.keys[0])
		rc.keys = rc.keys[1:]
	}
	rc.cache[key] = entry
	rc.keys = append(rc.keys, key)
	atomic.StoreUint32(&rc.count, uint32(len(rc.cache)))
	rc.hot.Store(&routeCacheHotEntry{key: key, entry: entry})
}

func (rc *RouteCache) clear() {
	if rc == nil || atomic.LoadUint32(&rc.count) == 0 {
		if rc != nil {
			rc.hot.Store(nil)
		}
		return
	}
	rc.mu.Lock()
	rc.cache = nil
	rc.keys = nil
	atomic.StoreUint32(&rc.count, 0)
	rc.mu.Unlock()
	rc.hot.Store(nil)
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

func StringHandler(str string) HandlerFunc {
	return func(c *Context) error {
		return c.String(str)
	}
}
