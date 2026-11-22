package enterpriseactionsrunnergroup

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_enterprise_actions_runner_group resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_enterprise_actions_runner_group", func(r *config.Resource) {
		r.Kind = "ActionsRunnerGroup"
		r.ShortGroup = "enterprise"

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "selected_organization_ids")
	})
}
