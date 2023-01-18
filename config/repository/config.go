package repository

import "github.com/upbound/upjet/pkg/config"

// Configure github_repository resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.ShortGroup = "repo"

		r.LateInitializer = config.LateInitializer{
			IgnoredFields: []string{"private", "default_branch"},
		}
	})
}
