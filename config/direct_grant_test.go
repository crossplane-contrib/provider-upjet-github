package config

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// collaboratorTestResourceData builds a *schema.ResourceData shaped like
// github_repository_collaborator's schema (repository, username, permission,
// invitation_id), pre-populated with a repository/username pair and, unless
// id is "", an ID.
func collaboratorTestResourceData(t *testing.T, id string) *schema.ResourceData {
	t.Helper()
	r := &schema.Resource{Schema: map[string]*schema.Schema{
		"repository":    {Type: schema.TypeString, Optional: true},
		"username":      {Type: schema.TypeString, Optional: true},
		"permission":    {Type: schema.TypeString, Optional: true},
		"invitation_id": {Type: schema.TypeString, Optional: true},
	}}
	d := r.TestResourceData()
	if id != "" {
		d.SetId(id)
	}
	if err := d.Set("repository", "some-owner/some-repo"); err != nil {
		t.Fatalf("set repository: %v", err)
	}
	if err := d.Set("username", "some-user"); err != nil {
		t.Fatalf("set username: %v", err)
	}
	return d
}

// passthroughOrig simulates upstream's Read having already run successfully
// and left d exactly as it found it (the common case: nothing changed).
func passthroughOrig(_ *schema.ResourceData, _ interface{}) error {
	return nil
}

// A direct grant gone, with team-inherited access retained (lookup ->
// Exists: false), must clear the ID so Observe reports ResourceExists: false
// and deletion (or re-Create) can proceed, instead of wedging forever on
// access that was never a direct grant.
func TestWrapReadForDirectGrant_ExistsFalseClearsID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: false}, nil
	}

	wrapped := wrapReadForDirectGrant(passthroughOrig, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to be cleared, got %q", d.Id())
	}
}

// The identical fixture as ExistsFalseClearsID, but with the lookup
// reporting Exists: true, must keep its ID. Paired with that test, this
// proves the wrapper's SetId("") is actually conditioned on the lookup
// result rather than firing unconditionally (or never). The permission
// assertion is also load-bearing here, not incidental: against a passthrough
// (orig-only, no wrapper logic) permission would stay unset, so this is what
// makes the test fail without a real wrapper rather than passing trivially
// because nothing ever touches the ID either way.
func TestWrapReadForDirectGrant_ExistsTrueKeepsID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: rolePermissionPush}, nil
	}

	wrapped := wrapReadForDirectGrant(passthroughOrig, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
	}
	if got := d.Get("permission").(string); got != rolePermissionPush {
		t.Fatalf("expected permission %q, got %q", rolePermissionPush, got)
	}
}

// A direct grant with a non-mapped role name (maintain is not one of
// upstream's read/write REST vocabulary) is written to state verbatim.
func TestWrapReadForDirectGrant_SetsRoleNameAsPermission(t *testing.T) {
	d := collaboratorTestResourceData(t, "some-owner/some-repo:some-user")

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: "maintain"}, nil
	}

	wrapped := wrapReadForDirectGrant(passthroughOrig, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got := d.Get("permission").(string); got != "maintain" {
		t.Fatalf("expected permission %q, got %q", "maintain", got)
	}
}

// The role-name mapping must mirror upstream's getPermission
// (github/util_permissions.go:10 at terraform-provider-github v6.13.0):
// read->pull, write->push, everything else (including custom repository
// roles) passed through unchanged. Compared case-insensitively.
func TestWrapReadForDirectGrant_RoleNameMapping(t *testing.T) {
	cases := map[string]struct {
		roleName string
		want     string
	}{
		"read maps to pull":                       {roleName: "read", want: "pull"},
		"write maps to push":                      {roleName: "write", want: rolePermissionPush},
		"custom role name passes through":         {roleName: "Custom Repo Role", want: "Custom Repo Role"},
		"uppercase READ maps case-insensitively":  {roleName: "READ", want: "pull"},
		"uppercase WRITE maps case-insensitively": {roleName: "WRITE", want: rolePermissionPush},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := collaboratorTestResourceData(t, "some-owner/some-repo:some-user")

			lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
				return directgrant.DirectGrant{Exists: true, RoleName: tc.roleName}, nil
			}

			wrapped := wrapReadForDirectGrant(passthroughOrig, lookup)
			if err := wrapped(d, nil); err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
			if got := d.Get("permission").(string); got != tc.want {
				t.Fatalf("expected permission %q, got %q", tc.want, got)
			}
		})
	}
}

// The highest-stakes case: a lookup error must never clear the ID. Clearing
// on error would drop the delete finalizer across every collaborator
// managed resource on a single flaky GraphQL call. The wrapper must leave
// state exactly as orig set it: ID preserved, and permission left at
// whatever orig wrote (here, simulating upstream's own REST-derived
// permission value that this workaround exists to correct -- but on lookup
// failure we degrade to that known-buggy value rather than guess).
//
// The lookupCalled assertion is load-bearing: without it, a passthrough that
// never calls lookup at all would satisfy the ID/permission checks
// trivially (nothing ever touched them). Asserting the lookup really ran
// before failing is what proves the preservation is deliberate, not absent.
func TestWrapReadForDirectGrant_LookupErrorPreservesState(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	origSetPermission := func(d *schema.ResourceData, _ interface{}) error {
		// Simulates upstream's Read having already set permission from its
		// own (effective-role) REST lookup before our wrapper runs.
		return d.Set("permission", "admin")
	}

	lookupCalled := false
	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		lookupCalled = true
		return directgrant.DirectGrant{}, errors.New("GraphQL request failed: 503 Service Unavailable")
	}

	wrapped := wrapReadForDirectGrant(origSetPermission, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error (fail-safe degrade, not a hard error), got %v", err)
	}
	if !lookupCalled {
		t.Fatalf("expected the lookup to have been called")
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q -- a lookup error must never clear the ID", startingID, d.Id())
	}
	if got := d.Get("permission").(string); got != "admin" {
		t.Fatalf("expected permission to remain %q (whatever orig set), got %q", "admin", got)
	}
}

// orig returning an error must be returned unchanged, with no lookup call at
// all -- this wrapper is only responsible for correcting a *successful*
// Read's result against the direct-grant lookup, not for reinterpreting
// orig's own errors.
//
// Deliberately passes against a bare `return orig` stub: on this input (orig
// errored), the required behaviour -- return orig's error unchanged, call
// lookup zero times -- is identical to a no-op, so no assertion here can
// distinguish the two. It was instead verified against a stub that
// implements the mapping/clearing logic but omits this ordering guard
// (ignores orig's error and proceeds anyway), which is the realistic bug
// this test catches.
func TestWrapReadForDirectGrant_OrigErrorPassedThroughUnchanged(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	wantErr := errors.New("some transport failure")
	origErr := func(_ *schema.ResourceData, _ interface{}) error {
		return wantErr
	}

	lookupCalled := false
	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		lookupCalled = true
		return directgrant.DirectGrant{}, nil
	}

	wrapped := wrapReadForDirectGrant(origErr, lookup)
	gotErr := wrapped(d, nil)

	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("expected error %v to be returned unchanged, got %v", wantErr, gotErr)
	}
	if lookupCalled {
		t.Fatalf("expected lookup not to be called when orig errors")
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
	}
}

// When orig itself already cleared the ID (upstream decided the resource is
// gone via its own REST lookup), the wrapper must not call the lookup at all
// and must leave the ID empty.
//
// Same exception as OrigErrorPassedThroughUnchanged, for the same reason: on
// this input, "do nothing" is the only correct behaviour, so a `return orig`
// stub cannot be distinguished from a correct implementation here. Verified
// instead against the same guard-less stub (there, it proceeds to call
// lookup even though the ID is already empty).
func TestWrapReadForDirectGrant_OrigClearedID_LookupNeverCalled(t *testing.T) {
	d := collaboratorTestResourceData(t, "some-owner/some-repo:some-user")

	origClearsID := func(d *schema.ResourceData, _ interface{}) error {
		d.SetId("")
		return nil
	}

	lookupCalled := false
	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		lookupCalled = true
		return directgrant.DirectGrant{}, nil
	}

	wrapped := wrapReadForDirectGrant(origClearsID, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if lookupCalled {
		t.Fatalf("expected lookup not to be called when orig already cleared the ID")
	}
	if d.Id() != "" {
		t.Fatalf("expected ID to remain empty, got %q", d.Id())
	}
}

// directgrant.match models roleName as a nullable GraphQL field, and a null
// decodes to Exists: true, RoleName: "". permission is ForceNew, so writing
// "" would diff against any real spec value and force a delete and recreate
// of a live collaborator. This must fail safe the same direction as a lookup
// error: leave permission exactly as orig set it.
func TestWrapReadForDirectGrant_EmptyRoleNamePreservesPermission(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	origSetPermission := func(d *schema.ResourceData, _ interface{}) error {
		return d.Set("permission", rolePermissionPush)
	}

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: ""}, nil
	}

	wrapped := wrapReadForDirectGrant(origSetPermission, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
	}
	if got := d.Get("permission").(string); got != rolePermissionPush {
		t.Fatalf("expected permission to remain %q (whatever orig set), got %q -- an empty role name must not overwrite it", rolePermissionPush, got)
	}
}

// invitation_id, once set by upstream's findRepoInvitation branch, is never
// Set back to empty by upstream's own Read once the invitation is accepted
// (ListInvitations only lists still-open invitations, so the
// accepted-collaborator branch never touches the field again). Plugin-SDK
// ResourceData semantics mean a field a Read never Sets falls back to the
// persisted state value, not to empty, so invitation_id stays non-empty in
// state indefinitely after acceptance -- skipping the lookup whenever it is
// set would therefore read as "still pending" forever and silently restore
// the upstream bug for every collaborator that was ever invited.
//
// The lookup must always run and, on a confirmed direct grant, clear the
// stale invitation_id alongside setting permission -- a confirmed grant
// means any recorded invitation has been accepted and is gone.
func TestWrapReadForDirectGrant_PreviouslyPendingNowAccepted_ClearsStaleInvitationID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)
	if err := d.Set("invitation_id", "123456"); err != nil {
		t.Fatalf("set invitation_id: %v", err)
	}

	// orig here represents upstream's Read having taken the
	// accepted-collaborator branch (ListCollaborators found the user), which
	// never Sets invitation_id -- so it falls back to the persisted state
	// value above, exactly as ResourceData does for an untouched field.
	origAcceptedCollaborator := passthroughOrig

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: rolePermissionPush}, nil
	}

	wrapped := wrapReadForDirectGrant(origAcceptedCollaborator, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q", startingID, d.Id())
	}
	if got := d.Get("permission").(string); got != rolePermissionPush {
		t.Fatalf("expected permission %q, got %q", rolePermissionPush, got)
	}
	if got := d.Get("invitation_id").(string); got != "" {
		t.Fatalf("expected stale invitation_id to be cleared once a direct grant is confirmed, got %q", got)
	}
}

// A genuinely still-pending invitation (Exists: false from the lookup,
// invitation_id still non-empty) must keep both its ID and its
// invitation_id untouched: the lookup ran, but a still-pending invitee
// legitimately has no Repository-sourced permission source yet, so "not
// found" here does not mean "gone."
func TestWrapReadForDirectGrant_StillPending_KeepsIDAndInvitationID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)
	if err := d.Set("invitation_id", "123456"); err != nil {
		t.Fatalf("set invitation_id: %v", err)
	}

	origSetsInvitation := func(d *schema.ResourceData, _ interface{}) error {
		return d.Set("permission", rolePermissionPush)
	}

	lookupCalled := false
	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		lookupCalled = true
		return directgrant.DirectGrant{Exists: false}, nil
	}

	wrapped := wrapReadForDirectGrant(origSetsInvitation, lookup)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !lookupCalled {
		t.Fatalf("expected the lookup to have been called")
	}
	if d.Id() != startingID {
		t.Fatalf("expected ID to remain %q, got %q -- a still-pending invitation must not be cleared", startingID, d.Id())
	}
	if got := d.Get("invitation_id").(string); got != "123456" {
		t.Fatalf("expected invitation_id to remain %q, got %q", "123456", got)
	}
}

// The property that makes the invitation_id guard self-resetting rather than
// stale-persistent: a resource that was once pending, then had its grant
// confirmed (clearing invitation_id), must -- on a later read where the
// lookup reports Exists: false -- reap normally: ID cleared, same as any
// collaborator with no invitation history. Run as two sequential wrapped
// calls against the same ResourceData to model the two reconciles in
// sequence.
func TestWrapReadForDirectGrant_StaleInvitationClearedThenRevoked_ClearsID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)
	if err := d.Set("invitation_id", "123456"); err != nil {
		t.Fatalf("set invitation_id: %v", err)
	}

	confirmedGrant := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: rolePermissionPush}, nil
	}
	wrapped := wrapReadForDirectGrant(passthroughOrig, confirmedGrant)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("first read: expected nil error, got %v", err)
	}
	if got := d.Get("invitation_id").(string); got != "" {
		t.Fatalf("first read: expected invitation_id to be cleared, got %q", got)
	}

	revoked := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: false}, nil
	}
	wrapped = wrapReadForDirectGrant(passthroughOrig, revoked)
	if err := wrapped(d, nil); err != nil {
		t.Fatalf("second read: expected nil error, got %v", err)
	}
	if d.Id() != "" {
		t.Fatalf("second read: expected ID to be cleared now that the grant is revoked, got %q", d.Id())
	}
}

// A confirmed direct grant with an empty roleName must still clear a stale
// invitation_id. The empty-roleName guard exists only to protect the
// ForceNew permission attribute from a guessed value -- it says nothing
// about the invitation, which the confirmed grant has already proven
// accepted. Returning before the clearing step leaves invitation_id set
// forever, and the still-pending branch above then reads that leftover
// value as "still pending" and never reaps the resource: exactly the wedge
// this wrapper exists to remove, reachable whenever GraphQL reports a null
// roleName.
func TestWrapReadForDirectGrant_EmptyRoleNameStillClearsStaleInvitationID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)
	if err := d.Set("permission", rolePermissionPush); err != nil {
		t.Fatalf("set permission: %v", err)
	}
	if err := d.Set("invitation_id", "12345678"); err != nil {
		t.Fatalf("set invitation_id: %v", err)
	}

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: ""}, nil
	}

	if err := wrapReadForDirectGrant(passthroughOrig, lookup)(d, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Id() != startingID {
		t.Fatalf("ID = %q, want it preserved as %q", d.Id(), startingID)
	}
	// permission is left as orig set it: an empty roleName is not actionable
	// against a ForceNew attribute.
	if got := d.Get("permission").(string); got != rolePermissionPush {
		t.Fatalf("permission = %q, want it left as %q", got, rolePermissionPush)
	}
	if got := d.Get("invitation_id").(string); got != "" {
		t.Fatalf("invitation_id = %q, want it cleared: the direct grant is confirmed, so the invitation was accepted and the field is stale", got)
	}
}

// Once the stale invitation_id has been cleared on an empty-roleName read, a
// later revocation must reap the resource rather than reading the leftover
// invitation as "still pending".
func TestWrapReadForDirectGrant_EmptyRoleNameThenRevoked_ClearsID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)
	if err := d.Set("invitation_id", "12345678"); err != nil {
		t.Fatalf("set invitation_id: %v", err)
	}

	granted := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: ""}, nil
	}
	if err := wrapReadForDirectGrant(passthroughOrig, granted)(d, nil); err != nil {
		t.Fatalf("first read: unexpected error: %v", err)
	}

	revoked := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{}, nil
	}
	if err := wrapReadForDirectGrant(passthroughOrig, revoked)(d, nil); err != nil {
		t.Fatalf("second read: unexpected error: %v", err)
	}
	if d.Id() != "" {
		t.Fatalf("ID = %q, want it cleared after the grant was revoked", d.Id())
	}
}

// The fail-safe branch must say something. Swallowing every lookup error
// with no log line makes a credential withdrawal safe but indistinguishable
// from the fix working. The line has to carry enough identity to be
// actionable.
func TestWrapReadForDirectGrant_LookupErrorIsLogged(t *testing.T) {
	d := collaboratorTestResourceData(t, "some-owner/some-repo:some-user")

	var logged []string
	directgrant.SetLogger(func(msg string, keysAndValues ...any) {
		logged = append(logged, fmt.Sprint(append([]any{msg}, keysAndValues...)...))
	})
	t.Cleanup(func() { directgrant.SetLogger(nil) })

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{}, errors.New("no GraphQL client registered for this ProviderConfig")
	}
	if err := wrapReadForDirectGrant(passthroughOrig, lookup)(d, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logged) == 0 {
		t.Fatal("the fail-safe branch logged nothing: a credential withdrawal would silently disable the fix")
	}
	line := logged[0]
	for _, want := range []string{"some-owner/some-repo", "some-user", "no GraphQL client registered"} {
		if !strings.Contains(line, want) {
			t.Fatalf("fail-safe log line %q does not contain %q", line, want)
		}
	}
}

// The happy path must not log. A warning on every successful
// Observe would be noise at this provider's reconcile rate and would train
// operators to ignore the line that matters.
func TestWrapReadForDirectGrant_SuccessDoesNotLog(t *testing.T) {
	d := collaboratorTestResourceData(t, "some-owner/some-repo:some-user")

	var logged int
	directgrant.SetLogger(func(string, ...any) { logged++ })
	t.Cleanup(func() { directgrant.SetLogger(nil) })

	lookup := func(_ context.Context, _ any, _, _ string) (directgrant.DirectGrant, error) {
		return directgrant.DirectGrant{Exists: true, RoleName: "write"}, nil
	}
	if err := wrapReadForDirectGrant(passthroughOrig, lookup)(d, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if logged != 0 {
		t.Fatalf("expected no log output on the success path, got %d calls", logged)
	}
}

// The property that must not regress, asserted end to end against the real
// directgrant.Lookup rather than a stub. Every GraphQL failure mode
// -- transport, status, body, GraphQL errors array, credential, truncation --
// must leave the ID standing. A single one of these reaching SetId("") drops
// the delete finalizer across every collaborator managed resource at once.
//
// The per-layer tests each assert half of this; only wiring the two together
// shows that no combination of them clears an ID.
func TestWrapReadForDirectGrant_NoGraphQLFailureClearsID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"

	cases := map[string]struct {
		status int
		body   string
		token  directgrant.TokenProvider
	}{
		"HTTP 500": {
			status: http.StatusInternalServerError, body: `{"message":"Internal Server Error"}`,
		},
		"HTTP 401 after a credential withdrawal": {
			status: http.StatusUnauthorized, body: `{"message":"Bad credentials"}`,
		},
		"HTTP 403 secondary rate limit": {
			status: http.StatusForbidden, body: `{"message":"You have exceeded a secondary rate limit"}`,
		},
		"GraphQL INSUFFICIENT_SCOPES": {
			status: http.StatusOK,
			body:   `{"data":{"repository":null},"errors":[{"type":"INSUFFICIENT_SCOPES","message":"admin:org required"}]}`,
		},
		"GraphQL NOT_FOUND": {
			status: http.StatusOK,
			body:   `{"data":{"repository":null},"errors":[{"type":"NOT_FOUND","message":"Could not resolve to a Repository"}]}`,
		},
		"null repository with no errors array": {
			status: http.StatusOK, body: `{"data":{"repository":null}}`,
		},
		"an intermediary's unrelated 200": {
			status: http.StatusOK, body: `{"message":"proxy is warming up"}`,
		},
		"a non-JSON body": {
			status: http.StatusOK, body: `<html>502 Bad Gateway</html>`,
		},
		"a truncated result with no exact match": {
			status: http.StatusOK,
			body: `{"data":{"repository":{"collaborators":{"pageInfo":{"hasNextPage":true},"edges":[
				{"node":{"login":"some-user-bot"},"permissionSources":[{"roleName":"write","source":{"__typename":"Repository"}}]}
			]}}}}`,
		},
		"a token provider that fails": {
			status: http.StatusOK,
			body:   `{"data":{"repository":{"collaborators":{"edges":[]}}}}`,
			token: func(context.Context) (string, error) {
				return "", errors.New("cannot mint a GitHub App installation token")
			},
		},
		"a token provider yielding nothing": {
			status: http.StatusOK,
			body:   `{"data":{"repository":{"collaborators":{"edges":[]}}}}`,
			token:  func(context.Context) (string, error) { return "", nil },
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			t.Cleanup(srv.Close)

			token := tc.token
			if token == nil {
				token = directgrant.StaticToken("test-token")
			}
			meta := new(int)
			directgrant.Register(meta, srv.URL, token, "some-owner")
			t.Cleanup(func() { directgrant.Deregister(meta) })

			d := collaboratorTestResourceData(t, startingID)
			if err := d.Set("permission", rolePermissionPush); err != nil {
				t.Fatalf("set permission: %v", err)
			}

			if err := wrapReadForDirectGrant(passthroughOrig, directgrant.Lookup)(d, meta); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if d.Id() != startingID {
				t.Fatalf("ID = %q, want it preserved as %q: a GraphQL failure must never read as a revoked grant", d.Id(), startingID)
			}
			if got := d.Get("permission").(string); got != rolePermissionPush {
				t.Fatalf("permission = %q, want it left as %q", got, rolePermissionPush)
			}
		})
	}
}

// An unregistered meta must also preserve the ID, and say so.
func TestWrapReadForDirectGrant_UnregisteredMetaPreservesID(t *testing.T) {
	const startingID = "some-owner/some-repo:some-user"
	d := collaboratorTestResourceData(t, startingID)

	var logged []string
	directgrant.SetLogger(func(msg string, keysAndValues ...any) {
		logged = append(logged, fmt.Sprint(append([]any{msg}, keysAndValues...)...))
	})
	t.Cleanup(func() { directgrant.SetLogger(nil) })

	if err := wrapReadForDirectGrant(passthroughOrig, directgrant.Lookup)(d, new(int)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Id() != startingID {
		t.Fatalf("ID = %q, want it preserved as %q", d.Id(), startingID)
	}
	if len(logged) == 0 || !strings.Contains(logged[0], "no GraphQL client registered") {
		t.Fatalf("expected the fail-safe to log the unregistered-meta cause, got %v", logged)
	}
}
