package actionsrepositoryaccesslevel

import "github.com/crossplane/upjet/pkg/config"

// Configure github_actions_repository_access_level resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_repository_access_level", func(r *config.Resource) {

		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
