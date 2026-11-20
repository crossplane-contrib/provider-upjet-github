package repositorytopics

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_topics resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_topics", func(r *config.Resource) {
		r.Kind = "RepositoryTopics"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
