package teammembership

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_team_membership resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_team_membership", func(r *config.Resource) {
		r.Kind = "TeamMembership"
		r.ShortGroup = "team"

		r.References["team_id"] = config.Reference{
			TerraformName: "github_team",
		}
	})
}
