package organizationrole

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_role resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_role", func(r *config.Resource) {
		r.Kind = "OrganizationRole"
		r.ShortGroup = "organization"
	})
}
