package membership

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_membership resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_membership", func(r *config.Resource) {
		r.Kind = "Membership"
		r.ShortGroup = "user"

	})
}
