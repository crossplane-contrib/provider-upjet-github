package appinstallationrepositories

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_app_installation_repositories resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_app_installation_repositories", func(r *config.Resource) {
		r.Kind = "AppInstallationRepositories"
		r.ShortGroup = "app"

		r.References["selected_repositories"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
