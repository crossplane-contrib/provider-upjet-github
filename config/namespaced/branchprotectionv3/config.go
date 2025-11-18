package branchprotectionv3

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_branch_protection_v3 resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_branch_protection_v3", func(r *config.Resource) {
		r.Kind = "BranchProtectionv3"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
