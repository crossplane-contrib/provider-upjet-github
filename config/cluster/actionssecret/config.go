package actionssecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_secret", func(r *config.Resource) {

		r.Kind = "ActionsSecret"
		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
