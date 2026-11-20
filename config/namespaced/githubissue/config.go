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

		// A number, which a reference can't populate: references resolve to strings
		delete(r.References, "milestone_number")
	})
}
