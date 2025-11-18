package repositoryruleset

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_ruleset resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_ruleset", func(r *config.Resource) {
		// We need to override the default group that upjet generated for
		// this resource, which would be "github"
		r.Kind = "RepositoryRuleset"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
