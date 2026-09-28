/*
Copyright 2022 Upbound Inc.
*/

package clients

import (
	"net/http"
	"strconv"
	"time"
)

const (
	// secondaryLimitCooldownCap bounds how long the client will stop sending
	// requests after a secondary rate limit that GitHub gave no Retry-After for.
	secondaryLimitCooldownCap = 15 * time.Second

	headerRetryAfter         = "Retry-After"
	headerRateLimitReset     = "X-RateLimit-Reset"
	headerRateLimitRemaining = "X-RateLimit-Remaining"
)

// boundSecondaryLimitCooldown caps the cooldown the rate limiter above it derives
// from a secondary rate limit response.
//
// That limiter parks every subsequent request until the cooldown ends. Lacking a
// Retry-After it falls back to X-RateLimit-Reset, which carries the *primary*
// hourly reset and so can be most of an hour away. Every request then outlives its
// reconcile deadline, and the limiter issues them anyway once the wait is cut
// short, so the provider reports a rate limit as "context deadline exceeded" while
// no request reaches GitHub at all. One repository's write loop takes down reads
// for every GitHub resource on the control plane.
//
// Rewriting the reset rather than setting a Retry-After is deliberate: the retrying
// transport in the same chain honours Retry-After, so supplying one would turn a
// single wait into several and spend the deadline that way instead.
type boundSecondaryLimitCooldown struct {
	base http.RoundTripper
	cap  time.Duration
	now  func() time.Time
}

// withBoundedSecondaryLimitCooldown caps the cooldown the rate limiter above the
// chain derives from a secondary rate limit response.
func withBoundedSecondaryLimitCooldown(cooldownCap time.Duration) transportWrapper {
	return func(base http.RoundTripper) http.RoundTripper {
		return newBoundSecondaryLimitCooldown(base, cooldownCap)
	}
}

func newBoundSecondaryLimitCooldown(base http.RoundTripper, cooldownCap time.Duration) *boundSecondaryLimitCooldown {
	return &boundSecondaryLimitCooldown{base: base, cap: cooldownCap, now: time.Now}
}

func (b *boundSecondaryLimitCooldown) unwrap() http.RoundTripper { return b.base }

func (b *boundSecondaryLimitCooldown) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := b.base.RoundTrip(req)
	if err != nil || resp == nil {
		return resp, err
	}

	if reset, ok := b.overlongSecondaryCooldown(resp); ok {
		resp.Header.Set(headerRateLimitReset, strconv.FormatInt(reset.Unix(), 10))
	}

	return resp, nil
}

// overlongSecondaryCooldown reports the capped reset for a secondary rate limit
// response whose reset is further out than the cap, and whether to apply it.
//
// A response carrying a Retry-After is left alone: GitHub said how long to wait,
// and that value is the one to honour. A response with no requests remaining is a
// *primary* rate limit, whose reset is authoritative and drives a different
// limiter, so it is left alone too.
//
// The body is deliberately not inspected, although that is part of how the limiter
// above identifies a secondary rate limit. Being wrong here is inert: a response
// that limiter does not classify as a secondary rate limit never has its reset read
// for a cooldown.
func (b *boundSecondaryLimitCooldown) overlongSecondaryCooldown(resp *http.Response) (time.Time, bool) {
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusTooManyRequests:
	default:
		return time.Time{}, false
	}

	if resp.Header == nil || resp.Header.Get(headerRetryAfter) != "" {
		return time.Time{}, false
	}

	if remaining, err := strconv.ParseInt(resp.Header.Get(headerRateLimitRemaining), 10, 64); err == nil && remaining <= 0 {
		return time.Time{}, false
	}

	seconds, err := strconv.ParseInt(resp.Header.Get(headerRateLimitReset), 10, 64)
	if err != nil || seconds <= 0 {
		return time.Time{}, false
	}

	capped := b.now().Add(b.cap)
	if !time.Unix(seconds, 0).After(capped) {
		return time.Time{}, false
	}

	return capped, true
}
