package githubissue

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_issue resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_issue", func(r *config.Resource) {
		r.Kind = "GithubIssue"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
