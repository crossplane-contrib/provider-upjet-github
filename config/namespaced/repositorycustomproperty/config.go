package repositorycustomproperty

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_custom_property resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_custom_property", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "RepositoryCustomProperty"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
