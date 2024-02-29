package emugroupmapping

import "github.com/crossplane/upjet/pkg/config"

// Configure github_emu_group_mapping resources
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_emu_group_mapping", func(r *config.Resource) {
		r.Kind = "EmuGroupMapping"
		r.ShortGroup = "team"

		r.References["team_slug"] = config.Reference{
			Type: "Team",
		}
	})
}
