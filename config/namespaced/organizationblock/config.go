package organizationblock

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_block resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_block", func(r *config.Resource) {
		r.Kind = "OrganizationBlock"
		r.ShortGroup = "organization"
	})
}
