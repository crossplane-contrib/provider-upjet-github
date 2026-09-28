/*
Copyright 2022 Upbound Inc.
*/

package clients

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// chainTypes names each member of a RoundTripper chain, outermost first, so a test
// can assert composition order without reaching into any wrapper's fields.
func chainTypes(rt http.RoundTripper) []string {
	var types []string
	for rt != nil {
		types = append(types, fmt.Sprintf("%T", rt))
		next, ok := rt.(unwrapper)
		if !ok {
			break
		}
		rt = next.unwrap()
	}
	return types
}

// assertChainOrder checks that want is the leading prefix of the chain's types.
func assertChainOrder(t *testing.T, rt http.RoundTripper, want ...string) {
	t.Helper()
	got := chainTypes(rt)
	if len(got) < len(want) {
		t.Fatalf("chain is %v, want it to start with %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("chain is %v, want it to start with %v", got, want)
		}
	}
}

type namedTransport struct {
	name  string
	inner http.RoundTripper
	order *[]string
}

func (n *namedTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	*n.order = append(*n.order, n.name)
	if n.inner == nil {
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	}
	return n.inner.RoundTrip(r)
}

func (n *namedTransport) unwrap() http.RoundTripper { return n.inner }

func named(name string, order *[]string) transportWrapper {
	return func(base http.RoundTripper) http.RoundTripper {
		return &namedTransport{name: name, inner: base, order: order}
	}
}

// The first wrapper listed must end up outermost, so the declaration reads in the
// same order a request travels.
func TestChainTransportsWrapsFirstListedOutermost(t *testing.T) {
	var order []string
	base := &namedTransport{name: "base", order: &order}

	rt := chainTransports(base, named("outer", &order), named("inner", &order))

	resp, err := rt.RoundTrip(httptest.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.github.com/", nil))
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	want := []string{"outer", "inner", "base"}
	if len(order) != len(want) {
		t.Fatalf("request path = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("request path = %v, want %v", order, want)
		}
	}
}

func TestChainTransportsWithNoWrappersReturnsBase(t *testing.T) {
	var order []string
	base := &namedTransport{name: "base", order: &order}

	if got := chainTransports(base); got != http.RoundTripper(base) {
		t.Errorf("chainTransports(base) = %T, want the base unchanged", got)
	}
}

// cloneTransportRetains mirrors the type switch in terraform-provider-github's
// cloneTransport (internal/ghclient/transport.go): it tunes a concrete
// *http.Transport by cloning it, and returns any other RoundTripper unchanged.
// A wrapper is only reachable from the new client's chain if it is returned
// unchanged, so this is the property the whole seam depends on.
func cloneTransportRetains(rt http.RoundTripper) bool {
	_, concrete := rt.(*http.Transport)
	return !concrete
}

func TestInstallNewClientTransport_SurvivesCloneTransport(t *testing.T) {
	original := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = original })

	installNewClientTransport()

	assertChainOrder(t, http.DefaultTransport,
		"*clients.boundSecondaryLimitCooldown")
	if !cloneTransportRetains(http.DefaultTransport) {
		t.Error("cloneTransport would swap the wrapper for a tuned clone, dropping the counter out of the new client's chain")
	}
}

// The wiring itself: DefaultTransport must be wrapped when the new client is in
// use and left concrete when it is not. This is what was missing -- the counter
// was installed only where the legacy client would find it, so on a
// legacy_client=false control plane it reported nothing at all.
func TestInstallTransports_WrapsDefaultTransportOnlyForNewClient(t *testing.T) {
	for _, tc := range []struct {
		name        string
		legacy      string
		wantWrapped bool
	}{
		{name: "new client wraps DefaultTransport so its chain is reachable", legacy: "false", wantWrapped: true},
		{name: "legacy client leaves DefaultTransport concrete so the provider's unchecked assertion holds", legacy: "true", wantWrapped: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			originalTransport := http.DefaultTransport
			originalClient := http.DefaultClient.Transport
			t.Cleanup(func() {
				http.DefaultTransport = originalTransport
				http.DefaultClient.Transport = originalClient
			})
			t.Setenv(envLegacyClient, tc.legacy)

			installTransportChains()

			_, concrete := http.DefaultTransport.(*http.Transport)
			if wrapped := !concrete; wrapped != tc.wantWrapped {
				t.Errorf("http.DefaultTransport is %T (wrapped=%v), want wrapped=%v", http.DefaultTransport, wrapped, tc.wantWrapped)
			}

			if tc.wantWrapped {
				assertChainOrder(t, http.DefaultTransport, "*clients.boundSecondaryLimitCooldown")
			}
		})
	}
}

func TestTunedBaseTransport_CarriesRESTTuning(t *testing.T) {
	base := &http.Transport{MaxIdleConns: 1, MaxIdleConnsPerHost: 1, IdleConnTimeout: time.Second}

	tuned, ok := tunedBaseTransport(base).(*http.Transport)
	if !ok {
		t.Fatalf("tunedBaseTransport returned %T, want *http.Transport", tunedBaseTransport(base))
	}

	if !tuned.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 = false, want true")
	}
	if tuned.MaxIdleConns != newClientMaxIdleConns {
		t.Errorf("MaxIdleConns = %d, want %d", tuned.MaxIdleConns, newClientMaxIdleConns)
	}
	if tuned.MaxIdleConnsPerHost != newClientMaxIdleConns {
		t.Errorf("MaxIdleConnsPerHost = %d, want %d", tuned.MaxIdleConnsPerHost, newClientMaxIdleConns)
	}
	if tuned.IdleConnTimeout != newClientIdleConnTimeout {
		t.Errorf("IdleConnTimeout = %s, want %s", tuned.IdleConnTimeout, newClientIdleConnTimeout)
	}

	// The caller's transport must not be mutated; it is process-wide.
	if base.MaxIdleConns != 1 {
		t.Errorf("base transport was mutated: MaxIdleConns = %d, want 1", base.MaxIdleConns)
	}
}
