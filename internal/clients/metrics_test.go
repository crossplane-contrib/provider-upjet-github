/*
Copyright 2021 Upbound Inc.
*/

package clients

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	tpg "github.com/integrations/terraform-provider-github/v6/github"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// stubRoundTripper returns a canned status (or error) without touching the network.
type stubRoundTripper struct {
	status int
	err    error
	calls  atomic.Int64
}

func (s *stubRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	s.calls.Add(1)
	if s.err != nil {
		return nil, s.err
	}
	rec := httptest.NewRecorder()
	rec.WriteHeader(s.status)
	return rec.Result(), nil
}

// newTestCounters returns fresh, unregistered counters so tests never depend on
// package-global state or on each other's ordering.
func newTestCounters() (requests, extAPI *prometheus.CounterVec) {
	return prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_github_api_requests_total",
			Help: "test",
		}, []string{"method", "route", "status"}),
		prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_external_api_calls_total",
			Help: "test",
		}, []string{"service", "operation"})
}

func do(t *testing.T, rt http.RoundTripper, method, url string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		t.Fatalf("cannot build request: %v", err)
	}
	resp, err := rt.RoundTrip(req)
	if err != nil && !errors.Is(err, errStub) {
		t.Fatalf("unexpected RoundTrip error: %v", err)
	}
	closeBody(resp)
}

// closeBody drains nothing and just closes, but a RoundTripper's contract is
// that the caller owns the body -- and bodyclose enforces it in tests too.
func closeBody(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
}

var errStub = errors.New("stub transport failure")

// TestMetricsRoundTripper_CountsByStatus is the core assertion: a 200, a 304 and
// a 403 against the same route must land in three separate series. The 304 split
// is the point of the metric — only 304 responses are exempt from GitHub's
// primary rate limit, so a 200-vs-304 ratio is what makes conditional-request
// effectiveness measurable.
func TestMetricsRoundTripper_CountsByStatus(t *testing.T) {
	const url = "https://api.github.com/repos/some-owner/some-repo"

	for _, status := range []int{http.StatusOK, http.StatusNotModified, http.StatusForbidden} {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			requests, extAPI := newTestCounters()
			stub := &stubRoundTripper{status: status}
			rt := newMetricsRoundTripper(stub, requests, extAPI)

			do(t, rt, http.MethodGet, url)

			got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, "/repos/:x/:x", fmt.Sprint(status)))
			if got != 1 {
				t.Errorf("counter for status %d = %v, want 1", status, got)
			}
			if got := stub.calls.Load(); got != 1 {
				t.Errorf("base transport calls = %d, want 1", got)
			}
		})
	}
}

// TestMetricsRoundTripper_StatusesAreSeparateSeries pins that a 200 and a 304 on
// the same route do not collapse into one another.
func TestMetricsRoundTripper_StatusesAreSeparateSeries(t *testing.T) {
	requests, extAPI := newTestCounters()

	ok := newMetricsRoundTripper(&stubRoundTripper{status: http.StatusOK}, requests, extAPI)
	notModified := newMetricsRoundTripper(&stubRoundTripper{status: http.StatusNotModified}, requests, extAPI)

	const url = "https://api.github.com/repos/some-owner/some-repo/branches/main/protection"
	do(t, ok, http.MethodGet, url)
	do(t, notModified, http.MethodGet, url)
	do(t, notModified, http.MethodGet, url)

	const route = "/repos/:x/:x/branches/:x/protection"
	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, route, "200")); got != 1 {
		t.Errorf("200 counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, route, "304")); got != 2 {
		t.Errorf("304 counter = %v, want 2", got)
	}
}

// TestMetricsRoundTripper_TransportError records a request that never got a
// response, so failures are not silently uncounted.
func TestMetricsRoundTripper_TransportError(t *testing.T) {
	requests, extAPI := newTestCounters()
	rt := newMetricsRoundTripper(&stubRoundTripper{err: errStub}, requests, extAPI)

	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, "https://api.github.com/graphql", nil)
	resp, err := rt.RoundTrip(req)
	closeBody(resp)
	if err == nil {
		t.Fatal("expected the base transport error to be propagated")
	}
	if resp != nil {
		t.Errorf("expected no response, got %v", resp)
	}

	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodPost, "/graphql", statusTransportError)); got != 1 {
		t.Errorf("error counter = %v, want 1", got)
	}
}

// TestMetricsRoundTripper_FeedsUpjetCounter checks the second, already-allowlisted
// metric keeps the (service, operation) shape the AWS/Azure/GCP providers use.
func TestMetricsRoundTripper_FeedsUpjetCounter(t *testing.T) {
	requests, extAPI := newTestCounters()
	rt := newMetricsRoundTripper(&stubRoundTripper{status: http.StatusOK}, requests, extAPI)

	do(t, rt, http.MethodPatch, "https://api.github.com/orgs/some-org/teams/some-team")

	if got := testutil.ToFloat64(extAPI.WithLabelValues("orgs", http.MethodPatch)); got != 1 {
		t.Errorf("upjet counter = %v, want 1", got)
	}
}

func TestTemplateRoute(t *testing.T) {
	for _, tc := range []struct {
		path string
		want string
	}{
		{"/repos/some-owner/some-repo", "/repos/:x/:x"},
		{"/repos/some-owner/some-repo/issues/1234/labels", "/repos/:x/:x/issues/:x/labels"},
		{"/repos/some-owner/some-repo/branches/feature-x/protection/required_status_checks", "/repos/:x/:x/branches/:x/protection/required_status_checks"},
		{"/orgs/some-org/teams/some-team/memberships/some-user", "/orgs/:x/teams/:x/memberships/:x"},
		{"/graphql", "/graphql"},
		{"/api/v3/repos/some-owner/some-repo", "/api/v3/repos/:x/:x"},
		{"/user", "/user"},
		{"/", "/"},
		// Different owners and repositories must collapse onto one route: this
		// is what keeps the label from growing with the estate.
		{"/repos/another-owner/another-repo", "/repos/:x/:x"},
		// Depth is bounded: only maxRouteSegments segments are kept.
		{"/repos/o/r/a/b/c/d/e/f/g/h", "/repos/:x/:x/:x/:x/:x/:x/:x/..."},
	} {
		if got := templateRoute(tc.path); got != tc.want {
			t.Errorf("templateRoute(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// TestTemplateRoute_NeverContainsIdentifiers is the cardinality guarantee stated
// negatively: no owner, repository or ref name may reach a label value.
func TestTemplateRoute_NeverContainsIdentifiers(t *testing.T) {
	route := templateRoute("/repos/secret-owner/secret-repo/git/refs/heads/secret-branch")
	for _, identifier := range []string{"secret-owner", "secret-repo", "secret-branch"} {
		if strings.Contains(route, identifier) {
			t.Errorf("route %q leaked identifier %q", route, identifier)
		}
	}
}

func TestRouteLimiter_CapsDistinctRoutes(t *testing.T) {
	l := newRouteLimiter(3)

	for i := range 3 {
		route := fmt.Sprintf("/route-%d", i)
		if got := l.label(route); got != route {
			t.Errorf("label(%q) = %q, want it kept", route, got)
		}
		// Already-seen routes stay themselves.
		if got := l.label(route); got != route {
			t.Errorf("repeat label(%q) = %q, want it kept", route, got)
		}
	}

	if got := l.label("/route-over-cap"); got != routeOther {
		t.Errorf("label past the cap = %q, want %q", got, routeOther)
	}
	// The cap does not evict what is already tracked.
	if got := l.label("/route-0"); got != "/route-0" {
		t.Errorf("label(%q) after the cap = %q, want it kept", "/route-0", got)
	}
}

// TestMetricsRoundTripper_Concurrent guards the route table against a data race;
// meaningful under -race.
func TestMetricsRoundTripper_Concurrent(t *testing.T) {
	requests, extAPI := newTestCounters()
	rt := newMetricsRoundTripper(&stubRoundTripper{status: http.StatusOK}, requests, extAPI)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, fmt.Sprintf("https://api.github.com/repos/owner-%d/repo", i), nil)
			resp, err := rt.RoundTrip(req)
			closeBody(resp)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, "/repos/:x/:x", "200")); got != 50 {
		t.Errorf("counter = %v, want 50", got)
	}
}

// TestInstallMetricsTransport_LeavesDefaultTransportAlone pins the reason the
// hook is http.DefaultClient.Transport: the Terraform provider does an unchecked
// http.DefaultTransport.(*http.Transport) assertion on its anonymous client path,
// which would panic if http.DefaultTransport were replaced with a wrapper.
// With the legacy client in play, http.DefaultTransport must stay a concrete
// *http.Transport: the Terraform provider type-asserts it unchecked when it
// builds an anonymous client, and a wrapper there would panic the provider.
func TestInstallMetricsTransport_LegacyClientLeavesDefaultTransportAlone(t *testing.T) {
	if !legacyClientEnabled() {
		t.Skip("GITHUB_LEGACY_CLIENT is disabled here; DefaultTransport is wrapped by design")
	}

	installTransports()

	if _, ok := http.DefaultTransport.(*http.Transport); !ok {
		t.Fatalf("http.DefaultTransport is %T, want *http.Transport", http.DefaultTransport)
	}
	if _, ok := http.DefaultClient.Transport.(*metricsRoundTripper); !ok {
		t.Fatalf("http.DefaultClient.Transport is %T, want *metricsRoundTripper", http.DefaultClient.Transport)
	}

	// Idempotent: a second call must not wrap the wrapper.
	before := http.DefaultClient.Transport
	installTransports()
	if http.DefaultClient.Transport != before {
		t.Error("installTransports is not idempotent")
	}
}

func TestServiceLabel(t *testing.T) {
	for _, tc := range []struct {
		template string
		want     string
	}{
		{"/repos/:x/:x/issues/:x/labels", "repos"},
		{"/orgs/:x/teams/:x/memberships/:x", "orgs"},
		{"/api/v3/repos/:x/:x", "repos"},
		{"/graphql", "graphql"},
		{"/user", "user"},
		{"/:x/:x", serviceUnknown},
		{"/", serviceUnknown},
	} {
		if got := serviceLabel(tc.template); got != tc.want {
			t.Errorf("serviceLabel(%q) = %q, want %q", tc.template, got, tc.want)
		}
	}
}

// TestInstalledTransport_CountsTerraformProviderClient is the end-to-end check
// that matters: it drives the Terraform provider's own client constructor and
// asserts the request landed on the counter. Asserting that we installed the
// wrapper is not the same as asserting the provider's client routes through it,
// and this is what would break if the Terraform provider stopped building its
// client via oauth2.NewClient.
func TestInstalledTransport_CountsTerraformProviderClient(t *testing.T) {
	installTransports()

	rt, ok := http.DefaultClient.Transport.(*metricsRoundTripper)
	if !ok {
		t.Fatalf("http.DefaultClient.Transport is %T, want *metricsRoundTripper", http.DefaultClient.Transport)
	}

	// Count against the installed transport's own counters, isolated per test.
	requests, extAPI := newTestCounters()
	rt.requests, rt.externalAPICalls = requests, extAPI

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotModified)
	}))
	defer srv.Close()

	// A token makes Anonymous() false, which is the authenticated legacy path
	// that App authentication also resolves to once the installation token is
	// minted. Zero delays and retries keep the test fast and deterministic.
	cfg := &tpg.Config{Token: "not-a-real-token"}
	client := cfg.AuthenticatedHTTPClient()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+"/repos/some-owner/some-repo", nil)
	if err != nil {
		t.Fatalf("cannot build request: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request through the provider client failed: %v", err)
	}
	defer resp.Body.Close()

	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, "/repos/:x/:x", "304")); got != 1 {
		t.Errorf("304 counter = %v, want 1: the Terraform provider client did not route through the installed transport", got)
	}
	if got := testutil.ToFloat64(extAPI.WithLabelValues("repos", http.MethodGet)); got != 1 {
		t.Errorf("upjet counter = %v, want 1", got)
	}
}

func TestTunedBaseTransport_PassesThroughNonTransport(t *testing.T) {
	rt := &stubRoundTripper{status: http.StatusOK}

	if got := tunedBaseTransport(rt); got != http.RoundTripper(rt) {
		t.Errorf("tunedBaseTransport returned %T, want the original RoundTripper unchanged", got)
	}
}

// The whole point of wrapping DefaultTransport is that traffic reaching it is
// counted, including the 304s that make conditional reads free.
func TestNewClientMetricsTransport_Counts(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	requests, extAPI := newTestCounters()
	http.DefaultTransport = newMetricsRoundTripper(
		tunedBaseTransport(&stubRoundTripper{status: http.StatusNotModified}),
		requests, extAPI,
	)

	do(t, http.DefaultTransport, http.MethodGet, "https://api.github.com/repos/o/r")

	if got := testutil.ToFloat64(requests.WithLabelValues(http.MethodGet, "/repos/:x/:x", "304")); got != 1 {
		t.Errorf("304 count = %v, want 1", got)
	}
}
