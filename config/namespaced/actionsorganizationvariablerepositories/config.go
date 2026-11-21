package actionsorganizationvariablerepositories

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_variable_repositories resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_variable_repositories", func(r *config.Resource) {
		r.Kind = "OrganizationActionsVariableRepositorySet"
		r.ShortGroup = "actions"

		r.References["variable_name"] = config.Reference{
			TerraformName: "github_actions_organization_variable",
		}

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "selected_repository_ids")
	})
}
