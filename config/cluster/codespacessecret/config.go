package codespacessecret

import "github.com/crossplane/upjet/v2/pkg/config"

// Configure github_codespaces_secret resource.
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("github_codespaces_secret", func(r *config.Resource) {
		r.Kind = "CodespacesSecret"
		r.ShortGroup = "codespaces"

		r.References["repository"] = config.Reference{
			TerraformName: "github_repository",
		}
	})
}
