package codespacesorganizationsecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_codespaces_organization_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_codespaces_organization_secret", func(r *config.Resource) {
		r.Kind = "OrganizationCodespacesSecret"
		r.ShortGroup = "codespaces"
	})
}
