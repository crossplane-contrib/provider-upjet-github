package emugroupmapping

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure github_emu_group_mapping resources
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_emu_group_mapping", func(r *config.Resource) {
		r.Kind = "EmuGroupMapping"
		r.ShortGroup = "team"

		r.References["team_slug"] = config.Reference{
			TerraformName: "github_team",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("slug",true)`,
		}
	})
}
