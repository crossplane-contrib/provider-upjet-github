package actionsorganizationsecret

import "github.com/crossplane/upjet/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_actions_organization_secret", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "ActionsOrganizationSecret"
		r.ShortGroup = "actions"
	})
}
