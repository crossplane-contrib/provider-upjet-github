package actionshostedrunner

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_hosted_runner resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_hosted_runner", func(r *config.Resource) {
		r.Kind = "HostedRunner"
		r.ShortGroup = "actions"

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "runner_group_id")
	})
}
