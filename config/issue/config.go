package issue

import "github.com/crossplane/upjet/pkg/config"

// Configure github_issue_label resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_issue", func(r *config.Resource) {

		// We need to override the default group that upjet generated for
		// this resource, which would be "github"

		r.Kind = "Issue"
		r.ShortGroup = "repo"

		// This resource need the repository in which branch would be created
		// as an input. And by defining it as a reference to Repository
		// object, we can build cross resource referencing. See
		// repositoryRef in the example in the Testing section below.
		r.References["repository "] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
