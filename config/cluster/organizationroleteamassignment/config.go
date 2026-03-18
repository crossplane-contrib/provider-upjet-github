package organizationroleteamassignment

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
//
// Deprecated: Use OrganizationRoleTeam instead.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_role_team_assignment", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "OrganizationRoleTeamAssignment"
		r.ShortGroup = "organization"

		// Add reference to Team resource
		r.References["team_slug"] = config.Reference{
			TerraformName: "github_team",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("slug",true)`,
		}
	})
}
