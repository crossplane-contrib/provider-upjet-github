package repositoryenvironmentdeploymentpolicy

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure github_repository_environment_deployment_policy resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_environment_deployment_policy", func(r *config.Resource) {
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
		r.References["environment"] = config.Reference{
			TerraformName: "github_repository_environment",
		}
	})
}
