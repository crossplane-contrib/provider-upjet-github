package organizationrepositoryrole

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_repository_role resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_repository_role", func(r *config.Resource) {
		r.Kind = "OrganizationRepositoryRole"
		r.ShortGroup = "organization"
	})
}
