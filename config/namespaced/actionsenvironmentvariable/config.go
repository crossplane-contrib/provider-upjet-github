package actionsenvironmentvariable

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_environment_variable resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_environment_variable", func(r *config.Resource) {

		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
		r.References["environment"] = config.Reference{
			TerraformName: "github_repository_environment",
		}
	})
}
