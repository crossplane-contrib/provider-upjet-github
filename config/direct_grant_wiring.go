/*
Copyright 2021 Upbound Inc.
*/

package config

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane-contrib/provider-upjet-github/internal/directgrant"
)

// directGrantWorkaroundResourceName is the sole resource
// wrapReadForDirectGrant (direct_grant.go) applies to:
// github_repository_collaborator's Read reports effective (not direct)
// access, and can write the effective role into permission instead of the
// direct grant's own role -- see direct_grant.go's doc comment for the full
// rationale.
const directGrantWorkaroundResourceName = "github_repository_collaborator"

// withDirectGrantWorkaround wraps github_repository_collaborator's Read on
// the in-memory *schema.Provider so its result is corrected against a
// direct-grant GraphQL lookup (directgrant.Lookup) rather than trusted
// as-is. It mutates and returns the same provider so it can be composed
// inline at the ujconfig.WithTerraformProvider call site.
//
// A missing resource, a nil *schema.Resource, or a nil legacy Read field is
// skipped rather than panicking, so an upstream rename or a switch to
// ReadContext is a silent no-op here, not a crash.
func withDirectGrantWorkaround(p *schema.Provider) *schema.Provider {
	r, ok := p.ResourcesMap[directGrantWorkaroundResourceName]
	if !ok || r == nil || r.Read == nil { //nolint:staticcheck // SA1019: intentionally wrapping the legacy Read field this resource defines.
		return p
	}
	r.Read = wrapReadForDirectGrant(r.Read, directgrant.Lookup) //nolint:staticcheck // SA1019: see above; the SDK forbids setting ReadContext alongside Read.
	return p
}
