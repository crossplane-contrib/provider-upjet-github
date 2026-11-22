package repositorymilestone

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_milestone resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_milestone", func(r *config.Resource) {
		r.Kind = "RepositoryMilestone"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
