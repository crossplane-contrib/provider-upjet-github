package repositorypages

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_pages resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_pages", func(r *config.Resource) {
		r.Kind = "RepositoryPages"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
