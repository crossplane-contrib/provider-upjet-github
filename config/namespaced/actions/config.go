package actions

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_runner_group", func(r *config.Resource) {

		r.ShortGroup = "actions"
		//TODO: implemant an array of references
	})
}
