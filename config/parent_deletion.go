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

// withParentDeletionWorkaround wraps the affected resources' read functions on the
// in-memory provider. It mutates and returns the same provider so it can be used
// inline at the ujconfig.WithTerraformProvider call site. A resource the pinned
// provider no longer defines, or one that has moved off the field its wrapper
// replaces, is skipped.
func withParentDeletionWorkaround(p *schema.Provider) *schema.Provider {
	wrapRepositoryChildReads(p)
	wrapTeamChildReads(p)
	wrapCommitLookupReads(p)
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
