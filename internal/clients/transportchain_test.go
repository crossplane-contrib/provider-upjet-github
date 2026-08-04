/*
Copyright 2022 Upbound Inc.
*/

package clients

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
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

	resp, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, "https://api.github.com/", nil))
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
