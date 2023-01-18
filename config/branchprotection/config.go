package branchprotection

import "github.com/upbound/upjet/pkg/config"

// Configure github_branch_protection resource
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_branch_protection", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "BranchProtection"
		r.ShortGroup = "repo"
		// This resource need the repository in which branch would be created
		// as an input. And by defining it as a reference to Repository
		// object, we can build cross resource referencing. See
		// repositoryRef in the example in the Testing section below.
		r.References["repository_id"] = config.Reference{
			Type: "github.com/coopnorge/provider-github/apis/repo/v1alpha1.Repository",
		}
	})
}
