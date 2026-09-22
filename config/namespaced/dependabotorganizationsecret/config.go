package dependabotorganizationsecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_dependabot_organization_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_dependabot_organization_secret", func(r *config.Resource) {
		r.Kind = "DependabotOrganizationSecret"
		r.ShortGroup = "dependabot"
	})
}
