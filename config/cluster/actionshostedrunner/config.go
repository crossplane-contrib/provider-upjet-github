package actionshostedrunner

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_hosted_runner resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_hosted_runner", func(r *config.Resource) {
		r.Kind = "ActionsHostedRunner"
		r.ShortGroup = "actions"

		// runner_group_id is an integer in Terraform, but references resolve to strings.
		// To avoid type mismatch in generated resolvers, we omit the reference for now.
	})
}
