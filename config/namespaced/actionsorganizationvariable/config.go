package actionsorganizationvariable

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_variable resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_variable", func(r *config.Resource) {
		r.Kind = "OrganizationActionsVariable"
		r.ShortGroup = "actions"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
