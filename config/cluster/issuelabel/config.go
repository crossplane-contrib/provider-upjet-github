package issuelabel

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_issue_label resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_issue_label", func(r *config.Resource) {
		r.Kind = "RepositoryLabel"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
