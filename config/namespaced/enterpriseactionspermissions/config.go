package enterpriseactionspermissions

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure configures individual resources by adding custom ResourceConfigurators.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_enterprise_actions_permissions", func(r *config.Resource) {
		r.Kind = "ActionsPermissions"
		r.ShortGroup = "enterprise"
	})
}
