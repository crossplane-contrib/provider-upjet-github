package repositoryautolinkreference

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_autolink_reference resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_autolink_reference", func(r *config.Resource) {
		r.Kind = "RepositoryAutolinkReference"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
