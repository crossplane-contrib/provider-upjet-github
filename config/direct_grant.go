/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// directGrantLookup matches directgrant.Lookup. It is injected so
// wrapReadForDirectGrant is testable without an HTTP fixture; production
// wiring supplies directgrant.Lookup itself.
type directGrantLookup func(ctx context.Context, meta any, repository, login string) (directgrant.DirectGrant, error)

// directGrantLookupTimeout bounds the GraphQL lookup this wrapper makes on
// every Observe for every collaborator managed resource. The legacy
// schema.ReadFunc signature gives this wrapper no caller context to inherit
// a deadline from, but nothing prevents bounding its own call: without a
// timeout, a hung request blocks the reconciling goroutine indefinitely. A
// timeout surfaces as an ordinary lookup error, which the fail-safe path
// below already handles correctly (state left untouched, ID never cleared).
// 30s is generous for a single GraphQL request and short enough that a
// stuck call does not stall a reconcile for long.
const directGrantLookupTimeout = 30 * time.Second

// wrapReadForDirectGrant wraps github_repository_collaborator's Read so that
// its result is corrected against a direct-grant lookup (directgrant.Lookup)
// rather than trusted as-is.
//
// Upstream's own Read lists collaborators with no Affiliation filter, so it
// reports every login with any access to the repository -- including access
// inherited purely through team or organization membership -- as though it
// were the direct grant this managed resource represents. Two consequences
// follow: (1) Observe never reports the resource as gone once team-inherited
// access exists, wedging deletion (and blocking Create from ever firing) for
// an MR whose direct grant was already revoked; and (2) REST's role_name is
// the *effective* role across every source, not the direct grant's own role,
// so upstream can write the wrong value into the permission attribute
// whenever a broader team/org role masks a narrower (or different) direct
// grant.
//
// This wrapper corrects both: if no direct grant exists, it clears the ID
// (SetId("")) so the managed resource can delete cleanly or be recreated; if
// one does exist, it overwrites permission with the direct grant's own role
// name, mapped the way upstream's getPermission maps it
// (github/util_permissions.go:10 at terraform-provider-github v6.13.0):
// read -> pull, write -> push, every other role name (including custom
// repository roles) unchanged.
//
// Fail-safe direction: a lookup error must never clear the ID. Doing so
// would drop the delete finalizer across every collaborator managed
// resource on a single flaky GraphQL call. On lookup error this wrapper
// returns nil and leaves state exactly as orig set it.
//
// invitation_id complicates what "no direct grant found" means. Upstream
// sets it when AddCollaborator created a pending invitation rather than a
// completed collaborator, but findRepoInvitation only lists still-open
// invitations, so once accepted, upstream's Read never touches the field
// again -- and Plugin-SDK semantics mean a field Read never Sets falls back
// to the persisted state value, not to empty, so it stays non-empty in state
// indefinitely after acceptance. The lookup cannot be skipped just because
// invitation_id is set: that would restore the effective-role bug for every
// collaborator ever invited. Instead:
//   - Exists: true clears invitation_id unconditionally, before the
//     empty-roleName guard below runs -- the grant existing proves the
//     invitation was accepted regardless of whether roleName is usable, and
//     this is what makes the signal self-resetting for the next revocation.
//   - Exists: false with invitation_id still set is ambiguous: a pending
//     invitee is believed not to appear as a Repository-sourced
//     permissionSources edge -- inferred from the resource code, not
//     verified against live GraphQL, since no such invitation was available
//     to test against -- so it is treated as still-pending -- keep the ID,
//     change nothing.
//   - Exists: false with invitation_id empty is unambiguous: SetId("").
//
// nolint:staticcheck // SA1019: github_repository_collaborator defines the
// legacy schema.Resource.Read field (not ReadContext), and the SDK forbids
// setting both, so the wrapper must operate on the deprecated ReadFunc.
func wrapReadForDirectGrant(orig schema.ReadFunc, lookup directGrantLookup) schema.ReadFunc { //nolint:staticcheck // SA1019: see above.
	return func(d *schema.ResourceData, meta interface{}) error {
		if err := orig(d, meta); err != nil {
			return err
		}
		if d.Id() == "" {
			// Upstream's own Read already decided the resource is gone
			// (e.g. neither an invitation nor a collaborator was found, or
			// the parent repository 404'd). Nothing left to correct.
			return nil
		}

		repository := d.Get("repository").(string)
		login := d.Get("username").(string)

		ctx, cancel := context.WithTimeout(context.Background(), directGrantLookupTimeout)
		defer cancel()

		grant, err := lookup(ctx, meta, repository, login)
		if err != nil {
			// Fail-safe: leave state exactly as orig set it, and log it --
			// otherwise a lookup that always fails is indistinguishable from
			// this fix working, since both leave upstream's answer standing.
			// directgrant.Warn routes to the crossplane-runtime Logger
			// TerraformSetupBuilder installs (internal/clients/github.go); a
			// legacy schema.ReadFunc has no logger of its own to use instead.
			directgrant.Warn("direct-grant lookup failed; leaving upstream's collaborator read unchanged (permission may be the effective role, and a revoked direct grant will not be detected)",
				"login", login, "repository", repository, "error", err)
			return nil
		}
		if !grant.Exists {
			if invitationID, _ := d.Get("invitation_id").(string); invitationID != "" {
				// A still-pending invitee is believed to look exactly like
				// this, so treat Exists: false as pending rather than gone.
				return nil
			}
			d.SetId("")
			return nil
		}
		// Clearing a stale invitation_id must run before the empty-roleName
		// guard below, and unconditionally: the grant existing proves any
		// recorded invitation was accepted regardless of whether roleName is
		// usable, and clearing it here is what makes the Exists: false branch
		// above self-resetting.
		if invitationID, _ := d.Get("invitation_id").(string); invitationID != "" {
			if err := d.Set("invitation_id", ""); err != nil {
				return err
			}
		}
		if grant.RoleName == "" {
			// permission is ForceNew; writing "" here would diff against any
			// real spec value and force a delete+recreate of a live
			// collaborator. Fail-safe: leave it as orig set it.
			return nil
		}
		return d.Set("permission", mapDirectGrantRoleName(grant.RoleName))
	}
}

// mapDirectGrantRoleName mirrors upstream's getPermission
// (github/util_permissions.go:10 at terraform-provider-github v6.13.0):
// some GitHub API routes express permission as "read"/"write"/"admin",
// others as "pull"/"push"/"admin". A direct grant's role name arrives in the
// former vocabulary for the two that differ; every other role name --
// "admin", "triage", "maintain", and any custom repository role -- passes
// through unchanged. Compared case-insensitively, since GitHub role names
// (like logins) are case-insensitive.
// rolePermissionPush is the schema `permission` value upstream's
// getPermission maps role "write" to (github/util_permissions.go:10 at
// terraform-provider-github v6.13.0). Named so it can be shared with
// direct_grant_test.go's many assertions against it, rather than repeating
// the literal past goconst's threshold.
const rolePermissionPush = "push"

func mapDirectGrantRoleName(roleName string) string {
	switch strings.ToLower(roleName) {
	case "read":
		return "pull"
	case "write":
		return rolePermissionPush
	default:
		return roleName
	}
}
