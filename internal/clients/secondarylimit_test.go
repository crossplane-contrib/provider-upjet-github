/*
Copyright 2022 Upbound Inc.
*/

package clients

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// stubTransport answers with a fixed status and headers. It builds the response
// itself rather than taking one from a helper, so no function in this file returns
// an *http.Response for the caller to have to close.
type stubTransport struct {
	status int
	header http.Header
}

func (s *stubTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: s.status, Header: s.header, Body: http.NoBody}, nil
}

func header(pairs ...string) http.Header {
	h := make(http.Header)
	for i := 0; i < len(pairs); i += 2 {
		h.Set(pairs[i], pairs[i+1])
	}
	return h
}

func epoch(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

func TestBoundSecondaryLimitCooldown(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	cap := 15 * time.Second
	farOut := epoch(now.Add(50 * time.Minute))
	soon := epoch(now.Add(5 * time.Second))

	for _, tc := range []struct {
		name      string
		status    int
		header    http.Header
		wantReset string
	}{
		{
			// The live case: the reset carries the primary hourly window, so the
			// limiter above would park for the best part of an hour.
			name:      "caps a secondary limit reset that is further out than the cap",
			status:    http.StatusTooManyRequests,
			header:    header("X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", farOut),
			wantReset: epoch(now.Add(cap)),
		},
		{
			name:      "caps a 403 secondary limit too",
			status:    http.StatusForbidden,
			header:    header("X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", farOut),
			wantReset: epoch(now.Add(cap)),
		},
		{
			// GitHub said how long to wait, so that is the value to honour.
			name:      "leaves a response carrying Retry-After alone",
			status:    http.StatusTooManyRequests,
			header:    header("Retry-After", "60", "X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", farOut),
			wantReset: farOut,
		},
		{
			// No requests remaining is a primary rate limit, whose reset is
			// authoritative and drives a different limiter.
			name:      "leaves a primary rate limit alone",
			status:    http.StatusForbidden,
			header:    header("X-RateLimit-Remaining", "0", "X-RateLimit-Reset", farOut),
			wantReset: farOut,
		},
		{
			name:      "leaves a reset already inside the cap alone",
			status:    http.StatusTooManyRequests,
			header:    header("X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", soon),
			wantReset: soon,
		},
		{
			name:      "leaves a success alone",
			status:    http.StatusOK,
			header:    header("X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", farOut),
			wantReset: farOut,
		},
		{
			name:      "does nothing when there is no reset to cap",
			status:    http.StatusTooManyRequests,
			header:    header("X-RateLimit-Remaining", "4998"),
			wantReset: "",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rt := newBoundSecondaryLimitCooldown(&stubTransport{status: tc.status, header: tc.header}, cap)
			rt.now = func() time.Time { return now }

			resp, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, "https://api.github.com/repos/o/r", nil))
			if err != nil {
				t.Fatalf("RoundTrip() error = %v", err)
			}
			defer resp.Body.Close()

			if got := resp.Header.Get("X-RateLimit-Reset"); got != tc.wantReset {
				t.Errorf("X-RateLimit-Reset = %q, want %q", got, tc.wantReset)
			}
		})
	}
}

// Supplying a Retry-After would be the obvious way to bound the wait, and it is
// wrong: the retrying transport in the same chain honours Retry-After, so it would
// turn one wait into several and spend the reconcile deadline that way instead.
func TestBoundSecondaryLimitCooldownNeverSetsRetryAfter(t *testing.T) {
	now := time.Now()
	h := header("X-RateLimit-Remaining", "4998", "X-RateLimit-Reset", epoch(now.Add(time.Hour)))
	rt := newBoundSecondaryLimitCooldown(&stubTransport{status: http.StatusTooManyRequests, header: h}, 15*time.Second)

	resp, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, "https://api.github.com/repos/o/r", nil))
	if err != nil {
		t.Fatalf("RoundTrip() error = %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Retry-After"); got != "" {
		t.Errorf("Retry-After = %q, want it left unset", got)
	}
}
