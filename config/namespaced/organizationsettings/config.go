package organizationsettings

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_settings resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_settings", func(r *config.Resource) {
		r.Kind = "OrganizationSettings"
		r.ShortGroup = "organization"
	})
}
