package repositoryenvironment

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure github_repository_environment resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_environment", func(r *config.Resource) {

		r.Kind = "Environment"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}

	})
}
