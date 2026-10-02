package organizationroleteam

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_organization_role_team resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_organization_role_team", func(r *config.Resource) {
		r.Kind = "OrganizationRoleTeam"
		r.ShortGroup = "team"
	})
}
