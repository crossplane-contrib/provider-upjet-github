package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v88/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	tpg "github.com/integrations/terraform-provider-github/v6/github"
)

// The wrappers skip unknown keys, so a typo or a rename would be a silent no-op
// with no build or test failure. Assert every listed resource exists in the real
// provider and exposes the field its wrapper replaces.
func TestParentDeletionWorkaroundResourcesAreWired(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	for _, name := range parentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok {
			t.Errorf("%q is not a registered terraform-provider-github resource (silently skipped — typo or renamed key?)", name)
			continue
		}
		if r.Read == nil { //nolint:staticcheck // SA1019: verifying the legacy Read field the wrapper depends on is present.
			t.Errorf("%q has no legacy Read func; withParentDeletionWorkaround would silently skip it", name)
		}
	}
}

// A typed 404 becomes the "gone" signal; every other error propagates unchanged.
func TestWrapReadForParentDeletion(t *testing.T) {
	const startingID = "some-owner/some-repo"

	notFound := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusNotFound},
		Message:  "Not Found",
	}
	forbidden := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusForbidden},
		Message:  "Forbidden",
	}
	serverErr := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusInternalServerError},
		Message:  "Server Error",
	}
	genericErr := errors.New("dial tcp: connection refused")

	cases := map[string]struct {
		underlyingErr error
		wantErr       error
		wantClearedID bool
	}{
		"typed 404 clears ID and returns nil": {
			underlyingErr: notFound,
			wantErr:       nil,
			wantClearedID: true,
		},
		"typed 403 is propagated and keeps ID": {
			underlyingErr: forbidden,
			wantErr:       forbidden,
			wantClearedID: false,
		},
		"typed 500 is propagated and keeps ID": {
			underlyingErr: serverErr,
			wantErr:       serverErr,
			wantClearedID: false,
		},
		"non-github error is propagated and keeps ID": {
			underlyingErr: genericErr,
			wantErr:       genericErr,
			wantClearedID: false,
		},
		"nil error passes through and keeps ID": {
			underlyingErr: nil,
			wantErr:       nil,
			wantClearedID: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			wrapped := wrapReadForParentDeletion(func(_ *schema.ResourceData, _ interface{}) error {
				return tc.underlyingErr
			})

			gotErr := wrapped(d, nil)

			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("error mismatch: got %v, want %v", gotErr, tc.wantErr)
			}
			if tc.wantClearedID {
				if d.Id() != "" {
					t.Fatalf("expected ID to be cleared, got %q", d.Id())
				}
			} else {
				if d.Id() != startingID {
					t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
				}
			}
		})
	}
}

// Still detected if an upstream Read starts wrapping rather than returning the
// raw client error.
func TestWrapReadForParentDeletion_WrappedTyped404(t *testing.T) {
	notFound := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusNotFound},
		Message:  "Not Found",
	}
	d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
	d.SetId("some-owner/some-repo")

	wrapped := wrapReadForParentDeletion(func(_ *schema.ResourceData, _ interface{}) error {
		return fmt.Errorf("reading issue labels: %w", notFound)
	})
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected wrapped typed 404 to be treated as gone (nil error), got %v", err)
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to be cleared for wrapped typed 404, got %q", d.Id())
	}
}

// ghErrDiag renders an error the way the upstream ReadContext funcs do. The
// Request is attached so the summary carries the "<method> <url>: <status>" form.
func ghErrDiag(status int, message string) diag.Diagnostics {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.github.com/orgs/o/teams/some-team", nil)
	return diag.FromErr(&github.ErrorResponse{
		Response: &http.Response{StatusCode: status, Request: req},
		Message:  message,
	})
}

// A 404 surfaced as an error diagnostic becomes the "gone" signal; every other
// error propagates unchanged.
func TestWrapReadContextForParentDeletion(t *testing.T) {
	const startingID = "some-team:some-repo"

	cases := map[string]struct {
		diags         diag.Diagnostics
		wantErr       bool
		wantClearedID bool
	}{
		"404 clears ID and returns no error": {
			diags:         ghErrDiag(http.StatusNotFound, "Not Found"),
			wantErr:       false,
			wantClearedID: true,
		},
		"403 is propagated and keeps ID": {
			diags:         ghErrDiag(http.StatusForbidden, "Forbidden"),
			wantErr:       true,
			wantClearedID: false,
		},
		"500 is propagated and keeps ID": {
			diags:         ghErrDiag(http.StatusInternalServerError, "Server Error"),
			wantErr:       true,
			wantClearedID: false,
		},
		"non-github error is propagated and keeps ID": {
			diags:         diag.FromErr(errors.New("dial tcp: connection refused")),
			wantErr:       true,
			wantClearedID: false,
		},
		"no error passes through and keeps ID": {
			diags:         nil,
			wantErr:       false,
			wantClearedID: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			wrapped := wrapReadContextForParentDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
				return tc.diags
			})

			gotDiags := wrapped(context.Background(), d, nil)

			if gotDiags.HasError() != tc.wantErr {
				t.Fatalf("HasError() = %v, want %v (diags: %v)", gotDiags.HasError(), tc.wantErr, gotDiags)
			}
			// When propagating, the diagnostics must be returned unchanged (not
			// swapped for a different error).
			if tc.wantErr {
				if len(gotDiags) != len(tc.diags) || (len(gotDiags) > 0 && gotDiags[0].Summary != tc.diags[0].Summary) {
					t.Fatalf("expected diags propagated unchanged, got %v want %v", gotDiags, tc.diags)
				}
			}
			if tc.wantClearedID {
				if d.Id() != "" {
					t.Fatalf("expected ID to be cleared, got %q", d.Id())
				}
			} else {
				if d.Id() != startingID {
					t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
				}
			}
		})
	}
}

// Wiring guard for the team list, as above.
func TestTeamParentDeletionWorkaroundResourcesAreWired(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	for _, name := range teamParentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok {
			t.Errorf("%q is not a registered terraform-provider-github resource (silently skipped — typo or renamed key?)", name)
			continue
		}
		if r.ReadContext == nil {
			t.Errorf("%q has no ReadContext func; withParentDeletionWorkaround would silently skip it", name)
		}
	}
}

// Wiring guard for the commit-lookup list, as above.
func TestCommitLookupParentDeletionWorkaroundResourcesAreWired(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	for _, name := range commitLookupParentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok {
			t.Errorf("%q is not a registered terraform-provider-github resource (silently skipped — typo or renamed key?)", name)
			continue
		}
		if r.ReadContext == nil {
			t.Errorf("%q has no ReadContext func; withParentDeletionWorkaround would silently skip it", name)
		}
	}
}

// The wiring guard only proves the resource is present, so compare the function
// pointer to prove it is actually replaced.
func TestWithParentDeletionWorkaroundReplacesCommitLookupReadContext(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()

	before := map[string]string{}
	for _, name := range commitLookupParentDeletionWorkaroundResources {
		before[name] = fmt.Sprintf("%p", p.ResourcesMap[name].ReadContext)
	}

	withParentDeletionWorkaround(p)

	for _, name := range commitLookupParentDeletionWorkaroundResources {
		after := fmt.Sprintf("%p", p.ResourcesMap[name].ReadContext)
		if after == before[name] {
			t.Errorf("%q ReadContext was not wrapped (pointer unchanged: %s)", name, after)
		}
	}
}

// Both go-github renderings must be recognised: with a request attached
// ("<method> <url>: 404 ...") and without ("404 ...").
func TestWrapReadContextForParentDeletion_BothRenderedForms(t *testing.T) {
	cases := map[string]diag.Diagnostics{
		"with request":    ghErrDiag(http.StatusNotFound, "Not Found"),
		"without request": diag.FromErr(&github.ErrorResponse{Response: &http.Response{StatusCode: http.StatusNotFound}, Message: "Not Found"}),
	}

	for name, diags := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId("some-repo/some-file:main")

			wrapped := wrapReadContextForParentDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
				return diags
			})
			got := wrapped(context.Background(), d, nil)

			if got.HasError() {
				t.Fatalf("expected 404 to be treated as gone, got %v", got)
			}
			if d.Id() != "" {
				t.Fatalf("expected ID to be cleared, got %q", d.Id())
			}
		})
	}
}

// A file path containing a bare "404" is not an HTTP status. Relaxing the anchor
// in isGitHubNotFoundDiag to any space-delimited "404" fails this test.
func TestWrapReadContextForParentDeletion_PathContaining404IsNotGone(t *testing.T) {
	const startingID = "some-repo/docs/error 404 page.html:main"

	cases := map[string]string{
		"cannot find file with 404 in the path": "cannot find file docs/error 404 page.html in repo some-owner/some-repo",
		"bare 404 token mid-message":            "unexpected failure reading 404 handler config",
	}

	for name, summary := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			in := diag.FromErr(errors.New(summary))
			wrapped := wrapReadContextForParentDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
				return in
			})
			got := wrapped(context.Background(), d, nil)

			if !got.HasError() {
				t.Fatalf("expected error to propagate, got %v", got)
			}
			if len(got) != len(in) || got[0].Summary != in[0].Summary {
				t.Fatalf("expected diags propagated unchanged, got %v want %v", got, in)
			}
			if d.Id() != startingID {
				t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
			}
		})
	}
}

// An exhausted commit list carries no HTTP status, so it must be treated as gone
// alongside a 404.
func TestWrapReadContextForCommitLookupDeletion(t *testing.T) {
	const startingID = "some-repo/docs/index.md:main"

	exhausted := diag.FromErr(errors.New("cannot find file docs/index.md in repo some-owner/some-repo"))
	wrappedExhausted := diag.FromErr(fmt.Errorf("looking for commit: %w", errors.New("cannot find file docs/index.md in repo some-owner/some-repo")))

	cases := map[string]struct {
		diags         diag.Diagnostics
		wantGone      bool
		wantPropagate bool
	}{
		"commit-list exhaustion clears ID": {
			diags:    exhausted,
			wantGone: true,
		},
		"wrapped commit-list exhaustion clears ID": {
			diags:    wrappedExhausted,
			wantGone: true,
		},
		"404 still clears ID": {
			diags:    ghErrDiag(http.StatusNotFound, "Not Found"),
			wantGone: true,
		},
		"403 is propagated and keeps ID": {
			diags:         ghErrDiag(http.StatusForbidden, "Forbidden"),
			wantPropagate: true,
		},
		"500 is propagated and keeps ID": {
			diags:         ghErrDiag(http.StatusInternalServerError, "Server Error"),
			wantPropagate: true,
		},
		"network error is propagated and keeps ID": {
			diags:         diag.FromErr(errors.New("dial tcp: connection refused")),
			wantPropagate: true,
		},
		"no error passes through and keeps ID": {
			diags: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			wrapped := wrapReadContextForCommitLookupDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
				return tc.diags
			})
			got := wrapped(context.Background(), d, nil)

			if tc.wantGone {
				if got.HasError() {
					t.Fatalf("expected gone (no diags), got %v", got)
				}
				if d.Id() != "" {
					t.Fatalf("expected ID to be cleared, got %q", d.Id())
				}
				return
			}
			if tc.wantPropagate {
				if !got.HasError() {
					t.Fatalf("expected error to propagate, got %v", got)
				}
				if len(got) != len(tc.diags) || got[0].Summary != tc.diags[0].Summary {
					t.Fatalf("expected diags propagated unchanged, got %v want %v", got, tc.diags)
				}
			}
			if d.Id() != startingID {
				t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
			}
		})
	}
}

// These quote the exhaustion phrase without being the exhaustion error. Relaxing
// afterFileCommitNotFoundPrefix to a plain substring match fails all three.
func TestWrapReadContextForCommitLookupDeletion_NearMissesPropagate(t *testing.T) {
	const startingID = "some-repo/docs/index.md:main"

	cases := map[string]string{
		// Phrase mid-summary, so the opening anchor rejects it.
		"phrase mid-summary from a user-supplied name": "branch cannot find file x in repo y/z not found in some-owner/some-repo or repository is not readable",
		// Prefix without the separator that follows it upstream.
		"prefix without the in-repo separator": "cannot find file docs/index.md",
		// Phrase quoted inside a genuine API failure.
		"phrase inside a 403 summary": "GET https://api.github.com/repos/o/r/commits: 403 cannot find file x in repo y/z []",
	}

	for name, summary := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			in := diag.FromErr(errors.New(summary))
			wrapped := wrapReadContextForCommitLookupDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
				return in
			})
			got := wrapped(context.Background(), d, nil)

			if !got.HasError() {
				t.Fatalf("expected error to propagate, got %v", got)
			}
			if len(got) != len(in) || got[0].Summary != in[0].Summary {
				t.Fatalf("expected diags propagated unchanged, got %v want %v", got, in)
			}
			if d.Id() != startingID {
				t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
			}
		})
	}
}

// Deliberately asserted in the positive direction: a file whose own path contains
// the phrase still renders as a genuine exhaustion error when it is really gone.
func TestWrapReadContextForCommitLookupDeletion_PathContainingPhraseIsGone(t *testing.T) {
	d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
	d.SetId("some-repo/cannot find file x in repo y/z:main")

	in := diag.FromErr(errors.New("cannot find file cannot find file x in repo y/z in repo some-owner/some-repo"))
	wrapped := wrapReadContextForCommitLookupDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
		return in
	})
	got := wrapped(context.Background(), d, nil)

	if got.HasError() {
		t.Fatalf("expected gone (no diags), got %v", got)
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to be cleared, got %q", d.Id())
	}
}

// Keeping the two predicates separate bounds each workaround to the resources that
// need it, so the team wrapper must not treat commit-list exhaustion as gone.
func TestWrapReadContextForParentDeletion_IgnoresFileCommitExhaustion(t *testing.T) {
	const startingID = "some-team:some-repo"

	d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
	d.SetId(startingID)

	in := diag.FromErr(errors.New("cannot find file docs/index.md in repo some-owner/some-repo"))
	wrapped := wrapReadContextForParentDeletion(func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
		return in
	})
	got := wrapped(context.Background(), d, nil)

	if !got.HasError() {
		t.Fatalf("expected error to propagate through the team wrapper, got %v", got)
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
	}
}

// As above, for the team list.
func TestWithParentDeletionWorkaroundReplacesTeamReadContext(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()

	before := map[string]string{}
	for _, name := range teamParentDeletionWorkaroundResources {
		before[name] = fmt.Sprintf("%p", p.ResourcesMap[name].ReadContext)
	}

	withParentDeletionWorkaround(p)

	for _, name := range teamParentDeletionWorkaroundResources {
		after := fmt.Sprintf("%p", p.ResourcesMap[name].ReadContext)
		if after == before[name] {
			t.Errorf("%q ReadContext was not wrapped (pointer unchanged: %s)", name, after)
		}
	}
}

func TestBranchNotProtectedWorkaroundResourcesAreWired(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()
	for _, name := range branchNotProtectedWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok {
			t.Errorf("%q is not a registered terraform-provider-github resource (silently skipped — typo or renamed key?)", name)
			continue
		}
		if r.Read == nil { //nolint:staticcheck // SA1019: verifying the legacy Read field the wrapper depends on is present.
			t.Errorf("%q has no legacy Read func; withParentDeletionWorkaround would silently skip it (migrated to ReadContext upstream?)", name)
		}
	}
}

// wrapReadForBranchNotProtected must translate go-github's
// ErrBranchNotProtected sentinel -- which upstream's errors.As on
// *github.ErrorResponse cannot see, leaving its SetId("") arm unreachable -- into
// the SDK "gone" signal, while propagating every other error unchanged.
//
// The typed 404 case is asserted as *propagated*, not translated: this wrapper
// deliberately has no StatusNotFound arm, because a real 404 keeps its
// *github.ErrorResponse and upstream clears the ID itself.
func TestWrapReadForBranchNotProtected(t *testing.T) {
	const startingID = "some-repo:main"

	notFound := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusNotFound},
		Message:  "Not Found",
	}
	forbidden := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusForbidden},
		Message:  "Forbidden",
	}
	serverErr := &github.ErrorResponse{
		Response: &http.Response{StatusCode: http.StatusInternalServerError},
		Message:  "Server Error",
	}
	genericErr := errors.New("dial tcp: connection refused")

	cases := map[string]struct {
		underlyingErr error
		wantErr       error
		wantClearedID bool
	}{
		"ErrBranchNotProtected sentinel clears ID and returns nil": {
			underlyingErr: github.ErrBranchNotProtected,
			wantErr:       nil,
			wantClearedID: true,
		},
		"typed 404 is propagated and keeps ID (upstream clears it itself)": {
			underlyingErr: notFound,
			wantErr:       notFound,
			wantClearedID: false,
		},
		"typed 403 is propagated and keeps ID": {
			underlyingErr: forbidden,
			wantErr:       forbidden,
			wantClearedID: false,
		},
		"typed 500 is propagated and keeps ID": {
			underlyingErr: serverErr,
			wantErr:       serverErr,
			wantClearedID: false,
		},
		"non-github error is propagated and keeps ID": {
			underlyingErr: genericErr,
			wantErr:       genericErr,
			wantClearedID: false,
		},
		"nil error passes through and keeps ID": {
			underlyingErr: nil,
			wantErr:       nil,
			wantClearedID: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			wrapped := wrapReadForBranchNotProtected(func(_ *schema.ResourceData, _ interface{}) error {
				return tc.underlyingErr
			})

			gotErr := wrapped(d, nil)

			if !errors.Is(gotErr, tc.wantErr) {
				t.Fatalf("error mismatch: got %v, want %v", gotErr, tc.wantErr)
			}
			if tc.wantClearedID {
				if d.Id() != "" {
					t.Fatalf("expected ID to be cleared, got %q", d.Id())
				}
			} else {
				if d.Id() != startingID {
					t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
				}
			}
		})
	}
}

// The sentinel wrapped in a chain (fmt.Errorf("...: %w", ...)) must still be
// detected, since errors.Is walks the chain. Upstream returns it bare today; this
// guards against a regression if it starts annotating the error.
func TestWrapReadForBranchNotProtected_WrappedSentinel(t *testing.T) {
	d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
	d.SetId("some-repo:main")

	wrapped := wrapReadForBranchNotProtected(func(_ *schema.ResourceData, _ interface{}) error {
		return fmt.Errorf("reading branch protection: %w", github.ErrBranchNotProtected)
	})
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected wrapped sentinel to be treated as gone (nil error), got %v", err)
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to be cleared for wrapped sentinel, got %q", d.Id())
	}
}

// The near-miss for an identity match is a *different* error value carrying the
// byte-identical message. github.ErrBranchNotProtected is a bare errors.New, so
// only its identity means "protection is absent"; an unrelated error that happens
// to render the same text does not, and treating it as gone would drop the delete
// finalizer on a live resource.
//
// This is the test that fails if anyone ever "simplifies" the wrapper to
// strings.Contains(err.Error(), "branch is not protected") -- a real temptation
// here, since the other three wrappers in this file are string matchers.
func TestWrapReadForBranchNotProtected_IdenticalMessageIsNotTheSentinel(t *testing.T) {
	const startingID = "some-repo:main"

	// Same text as github.ErrBranchNotProtected, different error value.
	impostor := errors.New("branch is not protected")

	cases := map[string]error{
		"distinct error with identical message": impostor,
		"distinct error wrapped in a chain":     fmt.Errorf("reading branch protection: %w", impostor),
		"message embedded in a larger error":    errors.New("upstream says branch is not protected, retrying"),
	}

	for name, underlying := range cases {
		t.Run(name, func(t *testing.T) {
			d := (&schema.Resource{Schema: map[string]*schema.Schema{}}).TestResourceData()
			d.SetId(startingID)

			wrapped := wrapReadForBranchNotProtected(func(_ *schema.ResourceData, _ interface{}) error {
				return underlying
			})

			gotErr := wrapped(d, nil)
			if gotErr == nil {
				t.Fatalf("expected %v to propagate, but it was treated as gone", underlying)
			}
			if gotErr.Error() != underlying.Error() {
				t.Fatalf("error was altered: got %q, want %q", gotErr, underlying)
			}
			if d.Id() != startingID {
				t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
			}
		})
	}
}

// TestWithParentDeletionWorkaroundReplacesBranchNotProtectedRead asserts the
// composition: withParentDeletionWorkaround must actually replace the resource's
// legacy Read pointer, not merely leave it wired. Comparing the pointer
// before/after catches a missing or misordered loop that the wiring test above
// would not.
func TestWithParentDeletionWorkaroundReplacesBranchNotProtectedRead(t *testing.T) {
	p := tpg.NewProvider("dev", "none")()

	before := map[string]string{}
	for _, name := range branchNotProtectedWorkaroundResources {
		before[name] = fmt.Sprintf("%p", p.ResourcesMap[name].Read) //nolint:staticcheck // SA1019: the wrapper operates on the legacy Read field.
	}

	withParentDeletionWorkaround(p)

	for _, name := range branchNotProtectedWorkaroundResources {
		after := fmt.Sprintf("%p", p.ResourcesMap[name].Read) //nolint:staticcheck // SA1019: see above.
		if after == before[name] {
			t.Errorf("%q Read was not wrapped (pointer unchanged: %s)", name, after)
		}
	}
}
