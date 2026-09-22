package dependabotsecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_dependabot_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_dependabot_secret", func(r *config.Resource) {
		r.Kind = "DependabotSecret"
		r.ShortGroup = "dependabot"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
