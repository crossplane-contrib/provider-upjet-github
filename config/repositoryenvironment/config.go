package repositoryenvironment

import (
	"github.com/crossplane/upjet/pkg/config"
)

// Configure github_repository_environment resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_environment", func(r *config.Resource) {
		r.Kind = "RepositoryEnvironment"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}

		r.References["reviewers.teams"] = config.Reference{
			TerraformName: "github_team",
		}

		r.References["reviewers.users"] = config.Reference{
			TerraformName: "github_membership",
		}
	})
}
