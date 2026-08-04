/*
Copyright 2022 Upbound Inc.
*/

package clients

import "net/http"

// transportWrapper adds one concern to a RoundTripper chain.
type transportWrapper func(http.RoundTripper) http.RoundTripper

// chainTransports wraps base with each wrapper, the first listed ending up
// outermost and the last closest to the wire.
//
// Ordering is load-bearing here and reads inside-out when the constructors are
// nested directly, which is the opposite of how the chain is reasoned about.
// Listing the wrappers outermost-first keeps the declaration in the same order as
// the request path.
func chainTransports(base http.RoundTripper, wrappers ...transportWrapper) http.RoundTripper {
	for i := len(wrappers) - 1; i >= 0; i-- {
		base = wrappers[i](base)
	}
	return base
}

// unwrapper is implemented by a chain member that delegates to another
// RoundTripper, so a chain can be walked without knowing how it was assembled.
type unwrapper interface {
	unwrap() http.RoundTripper
}
