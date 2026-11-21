package appinstallationrepository

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_app_installation_repository resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_app_installation_repository", func(r *config.Resource) {
		r.Kind = "AppInstallationRepositoryAssociation"
		r.ShortGroup = "app"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
