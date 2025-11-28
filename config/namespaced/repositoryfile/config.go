package repositoryfile

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_file resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_file", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "RepositoryFile"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
		r.References["branch"] = config.Reference{
			TerraformName: "github_branch",
			Extractor:     `github.com/crossplane/upjet/v2/pkg/resource.ExtractParamPath("branch",true)`,
		}
	})
}
