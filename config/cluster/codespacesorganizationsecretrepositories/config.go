package codespacesorganizationsecretrepositories

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_codespaces_organization_secret_repositories resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_codespaces_organization_secret_repositories", func(r *config.Resource) {
		r.Kind = "OrganizationCodespacesSecretRepositorySet"
		r.ShortGroup = "codespaces"

		r.References["secret_name"] = config.Reference{
			TerraformName: "github_codespaces_organization_secret",
		}
	})
}
