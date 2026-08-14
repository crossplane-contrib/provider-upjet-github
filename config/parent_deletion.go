/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/go-github/v88/github"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Several terraform-provider-github resources violate the terraform-plugin-sdk
// Read contract: when their parent object is gone they return an error instead of
// clearing the ID and returning nil. upjet's SDK external client surfaces that as
// an error from Observe rather than ResourceExists:false, so crossplane-runtime
// never removes the delete finalizer and the child managed resource stays in
// Terminating forever.
//
// The wrappers below turn a missing parent into the SDK's "gone" signal
// (SetId("") + nil). Every other error -- 401/403/5xx, network -- is propagated
// unchanged so a transient failure never orphans a live resource. Resources that
// already clear the ID on a 404 are excluded, and every entry here should be
// dropped as upstream fixes land. Verified against terraform-provider-github
// v6.13.0.

// parentDeletionWorkaroundResources are repository children exposing the legacy
// Read field whose reads return a bare error on a parent-repository 404.
//
// Excluded because they already clear the ID themselves: github_branch_protection
// (via its GraphQL "Could not resolve to a node with the global id" branch),
// github_repository_collaborators, github_repository_custom_property,
// github_repository_ruleset, github_branch_default and
// github_app_installation_repository. github_actions_organization_permissions is
// org-scoped, so repository deletion never wedges it.
var parentDeletionWorkaroundResources = []string{
	"github_issue_labels",
	"github_repository_dependabot_security_updates",
	"github_actions_repository_access_level",
	"github_actions_repository_permissions",
	"github_workflow_repository_permissions",
}

// teamParentDeletionWorkaroundResources are team children whose ReadContext
// resolves the parent team by slug (getTeamID / lookupTeamID ->
// Teams.GetTeamBySlug) and returns diag.FromErr on the lookup 404. Each handles a
// 404 only later, past the point that wedges.
//
// Excluded because they already clear the ID on a 404: github_team (resolves by
// numeric ID), github_emu_group_mapping and github_team_sync_group_mapping.
// github_team_settings resolves the team over GraphQL, so there is no REST 404 to
// match; a missing team is believed to come back as a null Team, inferred from the
// resource code rather than confirmed against the live API.
var teamParentDeletionWorkaroundResources = []string{
	"github_team_repository",
	"github_team_membership",
	"github_team_members",
}

// commitLookupParentDeletionWorkaroundResources are resources that guard their
// primary lookup but then make a second, unguarded one. github_repository_file is
// the only one: its content lookup is guarded, but the commit lookup that follows
// (Repositories.GetCommit when commit_sha is in state, otherwise getFileCommit)
// ends in a bare diag.FromErr for both of its "gone" outcomes -- a 404, and an
// exhausted commit list.
var commitLookupParentDeletionWorkaroundResources = []string{
	"github_repository_file",
}

// branchNotProtectedWorkaroundResources lists resources that wedge on their own
// deletion rather than a parent's. Despite living in this file, this is not a
// parent-404 family at all: nothing about the repository, branch or any parent
// is missing. The resource's own Read cannot recognise the resource as gone
// immediately after its own successful Delete.
//
// github_branch_protection_v3 is the only member. Verified against
// terraform-provider-github v6.13.0 and go-github v88.0.0:
//
//   - GitHub answers GET .../branches/<branch>/protection with 404 and the
//     message "Branch not protected" once protection is removed.
//   - go-github's GetBranchProtection detects exactly that message
//     (isBranchNotProtected -> errorResponse.Message == "Branch not protected")
//     and *substitutes* the bare sentinel github.ErrBranchNotProtected -- a
//     plain errors.New with no Response and no wrapping -- discarding the
//     *github.ErrorResponse.
//   - resourceGithubBranchProtectionV3Read then does
//     errors.As(err, &ghErr) on *github.ErrorResponse. The sentinel is not one,
//     so errors.As fails, its http.StatusNotFound arm that would SetId("") is
//     unreachable, and it falls through to `return err`.
//
// So after the provider's own Delete succeeds, the confirming Observe hard-errors
// forever. crossplane-runtime can never conclude the external resource is absent,
// never removes the delete finalizer, and the MR stays in Terminating
// permanently. This is an upstream Read-contract bug. The upstream fix would be
// an added errors.Is(err, github.ErrBranchNotProtected) check alongside the
// existing errors.As. No issue or PR has been filed as of this commit -- unlike
// commitLookupParentDeletionWorkaroundResources above, which cites a concrete
// upstream issue, this is a recommendation with nothing tracking it yet.
//
// Only Read reaches the sentinel. requireSignedCommitsRead calls
// GetSignaturesProtectedBranch, which can also return it, but that function
// swallows every error (`return nil //nolint:nilerr`), so it never propagates.
// GetRequiredStatusChecks and ListRequiredStatusChecksContexts are the other two
// go-github functions that substitute the sentinel; no terraform-provider-github
// resource calls either at this version.
//
// Deliberately NOT handled here, and why:
//   - The parent-repository-gone case for this same resource. When the repository
//     is gone, GitHub's message is "Not Found", not "Branch not protected", so
//     isBranchNotProtected does not match, the *github.ErrorResponse survives,
//     and upstream's own http.StatusNotFound arm clears the ID correctly. Note
//     that this rests on GitHub returning "Not Found" for an absent repository:
//     that is inferred from the API's documented behaviour and from
//     isBranchNotProtected's exact-message match, NOT verified against a live
//     deleted repository. A branch deleted while the repository survives is
//     believed to behave the same way ("Branch not found"), likewise unverified.
//     If either assumption is wrong, the symptom is a wedge on parent deletion,
//     not a dropped finalizer on a live resource.
//   - github_branch_protection (the non-v3, GraphQL resource). It is already
//     excluded from parentDeletionWorkaroundResources above, but for a different
//     question (its GraphQL repo-gone handling) -- do not look there for sentinel
//     reasoning. It does not call this REST endpoint at all, so it can never see
//     this sentinel.
var branchNotProtectedWorkaroundResources = []string{
	"github_branch_protection_v3",
}

// wrapReadForParentDeletion translates a typed GitHub 404 (parent repository gone)
// into the SDK's "gone" signal. It runs before the SDK flattens the returned error
// into a diagnostic, so the typed *github.ErrorResponse is still available to
// errors.As.
//
// A 404 from the actions/*-permissions endpoints is an unambiguous "repo gone":
// GitHub returns enabled:false payloads, not 404, when a feature is merely
// disabled.
// nolint:staticcheck // SA1019: the affected upstream resources define the
// legacy schema.Resource.Read field (not ReadContext), and the SDK forbids
// setting both, so the wrapper must operate on the deprecated ReadFunc.
func wrapReadForParentDeletion(orig schema.ReadFunc) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		err := orig(d, meta)
		if err == nil {
			return nil
		}
		var ghErr *github.ErrorResponse
		if errors.As(err, &ghErr) && ghErr.Response != nil && ghErr.Response.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return err
	}
}

// isGitHubNotFoundDiag reports whether diags contains an HTTP 404 from the GitHub
// API. ReadContext funcs return errors via diag.FromErr, which keeps only
// err.Error() and discards the wrapped error, so the typed *github.ErrorResponse
// is unavailable here and the rendered status has to be matched instead.
//
// Both anchors are load-bearing. (*github.ErrorResponse).Error() renders the
// status either after ": " or at the start of the summary, and some upstream reads
// interpolate a caller-supplied name into the message -- matching any
// space-delimited "404" would read a path like "docs/error 404 page.html" as an
// HTTP 404 and drop the delete finalizer on a live resource.
func isGitHubNotFoundDiag(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if d.Severity != diag.Error {
			continue
		}
		if strings.Contains(d.Summary, ": 404 ") || strings.HasPrefix(d.Summary, "404 ") {
			return true
		}
	}
	return false
}

// wrapReadContextForParentDeletion translates a GitHub 404 surfaced as an error
// diagnostic (parent team gone) into the SDK's "gone" signal.
//
// GitHub returns 404, not 403, for objects a token cannot see, so a permissions or
// scope blip is indistinguishable from a genuine deletion here and is treated as
// gone.
func wrapReadContextForParentDeletion(orig schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		diags := orig(ctx, d, meta)
		if isGitHubNotFoundDiag(diags) {
			d.SetId("")
			return nil
		}
		return diags
	}
}

// Mirrors getFileCommit's exhaustion error, which upstream renders as
// fmt.Errorf("cannot find file %s in repo %s/%s", file, owner, repo).
const (
	fileCommitNotFoundPrefix    = "cannot find file "
	fileCommitNotFoundSeparator = " in repo "
)

// isFileCommitNotFoundDiag reports whether diags says no commit reachable from the
// ref still contains the file. That is a "gone" signal, but it carries no HTTP
// status, so isGitHubNotFoundDiag cannot see it.
func isFileCommitNotFoundDiag(diags diag.Diagnostics) bool {
	for _, d := range diags {
		if d.Severity != diag.Error {
			continue
		}
		if rest, ok := afterFileCommitNotFoundPrefix(d.Summary); ok &&
			strings.Contains(rest, fileCommitNotFoundSeparator) {
			return true
		}
	}
	return false
}

// afterFileCommitNotFoundPrefix returns what follows the exhaustion-error prefix,
// allowing for a ": " wrapping boundary the way isGitHubNotFoundDiag does.
//
// The prefix must open the message rather than appear anywhere in it: the path,
// owner and repo are all interpolated into that error, so a substring match would
// read an unrelated failure that merely quotes a user-supplied name as "gone".
func afterFileCommitNotFoundPrefix(summary string) (string, bool) {
	if i := strings.Index(summary, ": "+fileCommitNotFoundPrefix); i >= 0 {
		summary = summary[i+len(": "):]
	}
	return strings.CutPrefix(summary, fileCommitNotFoundPrefix)
}

// wrapReadContextForCommitLookupDeletion turns both ways the commit lookup reports
// a missing resource -- a GitHub 404, and an exhausted commit list -- into the
// SDK's "gone" signal.
//
// The string match is a bridge until integrations/terraform-provider-github#3587
// lands a sentinel error matchable with errors.Is; drop this wrapper,
// isFileCommitNotFoundDiag and the list it serves once the pinned version carries
// it.
func wrapReadContextForCommitLookupDeletion(orig schema.ReadContextFunc) schema.ReadContextFunc {
	return func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
		diags := orig(ctx, d, meta)
		if isGitHubNotFoundDiag(diags) || isFileCommitNotFoundDiag(diags) {
			d.SetId("")
			return nil
		}
		return diags
	}
}

// wrapReadForBranchNotProtected wraps an old-style schema.ReadFunc so that
// go-github's github.ErrBranchNotProtected sentinel is translated into the SDK's
// "resource no longer exists" signal (SetId("") + nil error) instead of a hard
// error. Every other error is propagated unchanged.
//
// This wraps the legacy Read field, so it runs before the SDK flattens the Go
// error into a diagnostic and can match the error *identity* with errors.Is.
// Match the sentinel only -- never its message. github.ErrBranchNotProtected is
// a bare errors.New, so a different error value carrying the identical text is
// not this condition, and a string match would treat it as one.
//
// No http.StatusNotFound arm here, deliberately. go-github only substitutes the
// sentinel when the repository and branch both exist and protection is absent, so
// a genuine 404 (repository or branch gone) still arrives as a
// *github.ErrorResponse and upstream's own StatusNotFound arm already clears the
// ID. Adding a 404 arm would be dead code.
//
// Caveats (honest scope):
//
//   - Unlike wrapReadContextForParentDeletion, this wrapper does NOT carry the
//     "GitHub returns 404 rather than 403 for resources a token cannot see" risk.
//     An unreadable repository yields message "Not Found", never "Branch not
//     protected", so a permissions blip cannot reach this sentinel and cannot
//     drop the delete finalizer on a live resource. The sentinel means protection
//     is genuinely absent.
//   - The create-retry consequence also inverts relative to the other families.
//     If protection is removed out-of-band, Observe now reports
//     ResourceExists:false and crossplane-runtime calls Create, which re-applies
//     protection and succeeds, because the branch still exists. That is drift
//     correction, not the create-retry loop the parent-404 families produce
//     (there the parent is gone, so Create cannot succeed).
//   - It does not fix the underlying upstream bug: `terraform plan` against the
//     unwrapped provider still errors. Only this provider's Read path is repaired.
//
// nolint:staticcheck // SA1019: github_branch_protection_v3 defines the legacy
// schema.Resource.Read field (not ReadContext), and the SDK forbids setting both,
// so the wrapper must operate on the deprecated ReadFunc.
func wrapReadForBranchNotProtected(orig schema.ReadFunc) schema.ReadFunc {
	return func(d *schema.ResourceData, meta interface{}) error {
		err := orig(d, meta)
		if err == nil {
			return nil
		}
		if errors.Is(err, github.ErrBranchNotProtected) {
			d.SetId("")
			return nil
		}
		return err
	}
}

// withParentDeletionWorkaround wraps the affected resources' read functions on the
// in-memory provider. It mutates and returns the same provider so it can be used
// inline at the ujconfig.WithTerraformProvider call site. A resource the pinned
// provider no longer defines, or one that has moved off the field its wrapper
// replaces, is skipped.
//
// The fourth loop is not a parent-deletion case despite the name: it lets
// github_branch_protection_v3 recognise its own successful deletion. See
// branchNotProtectedWorkaroundResources.
func withParentDeletionWorkaround(p *schema.Provider) *schema.Provider {
	wrapRepositoryChildReads(p)
	wrapTeamChildReads(p)
	wrapCommitLookupReads(p)
	wrapBranchNotProtectedReads(p)
	return p
}

func wrapRepositoryChildReads(p *schema.Provider) {
	for _, name := range parentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok || r == nil || r.Read == nil { //nolint:staticcheck // SA1019: intentionally wrapping the legacy Read field these resources define.
			continue
		}
		r.Read = wrapReadForParentDeletion(r.Read) //nolint:staticcheck // SA1019: see above; the SDK forbids setting ReadContext alongside Read.
	}
}

func wrapTeamChildReads(p *schema.Provider) {
	for _, name := range teamParentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok || r == nil || r.ReadContext == nil {
			continue
		}
		r.ReadContext = wrapReadContextForParentDeletion(r.ReadContext)
	}
}

func wrapCommitLookupReads(p *schema.Provider) {
	for _, name := range commitLookupParentDeletionWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok || r == nil || r.ReadContext == nil {
			continue
		}
		r.ReadContext = wrapReadContextForCommitLookupDeletion(r.ReadContext)
	}
}

func wrapBranchNotProtectedReads(p *schema.Provider) {
	for _, name := range branchNotProtectedWorkaroundResources {
		r, ok := p.ResourcesMap[name]
		if !ok || r == nil || r.Read == nil { //nolint:staticcheck // SA1019: intentionally wrapping the legacy Read field this resource defines.
			continue
		}
		r.Read = wrapReadForBranchNotProtected(r.Read) //nolint:staticcheck // SA1019: see above; the SDK forbids setting ReadContext alongside Read.
	}
}
