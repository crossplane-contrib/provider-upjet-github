package repositorytagprotection

import (
	"github.com/crossplane/upjet/pkg/config"
)

// Configure github_repository_tag_protection resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_repository_tag_protection", func(r *config.Resource) {

		r.ShortGroup = "repo"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
