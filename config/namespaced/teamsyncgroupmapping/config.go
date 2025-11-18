package teamsyncgroupmapping

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_team_sync_group_mapping resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_team_sync_group_mapping", func(r *config.Resource) {
		r.Kind = "TeamSyncGroupMapping"
		r.ShortGroup = "team"

		r.References["team_id"] = config.Reference{
			TerraformName: "github_team",
		}
	})
}
