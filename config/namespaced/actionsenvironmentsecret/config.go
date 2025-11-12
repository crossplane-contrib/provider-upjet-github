package actionsenvironmentsecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_environment_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_environment_secret", func(r *config.Resource) {
		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
		r.References["environment"] = config.Reference{
			TerraformName: "github_repository_environment",
		}
	})
}
