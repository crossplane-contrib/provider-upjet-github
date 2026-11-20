package actionsorganizationvariablerepository

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_variable_repository resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_variable_repository", func(r *config.Resource) {
		r.Kind = "OrganizationActionsVariableRepository"
		r.ShortGroup = "actions"

		r.References["variable_name"] = config.Reference{
			TerraformName: "github_actions_organization_variable",
		}

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "repository_id")
	})
}
