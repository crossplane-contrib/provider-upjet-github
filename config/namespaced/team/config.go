package team

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_team
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_team", func(r *config.Resource) {
		r.Kind = "Team"
		r.ShortGroup = "team"
	})
}
