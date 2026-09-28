/*
Copyright 2022 Upbound Inc.
*/

package clients

import (
	"net/http"
	"sync"
	"time"

	upjetmetrics "github.com/crossplane/upjet/v2/pkg/metrics"
	ctrlmetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// newClientMaxIdleConns and newClientIdleConnTimeout mirror the REST values
	// terraform-provider-github applies in its own cloneTransport, which it skips
	// for a wrapped transport.
	newClientMaxIdleConns    = 100
	newClientIdleConnTimeout = 90 * time.Second
)

// transportWrapper adds one concern to a RoundTripper chain.
type transportWrapper func(http.RoundTripper) http.RoundTripper

// chainTransports wraps base with each wrapper, the first listed ending up
// outermost and the last closest to the wire.
//
// Nesting the constructors directly reads inside-out, which is the reverse of the
// order a request travels, and that order is load-bearing. Listing the wrappers
// outermost-first keeps the declaration in the same order as the request path.
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

var installTransportsOnce sync.Once

// installTransports inserts this provider's RoundTrippers underneath the Terraform
// provider's client stack. It is safe to call more than once; only the first call
// installs.
func installTransports() {
	installTransportsOnce.Do(func() {
		ctrlmetrics.Registry.MustRegister(githubAPIRequests)
		installTransportChains()
	})
}

// installTransportChains does the wiring. It is separate from the sync.Once so
// tests can exercise both the legacy and the new-client path, which the Once would
// otherwise let only one of them reach.
// The legacy client never consults http.DefaultTransport -- it resolves its base
// RoundTripper from http.DefaultClient, which the Terraform provider builds its
// authenticated client from -- so that seam is wrapped too, and unconditionally.
func installTransportChains() {
	base := http.DefaultClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	http.DefaultClient.Transport = newMetricsRoundTripper(base, githubAPIRequests, upjetmetrics.ExternalAPICalls)

	if !legacyClientEnabled() {
		installNewClientTransport()
	}
}

// installNewClientTransport wraps http.DefaultTransport, which is where the new
// client (legacy_client = false) builds its chain from.
//
// This is only safe with the legacy client disabled. The Terraform provider
// asserts http.DefaultTransport to *http.Transport unchecked when it builds an
// anonymous client, which would panic against a wrapper -- but that call sits
// inside the provider's `if c.LegacyClient` branch, so it cannot be reached from
// here.
//
// The wrapped base has to carry its own connection tuning. The provider's
// cloneTransport tunes a value only when it is already a *http.Transport and
// returns anything else untouched. That is exactly what keeps this wrapper in the
// chain, and equally what stops the settings it would have applied from being
// applied.
func installNewClientTransport() {
	http.DefaultTransport = chainTransports(tunedBaseTransport(http.DefaultTransport),
		withBoundedSecondaryLimitCooldown(secondaryLimitCooldownCap),
		withRequestMetrics(githubAPIRequests, upjetmetrics.ExternalAPICalls),
	)
}

// tunedBaseTransport clones tr and applies the connection settings the Terraform
// provider's cloneTransport applies to a REST client. Its GraphQL clients are
// tuned a little differently there (fewer idle connections, longer idle timeout)
// and inherit these REST values instead, which affects connection reuse only.
func tunedBaseTransport(tr http.RoundTripper) http.RoundTripper {
	base, ok := tr.(*http.Transport)
	if !ok {
		return tr
	}

	tuned := base.Clone()
	tuned.ForceAttemptHTTP2 = true
	tuned.MaxIdleConns = newClientMaxIdleConns
	tuned.MaxIdleConnsPerHost = newClientMaxIdleConns
	tuned.IdleConnTimeout = newClientIdleConnTimeout
	return tuned
}
