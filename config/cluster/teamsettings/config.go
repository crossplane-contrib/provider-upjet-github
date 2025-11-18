package teamsettings

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_team_settings resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_team_settings", func(r *config.Resource) {
		r.Kind = "TeamSettings"
		r.ShortGroup = "team"

		r.References["team_id"] = config.Reference{
			TerraformName: "github_team",
		}
	})
}
