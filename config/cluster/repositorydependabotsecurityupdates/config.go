package repositorydependabotsecurityupdates

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_repository_dependabot_security_updates resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_dependabot_security_updates", func(r *config.Resource) {
		r.Kind = "RepositoryDependabotSecurityUpdates"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
