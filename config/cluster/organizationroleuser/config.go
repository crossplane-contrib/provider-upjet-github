package organizationroleuser

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_role_user resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_role_user", func(r *config.Resource) {
		r.Kind = "OrganizationRoleUser"
		r.ShortGroup = "user"
	})
}
