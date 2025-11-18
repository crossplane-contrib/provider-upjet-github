package actionsorganizationsecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_secret", func(r *config.Resource) {
		r.Kind = "OrganizationActionsSecret"
		r.ShortGroup = "actions"
	})
}
