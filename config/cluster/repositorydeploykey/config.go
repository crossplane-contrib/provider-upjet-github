package repositorydeploykey

import (
	"github.com/crossplane/upjet/v2/pkg/config"
)

// Configure github_repository_deploy_key resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_deploy_key", func(r *config.Resource) {
		r.Kind = "DeployKey"
		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}

		r.TerraformResource.Schema["key"].Required = true
		r.TerraformResource.Schema["read_only"].Required = true
		r.TerraformResource.Schema["title"].Required = true

		// Setting the field as sensitive to be able to pass the content from a k8s secret
		r.TerraformResource.Schema["key"].Sensitive = true
	})
}
