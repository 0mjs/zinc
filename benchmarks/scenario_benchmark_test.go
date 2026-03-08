package benchmarks

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	. "github.com/0mjs/zinc"
	"github.com/gin-gonic/gin"
	"github.com/go-chi/chi/v5"
	"github.com/julienschmidt/httprouter"
	"github.com/labstack/echo/v5"
)

var benchmarkSinkInt int

type scenarioRouteSpec struct {
	method  string
	pattern string
}

type scenarioRoute struct {
	method      string
	pattern     string
	requestPath string
	paramNames  []string
}

type benchmarkScenario struct {
	name           string
	routes         []scenarioRoute
	allRequests    []*http.Request
	staticRequest  *http.Request
	paramRequest   *http.Request
	withServeMux   bool
	withHTTPRouter bool
}

var scenarioBenchmarks = []benchmarkScenario{
	scenarioWithRouters(newBenchmarkScenario("Static157", staticRouteScenarioSpecs(), 157), true, true),
	scenarioWithRouters(newBenchmarkScenario("GitHubAPI203", gitHubRouteScenarioSpecs(), 203), true, false),
	scenarioWithRouters(newBenchmarkScenario("GPlusAPI13", gplusRouteScenarioSpecs(), 13), true, false),
	scenarioWithRouters(newBenchmarkScenario("ParseAPI26", parseRouteScenarioSpecs(), 26), true, true),
}

func scenarioWithRouters(scenario benchmarkScenario, withServeMux, withHTTPRouter bool) benchmarkScenario {
	scenario.withServeMux = withServeMux
	scenario.withHTTPRouter = withHTTPRouter
	return scenario
}

func newBenchmarkScenario(name string, specs []scenarioRouteSpec, expected int) benchmarkScenario {
	if len(specs) != expected {
		panic(fmt.Sprintf("benchmark scenario %s expected %d routes, got %d", name, expected, len(specs)))
	}

	routes := make([]scenarioRoute, len(specs))
	seen := make(map[string]struct{}, len(specs))
	for i, spec := range specs {
		key := spec.method + " " + spec.pattern
		if _, exists := seen[key]; exists {
			panic(fmt.Sprintf("duplicate scenario route %s in %s", key, name))
		}
		seen[key] = struct{}{}

		routes[i] = scenarioRoute{
			method:      spec.method,
			pattern:     spec.pattern,
			requestPath: materializeScenarioPath(spec.pattern),
			paramNames:  scenarioParamNames(spec.pattern),
		}
	}

	var (
		staticRoute *scenarioRoute
		paramRoute  *scenarioRoute
	)
	allRequests := make([]*http.Request, len(routes))
	for i := range routes {
		route := &routes[i]
		allRequests[i] = httptest.NewRequest(route.method, route.requestPath, nil)

		if len(route.paramNames) == 0 {
			if staticRoute == nil || len(route.pattern) > len(staticRoute.pattern) {
				staticRoute = route
			}
			continue
		}

		if paramRoute == nil ||
			len(route.paramNames) > len(paramRoute.paramNames) ||
			(len(route.paramNames) == len(paramRoute.paramNames) && len(route.pattern) > len(paramRoute.pattern)) {
			paramRoute = route
		}
	}

	scenario := benchmarkScenario{
		name:        name,
		routes:      routes,
		allRequests: allRequests,
	}
	if staticRoute != nil {
		scenario.staticRequest = httptest.NewRequest(staticRoute.method, staticRoute.requestPath, nil)
	}
	if paramRoute != nil {
		scenario.paramRequest = httptest.NewRequest(paramRoute.method, paramRoute.requestPath, nil)
	}
	return scenario
}

func scenarioParamNames(pattern string) []string {
	if !strings.Contains(pattern, ":") {
		return nil
	}

	segments := strings.Split(pattern, "/")
	names := make([]string, 0, 4)
	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			names = append(names, segment[1:])
		}
	}
	return names
}

func materializeScenarioPath(pattern string) string {
	if !strings.Contains(pattern, ":") {
		return pattern
	}

	segments := strings.Split(pattern, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = scenarioSampleValue(segment[1:])
		}
	}
	return strings.Join(segments, "/")
}

func scenarioBracePattern(pattern string) string {
	if !strings.Contains(pattern, ":") {
		return pattern
	}

	segments := strings.Split(pattern, "/")
	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			segments[i] = "{" + segment[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}

func scenarioSampleValue(name string) string {
	switch name {
	case "activityId":
		return "activity-42"
	case "assignee":
		return "alice"
	case "base":
		return "main"
	case "branch":
		return "main"
	case "className":
		return "Todo"
	case "collection":
		return "public"
	case "commentId":
		return "101"
	case "communityId":
		return "202"
	case "deliveryId":
		return "303"
	case "downloadId":
		return "404"
	case "entry":
		return "readme"
	case "eventId":
		return "505"
	case "eventName":
		return "signup"
	case "fileName":
		return "avatar.png"
	case "functionName":
		return "notify"
	case "gistId":
		return "aa11bb22"
	case "head":
		return "feature"
	case "hookId":
		return "606"
	case "installationId":
		return "707"
	case "invitationId":
		return "808"
	case "key":
		return "mit"
	case "keyId":
		return "909"
	case "milestone":
		return "4"
	case "number":
		return "42"
	case "objectId":
		return "obj123"
	case "org":
		return "open-source"
	case "owner":
		return "matt"
	case "projectId":
		return "1001"
	case "ref":
		return "heads-main"
	case "releaseId":
		return "1002"
	case "repo":
		return "zinc"
	case "sha":
		return "deadbeef"
	case "target":
		return "bob"
	case "teamId":
		return "1003"
	case "teamSlug":
		return "platform"
	case "user":
		return "alice"
	case "userId":
		return "me"
	default:
		return name + "-1"
	}
}

func scenarioParamScoreZinc(names []string, c *Context) int {
	score := 0
	for _, name := range names {
		score += len(c.Param(name))
	}
	benchmarkSinkInt = score
	return score
}

func scenarioParamScoreServeMux(names []string, r *http.Request) int {
	score := 0
	for _, name := range names {
		score += len(r.PathValue(name))
	}
	benchmarkSinkInt = score
	return score
}

func scenarioParamScoreChi(names []string, req *http.Request) int {
	score := 0
	for _, name := range names {
		score += len(chi.URLParam(req, name))
	}
	benchmarkSinkInt = score
	return score
}

func scenarioParamScoreEcho(names []string, c *echo.Context) int {
	score := 0
	for _, name := range names {
		score += len(c.Param(name))
	}
	benchmarkSinkInt = score
	return score
}

func scenarioParamScoreGin(names []string, c *gin.Context) int {
	score := 0
	for _, name := range names {
		score += len(c.Param(name))
	}
	benchmarkSinkInt = score
	return score
}

func scenarioParamScoreHTTPRouter(names []string, ps httprouter.Params) int {
	score := 0
	for _, name := range names {
		score += len(ps.ByName(name))
	}
	benchmarkSinkInt = score
	return score
}

func buildZincScenarioHandler(routes []scenarioRoute) http.Handler {
	app := New()
	for _, route := range routes {
		route := route
		mustNoErr(app.Add(route.method, route.pattern, func(c *Context) error {
			scenarioParamScoreZinc(route.paramNames, c)
			return c.String(benchmarkOKResponse)
		}))
	}
	return app
}

func buildServeMuxScenarioHandler(routes []scenarioRoute) http.Handler {
	mux := http.NewServeMux()
	for _, route := range routes {
		route := route
		mux.HandleFunc(route.method+" "+scenarioBracePattern(route.pattern), func(w http.ResponseWriter, r *http.Request) {
			scenarioParamScoreServeMux(route.paramNames, r)
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return mux
}

func buildChiScenarioHandler(routes []scenarioRoute) http.Handler {
	r := chi.NewRouter()
	for _, route := range routes {
		route := route
		r.MethodFunc(route.method, scenarioBracePattern(route.pattern), func(w http.ResponseWriter, req *http.Request) {
			scenarioParamScoreChi(route.paramNames, req)
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func buildEchoScenarioHandler(routes []scenarioRoute) http.Handler {
	e := echo.New()
	for _, route := range routes {
		route := route
		e.Add(route.method, route.pattern, func(c *echo.Context) error {
			scenarioParamScoreEcho(route.paramNames, c)
			return c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return e
}

func buildGinScenarioHandler(routes []scenarioRoute) http.Handler {
	r := gin.New()
	for _, route := range routes {
		route := route
		r.Handle(route.method, route.pattern, func(c *gin.Context) {
			scenarioParamScoreGin(route.paramNames, c)
			c.String(http.StatusOK, benchmarkOKResponse)
		})
	}
	return r
}

func buildHTTPRouterScenarioHandler(routes []scenarioRoute) http.Handler {
	r := httprouter.New()
	for _, route := range routes {
		route := route
		r.Handle(route.method, route.pattern, func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
			scenarioParamScoreHTTPRouter(route.paramNames, ps)
			_, _ = io.WriteString(w, benchmarkOKResponse)
		})
	}
	return r
}

func scenarioCases(scenario benchmarkScenario) []benchmarkCase {
	cases := []benchmarkCase{
		{name: "Zinc", build: func() http.Handler { return buildZincScenarioHandler(scenario.routes) }},
	}
	if scenario.withServeMux {
		cases = append(cases, benchmarkCase{name: "ServeMux", build: func() http.Handler { return buildServeMuxScenarioHandler(scenario.routes) }})
	}
	if scenario.withHTTPRouter {
		cases = append(cases, benchmarkCase{name: "HttpRouter", build: func() http.Handler { return buildHTTPRouterScenarioHandler(scenario.routes) }})
	}
	cases = append(cases,
		benchmarkCase{name: "Chi", build: func() http.Handler { return buildChiScenarioHandler(scenario.routes) }},
		benchmarkCase{name: "Echo", build: func() http.Handler { return buildEchoScenarioHandler(scenario.routes) }},
		benchmarkCase{name: "Gin", build: func() http.Handler { return buildGinScenarioHandler(scenario.routes) }},
	)
	return cases
}

func runScenarioSingleRequestBenchmark(b *testing.B, request *http.Request, cases []benchmarkCase) {
	runServeHTTPRequestSetBenchmarks(b, cases, []*http.Request{request})
}

func BenchmarkScenarioRouteSetBuild(b *testing.B) {
	for _, scenario := range scenarioBenchmarks {
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			runRegistrationBenchmarks(b, scenarioCases(scenario))
		})
	}
}

func BenchmarkScenarioRouteSetStatic(b *testing.B) {
	for _, scenario := range scenarioBenchmarks {
		if scenario.staticRequest == nil {
			continue
		}
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			runScenarioSingleRequestBenchmark(b, scenario.staticRequest, scenarioCases(scenario))
		})
	}
}

func BenchmarkScenarioRouteSetParam(b *testing.B) {
	for _, scenario := range scenarioBenchmarks {
		if scenario.paramRequest == nil {
			continue
		}
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			runScenarioSingleRequestBenchmark(b, scenario.paramRequest, scenarioCases(scenario))
		})
	}
}

func BenchmarkScenarioRouteSetAll(b *testing.B) {
	for _, scenario := range scenarioBenchmarks {
		scenario := scenario
		b.Run(scenario.name, func(b *testing.B) {
			runServeHTTPRequestSetBenchmarks(b, scenarioCases(scenario), scenario.allRequests)
		})
	}
}

func staticRouteScenarioSpecs() []scenarioRouteSpec {
	routes := make([]scenarioRouteSpec, 0, 157)
	add := func(pattern string) {
		routes = append(routes, scenarioRouteSpec{method: http.MethodGet, pattern: pattern})
	}

	for _, pattern := range []string{
		"/",
		"/about",
		"/pricing",
		"/contact",
		"/careers",
		"/blog",
		"/docs",
		"/docs/faq",
		"/changelog",
		"/security",
		"/status",
		"/legal/terms",
		"/legal/privacy",
		"/customers",
		"/partners",
		"/integrations",
		"/download",
	} {
		add(pattern)
	}

	docsSections := []string{"getting-started", "guides", "reference", "cookbook", "tutorials"}
	docsPages := []string{"install", "routing", "middleware", "binding", "responses", "testing", "templates", "deploy", "security", "performance"}
	for _, section := range docsSections {
		for _, page := range docsPages {
			add("/docs/" + section + "/" + page)
		}
	}

	refSections := []string{"app", "context", "router", "middleware"}
	refPages := []string{"overview", "config", "routes", "params", "json", "errors", "testing", "benchmarks"}
	for _, section := range refSections {
		for _, page := range refPages {
			add("/reference/" + section + "/" + page)
		}
	}

	for year := 2022; year <= 2024; year++ {
		for month := 1; month <= 6; month++ {
			add("/blog/" + strconv.Itoa(year) + "/" + fmt.Sprintf("%02d", month) + "/release-notes")
		}
	}

	assetKinds := []string{"css", "js", "img", "fonts"}
	for _, kind := range assetKinds {
		for i := 1; i <= 5; i++ {
			add("/assets/" + kind + "/bundle-" + strconv.Itoa(i))
		}
	}

	solutionKinds := []string{"saas", "fintech", "commerce", "internal-tools"}
	solutionPages := []string{"overview", "architecture", "security", "observability", "launch"}
	for _, kind := range solutionKinds {
		for _, page := range solutionPages {
			add("/solutions/" + kind + "/" + page)
		}
	}

	return routes
}

func gplusRouteScenarioSpecs() []scenarioRouteSpec {
	return []scenarioRouteSpec{
		{method: http.MethodGet, pattern: "/plus/v1/people/me"},
		{method: http.MethodGet, pattern: "/plus/v1/people/:userId"},
		{method: http.MethodGet, pattern: "/plus/v1/people/:userId/activities/public"},
		{method: http.MethodGet, pattern: "/plus/v1/people/:userId/activities/:collection"},
		{method: http.MethodGet, pattern: "/plus/v1/people/:userId/people/:collection"},
		{method: http.MethodGet, pattern: "/plus/v1/people/:userId/moments/vault"},
		{method: http.MethodPost, pattern: "/plus/v1/people/:userId/moments/vault"},
		{method: http.MethodGet, pattern: "/plus/v1/activities/:activityId"},
		{method: http.MethodGet, pattern: "/plus/v1/activities/:activityId/people/:collection"},
		{method: http.MethodGet, pattern: "/plus/v1/comments/:commentId"},
		{method: http.MethodGet, pattern: "/plus/v1/communities/:communityId"},
		{method: http.MethodGet, pattern: "/plus/v1/communities/:communityId/members"},
		{method: http.MethodGet, pattern: "/plus/v1/communities/:communityId/activities"},
	}
}

func parseRouteScenarioSpecs() []scenarioRouteSpec {
	return []scenarioRouteSpec{
		{method: http.MethodPost, pattern: "/1/classes/:className"},
		{method: http.MethodGet, pattern: "/1/classes/:className"},
		{method: http.MethodGet, pattern: "/1/classes/:className/:objectId"},
		{method: http.MethodPut, pattern: "/1/classes/:className/:objectId"},
		{method: http.MethodDelete, pattern: "/1/classes/:className/:objectId"},
		{method: http.MethodPost, pattern: "/1/users"},
		{method: http.MethodGet, pattern: "/1/users/:objectId"},
		{method: http.MethodPut, pattern: "/1/users/:objectId"},
		{method: http.MethodGet, pattern: "/1/login"},
		{method: http.MethodPost, pattern: "/1/requestPasswordReset"},
		{method: http.MethodPost, pattern: "/1/events/:eventName"},
		{method: http.MethodPost, pattern: "/1/functions/:functionName"},
		{method: http.MethodGet, pattern: "/1/functions/:functionName"},
		{method: http.MethodGet, pattern: "/1/schemas"},
		{method: http.MethodGet, pattern: "/1/schemas/:className"},
		{method: http.MethodPut, pattern: "/1/schemas/:className"},
		{method: http.MethodDelete, pattern: "/1/schemas/:className"},
		{method: http.MethodPost, pattern: "/1/installations"},
		{method: http.MethodGet, pattern: "/1/installations/:objectId"},
		{method: http.MethodPut, pattern: "/1/installations/:objectId"},
		{method: http.MethodPost, pattern: "/1/files/:fileName"},
		{method: http.MethodGet, pattern: "/1/health"},
		{method: http.MethodGet, pattern: "/1/config"},
		{method: http.MethodPut, pattern: "/1/config"},
		{method: http.MethodGet, pattern: "/1/roles/:objectId"},
		{method: http.MethodPost, pattern: "/1/batch"},
	}
}

func gitHubRouteScenarioSpecs() []scenarioRouteSpec {
	routes := make([]scenarioRouteSpec, 0, 203)
	add := func(method, pattern string) {
		routes = append(routes, scenarioRouteSpec{method: method, pattern: pattern})
	}
	addAll := func(method string, patterns ...string) {
		for _, pattern := range patterns {
			add(method, pattern)
		}
	}

	addAll(http.MethodGet,
		"/gh/user",
		"/gh/user/emails",
		"/gh/user/followers",
		"/gh/user/following",
		"/gh/user/keys",
		"/gh/user/repos",
		"/gh/user/starred",
		"/gh/user/starred/:owner/:repo",
		"/gh/user/subscriptions",
		"/gh/user/orgs",
		"/gh/notifications",
		"/gh/gists",
		"/gh/emojis",
		"/gh/events",
		"/gh/feeds",
		"/gh/meta",
		"/gh/rate_limit",
		"/gh/search/code",
		"/gh/search/issues",
		"/gh/search/repositories",
		"/gh/search/users",
		"/gh/licenses",
	)
	addAll(http.MethodPatch, "/gh/user")
	addAll(http.MethodPost, "/gh/gists", "/gh/markdown", "/gh/markdown/raw")
	addAll(http.MethodPost, "/gh/user/emails", "/gh/user/keys")
	addAll(http.MethodDelete, "/gh/user/emails")
	addAll(http.MethodPut, "/gh/user/starred/:owner/:repo")
	addAll(http.MethodDelete, "/gh/user/starred/:owner/:repo")
	addAll(http.MethodGet, "/gh/licenses/:key")

	addAll(http.MethodGet,
		"/gh/users/:user",
		"/gh/users/:user/followers",
		"/gh/users/:user/following",
		"/gh/users/:user/following/:target",
		"/gh/users/:user/gists",
		"/gh/users/:user/keys",
		"/gh/users/:user/orgs",
		"/gh/users/:user/received_events",
		"/gh/users/:user/received_events/public",
		"/gh/users/:user/repos",
		"/gh/users/:user/starred",
		"/gh/users/:user/starred/:owner/:repo",
		"/gh/users/:user/subscriptions",
		"/gh/users/:user/subscriptions/:owner/:repo",
		"/gh/users/:user/events",
		"/gh/users/:user/events/public",
		"/gh/users/:user/events/orgs/:org",
		"/gh/users/:user/hovercard",
		"/gh/users/:user/installations",
		"/gh/users/:user/packages",
		"/gh/users/:user/projects",
	)

	addAll(http.MethodGet,
		"/gh/orgs/:org",
		"/gh/orgs/:org/members",
		"/gh/orgs/:org/members/:user",
		"/gh/orgs/:org/public_members",
		"/gh/orgs/:org/public_members/:user",
		"/gh/orgs/:org/repos",
		"/gh/orgs/:org/issues",
		"/gh/orgs/:org/hooks",
		"/gh/orgs/:org/hooks/:hookId",
		"/gh/orgs/:org/teams",
		"/gh/orgs/:org/teams/:teamSlug",
		"/gh/orgs/:org/invitations",
		"/gh/orgs/:org/installations",
		"/gh/orgs/:org/packages",
		"/gh/orgs/:org/actions/secrets",
	)
	addAll(http.MethodPatch, "/gh/orgs/:org", "/gh/orgs/:org/hooks/:hookId")
	addAll(http.MethodPost, "/gh/orgs/:org/hooks")
	addAll(http.MethodDelete, "/gh/orgs/:org/members/:user", "/gh/orgs/:org/hooks/:hookId", "/gh/orgs/:org/public_members/:user")
	addAll(http.MethodPut, "/gh/orgs/:org/public_members/:user")

	addAll(http.MethodGet,
		"/gh/repos/:owner/:repo",
		"/gh/repos/:owner/:repo/branches",
		"/gh/repos/:owner/:repo/branches/:branch",
		"/gh/repos/:owner/:repo/tags",
		"/gh/repos/:owner/:repo/contributors",
		"/gh/repos/:owner/:repo/stargazers",
		"/gh/repos/:owner/:repo/subscribers",
		"/gh/repos/:owner/:repo/subscription",
		"/gh/repos/:owner/:repo/languages",
		"/gh/repos/:owner/:repo/teams",
		"/gh/repos/:owner/:repo/topics",
		"/gh/repos/:owner/:repo/readme",
		"/gh/repos/:owner/:repo/license",
		"/gh/repos/:owner/:repo/releases",
		"/gh/repos/:owner/:repo/releases/latest",
		"/gh/repos/:owner/:repo/releases/:releaseId",
		"/gh/repos/:owner/:repo/deployments",
		"/gh/repos/:owner/:repo/environments",
		"/gh/repos/:owner/:repo/actions/runs",
		"/gh/repos/:owner/:repo/pages",
		"/gh/repos/:owner/:repo/contents/:entry",
		"/gh/repos/:owner/:repo/commits",
		"/gh/repos/:owner/:repo/commits/:sha",
		"/gh/repos/:owner/:repo/commits/:sha/status",
		"/gh/repos/:owner/:repo/commits/:sha/check-runs",
		"/gh/repos/:owner/:repo/comments",
		"/gh/repos/:owner/:repo/comments/:commentId",
		"/gh/repos/:owner/:repo/git/blobs/:sha",
		"/gh/repos/:owner/:repo/git/trees/:sha",
		"/gh/repos/:owner/:repo/git/commits/:sha",
		"/gh/repos/:owner/:repo/git/refs",
		"/gh/repos/:owner/:repo/git/ref/:ref",
		"/gh/repos/:owner/:repo/compare/:base/:head",
		"/gh/repos/:owner/:repo/downloads",
		"/gh/repos/:owner/:repo/downloads/:downloadId",
		"/gh/repos/:owner/:repo/stats/contributors",
		"/gh/repos/:owner/:repo/stats/commit_activity",
		"/gh/repos/:owner/:repo/issues",
		"/gh/repos/:owner/:repo/issues/:number",
		"/gh/repos/:owner/:repo/issues/:number/comments",
		"/gh/repos/:owner/:repo/issue-comments/:commentId",
		"/gh/repos/:owner/:repo/issues/events",
		"/gh/repos/:owner/:repo/issue-events/:eventId",
		"/gh/repos/:owner/:repo/issues/:number/events",
		"/gh/repos/:owner/:repo/assignees",
		"/gh/repos/:owner/:repo/assignees/:assignee",
		"/gh/repos/:owner/:repo/labels",
		"/gh/repos/:owner/:repo/labels/:name",
		"/gh/repos/:owner/:repo/milestones",
		"/gh/repos/:owner/:repo/milestones/:milestone",
		"/gh/repos/:owner/:repo/pulls",
		"/gh/repos/:owner/:repo/pulls/:number",
		"/gh/repos/:owner/:repo/pulls/:number/comments",
		"/gh/repos/:owner/:repo/pulls/:number/commits",
		"/gh/repos/:owner/:repo/pulls/:number/files",
		"/gh/repos/:owner/:repo/pull-comments/:commentId",
		"/gh/repos/:owner/:repo/hooks",
		"/gh/repos/:owner/:repo/hooks/:hookId",
		"/gh/repos/:owner/:repo/hooks/:hookId/deliveries",
		"/gh/repos/:owner/:repo/hooks/:hookId/deliveries/:deliveryId",
		"/gh/repos/:owner/:repo/collaborators",
		"/gh/repos/:owner/:repo/collaborators/:user",
		"/gh/repos/:owner/:repo/invitations",
		"/gh/repos/:owner/:repo/invitations/:invitationId",
		"/gh/repos/:owner/:repo/keys",
		"/gh/repos/:owner/:repo/keys/:keyId",
		"/gh/repos/:owner/:repo/projects",
		"/gh/repos/:owner/:repo/projects/:projectId",
		"/gh/repos/:owner/:repo/actions/secrets",
	)
	addAll(http.MethodPatch,
		"/gh/repos/:owner/:repo",
		"/gh/repos/:owner/:repo/releases/:releaseId",
		"/gh/repos/:owner/:repo/comments/:commentId",
		"/gh/repos/:owner/:repo/issues/:number",
		"/gh/repos/:owner/:repo/issue-comments/:commentId",
		"/gh/repos/:owner/:repo/labels/:name",
		"/gh/repos/:owner/:repo/milestones/:milestone",
		"/gh/repos/:owner/:repo/pulls/:number",
		"/gh/repos/:owner/:repo/pull-comments/:commentId",
		"/gh/repos/:owner/:repo/hooks/:hookId",
		"/gh/repos/:owner/:repo/projects/:projectId",
	)
	addAll(http.MethodPost,
		"/gh/repos/:owner/:repo/releases",
		"/gh/repos/:owner/:repo/deployments",
		"/gh/repos/:owner/:repo/git/refs",
		"/gh/repos/:owner/:repo/issues",
		"/gh/repos/:owner/:repo/issues/:number/comments",
		"/gh/repos/:owner/:repo/labels",
		"/gh/repos/:owner/:repo/milestones",
		"/gh/repos/:owner/:repo/pulls",
		"/gh/repos/:owner/:repo/hooks",
		"/gh/repos/:owner/:repo/keys",
		"/gh/repos/:owner/:repo/projects",
	)
	addAll(http.MethodPut,
		"/gh/repos/:owner/:repo/subscription",
		"/gh/repos/:owner/:repo/topics",
		"/gh/repos/:owner/:repo/contents/:entry",
		"/gh/repos/:owner/:repo/pages",
		"/gh/repos/:owner/:repo/collaborators/:user",
	)
	addAll(http.MethodDelete,
		"/gh/repos/:owner/:repo/subscription",
		"/gh/repos/:owner/:repo/releases/:releaseId",
		"/gh/repos/:owner/:repo/contents/:entry",
		"/gh/repos/:owner/:repo/downloads/:downloadId",
		"/gh/repos/:owner/:repo/issue-comments/:commentId",
		"/gh/repos/:owner/:repo/labels/:name",
		"/gh/repos/:owner/:repo/milestones/:milestone",
		"/gh/repos/:owner/:repo/pull-comments/:commentId",
		"/gh/repos/:owner/:repo/hooks/:hookId",
		"/gh/repos/:owner/:repo/collaborators/:user",
		"/gh/repos/:owner/:repo/invitations/:invitationId",
		"/gh/repos/:owner/:repo/keys/:keyId",
		"/gh/repos/:owner/:repo/pages",
		"/gh/repos/:owner/:repo/projects/:projectId",
	)

	addAll(http.MethodGet,
		"/gh/gists/:gistId",
		"/gh/gists/:gistId/comments",
		"/gh/gists/:gistId/comments/:commentId",
		"/gh/gists/:gistId/forks",
		"/gh/app/installations/:installationId",
		"/gh/app/installations/:installationId/repositories",
		"/gh/teams/:teamId",
		"/gh/teams/:teamId/repos",
		"/gh/teams/:teamId/members",
		"/gh/teams/:teamId/members/:user",
	)
	addAll(http.MethodPatch, "/gh/gists/:gistId", "/gh/gists/:gistId/comments/:commentId")
	addAll(http.MethodPost, "/gh/gists/:gistId/comments", "/gh/gists/:gistId/forks")
	addAll(http.MethodPut, "/gh/gists/:gistId/star")
	addAll(http.MethodDelete, "/gh/gists/:gistId", "/gh/gists/:gistId/comments/:commentId", "/gh/gists/:gistId/star")

	return routes
}

func TestRunScenarioBenchmarks(t *testing.T) {
	t.Skip(`
From benchmarks/:
    go test -run=^$ -bench '^BenchmarkScenarioRouteSet' -benchmem

To run from the repo root:
    cd benchmarks && go test -run=^$ -bench '^BenchmarkScenarioRouteSet' -benchmem

To keep local runs fast while iterating from benchmarks/:
    go test -run=^$ -bench '^BenchmarkScenarioRouteSet(All|Param|Static)$' -benchmem -benchtime=200ms

Notes:
- These route corpora are benchmark-oriented recreations inspired by the classic Go HTTP router benchmark shapes.
- The scenario suite is intentionally separate from comp_benchmark_test.go so micro and scenario results stay comparable over time.
- HttpRouter is excluded from the overlapping GitHub and GPlus corpora because it rejects those static-vs-param branches at registration time.
`)
}
