package branchdefault

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_branch_default resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_branch_default", func(r *config.Resource) {
		r.Kind = "DefaultBranch"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
		r.References["branch"] = config.Reference{
			TerraformName: "github_branch",
		}
	})
}
