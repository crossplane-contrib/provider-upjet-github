package actionsorganizationsecretrepositories

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_secret_repositories resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_secret_repositories", func(r *config.Resource) {
		r.Kind = "OrganizationActionsSecretRepositorySet"
		r.ShortGroup = "actions"

		r.References["secret_name"] = config.Reference{
			TerraformName: "github_actions_organization_secret",
		}

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "selected_repository_ids")
	})
}
