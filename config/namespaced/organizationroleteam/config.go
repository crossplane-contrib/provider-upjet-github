package organizationroleteam

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_role_team", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "OrganizationRoleTeam"
		r.ShortGroup = "organization"

		// Add reference to Team resource
		r.References["team_slug"] = config.Reference{
			TerraformName: "github_team",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("slug",true)`,
		}
	})
}
