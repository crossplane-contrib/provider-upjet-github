package branchprotection

import "github.com/crossplane/upjet/pkg/config"

// Configure github_branch_protection resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_branch_protection", func(r *config.Resource) {
		r.Kind = "BranchProtection"
		r.ShortGroup = "repo"

		r.References["repository_id"] = config.Reference{
			Type: "github.com/coopnorge/provider-github/apis/repo/v1alpha1.Repository",
		}
	})
}
