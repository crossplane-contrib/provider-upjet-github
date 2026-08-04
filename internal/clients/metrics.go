/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// routePlaceholder replaces a path segment that is not a known GitHub API
	// literal (an owner, repository, team slug, ref, ID, ...).
	routePlaceholder = ":x"

	// routeOther is the label used once maxRouteLabels distinct routes have
	// been seen. It is the hard bound on the route label's cardinality.
	routeOther = "other"

	// serviceUnknown is the service label for a path with no recognisable
	// leading GitHub API literal.
	serviceUnknown = "unknown"

	// maxRouteLabels bounds the number of distinct route label values.
	maxRouteLabels = 200

	// maxRouteSegments bounds the depth of a route label.
	maxRouteSegments = 8

	// statusTransportError is the status label used when the request failed
	// before a response was received.
	statusTransportError = "error"
)

// githubAPIRequests counts GitHub API requests by method, templated route and
// response status. Status is a first-class label because only 304 Not Modified
// responses are exempt from GitHub's primary rate limit, so the 304-vs-200 split
// measures how much of the provider's read traffic is answered conditionally.
var githubAPIRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "github_api",
	Name:      "requests_total",
	Help:      "The number of GitHub API requests made by the provider, by method, route and response status.",
}, []string{"method", "route", "status"})

// metricsRoundTripper counts every request it forwards. It sits below the
// Terraform provider's conditional-request layer, so a 304 from GitHub is
// observed as a 304 rather than as the caller-visible unchanged read.
//
// Being installed on http.DefaultClient means any other traffic through that
// client is counted as well. In this process that is GitHub API traffic; the
// Kubernetes clients build their own transports.
type metricsRoundTripper struct {
	base             http.RoundTripper
	requests         *prometheus.CounterVec
	externalAPICalls *prometheus.CounterVec
	routes           *routeLimiter
}

// withRequestMetrics counts every request forwarded through the chain.
func withRequestMetrics(requests, externalAPICalls *prometheus.CounterVec) transportWrapper {
	return func(base http.RoundTripper) http.RoundTripper {
		return newMetricsRoundTripper(base, requests, externalAPICalls)
	}
}

func newMetricsRoundTripper(base http.RoundTripper, requests, externalAPICalls *prometheus.CounterVec) *metricsRoundTripper {
	return &metricsRoundTripper{
		base:             base,
		requests:         requests,
		externalAPICalls: externalAPICalls,
		routes:           newRouteLimiter(maxRouteLabels),
	}
}

func (m *metricsRoundTripper) unwrap() http.RoundTripper { return m.base }

func (m *metricsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := m.base.RoundTrip(req)

	template := ""
	if req.URL != nil {
		template = templateRoute(req.URL.Path)
	}

	status := statusTransportError
	if resp != nil {
		status = strconv.Itoa(resp.StatusCode)
	}

	m.requests.WithLabelValues(req.Method, m.routes.label(template), status).Inc()

	// Also feed upjet's generic external-API counter, whose timeseries budget is
	// shared with the other upjet providers. It gets the coarse service name and
	// the method, which is the (service, operation) shape the AWS, Azure and GCP
	// providers report; the full route template stays on githubAPIRequests.
	if m.externalAPICalls != nil {
		m.externalAPICalls.WithLabelValues(serviceLabel(template), req.Method).Inc()
	}

	return resp, err
}

// serviceLabel reduces a route template to a coarse GitHub API service name, so
// the shared upjet counter stays at roughly the cardinality the other providers
// give it rather than gaining a value per route.
func serviceLabel(template string) string {
	for _, s := range strings.Split(strings.Trim(template, "/"), "/") {
		switch s {
		case "", routePlaceholder, "api", "v3":
			continue
		default:
			return s
		}
	}
	return serviceUnknown
}

// routeLimiter caps how many distinct route labels may be emitted, collapsing
// everything beyond the cap into routeOther.
type routeLimiter struct {
	mu   sync.Mutex
	seen map[string]struct{}
	max  int
}

func newRouteLimiter(max int) *routeLimiter {
	return &routeLimiter{seen: make(map[string]struct{}, max), max: max}
}

func (l *routeLimiter) label(route string) string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if _, ok := l.seen[route]; ok {
		return route
	}
	if len(l.seen) >= l.max {
		return routeOther
	}
	l.seen[route] = struct{}{}
	return route
}

// templateRoute turns a request path into a bounded route template by keeping
// path segments that are known GitHub API literals and replacing every other
// segment with routePlaceholder.
//
// Doing it this way round is what bounds the label. A literal missing from
// knownRouteSegments merges two routes into one, which loses detail but cannot
// grow the label set. The only way to grow it is a variable segment that
// happens to equal a known literal (a repository actually named "issues"), and
// routeLimiter's cap bounds that.
func templateRoute(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) == 1 && segments[0] == "" {
		return "/"
	}

	truncated := false
	if len(segments) > maxRouteSegments {
		segments = segments[:maxRouteSegments]
		truncated = true
	}

	out := make([]string, 0, len(segments)+1)
	for _, s := range segments {
		lower := strings.ToLower(s)
		if _, ok := knownRouteSegments[lower]; ok {
			out = append(out, lower)
			continue
		}
		out = append(out, routePlaceholder)
	}
	if truncated {
		out = append(out, "...")
	}

	return "/" + strings.Join(out, "/")
}

// knownRouteSegments are the static path segments of the GitHub API endpoints
// this provider uses. It does not need to be exhaustive; see templateRoute.
var knownRouteSegments = map[string]struct{}{
	// GHES / GraphQL prefixes
	"api": {}, "v3": {}, "graphql": {}, "rate_limit": {}, "meta": {},

	// top level
	"app": {}, "applications": {}, "enterprises": {}, "installation": {},
	"installations": {}, "licenses": {}, "orgs": {}, "repos": {}, "search": {},
	"teams": {}, "user": {}, "users": {},

	// repository sub-resources
	"actions": {}, "activity": {}, "assignees": {}, "attestations": {},
	"autolinks": {}, "branches": {}, "check-runs": {}, "check-suites": {},
	"code-scanning": {}, "codeowners": {}, "collaborators": {}, "comments": {},
	"commits": {}, "compare": {}, "contents": {}, "contributors": {},
	"custom-properties": {}, "dependabot": {}, "dependency-graph": {},
	"deployments": {}, "dispatches": {}, "environments": {}, "events": {},
	"forks": {}, "generate": {}, "git": {}, "hooks": {}, "invitations": {},
	"issues": {}, "keys": {}, "labels": {}, "languages": {}, "merges": {},
	"milestones": {}, "notifications": {}, "pages": {}, "projects": {},
	"properties": {}, "pulls": {}, "readme": {}, "releases": {},
	"rule-suites": {}, "rules": {}, "rulesets": {}, "secret-scanning": {},
	"stargazers": {}, "statuses": {}, "subscribers": {}, "subscription": {},
	"tags": {}, "topics": {}, "transfer": {}, "vulnerability-alerts": {},
	"private-vulnerability-reporting": {},

	// git database
	"blobs": {}, "matching-refs": {}, "refs": {}, "trees": {},

	// branch protection
	"protection": {}, "rename": {}, "enforce_admins": {},
	"required_pull_request_reviews": {}, "required_signatures": {},
	"required_status_checks": {}, "restrictions": {}, "contexts": {},

	// actions
	"artifacts": {}, "caches": {}, "jobs": {}, "logs": {}, "oidc": {},
	"permissions": {}, "public-key": {}, "registration-token": {},
	"remove-token": {}, "runner-groups": {}, "runners": {}, "runs": {},
	"secrets": {}, "selected-actions": {}, "sub": {}, "variables": {},
	"workflows": {}, "access": {}, "customization": {},
	"deployment-branch-policies": {}, "deployment-protection-rules": {},

	// org / team sub-resources
	"audit-log": {}, "blocks": {}, "copilot": {},
	"credential-authorizations": {}, "external-group": {},
	"external-groups": {}, "failed_invitations": {}, "memberships": {},
	"members": {}, "outside_collaborators": {},
	"personal-access-tokens": {}, "public_members": {},
	"security-managers": {}, "settings": {}, "schema": {}, "team-sync": {},
	"group-mappings": {}, "discussions": {},

	// user sub-resources
	"emails": {}, "followers": {}, "following": {}, "gpg_keys": {},
	"ssh_signing_keys": {}, "starred": {}, "subscriptions": {},

	// app
	"config": {}, "deliveries": {}, "hook": {}, "manifests": {},
	"access_tokens": {},
}
