package actionsorganizationsecretrepository

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_actions_organization_secret_repository resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_secret_repository", func(r *config.Resource) {
		r.Kind = "OrganizationActionsSecretRepository"
		r.ShortGroup = "actions"

		r.References["secret_name"] = config.Reference{
			TerraformName: "github_actions_organization_secret",
		}

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "repository_id")
	})
}
