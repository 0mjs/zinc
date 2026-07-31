package zinc

import (
	"fmt"
	"math/bits"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unicode"
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
	radixRegexp
	radixCatchAll
)

type radixRoute struct {
	handler          HandlerFunc
	extraParamNames  []string
	inlineParamNames [2]string
	paramIndices     map[string]uint8
	infoIndex        uint32
	paramCount       uint16
}

type radixNode struct {
	kind           radixNodeKind
	prefix         string
	regexpRaw      string
	regexp         *regexp.Regexp
	route          *radixRoute
	indices        []byte
	indexTable     *[256]uint16
	children       []*radixNode
	regexpChildren []*radixNode
	paramChild     *radixNode
	catchAllChild  *radixNode
}

type Router struct {
	cache           *RouteCache
	config          *Config
	routes          RouteMap
	namedRoutes     map[string]uint32
	staticRoutes    [routeMethodCount]map[string]*Route
	staticAllowed   map[string]allowedMethodSet
	hasCustomStatic bool
	// Dynamic routes are partitioned by method, so each leaf owns one route.
	dynamicRoots      [routeMethodCount]*radixNode
	dynamicTrees      dynamicMethodTrees
	routeInfos        []routeMeta
	dynamicRouteCount int
	staticRouteLens   [routeMethodCount]uint64
	staticLongMethods methodMask
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
	methodCache  [routeMethodCount]map[string]routeCacheEntry
	extraCache   map[routeCacheKey]routeCacheEntry
	overlay      [routeMethodCount]map[string]routeCacheEntry
	extraOverlay map[routeCacheKey]routeCacheEntry
	mu           sync.RWMutex
	size         int
	keys         []routeCacheKey
	next         int
	count        uint32
	hits         uint32
	frozenCount  uint32
	admit        [routeCacheAdmissionShards]uint32
	dirty        uint32
	hot          atomic.Pointer[routeCacheHotEntry]
	snapshot     atomic.Pointer[routeCacheReadSnapshot]
	overlayHits  uint32
}

type routeCacheHotEntry struct {
	key   routeCacheKey
	entry routeCacheEntry
}

type routeCacheReadSnapshot struct {
	methodCache [routeMethodCount]map[string]routeCacheEntry
	extraCache  map[routeCacheKey]routeCacheEntry
}

type collectedRouteParams struct {
	count  int
	inline [2]string
	extra  []string
}

type parsedDynamicSegment struct {
	kind  radixNodeKind
	name  string
	expr  string
	width int
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

// A full cache admits one miss per interval in each shard. This protects hot
// entries from one-hit paths while still letting repeatedly missed paths enter.
const routeCacheAdmissionInterval = 4

// Keep this a power of two so routeCacheAdmissionShard can use a mask.
const routeCacheAdmissionShards = 16

// Promote a stable, partially filled cache after two complete hit cycles.
// Keeping one quarter of the cache free prevents read-mostly promotion from
// disabling adaptation when the concrete-path working set exceeds capacity.
const routeCacheFreezeHitCycles = 2
const routeCacheFreezeCapacityNumerator = 3
const routeCacheFreezeCapacityDenominator = 4

// Replace a frozen snapshot when a new working set is at least half its size
// and remains unchanged for the same two complete hit cycles. New overlay
// insertions reset the stability counter, preventing partial phase shifts from
// being promoted prematurely.
const routeCacheRebaseSizeNumerator = 1
const routeCacheRebaseSizeDenominator = 2

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
	return r.add(method, path, "", handlers...)
}

func (r *Router) AddNamed(method, path, name string, handlers ...HandlerFunc) error {
	return r.add(method, path, name, handlers...)
}

func (r *Router) add(method, path, name string, handlers ...HandlerFunc) error {
	if len(handlers) == 0 {
		return fmt.Errorf("no handler provided for %s %s", method, path)
	}
	if name != "" {
		if _, exists := r.namedRoutes[name]; exists {
			return fmt.Errorf("route name already registered: %s", name)
		}
	}

	path = r.normalizePath(path)
	if err := rejectLegacyRoutePattern(path); err != nil {
		return err
	}
	registeredPath := path
	compiledPath, bracePattern, paramNames, err := compileBraceRoutePattern(path)
	if err != nil {
		return err
	}
	if bracePattern {
		path = compiledPath
	}
	isDynamic := bracePattern
	if !bracePattern {
		if firstDynamic := strings.IndexAny(path, ":*"); firstDynamic >= 0 {
			isDynamic = true
			paramNames, err = collectRouteParams(path, firstDynamic, bracePattern)
			if err != nil {
				return err
			}
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
	info := newRouteMeta(method, registeredPath, name, finalHandler, paramNames.slice(), false)
	mask := methodMaskFor(method)

	if !isDynamic {
		if r.routes == nil {
			r.routes = make(map[string]map[string]*Route)
		}
		if mask == 0 {
			r.hasCustomStatic = true
		}
		methodRoutes := r.staticRoutesFor(method, mask)
		if methodRoutes == nil {
			methodRoutes = make(map[string]*Route)
			if slot := singleBitIndex(mask); slot >= 0 {
				r.staticRoutes[slot] = methodRoutes
			}
			r.routes[method] = methodRoutes
		}
		strictRouting := r.config != nil && r.config.StrictRouting
		caseSensitive := r.config == nil || r.config.CaseSensitive
		if staticRouteHasSingleCandidate(path, strictRouting, caseSensitive) {
			if methodRoutes[path] != nil {
				return fmt.Errorf("route already registered for %s", path)
			}
			route := &Route{
				handler:   precomposed,
				infoIndex: infoIndex,
			}
			methodRoutes[path] = route
			r.recordStaticRouteLength(mask, path)
			r.routeInfos = append(r.routeInfos, info)
			r.recordNamedRoute(name, infoIndex)
			r.invalidateCache()
			return nil
		}
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
			r.recordStaticRouteLength(mask, candidate)
			r.staticAllowed[candidate] = addAllowedMethod(r.staticAllowed[candidate], method, mask)
		}
		r.routeInfos = append(r.routeInfos, info)
		r.recordNamedRoute(name, infoIndex)
		r.invalidateCache()
		return nil
	}

	route := newRadixRoute(precomposed, infoIndex, paramNames)
	if err := r.ensureDynamicTree(method, mask).addWithCase(path, route, r.config == nil || !r.config.CaseSensitive); err != nil {
		return err
	}
	r.routeInfos = append(r.routeInfos, info)
	r.recordNamedRoute(name, infoIndex)
	r.dynamicRouteCount++
	r.invalidateCache()
	return nil
}

func (r *Router) staticRoutesFor(method string, mask methodMask) map[string]*Route {
	if slot := singleBitIndex(mask); slot >= 0 {
		return r.staticRoutes[slot]
	}
	return r.routes[method]
}

func (r *Router) recordStaticRouteLength(mask methodMask, path string) {
	slot := singleBitIndex(mask)
	if slot < 0 {
		return
	}
	length := len(path)
	if length >= 64 {
		r.staticLongMethods |= mask
		return
	}
	r.staticRouteLens[slot] |= uint64(1) << length
}

func (r *Router) hasStaticRouteLength(slot int, mask methodMask, length int) bool {
	if length >= 64 {
		return r.staticLongMethods&mask != 0
	}
	return r.staticRouteLens[slot]&(uint64(1)<<length) != 0
}

func (r *Router) recordNamedRoute(name string, index uint32) {
	if name == "" {
		return
	}
	if r.namedRoutes == nil {
		r.namedRoutes = make(map[string]uint32)
	}
	r.namedRoutes[name] = index
}

func (r *Router) invalidateCache() {
	if r.cache != nil {
		r.cache.invalidate()
	}
}

func (r *Router) Routes() []RouteInfo {
	out := make([]RouteInfo, len(r.routeInfos))
	for i, info := range r.routeInfos {
		out[i] = info.export()
	}
	return out
}

func (r *Router) routeMetaByName(name string) (routeMeta, bool) {
	if r.namedRoutes == nil {
		return routeMeta{}, false
	}
	index, ok := r.namedRoutes[name]
	if !ok {
		return routeMeta{}, false
	}
	return r.routeMetaAt(index), true
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
	mask := methodMaskFor(method)
	slot := singleBitIndex(mask)
	var routes map[string]*Route
	if slot >= 0 {
		routes = r.staticRoutes[slot]
	} else {
		routes = r.routes[method]
	}
	staticLengthPossible := true
	if routes != nil && r.dynamicRouteCount != 0 && slot >= 0 {
		staticLengthPossible = r.hasStaticRouteLength(slot, mask, len(originalPath))
		if !staticLengthPossible && path != originalPath {
			staticLengthPossible = r.hasStaticRouteLength(slot, mask, len(path))
		}
	}
	if routes != nil && staticLengthPossible {
		if route := lookupStaticRouteExact(routes, originalPath, path); route != nil {
			ctx.setRouteIndex(route.infoIndex)
			return true, allowedMethodSet{}, route.handler(ctx)
		}
	}
	dispatchCacheEnabled := r.dispatchCacheEnabled()
	var key routeCacheKey
	if dispatchCacheEnabled {
		key = routeCacheKey{method: method, path: originalPath}
	}
	if dispatchCacheEnabled && routes != nil {
		if entry, ok := r.cache.getHot(key); ok && entry.route == nil {
			return false, entry.allowed, nil
		}
	}
	if dispatchCacheEnabled {
		if entry, ok := r.cache.getWithMask(key, mask); ok {
			if entry.route == nil {
				return false, entry.allowed, nil
			}
			ctx.applyRouteParams(originalPath, entry.route, entry.values)
			ctx.setRouteIndex(entry.route.infoIndex)
			return true, allowedMethodSet{}, entry.route.handler(ctx)
		}
	}
	captured := ctx.paramRangesScratch()
	entry := r.lookupDynamicDispatch(method, mask, path, needAllowed, captured)
	if !caseSensitive && entry.route == nil {
		if routes := r.staticRoutesFor(method, mask); routes != nil {
			if route := lookupStaticRouteLower(routes, originalPath, path); route != nil {
				ctx.setRouteIndex(route.infoIndex)
				return true, allowedMethodSet{}, route.handler(ctx)
			}
		}
		if lower, changed := lowercasePath(path); changed {
			lowerEntry := r.lookupDynamicDispatch(method, mask, lower, needAllowed, captured)
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
		r.cache.setMissWithMask(routeCacheKey{method: method, path: originalPath}, mask, entry)
	}
	if entry.route == nil {
		return false, entry.allowed, nil
	}
	ctx.applyRouteParams(originalPath, entry.route, entry.values)
	ctx.setRouteIndex(entry.route.infoIndex)
	return true, allowedMethodSet{}, entry.route.handler(ctx)
}

func (r *Router) findDynamicInto(method, path string, ctx *Context) HandlerFunc {
	mask := methodMaskFor(method)
	cacheEnabled := r.dynamicCacheEnabled()
	if cacheEnabled {
		key := routeCacheKey{method: method, path: path}
		if entry, ok := r.cache.getWithMask(key, mask); ok {
			if entry.route == nil {
				return nil
			}
			ctx.applyRouteParams(path, entry.route, entry.values)
			ctx.setRoute(r.routeMetaAt(entry.route.infoIndex))
			return entry.route.handler
		}
	}
	root := r.dynamicTree(method, mask)
	if root == nil {
		return nil
	}
	captured := ctx.paramRangesScratch()
	if captured == nil {
		captured = &paramRanges{}
	}
	matched := lookupDynamicRoute(root, path, captured)
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
		r.cache.setMissWithMask(key, mask, entry)
	}
	return matched.handler
}

func (r *Router) dynamicCacheEnabled() bool {
	return r.cache != nil && r.dynamicRouteCount >= routeCacheMinRoutes
}

func (r *Router) dispatchCacheEnabled() bool {
	return r.cache != nil
}

func (r *Router) lookupDynamicDispatch(method string, mask methodMask, path string, needAllowed bool, captured *paramRanges) routeCacheEntry {
	if r.dynamicRouteCount == 0 {
		return routeCacheEntry{}
	}
	if captured == nil {
		captured = &paramRanges{}
	}
	if root := r.dynamicTree(method, mask); root != nil {
		if route := lookupDynamicRoute(root, path, captured); route != nil {
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

func lookupDynamicRoute(root *radixNode, path string, captured *paramRanges) *radixRoute {
	if root == nil {
		return nil
	}
	return root.lookup(path, 0, captured, 0)
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
	if len(r.staticAllowed) != 0 {
		if allowed := lookupStaticAllowed(r.staticAllowed, originalPath, path, caseSensitive); !allowed.empty() {
			return allowed
		}
	}
	return r.lookupStaticAllowedByScan(originalPath, path, caseSensitive)
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

func (r *Router) lookupStaticAllowedByScan(originalPath, path string, caseSensitive bool) allowedMethodSet {
	if len(r.routes) == 0 {
		return allowedMethodSet{}
	}
	var allowed allowedMethodSet
	for slot, method := range routeMethods {
		routes := r.staticRoutes[slot]
		if len(routes) == 0 {
			continue
		}
		if lookupStaticRouteExact(routes, originalPath, path) != nil {
			allowed.mask |= methodMaskFor(method)
			continue
		}
		if !caseSensitive && lookupStaticRouteLower(routes, originalPath, path) != nil {
			allowed.mask |= methodMaskFor(method)
		}
	}
	if !r.hasCustomStatic {
		return allowed
	}
	var custom []string
	for method := range r.routes {
		if methodMaskFor(method) == 0 {
			custom = append(custom, method)
		}
	}
	if len(custom) == 0 {
		return allowed
	}
	sort.Strings(custom)
	for _, method := range custom {
		routes := r.routes[method]
		if len(routes) == 0 {
			continue
		}
		if lookupStaticRouteExact(routes, originalPath, path) != nil {
			allowed.addMethod(method)
			continue
		}
		if !caseSensitive && lookupStaticRouteLower(routes, originalPath, path) != nil {
			allowed.addMethod(method)
		}
	}
	return allowed
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

func staticRouteHasSingleCandidate(path string, strictRouting, caseSensitive bool) bool {
	if !strictRouting && len(path) > 1 && path[len(path)-1] == '/' {
		return false
	}
	if caseSensitive {
		return true
	}
	for i := 0; i < len(path); i++ {
		c := path[i]
		if (c >= 'A' && c <= 'Z') || c >= 0x80 {
			return false
		}
	}
	return true
}

func addAllowedMethod(allowed allowedMethodSet, method string, mask methodMask) allowedMethodSet {
	if mask != 0 {
		allowed.mask |= mask
		return allowed
	}
	allowed.addMethod(method)
	return allowed
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

func rejectLegacyRoutePattern(path string) error {
	if !strings.ContainsAny(path, ":*") {
		return nil
	}

	for start := 0; start < len(path); {
		end := start
		firstParam := -1
		for end < len(path) && path[end] != '/' {
			switch path[end] {
			case paramIdentifier:
				if firstParam < 0 {
					firstParam = end
				}
			case wildcardIdentifier:
				for end < len(path) && path[end] != '/' {
					end++
				}
				return fmt.Errorf("legacy route wildcard %q in path %q: use {name...} syntax", path[start:end], path)
			}
			end++
		}
		if firstParam >= 0 && firstParam+1 < end {
			return fmt.Errorf("legacy route parameter %q in path %q: use {name} syntax", path[start:end], path)
		}
		start = end + 1
	}
	return nil
}

// compileBraceRoutePattern translates Zinc's public Go-style route syntax to
// the compact internal radix syntax. The registered pattern is retained in
// route metadata, so this translation is invisible outside the router.
func compileBraceRoutePattern(path string) (string, bool, collectedRouteParams, error) {
	var names collectedRouteParams
	if !strings.ContainsAny(path, "{}") {
		return path, false, names, nil
	}

	var builder strings.Builder
	builder.Grow(len(path))
	last := 0

	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '}':
			return "", false, names, fmt.Errorf("unmatched closing brace in route path %q", path)
		case '{':
			if i == 0 || path[i-1] != '/' {
				return "", false, names, fmt.Errorf("route parameter must occupy a complete segment in path %q", path)
			}
			endOffset := strings.IndexByte(path[i+1:], '}')
			if endOffset < 0 {
				return "", false, names, fmt.Errorf("unclosed route parameter in path %q", path)
			}
			end := i + 1 + endOffset
			if end+1 < len(path) && path[end+1] != '/' {
				return "", false, names, fmt.Errorf("route parameter must occupy a complete segment in path %q", path)
			}

			rawName := path[i+1 : end]
			catchAll := strings.HasSuffix(rawName, "...")
			name := strings.TrimSuffix(rawName, "...")
			if !validRouteParamName(name) {
				return "", false, names, fmt.Errorf("invalid route parameter %q in path %q", name, path)
			}
			if names.contains(name) {
				return "", false, names, fmt.Errorf("duplicate route parameter %q in path %q", name, path)
			}
			names.add(name)
			if catchAll && end != len(path)-1 {
				return "", false, names, fmt.Errorf("route wildcard %q must be final in path %q", name, path)
			}

			builder.WriteString(path[last:i])
			if catchAll {
				builder.WriteByte(wildcardIdentifier)
			} else {
				builder.WriteByte(paramIdentifier)
			}
			builder.WriteString(name)
			last = end + 1
			i = end
		}
	}
	builder.WriteString(path[last:])
	return builder.String(), true, names, nil
}

func validRouteParamName(name string) bool {
	if name == "" {
		return false
	}
	for index, r := range name {
		if index == 0 && unicode.IsDigit(r) {
			return false
		}
		if r != '_' && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func (c *collectedRouteParams) add(name string) {
	if c.count < len(c.inline) {
		c.inline[c.count] = name
		c.count++
		return
	}
	c.extra = append(c.extra, name)
	c.count++
}

func (c *collectedRouteParams) contains(name string) bool {
	inlineCount := c.count
	if inlineCount > len(c.inline) {
		inlineCount = len(c.inline)
	}
	for i := 0; i < inlineCount; i++ {
		if c.inline[i] == name {
			return true
		}
	}
	for _, existing := range c.extra {
		if existing == name {
			return true
		}
	}
	return false
}

func collectRouteParams(path string, startIndex int, canonicalWildcardName ...bool) (collectedRouteParams, error) {
	var names collectedRouteParams
	useCanonicalWildcardName := len(canonicalWildcardName) > 0 && canonicalWildcardName[0]
	for i := startIndex; i < len(path); i++ {
		if path[i] != paramIdentifier && path[i] != wildcardIdentifier {
			continue
		}
		segment, err := parseDynamicSegment(path[i:])
		if err != nil {
			return names, fmt.Errorf("%w in path %q", err, path)
		}
		switch segment.kind {
		case radixParam, radixRegexp:
			names.add(segment.name)
		case radixCatchAll:
			if useCanonicalWildcardName {
				names.add(segment.name)
			} else {
				names.add("*")
			}
			return names, nil
		}
		i += segment.width - 1
	}
	return names, nil
}

func parseDynamicSegment(path string) (parsedDynamicSegment, error) {
	if path == "" {
		return parsedDynamicSegment{}, fmt.Errorf("invalid route segment")
	}
	switch path[0] {
	case paramIdentifier:
		start := 1
		if start >= len(path) || path[start] == '/' {
			return parsedDynamicSegment{}, fmt.Errorf("invalid parameter")
		}
		end := start
		for end < len(path) && path[end] != '/' && path[end] != '<' {
			if path[end] == paramIdentifier || path[end] == wildcardIdentifier {
				return parsedDynamicSegment{}, fmt.Errorf("invalid parameter")
			}
			end++
		}
		if end == start {
			return parsedDynamicSegment{}, fmt.Errorf("invalid parameter")
		}
		segment := parsedDynamicSegment{
			kind:  radixParam,
			name:  path[start:end],
			width: end,
		}
		if end < len(path) && path[end] == '<' {
			constraintEnd, expr, err := parseParamConstraint(path, end)
			if err != nil {
				return parsedDynamicSegment{}, err
			}
			segment.kind = radixRegexp
			segment.expr = expr
			segment.width = constraintEnd
		}
		return segment, nil
	case wildcardIdentifier:
		start := 1
		if start >= len(path) || path[start] == '/' {
			return parsedDynamicSegment{}, fmt.Errorf("invalid wildcard")
		}
		end := start
		for end < len(path) && path[end] != '/' {
			if path[end] == paramIdentifier || path[end] == wildcardIdentifier {
				return parsedDynamicSegment{}, fmt.Errorf("invalid wildcard")
			}
			end++
		}
		if end != len(path) {
			return parsedDynamicSegment{}, fmt.Errorf("wildcard must be final")
		}
		return parsedDynamicSegment{kind: radixCatchAll, name: path[start:end], width: end}, nil
	default:
		return parsedDynamicSegment{}, fmt.Errorf("invalid route segment")
	}
}

func parseParamConstraint(path string, start int) (int, string, error) {
	if start >= len(path) || path[start] != '<' {
		return 0, "", fmt.Errorf("invalid regex constraint")
	}
	end := start + 1
	for end < len(path) && path[end] != '>' {
		if path[end] == '/' {
			return 0, "", fmt.Errorf("invalid regex constraint")
		}
		end++
	}
	if end >= len(path) || end == start+1 {
		return 0, "", fmt.Errorf("invalid regex constraint")
	}
	return end + 1, path[start+1 : end], nil
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

func (c collectedRouteParams) slice() []string {
	if c.count == 0 {
		return nil
	}
	out := make([]string, 0, c.count)
	inlineCount := c.count
	if inlineCount > len(c.inline) {
		inlineCount = len(c.inline)
	}
	out = append(out, c.inline[:inlineCount]...)
	if len(c.extra) > 0 {
		out = append(out, c.extra...)
	}
	return out
}

func newRadixRoute(handler HandlerFunc, infoIndex uint32, names collectedRouteParams) *radixRoute {
	route := &radixRoute{
		handler:   handler,
		infoIndex: infoIndex,
	}
	count := names.count
	route.paramCount = uint16(count)
	inlineCount := count
	if inlineCount > len(route.inlineParamNames) {
		inlineCount = len(route.inlineParamNames)
	}
	for i := 0; i < inlineCount; i++ {
		route.inlineParamNames[i] = names.inline[i]
	}
	if len(names.extra) != 0 {
		route.extraParamNames = names.extra
	}
	if count >= indexedParamThreshold {
		route.paramIndices = make(map[string]uint8, count)
		for i := 0; i < count; i++ {
			route.paramIndices[route.paramNameAt(i)] = uint8(i)
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
	if r == nil {
		return 0, false
	}
	if len(r.paramIndices) != 0 {
		index, ok := r.paramIndices[name]
		return int(index), ok
	}
	count := int(r.paramCount)
	if count == 0 {
		return 0, false
	}
	if count > 0 && r.inlineParamNames[0] == name {
		return 0, true
	}
	if count > 1 && r.inlineParamNames[1] == name {
		return 1, true
	}
	for i := 2; i < count; i++ {
		if r.extraParamNames[i-len(r.inlineParamNames)] == name {
			return i, true
		}
	}
	return 0, false
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
	return n.addWithCase(path, route, false)
}

func (n *radixNode) addWithCase(path string, route *radixRoute, caseInsensitive bool) error {
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
		segment, err := parseDynamicSegment(remaining)
		if err != nil {
			return fmt.Errorf("%w in path %q", err, path)
		}
		switch segment.kind {
		case radixParam:
			current = current.addParamChild()
		case radixRegexp:
			current, err = current.addRegexpChild(segment.expr, caseInsensitive)
			if err != nil {
				return fmt.Errorf("%w in path %q", err, path)
			}
		case radixCatchAll:
			current = current.addCatchAllChild()
		default:
			return fmt.Errorf("invalid route segment in path %q", path)
		}
		remaining = remaining[segment.width:]
	}
	if !current.trySetRoute(route) {
		return fmt.Errorf("route already registered for %s", path)
	}
	return nil
}

func (n *radixNode) trySetRoute(route *radixRoute) bool {
	if route == nil || n.route != nil {
		return false
	}
	n.route = route
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
			regexpRaw:      child.regexpRaw,
			regexp:         child.regexp,
			route:          child.route,
			indices:        child.indices,
			indexTable:     child.indexTable,
			children:       child.children,
			regexpChildren: child.regexpChildren,
			paramChild:     child.paramChild,
			catchAllChild:  child.catchAllChild,
		}
		child.prefix = child.prefix[:common]
		child.regexpRaw = ""
		child.regexp = nil
		child.route = nil
		child.indices = nil
		child.indexTable = nil
		child.children = nil
		child.regexpChildren = nil
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

func (n *radixNode) addRegexpChild(expr string, caseInsensitive bool) (*radixNode, error) {
	for _, child := range n.regexpChildren {
		if child.regexpRaw == expr {
			return child, nil
		}
	}
	pattern := "^(?:" + expr + ")$"
	if caseInsensitive {
		pattern = "(?i)" + pattern
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex constraint %q: %w", expr, err)
	}
	child := &radixNode{
		kind:      radixRegexp,
		regexpRaw: expr,
		regexp:    compiled,
	}
	n.regexpChildren = append(n.regexpChildren, child)
	return child, nil
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
	case radixRegexp:
		if len(path) == 0 || path[0] == '/' {
			return nil
		}
		end := nextSlash(path)
		if end < 0 {
			end = len(path)
		}
		if n.regexp == nil || !n.regexp.MatchString(path[:end]) {
			return nil
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
	for _, child := range n.regexpChildren {
		if matched := child.lookup(path, offset, values, captured); matched != nil {
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
	case radixRegexp:
		if len(path) == 0 || path[0] == '/' {
			return false
		}
		end := nextSlash(path)
		if end < 0 {
			end = len(path)
		}
		if n.regexp == nil || !n.regexp.MatchString(path[:end]) {
			return false
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
	for _, child := range n.regexpChildren {
		if child.matchesPath(path, captured) {
			return true
		}
	}
	if n.paramChild != nil && n.paramChild.matchesPath(path, captured) {
		return true
	}
	if n.catchAllChild != nil && n.catchAllChild.matchesPath(path, captured) {
		return true
	}
	return false
}

func (n *radixNode) hasPathRoute(captured int) bool {
	return n.route != nil && captured == int(n.route.paramCount)
}

func (n *radixNode) matchRoute(captured int) *radixRoute {
	if n.route == nil || captured != int(n.route.paramCount) {
		return nil
	}
	return n.route
}

func (n *radixNode) hasRoutes() bool {
	return n.route != nil
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
	return rc.getWithMask(key, methodMaskFor(key.method))
}

func (rc *RouteCache) getWithMask(key routeCacheKey, mask methodMask) (routeCacheEntry, bool) {
	if rc == nil {
		return routeCacheEntry{}, false
	}
	if snapshot := rc.snapshot.Load(); snapshot != nil {
		if entry, ok := snapshot.get(key, mask); ok {
			return entry, true
		}
		if atomic.LoadUint32(&rc.count) == atomic.LoadUint32(&rc.frozenCount) {
			return routeCacheEntry{}, false
		}
		rc.mu.RLock()
		entry, ok := rc.getOverlayLocked(key, mask)
		rc.mu.RUnlock()
		if ok {
			rc.recordOverlayHit(snapshot)
		}
		return entry, ok
	}
	rc.ensureFresh()
	if hot := rc.hot.Load(); hot != nil && hot.key == key {
		return hot.entry, true
	}
	if atomic.LoadUint32(&rc.count) == 0 {
		return routeCacheEntry{}, false
	}
	rc.mu.RLock()
	entry, ok := rc.getLocked(key, mask)
	rc.mu.RUnlock()
	if ok {
		rc.recordHit()
	}
	return entry, ok
}

func (rc *RouteCache) getHot(key routeCacheKey) (routeCacheEntry, bool) {
	if rc == nil || atomic.LoadUint32(&rc.dirty) != 0 {
		return routeCacheEntry{}, false
	}
	if hot := rc.hot.Load(); hot != nil && hot.key == key {
		return hot.entry, true
	}
	return routeCacheEntry{}, false
}

func (rc *RouteCache) set(key routeCacheKey, entry routeCacheEntry) {
	rc.setWithMask(key, methodMaskFor(key.method), entry)
}

func (rc *RouteCache) setWithMask(key routeCacheKey, mask methodMask, entry routeCacheEntry) {
	if rc == nil || rc.size <= 0 {
		return
	}
	rc.ensureFresh()
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if snapshot := rc.snapshot.Load(); snapshot != nil {
		if _, exists := snapshot.get(key, mask); exists {
			return
		}
		rc.setOverlayLocked(key, mask, entry)
		return
	}
	if rc.keys == nil {
		rc.keys = make([]routeCacheKey, 0, rc.size)
		rc.next = 0
	}
	if _, exists := rc.getLocked(key, mask); exists {
		rc.setLocked(key, mask, entry)
		if hot := rc.hot.Load(); hot != nil && hot.key == key {
			rc.hot.Store(&routeCacheHotEntry{key: key, entry: entry})
		}
		return
	}
	count := int(atomic.LoadUint32(&rc.count))
	if count < rc.size {
		rc.setLocked(key, mask, entry)
		rc.keys = append(rc.keys, key)
		atomic.StoreUint32(&rc.count, uint32(count+1))
		rc.hot.Store(&routeCacheHotEntry{key: key, entry: entry})
		return
	}

	victim := rc.keys[rc.next]
	rc.deleteLocked(victim, methodMaskFor(victim.method))
	rc.setLocked(key, mask, entry)
	rc.keys[rc.next] = key
	rc.next++
	if rc.next == len(rc.keys) {
		rc.next = 0
	}
	if hot := rc.hot.Load(); hot != nil && hot.key == victim {
		rc.hot.Store(nil)
	}
}

func (rc *RouteCache) setMiss(key routeCacheKey, entry routeCacheEntry) {
	rc.setMissWithMask(key, methodMaskFor(key.method), entry)
}

func (rc *RouteCache) setMissWithMask(key routeCacheKey, mask methodMask, entry routeCacheEntry) {
	if rc == nil || rc.size <= 0 {
		return
	}
	admissionShard := routeCacheAdmissionShard(key)
	if atomic.LoadUint32(&rc.count) >= uint32(rc.size) &&
		atomic.AddUint32(&rc.admit[admissionShard], 1)%routeCacheAdmissionInterval != 0 {
		return
	}
	rc.setWithMask(key, mask, entry)
}

func (rc *RouteCache) recordHit() {
	count := atomic.LoadUint32(&rc.count)
	if count < routeCacheMinRoutes ||
		uint64(count)*routeCacheFreezeCapacityDenominator >
			uint64(rc.size)*routeCacheFreezeCapacityNumerator {
		return
	}
	hits := atomic.AddUint32(&rc.hits, 1)
	if uint64(hits) < uint64(count)*routeCacheFreezeHitCycles {
		return
	}
	rc.freezeReadSnapshot(count)
}

func (rc *RouteCache) recordOverlayHit(expectedSnapshot *routeCacheReadSnapshot) {
	if rc.snapshot.Load() != expectedSnapshot {
		return
	}
	count := atomic.LoadUint32(&rc.count)
	frozenCount := atomic.LoadUint32(&rc.frozenCount)
	if count <= frozenCount {
		return
	}
	overlayCount := count - frozenCount
	if overlayCount < routeCacheMinRoutes ||
		uint64(overlayCount)*routeCacheFreezeCapacityDenominator >
			uint64(rc.size)*routeCacheFreezeCapacityNumerator ||
		uint64(overlayCount)*routeCacheRebaseSizeDenominator <
			uint64(frozenCount)*routeCacheRebaseSizeNumerator {
		return
	}
	hits := atomic.AddUint32(&rc.overlayHits, 1)
	if uint64(hits) < uint64(overlayCount)*routeCacheFreezeHitCycles {
		return
	}
	rc.rebaseReadSnapshot(expectedSnapshot, overlayCount, frozenCount)
}

func (rc *RouteCache) freezeReadSnapshot(expectedCount uint32) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.snapshot.Load() != nil ||
		atomic.LoadUint32(&rc.dirty) != 0 ||
		atomic.LoadUint32(&rc.count) != expectedCount ||
		uint64(atomic.LoadUint32(&rc.hits)) < uint64(expectedCount)*routeCacheFreezeHitCycles {
		return
	}
	snapshot := &routeCacheReadSnapshot{
		methodCache: rc.methodCache,
		extraCache:  rc.extraCache,
	}
	rc.keys = nil
	rc.next = 0
	atomic.StoreUint32(&rc.frozenCount, expectedCount)
	rc.snapshot.Store(snapshot)
	rc.hot.Store(nil)
}

func (rc *RouteCache) rebaseReadSnapshot(
	expectedSnapshot *routeCacheReadSnapshot,
	expectedOverlayCount,
	expectedFrozenCount uint32,
) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	if rc.snapshot.Load() != expectedSnapshot ||
		atomic.LoadUint32(&rc.dirty) != 0 ||
		atomic.LoadUint32(&rc.frozenCount) != expectedFrozenCount ||
		atomic.LoadUint32(&rc.count) != expectedFrozenCount+expectedOverlayCount ||
		uint64(atomic.LoadUint32(&rc.overlayHits)) <
			uint64(expectedOverlayCount)*routeCacheFreezeHitCycles {
		return
	}

	rc.methodCache = rc.overlay
	rc.extraCache = rc.extraOverlay
	rc.overlay = [routeMethodCount]map[string]routeCacheEntry{}
	rc.extraOverlay = nil
	rc.keys = nil
	rc.next = 0
	atomic.StoreUint32(&rc.count, expectedOverlayCount)
	atomic.StoreUint32(&rc.hits, 0)
	atomic.StoreUint32(&rc.frozenCount, expectedOverlayCount)
	atomic.StoreUint32(&rc.overlayHits, 0)
	rc.snapshot.Store(&routeCacheReadSnapshot{
		methodCache: rc.methodCache,
		extraCache:  rc.extraCache,
	})
	rc.hot.Store(nil)
}

func (rc *RouteCache) getLocked(key routeCacheKey, mask methodMask) (routeCacheEntry, bool) {
	if slot := singleBitIndex(mask); slot >= 0 {
		entry, ok := rc.methodCache[slot][key.path]
		return entry, ok
	}
	entry, ok := rc.extraCache[key]
	return entry, ok
}

func (snapshot *routeCacheReadSnapshot) get(key routeCacheKey, mask methodMask) (routeCacheEntry, bool) {
	if slot := singleBitIndex(mask); slot >= 0 {
		entry, ok := snapshot.methodCache[slot][key.path]
		return entry, ok
	}
	entry, ok := snapshot.extraCache[key]
	return entry, ok
}

func (rc *RouteCache) getOverlayLocked(key routeCacheKey, mask methodMask) (routeCacheEntry, bool) {
	if slot := singleBitIndex(mask); slot >= 0 {
		entry, ok := rc.overlay[slot][key.path]
		return entry, ok
	}
	entry, ok := rc.extraOverlay[key]
	return entry, ok
}

func (rc *RouteCache) setLocked(key routeCacheKey, mask methodMask, entry routeCacheEntry) {
	if slot := singleBitIndex(mask); slot >= 0 {
		cache := rc.methodCache[slot]
		if cache == nil {
			cache = make(map[string]routeCacheEntry)
			rc.methodCache[slot] = cache
		}
		cache[key.path] = entry
		return
	}
	if rc.extraCache == nil {
		rc.extraCache = make(map[routeCacheKey]routeCacheEntry)
	}
	rc.extraCache[key] = entry
}

func (rc *RouteCache) setOverlayLocked(key routeCacheKey, mask methodMask, entry routeCacheEntry) {
	if _, exists := rc.getOverlayLocked(key, mask); exists {
		rc.storeOverlayLocked(key, mask, entry)
		return
	}
	count := int(atomic.LoadUint32(&rc.count))
	if count < rc.size {
		rc.storeOverlayLocked(key, mask, entry)
		rc.keys = append(rc.keys, key)
		atomic.StoreUint32(&rc.count, uint32(count+1))
		atomic.StoreUint32(&rc.overlayHits, 0)
		rc.hot.Store(&routeCacheHotEntry{key: key, entry: entry})
		return
	}
	if len(rc.keys) == 0 {
		return
	}
	victim := rc.keys[rc.next]
	rc.deleteOverlayLocked(victim, methodMaskFor(victim.method))
	rc.storeOverlayLocked(key, mask, entry)
	rc.keys[rc.next] = key
	atomic.StoreUint32(&rc.overlayHits, 0)
	rc.next++
	if rc.next == len(rc.keys) {
		rc.next = 0
	}
	if hot := rc.hot.Load(); hot != nil && hot.key == victim {
		rc.hot.Store(nil)
	}
}

func (rc *RouteCache) storeOverlayLocked(key routeCacheKey, mask methodMask, entry routeCacheEntry) {
	if slot := singleBitIndex(mask); slot >= 0 {
		cache := rc.overlay[slot]
		if cache == nil {
			cache = make(map[string]routeCacheEntry)
			rc.overlay[slot] = cache
		}
		cache[key.path] = entry
		return
	}
	if rc.extraOverlay == nil {
		rc.extraOverlay = make(map[routeCacheKey]routeCacheEntry)
	}
	rc.extraOverlay[key] = entry
}

func (rc *RouteCache) deleteLocked(key routeCacheKey, mask methodMask) {
	if slot := singleBitIndex(mask); slot >= 0 {
		delete(rc.methodCache[slot], key.path)
		return
	}
	delete(rc.extraCache, key)
}

func (rc *RouteCache) deleteOverlayLocked(key routeCacheKey, mask methodMask) {
	if slot := singleBitIndex(mask); slot >= 0 {
		delete(rc.overlay[slot], key.path)
		return
	}
	delete(rc.extraOverlay, key)
}

func routeCacheAdmissionShard(key routeCacheKey) int {
	path := key.path
	hash := uint32(len(path))*16777619 ^ uint32(len(key.method))
	if len(path) != 0 {
		hash = (hash ^ uint32(path[0])) * 16777619
		hash = (hash ^ uint32(path[len(path)/2])) * 16777619
		hash = (hash ^ uint32(path[len(path)-1])) * 16777619
	}
	if len(key.method) != 0 {
		hash = (hash ^ uint32(key.method[0])) * 16777619
	}
	return int(hash & (routeCacheAdmissionShards - 1))
}

func (rc *RouteCache) invalidate() {
	if rc == nil {
		return
	}
	atomic.StoreUint32(&rc.dirty, 1)
	rc.snapshot.Store(nil)
	rc.hot.Store(nil)
}

func (rc *RouteCache) ensureFresh() {
	if rc == nil || atomic.LoadUint32(&rc.dirty) == 0 {
		return
	}
	rc.mu.Lock()
	if atomic.LoadUint32(&rc.dirty) != 0 {
		rc.methodCache = [routeMethodCount]map[string]routeCacheEntry{}
		rc.extraCache = nil
		rc.overlay = [routeMethodCount]map[string]routeCacheEntry{}
		rc.extraOverlay = nil
		rc.keys = nil
		rc.next = 0
		atomic.StoreUint32(&rc.count, 0)
		atomic.StoreUint32(&rc.hits, 0)
		atomic.StoreUint32(&rc.frozenCount, 0)
		atomic.StoreUint32(&rc.overlayHits, 0)
		for i := range rc.admit {
			atomic.StoreUint32(&rc.admit[i], 0)
		}
		atomic.StoreUint32(&rc.dirty, 0)
		rc.snapshot.Store(nil)
		rc.hot.Store(nil)
	}
	rc.mu.Unlock()
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

func (a *App) Add(method, path string, handlers ...HandlerFunc) {
	mustRegister(a.router.Add(method, path, handlers...))
}

func (a *App) Get(path string, handlers ...HandlerFunc) {
	a.Add(MethodGet, path, handlers...)
}
func (a *App) Post(path string, handlers ...HandlerFunc) {
	a.Add(MethodPost, path, handlers...)
}
func (a *App) Put(path string, handlers ...HandlerFunc) {
	a.Add(MethodPut, path, handlers...)
}
func (a *App) Delete(path string, handlers ...HandlerFunc) {
	a.Add(MethodDelete, path, handlers...)
}
func (a *App) Patch(path string, handlers ...HandlerFunc) {
	a.Add(MethodPatch, path, handlers...)
}
func (a *App) Head(path string, handlers ...HandlerFunc) {
	a.Add(MethodHead, path, handlers...)
}
func (a *App) Options(path string, handlers ...HandlerFunc) {
	a.Add(MethodOptions, path, handlers...)
}
func (a *App) Connect(path string, handlers ...HandlerFunc) {
	a.Add(MethodConnect, path, handlers...)
}
func (a *App) Trace(path string, handlers ...HandlerFunc) {
	a.Add(MethodTrace, path, handlers...)
}

func (a *App) Match(methods []string, path string, handlers ...HandlerFunc) {
	for _, method := range methods {
		a.Add(method, path, handlers...)
	}
}

func (a *App) All(path string, handlers ...HandlerFunc) {
	a.Match(routeMethods, path, handlers...)
}

func (a *App) Any(path string, handlers ...HandlerFunc) {
	a.All(path, handlers...)
}
