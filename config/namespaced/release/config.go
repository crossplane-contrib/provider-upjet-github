package release

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_release resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_release", func(r *config.Resource) {
		r.Kind = "Release"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
