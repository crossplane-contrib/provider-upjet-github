package repositoryfile

import "github.com/upbound/upjet/pkg/config"

// Configure github_branch resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_file", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "RepositoryFile"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			Type: "Repository",
		}
		r.References["branch"] = config.Reference{
			Type: "Branch",
		}
	})
}
